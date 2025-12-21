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
	defaultKeywords = []string{"TODO", "FIXME", "NOTE", "WARNING", "HACK", "CHANGED", "REVIEW"}
	todoIDPattern   = regexp.MustCompile(`^[a-z0-9_-]+$`)
	commentLine     = regexp.MustCompile(`^\s*(//|#|--|/\*|{/\*|<!--|\*|"""|''')`)
)

var commentClosers = []string{"*/", "*/}", "-->", `"""`, `'''`}

func compileTodoPattern(keywords []string) (*regexp.Regexp, error) {
	if len(keywords) == 0 {
		return nil, fmt.Errorf("at least one keyword is required")
	}
	var parts []string
	for _, k := range keywords {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		parts = append(parts, regexp.QuoteMeta(k))
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("at least one keyword is required")
	}
	pat := fmt.Sprintf(`(?i)\b(%s)\b:?`, strings.Join(parts, "|"))
	return regexp.Compile(pat)
}

// ScanError provides contextual information for parse failures.
type ScanError struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Msg  string `json:"msg"`
}

// Scan walks the repository using default keywords.
func Scan(root string) (graph.Graph, []ScanError, error) {
	return ScanWithKeywords(root, defaultKeywords)
}

// ScanWithKeywords walks the repository, parses TODO blocks, and returns the graph.
// keywords defines which markers are recognized (e.g. TODO, FIXME).
func ScanWithKeywords(root string, keywords []string) (graph.Graph, []ScanError, error) {
	if len(keywords) == 0 {
		keywords = defaultKeywords
	}
	pattern, err := compileTodoPattern(keywords)
	if err != nil {
		return graph.Graph{}, nil, err
	}
	todos := make(map[string]graph.Todo)
	var edges []graph.Edge
	var errs []ScanError

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
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
		fileEdges, fileTodos, fileErrs, err := scanFile(path, rel, pattern)
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

func scanFile(path, rel string, todoPattern *regexp.Regexp) ([]graph.Edge, []graph.Todo, []ScanError, error) {
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
	todos := make(map[string]graph.Todo)

	type pending struct {
		line    int
		id      string
		deps    []string
		invalid bool
	}
	var current *pending

	inBlock := false
	blockEnd := ""

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		skip := false
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

		// inline comment markers with TODO/keyword after code
		if !comment {
			inlineMarkers := []string{"//", "#", "--", "/*", "{/*"}
			for _, m := range inlineMarkers {
				if idx := strings.Index(line, m); idx > 0 && todoPattern.MatchString(line[idx:]) {
					errs = append(errs, ScanError{File: rel, Line: i + 1, Msg: "TODO must start at a comment line"})
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		// close block if end appears on this line
		if inBlock && blockEnd != "" && strings.Contains(line, blockEnd) {
			inBlock = false
			blockEnd = ""
		}

		// open block only when comment start is at the beginning
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

		// break association on blank or non-comment
		if current != nil && (trimmed == "" || (!comment && !todoPattern.MatchString(trimmed))) {
			if current.id != "" && !current.invalid {
				todos[current.id] = graph.Todo{ID: current.id, File: rel, Line: current.line}
				for _, dep := range current.deps {
					edges = append(edges, graph.Edge{From: dep, To: current.id, Type: "blocks"})
				}
			}
			current = nil
		}

		// inline keyword not allowed (only if a comment marker is present)
		if todoPattern.MatchString(line) && !isCommentStart && !inBlock &&
			(strings.Contains(line, "//") || strings.Contains(line, "#") || strings.Contains(line, "--") || strings.Contains(line, "/*")) {
			errs = append(errs, ScanError{File: rel, Line: i + 1, Msg: "TODO must start at a comment line"})
			continue
		}

		if !comment {
			continue
		}

		trimmedNoPrefix := strings.TrimSpace(commentLine.ReplaceAllString(line, ""))
		lowerTrimmed := strings.ToLower(trimmedNoPrefix)
		if strings.HasPrefix(lowerTrimmed, "@todo-") {
			// metadata handled below
		} else if todoPattern.MatchString(trimmedNoPrefix) {
			// close previous pending
			if current != nil {
				if current.id != "" && !current.invalid {
					todos[current.id] = graph.Todo{ID: current.id, File: rel, Line: current.line}
					for _, dep := range current.deps {
						edges = append(edges, graph.Edge{From: dep, To: current.id, Type: "blocks"})
					}
				}
			}

			current = &pending{line: i + 1}
			continue
		}

		// metadata handling
		cleaned := strings.TrimSpace(commentLine.ReplaceAllString(line, ""))
		lower := strings.ToLower(cleaned)
		if current == nil {
			if strings.HasPrefix(lower, "@todo-id") || strings.HasPrefix(lower, "@todo-deps") {
				errs = append(errs, ScanError{File: rel, Line: i + 1, Msg: "metadata without a preceding TODO"})
			}
			continue
		}

		switch {
		case strings.HasPrefix(lower, "@todo-id"):
			val := strings.TrimSpace(strings.TrimPrefix(cleaned, "@todo-id"))
			val = strings.TrimPrefix(val, ":")
			val = strings.TrimSpace(cleanCommentSuffix(val))
			if val == "" {
				errs = append(errs, ScanError{File: rel, Line: i + 1, Msg: "TODO id must not be empty"})
				continue
			}
			if !todoIDPattern.MatchString(val) {
				errs = append(errs, ScanError{
					File: rel,
					Line: i + 1,
					Msg:  fmt.Sprintf("TODO id %q must use lowercase letters, digits, hyphens, or underscores", val),
				})
				current.invalid = true
				continue
			}
			current.id = val
		case strings.HasPrefix(lower, "@todo-deps"):
			raw := strings.TrimSpace(strings.TrimPrefix(cleaned, "@todo-deps"))
			raw = strings.TrimPrefix(raw, ":")
			raw = strings.TrimSpace(cleanCommentSuffix(raw))
			ids, idErrs := parseIDs(raw, i+1, rel)
			errs = append(errs, idErrs...)
			current.deps = append(current.deps, ids...)
		case strings.HasPrefix(lower, "@"):
			errs = append(errs, ScanError{File: rel, Line: i + 1, Msg: "unknown metadata (use @todo-id or @todo-deps)"})
		default:
			// plain comment allowed
		}
	}

	// flush pending
	if current != nil {
		if current.id != "" && !current.invalid {
			todos[current.id] = graph.Todo{ID: current.id, File: rel, Line: current.line}
			for _, dep := range current.deps {
				edges = append(edges, graph.Edge{From: dep, To: current.id, Type: "blocks"})
			}
		}
	}

	// convert todo map to slice
	var todoList []graph.Todo
	for _, t := range todos {
		todoList = append(todoList, t)
	}

	return edges, todoList, errs, nil
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
