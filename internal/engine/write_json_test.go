package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuri-sun/comment-graph/internal/graph"
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
	path := filepath.Join(dir, ".comment-graph.json")
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

func TestRenderGraphJSONWithReport(t *testing.T) {
	g := graph.Graph{
		Todos: map[string]graph.Todo{
			"a": {ID: "a", File: "a.go", Line: 1},
		},
	}
	report := CheckReport{
		ScanErrors: []ScanError{
			{File: "a.go", Line: 1, Msg: "oops"},
		},
		Isolated: []string{"a"},
	}

	data, err := RenderGraphPayloadJSON(g, &report)
	if err != nil {
		t.Fatalf("render json: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	graphPayload, ok := decoded["graph"].(map[string]any)
	if !ok {
		t.Fatalf("expected graph in payload, got: %v", decoded)
	}
	if _, ok := graphPayload["todos"]; !ok {
		t.Fatalf("expected todos in graph payload, got: %v", graphPayload)
	}
	rep, ok := decoded["report"].(map[string]any)
	if !ok {
		t.Fatalf("expected report in payload, got: %v", decoded)
	}
	if _, ok := rep["scanErrors"]; !ok {
		t.Fatalf("expected scanErrors in report, got: %v", rep)
	}
}
