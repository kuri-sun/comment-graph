package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kuri-sun/todo-graph/internal/graph"
)

var depsTodoPattern = mustDefaultPattern()

func mustDefaultPattern() *regexp.Regexp {
	p, err := compileTodoPattern(defaultKeywords)
	if err != nil {
		panic(err)
	}
	return p
}

// UpdateDeps updates @todo-deps for a given TODO id in its source file.
// It validates that the TODO and all parents exist in the scanned graph and
// rejects TODO blocks that already declare multiple @todo-deps lines.
func UpdateDeps(root string, g graph.Graph, target string, parents []string) error {
	return updateDeps(root, g, target, parents, false)
}

// UpdateDepsAllowEmpty allows clearing deps (used by detach).
func UpdateDepsAllowEmpty(root string, g graph.Graph, target string, parents []string) error {
	return updateDeps(root, g, target, parents, true)
}

func updateDeps(root string, g graph.Graph, target string, parents []string, allowEmpty bool) error {
	t, ok := g.Todos[target]
	if !ok {
		return fmt.Errorf("TODO %q not found", target)
	}
	for _, p := range parents {
		if _, ok := g.Todos[p]; !ok {
			return fmt.Errorf("parent TODO %q not found", p)
		}
	}

	if len(parents) == 0 && !allowEmpty {
		return fmt.Errorf("at least one parent is required")
	}

	path := filepath.Join(root, t.File)
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	if t.Line <= 0 || t.Line-1 >= len(lines) {
		return fmt.Errorf("invalid line for %q: %d", target, t.Line)
	}

	// find metadata block boundaries
	todoIdx := t.Line - 1
	insertIdx := todoIdx + 1
	var depsIdx int = -1
	var depsCount int

	for i := todoIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		cleaned := strings.TrimSpace(commentLine.ReplaceAllString(trimmed, ""))
		lower := strings.ToLower(cleaned)
		if trimmed == "" || (!strings.HasPrefix(lower, "@") && depsTodoPattern.MatchString(cleaned)) || !commentLine.MatchString(trimmed) {
			break
		}
		if strings.HasPrefix(lower, "@todo-id") {
			insertIdx = i + 1
		}
		if strings.HasPrefix(lower, "@todo-deps") {
			depsCount++
			if depsCount > 1 {
				return fmt.Errorf("multiple @todo-deps entries found for %q in %s:%d", target, t.File, i+1)
			}
			depsIdx = i
			insertIdx = i
		}
	}

	if depsCount > 1 {
		return fmt.Errorf("multiple @todo-deps entries found for %q in %s", target, t.File)
	}

	if len(parents) == 0 && allowEmpty {
		if depsIdx >= 0 {
			lines = append(lines[:depsIdx], lines[depsIdx+1:]...)
		}
		return writeLines(path, lines)
	}

	depsLine := formatDepsLine(lines[todoIdx], parents)

	if depsIdx >= 0 {
		lines[depsIdx] = depsLine
	} else {
		if insertIdx > len(lines) {
			insertIdx = len(lines)
		}
		lines = append(lines[:insertIdx], append([]string{depsLine}, lines[insertIdx:]...)...)
	}

	return writeLines(path, lines)
}

func formatDepsLine(todoLine string, parents []string) string {
	prefix, suffix := commentDelimiters(todoLine)
	return fmt.Sprintf("%s @todo-deps %s%s", prefix, strings.Join(parents, ", "), suffix)
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func writeLines(path string, lines []string) error {
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

// CurrentParents returns the list of parent ids (blocks edges) for a child.
func CurrentParents(g graph.Graph, child string) []string {
	var parents []string
	for _, e := range g.Edges {
		if e.To == child && e.Type == "blocks" {
			parents = append(parents, e.From)
		}
	}
	sort.Strings(parents)
	parents = dedupeStrings(parents)
	return parents
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
