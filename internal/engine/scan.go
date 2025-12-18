package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kuri-sun/todo-graph/internal/graph"
)

var (
	todoLinePattern = regexp.MustCompile(`TODO:?\s*(\[#([^\]]*)\])?(.*)`)
	todoIDPattern   = regexp.MustCompile(`^[a-z0-9_-]+$`)
	commentLine     = regexp.MustCompile(`^\s*(//|#|--|/\\*|\\*|<!--)`)
)

type blockComment struct {
	start string
	end   string
}

// ScanError provides contextual information for parse failures.
type ScanError struct {
	File string
	Line int
	Msg  string
}

// Scan walks the repository, parses TODO blocks, and returns the graph.
func Scan(root string) (graph.Graph, []ScanError, error) {
	todos := make(map[string]graph.Todo)
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
		if d.Name() == ".todo-graph" {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		fileEdges, fileTodos, fileErrs, err := scanFile(path, rel)
		if err != nil {
			return err
		}
		for _, e := range fileErrs {
			errs = append(errs, e)
		}
		for _, t := range fileTodos {
			if existing, ok := todos[t.ID]; ok {
				errs = append(errs, ScanError{
					File: rel,
					Line: t.Line,
					Msg:  fmt.Sprintf("duplicate TODO id %q (first defined in %s:%d)", t.ID, existing.File, existing.Line),
				})
				continue
			}
			todos[t.ID] = t
		}
		edges = append(edges, fileEdges...)
		return nil
	})
	if err != nil {
		return graph.Graph{}, nil, err
	}

	edges = dedupeEdges(edges)

	return graph.Graph{
		Todos: todos,
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

func scanFile(path, rel string) ([]graph.Edge, []graph.Todo, []ScanError, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if isBinary(content) {
		return nil, nil, nil, nil
	}

	lines := strings.Split(string(content), "\n")

	var edges []graph.Edge
	var todos []graph.Todo
	var errs []ScanError

	blockComments := []blockComment{
		{start: "/*", end: "*/"},
		{start: "<!--", end: "-->"},
		{start: `"""`, end: `"""`},
		{start: `'''`, end: `'''`},
	}
	inBlock := false
	var currentBlockEnd string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// determine if this line is in a comment
		comment := false
		if inBlock {
			comment = true
			if strings.Contains(line, currentBlockEnd) {
				inBlock = false
				currentBlockEnd = ""
			}
		}
		if !comment {
			switch {
			case strings.HasPrefix(trimmed, "//"),
				strings.HasPrefix(trimmed, "#"),
				strings.HasPrefix(trimmed, "--"):
				comment = true
			default:
				for _, bc := range blockComments {
					if idx := strings.Index(line, bc.start); idx != -1 {
						comment = true
						if strings.Index(line[idx+len(bc.start):], bc.end) == -1 {
							inBlock = true
							currentBlockEnd = bc.end
						}
						break
					}
				}
			}
		}

		if !comment {
			continue
		}

		match := todoLinePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		rawID := match[2]
		desc := strings.TrimSpace(match[3])
		if match[1] != "" && rawID == "" {
			errs = append(errs, ScanError{File: rel, Line: i + 1, Msg: "TODO id must not be empty"})
			continue
		}
		if rawID == "" {
			derived := deriveID(desc, i+1)
			if derived == "" {
				errs = append(errs, ScanError{File: rel, Line: i + 1, Msg: "TODO id must not be empty"})
				continue
			}
			rawID = derived
		}
		if !todoIDPattern.MatchString(rawID) {
			errs = append(errs, ScanError{
				File: rel,
				Line: i + 1,
				Msg:  fmt.Sprintf("TODO id %q must use lowercase letters, digits, hyphens, or underscores", rawID),
			})
			continue
		}

		todo := graph.Todo{
			ID:   rawID,
			File: rel,
			Line: i + 1,
		}

		depends, _, metaErrs, endIdx := parseMetadata(lines, i+1, rel)
		errs = append(errs, metaErrs...)

		for _, dep := range depends {
			edges = append(edges, graph.Edge{From: dep, To: rawID, Type: "blocks"})
		}

		todos = append(todos, todo)
		i = endIdx
	}

	return edges, todos, errs, nil
}

func parseMetadata(lines []string, start int, file string) (depends []string, blocks []string, errs []ScanError, end int) {
	end = start - 1
	for idx := start; idx < len(lines); idx++ {
		line := lines[idx]
		if todoLinePattern.MatchString(line) {
			end = idx - 1
			return
		}
		if !commentLine.MatchString(line) {
			end = idx - 1
			return
		}

		key, values := parseKeyValue(line)
		switch key {
		case "deps":
			ids, idErrs := parseIDs(values, idx+1, file)
			errs = append(errs, idErrs...)
			depends = append(depends, ids...)
		default:
			// ignore unknown keys
		}
		end = idx
	}
	return
}

func parseKeyValue(line string) (key string, value string) {
	trimmed := commentLine.ReplaceAllString(line, "")
	trimmed = strings.TrimSpace(trimmed)
	colon := strings.Index(trimmed, ":")
	if colon == -1 {
		return "", ""
	}
	key = strings.ToLower(strings.TrimSpace(trimmed[:colon]))
	value = strings.TrimSpace(trimmed[colon+1:])
	return
}

func parseIDs(raw string, line int, file string) ([]string, []ScanError) {
	if raw == "" {
		return nil, nil
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	var parts []string
	if strings.Contains(trimmed, ",") {
		parts = strings.Split(trimmed, ",")
	} else {
		fields := strings.Fields(trimmed)
		if len(fields) > 1 {
			return nil, []ScanError{{File: file, Line: line, Msg: "ids must be comma-separated (e.g. #a, #b)"}}
		}
		parts = []string{trimmed}
	}

	var ids []string
	var errs []ScanError
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, " ") {
			errs = append(errs, ScanError{File: file, Line: line, Msg: "ids must be comma-separated (e.g. #a, #b)"})
			continue
		}
		if strings.HasPrefix(p, "#") {
			p = strings.TrimPrefix(p, "#")
		} else {
			errs = append(errs, ScanError{File: file, Line: line, Msg: fmt.Sprintf("id %q must start with #", p)})
			continue
		}
		if !todoIDPattern.MatchString(p) {
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

func deriveID(desc string, line int) string {
	return fmt.Sprintf("todo-%d", line)
}
