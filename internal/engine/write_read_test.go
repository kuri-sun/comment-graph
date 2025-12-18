package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"todo-graph/internal/graph"
)

// First write/read test: round-trip a simple graph through .todo-graph.
func TestWriteReadGraphRoundTrip(t *testing.T) {
	dir := t.TempDir()
	g := graph.Graph{
		Todos: map[string]graph.Todo{
			"a": {ID: "a", File: "a.go", Line: 1},
			"b": {ID: "b", File: "b.go", Line: 2},
		},
		Edges: []graph.Edge{
			{From: "a", To: "b", Type: "blocks"},
		},
	}

	if err := WriteGraph(dir, g); err != nil {
		t.Fatalf("write graph: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".todo-graph")); err != nil {
		t.Fatalf(".todo-graph not written: %v", err)
	}

	read, err := ReadGraph(dir)
	if err != nil {
		t.Fatalf("read graph: %v", err)
	}
	if !GraphsEqual(g, read) {
		t.Fatalf("graphs not equal after round trip: %+v vs %+v", g, read)
	}
}

func TestWriteGraphEmptyFormatsSections(t *testing.T) {
	dir := t.TempDir()
	g := graph.Graph{
		Todos: map[string]graph.Todo{},
		Edges: nil,
	}

	if err := WriteGraph(dir, g); err != nil {
		t.Fatalf("write graph: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".todo-graph"))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "todos: {}") {
		t.Fatalf("expected empty todos map, got: %s", content)
	}
	if !strings.Contains(content, "edges:\n  []") {
		t.Fatalf("expected empty edges list, got: %s", content)
	}
}
