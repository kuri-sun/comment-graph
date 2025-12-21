package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuri-sun/todo-graph/internal/graph"
)

func TestWriteGraphJSON(t *testing.T) {
	dir := t.TempDir()
	g := graph.Graph{
		Todos: map[string]graph.Todo{
			"a": {ID: "a", File: "a.go", Line: 1},
		},
		Edges: []graph.Edge{
			{From: "a", To: "b", Type: "blocks"},
		},
	}

	if err := WriteGraphJSON(dir, "", g, nil); err != nil {
		t.Fatalf("write json: %v", err)
	}
	path := filepath.Join(dir, ".todo-graph.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded["version"] != float64(1) {
		t.Fatalf("expected version 1, got %v", decoded["version"])
	}
}
