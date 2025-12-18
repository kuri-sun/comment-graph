package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"todo-graph/internal/graph"
)

var (
	todoLinePattern = regexp.MustCompile(`TODO:?\s*(\[#([^\]]*)\])?(.*)`)
	todoIDPattern   = regexp.MustCompile(`^[a-z0-9_-]+$`)
	commentLine     = regexp.MustCompile(`^\s*(//|#)`)
)

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

	for i := 0; i < len(lines); i++ {
		line := lines[i]
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
		case "depends-on":
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
	clean := strings.ReplaceAll(raw, ",", " ")
	fields := strings.Fields(clean)
	var ids []string
	var errs []ScanError
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if strings.HasPrefix(f, "#") {
			f = strings.TrimPrefix(f, "#")
		} else {
			errs = append(errs, ScanError{File: file, Line: line, Msg: fmt.Sprintf("id %q must start with #", f)})
			continue
		}
		if !todoIDPattern.MatchString(f) {
			errs = append(errs, ScanError{
				File: file,
				Line: line,
				Msg:  fmt.Sprintf("id %q must use lowercase letters, digits, hyphens, or underscores", f),
			})
			continue
		}
		ids = append(ids, f)
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
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return fmt.Sprintf("todo-%d", line)
	}
	for _, field := range strings.Fields(desc) {
		id := slugify(field)
		if todoIDPattern.MatchString(id) && id != "" {
			return id
		}
	}
	id := slugify(desc)
	if id == "" {
		return fmt.Sprintf("todo-%d", line)
	}
	return id
}

func slugify(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
			lastDash = false
		} else {
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
