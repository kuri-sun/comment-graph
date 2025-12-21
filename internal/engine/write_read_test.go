package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuri-sun/todo-graph/internal/graph"
)

// First write/read test: round-trip a simple graph through .comment-graph.
func TestWriteReadGraphRoundTrip(t *testing.T) {
	dir := t.TempDir()
	g := graph.Graph{
		Nodes: map[string]graph.Node{
			"a": {ID: "a", File: "a.go", Line: 1},
			"b": {ID: "b", File: "b.go", Line: 2},
		},
		Edges: []graph.Edge{
			{From: "a", To: "b", Type: "blocks"},
		},
	}

	if err := WriteGraph(dir, "", g); err != nil {
		t.Fatalf("write graph: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".comment-graph")); err != nil {
		t.Fatalf(".comment-graph not written: %v", err)
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
		Nodes: map[string]graph.Node{},
		Edges: nil,
	}

	if err := WriteGraph(dir, "", g); err != nil {
		t.Fatalf("write graph: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".comment-graph"))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "nodes: {}") {
		t.Fatalf("expected empty nodes map, got: %s", content)
	}
	if !strings.Contains(content, "edges:\n  []") {
		t.Fatalf("expected empty edges list, got: %s", content)
	}
}

func TestWriteGraphCustomPath(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "artifacts", "custom.comment-graph")

	g := graph.Graph{
		Nodes: map[string]graph.Node{
			"custom": {ID: "custom", File: "a.go", Line: 1},
		},
		Edges: nil,
	}

	if err := WriteGraph(dir, output, g); err != nil {
		t.Fatalf("write graph: %v", err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected custom output path to exist: %v", err)
	}
	content := readFile(t, output)
	if !strings.Contains(content, "custom:") {
		t.Fatalf("expected content to contain todo id, got: %s", content)
	}
	if _, err := os.Stat(filepath.Join(dir, ".comment-graph")); err == nil {
		t.Fatalf("default .comment-graph should not be written when output path is provided")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat default .comment-graph: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return string(data)
}
