package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kuri-sun/comment-graph/internal/graph"
)

var (
	cgraphIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
	commentLine     = regexp.MustCompile(`^\s*(//|#|--|/\*|{/\*|<!--|\*|"""|''')`)
)

var commentClosers = []string{"*/", "*/}", "-->", `"""`, `'''`}

// ScanError provides contextual information for parse failures.
type ScanError struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Msg  string `json:"msg"`
}

// Scan walks the repository and builds a comment graph.
func Scan(root string) (graph.Graph, []ScanError, error) {
	nodes := make(map[string]graph.Node)
	var edges []graph.Edge
	var errs []ScanError

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() && shouldSkipDir(d.Name()) {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == ".comment-graph" {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		fileEdges, fileNodes, fileErrs, err := scanFile(path, rel)
		if err != nil {
			return err
		}
		for _, e := range fileErrs {
			errs = append(errs, e)
		}
		for _, n := range fileNodes {
			if existing, ok := nodes[n.ID]; ok {
				errs = append(errs, ScanError{
					File: rel,
					Line: n.Line,
					Msg:  fmt.Sprintf("duplicate comment-graph id %q (first defined in %s:%d)", n.ID, existing.File, existing.Line),
				})
				continue
			}
			nodes[n.ID] = n
		}
		edges = append(edges, fileEdges...)
		return nil
	})
	if err != nil {
		return graph.Graph{}, nil, err
	}

	edges = dedupeEdges(edges)

	return graph.Graph{
		Nodes: nodes,
		Edges: edges,
	}, errs, nil
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", ".idea", ".vscode":
		return true
	default:
		return false
	}
}

func scanFile(path, rel string) ([]graph.Edge, []graph.Node, []ScanError, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if isBinary(content) {
		return nil, nil, nil, nil
	}

	lines := strings.Split(string(content), "\n")

	var edges []graph.Edge
	var errs []ScanError
	nodes := make(map[string]graph.Node)

	type pending struct {
		line    int
		id      string
		deps    []string
		invalid bool
		hasMeta bool
	}
	var current *pending

	flush := func() {
		if current == nil {
			return
		}
		if current.invalid {
			current = nil
			return
		}
		if current.id == "" {
			if current.hasMeta {
				errs = append(errs, ScanError{File: rel, Line: current.line, Msg: "metadata without @cgraph-id"})
			}
			current = nil
			return
		}
		nodes[current.id] = graph.Node{ID: current.id, File: rel, Line: current.line}
		for _, dep := range current.deps {
			edges = append(edges, graph.Edge{From: dep, To: current.id, Type: "blocks"})
		}
		current = nil
	}

	inBlock := false
	blockEnd := ""

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		isCommentStart := strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "--") ||
			strings.HasPrefix(trimmed, "/*") ||
			strings.HasPrefix(trimmed, "{/*") ||
			strings.HasPrefix(trimmed, "<!--") ||
			strings.HasPrefix(trimmed, `"""`) ||
			strings.HasPrefix(trimmed, `'''`) ||
			(inBlock && strings.HasPrefix(trimmed, "*"))

		comment := inBlock || isCommentStart

		if inBlock && blockEnd != "" && strings.Contains(line, blockEnd) {
			inBlock = false
			blockEnd = ""
		}

		if !inBlock {
			switch {
			case strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "{/*"):
				if !strings.Contains(line, "*/") && !strings.Contains(line, "*/}") {
					inBlock = true
					blockEnd = "*/"
					if strings.HasPrefix(trimmed, "{/*") {
						blockEnd = "*/}"
					}
				}
			case strings.HasPrefix(trimmed, "<!--"):
				if !strings.Contains(line, "-->") {
					inBlock = true
					blockEnd = "-->"
				}
			case strings.HasPrefix(trimmed, `"""`):
				if strings.Count(line, `"""`) == 1 {
					inBlock = true
					blockEnd = `"""`
				}
			case strings.HasPrefix(trimmed, `'''`):
				if strings.Count(line, `'''`) == 1 {
					inBlock = true
					blockEnd = `'''`
				}
			}
		}

		if current != nil && (trimmed == "" || !comment) {
			flush()
		}

		if !comment {
			continue
		}

		cleaned := strings.TrimSpace(commentLine.ReplaceAllString(line, ""))
		lower := strings.ToLower(cleaned)

		switch {
		case strings.HasPrefix(lower, "@cgraph-id"):
			if current != nil {
				flush()
			}
			if current == nil {
				current = &pending{}
			}
			current.hasMeta = true
			val := strings.TrimSpace(strings.TrimPrefix(cleaned, "@cgraph-id"))
			val = strings.TrimSpace(cleanCommentSuffix(val))
			if val == "" {
				errs = append(errs, ScanError{File: rel, Line: i + 1, Msg: "@cgraph-id must not be empty"})
				current.invalid = true
				continue
			}
			if !cgraphIDPattern.MatchString(val) {
				errs = append(errs, ScanError{
					File: rel,
					Line: i + 1,
					Msg:  fmt.Sprintf("@cgraph-id %q must use lowercase letters, digits, hyphens, or underscores", val),
				})
				current.invalid = true
				continue
			}
			current.id = val
			current.line = i + 1
		case strings.HasPrefix(lower, "@cgraph-deps"):
			if current == nil {
				current = &pending{line: i + 1}
			}
			current.hasMeta = true
			raw := strings.TrimSpace(strings.TrimPrefix(cleaned, "@cgraph-deps"))
			raw = strings.TrimSpace(cleanCommentSuffix(raw))
			ids, idErrs := parseIDs(raw, i+1, rel)
			errs = append(errs, idErrs...)
			current.deps = append(current.deps, ids...)
		case strings.HasPrefix(lower, "@"):
			errs = append(errs, ScanError{File: rel, Line: i + 1, Msg: "unknown metadata (use @cgraph-id or @cgraph-deps)"})
		default:
			// plain comment line, keep association
		}
	}

	flush()

	var nodeList []graph.Node
	for _, n := range nodes {
		nodeList = append(nodeList, n)
	}

	return edges, nodeList, errs, nil
}

