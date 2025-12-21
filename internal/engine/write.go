package engine

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kuri-sun/todo-graph/internal/graph"
)

// RenderGraphYAML renders the graph to a YAML string.
func RenderGraphYAML(g graph.Graph) string {
	var b strings.Builder
	b.WriteString("version: 1\n\n")

	writeTodos(&b, g.Todos)
	writeEdges(&b, g.Edges)

	return b.String()
}

// WriteGraph renders the graph to .todo-graph (default) or a custom path.
// If outputPath is empty, it writes to root/.todo-graph. Relative paths are
// resolved against root.
func WriteGraph(root, outputPath string, g graph.Graph) error {
	path := outputPath
	if path == "" {
		path = filepath.Join(root, ".todo-graph")
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(RenderGraphYAML(g)), 0o644)
}

func writeTodos(b *strings.Builder, todos map[string]graph.Todo) {
	if len(todos) == 0 {
		b.WriteString("todos: {}\n\n")
		return
	}
	b.WriteString("todos:\n")
	ids := make([]string, 0, len(todos))
	for id := range todos {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		t := todos[id]
		b.WriteString("  " + id + ":\n")
		b.WriteString("    file: " + yamlQuote(t.File) + "\n")
		b.WriteString("    line: " + strconv.Itoa(t.Line) + "\n\n")
	}
}

func writeEdges(b *strings.Builder, edges []graph.Edge) {
	b.WriteString("edges:\n")
	if len(edges) == 0 {
		b.WriteString("  []\n")
		return
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})

	for _, e := range edges {
		b.WriteString("  - from: " + yamlQuote(e.From) + "\n")
		b.WriteString("    to: " + yamlQuote(e.To) + "\n")
		b.WriteString("    type: " + yamlQuote(e.Type) + "\n")
	}
}

func yamlQuote(value string) string {
	quoted := strconv.Quote(value)
	// strconv.Quote wraps with double quotes; YAML accepts that representation.
	return quoted
}