func cleanCommentSuffix(s string) string {
	s = strings.TrimSpace(s)
	for _, c := range commentClosers {
		s = strings.TrimSuffix(s, c)
	}
	return strings.TrimSpace(s)
}

func parseIDs(raw string, line int, file string) ([]string, []ScanError) {
	if raw == "" {
		return nil, nil
	}
	trimmed := strings.TrimSpace(cleanCommentSuffix(raw))
	if trimmed == "" {
		return nil, nil
	}

	if !strings.Contains(trimmed, ",") && strings.Contains(trimmed, " ") {
		return nil, []ScanError{{File: file, Line: line, Msg: "ids must be comma-separated (e.g. a, b)"}}
	}

	parts := strings.Split(trimmed, ",")

	var ids []string
	var errs []ScanError
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, " ") {
			errs = append(errs, ScanError{File: file, Line: line, Msg: "ids must be comma-separated (e.g. a, b)"})
			continue
		}
		if !cgraphIDPattern.MatchString(p) {
			errs = append(errs, ScanError{
				File: file,
				Line: line,
				Msg:  fmt.Sprintf("id %q must use lowercase letters, digits, hyphens, or underscores", p),
			})
			continue
		}
		ids = append(ids, p)
	}
	return ids, errs
}

func dedupeEdges(edges []graph.Edge) []graph.Edge {
	seen := make(map[string]bool)
	var out []graph.Edge
	for _, e := range edges {
		key := fmt.Sprintf("%s->%s|%s", e.From, e.To, e.Type)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, e)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].From == out[j].From {
			return out[i].To < out[j].To
		}
		return out[i].From < out[j].From
	})
	return out
}

func isBinary(content []byte) bool {
	// simple heuristic: treat files with NULL bytes as binary
	for _, b := range content {
		if b == 0 {
			return true
		}
	}
	return false
}
