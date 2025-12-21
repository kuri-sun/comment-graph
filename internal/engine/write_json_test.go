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
		Nodes: map[string]graph.Node{
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
		Nodes: map[string]graph.Node{
			"a": {ID: "a", File: "a.go", Line: 1},
		},
	}
	report := CheckReport{
		ScanErrors: []ScanError{
			{File: "a.go", Line: 1, Msg: "oops"},
		},
		Isolated: []string{"a"},
	}

	data, err := RenderGraphPayloadJSON(g, &report, false)
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
	if _, ok := graphPayload["nodes"]; !ok {
		t.Fatalf("expected nodes in graph payload, got: %v", graphPayload)
	}
	rep, ok := decoded["report"].(map[string]any)
	if !ok {
		t.Fatalf("expected report in payload, got: %v", decoded)
	}
	if _, ok := rep["scanErrors"]; !ok {
		t.Fatalf("expected scanErrors in report, got: %v", rep)
	}
}

func TestRenderGraphJSONWithNonDependants(t *testing.T) {
	g := graph.Graph{
		Nodes: map[string]graph.Node{
			"a": {ID: "a", File: "a.go", Line: 1},
			"b": {ID: "b", File: "b.go", Line: 2},
			"c": {ID: "c", File: "c.go", Line: 3},
		},
		Edges: []graph.Edge{
			{From: "a", To: "b", Type: "blocks"},
		},
	}

	data, err := RenderGraphPayloadJSON(g, nil, true)
	if err != nil {
		t.Fatalf("render json: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	raw, ok := decoded["nonDependantNodes"].([]any)
	if !ok {
		t.Fatalf("expected nonDependantNodes in payload, got: %v", decoded)
	}
	if len(raw) != 2 {
		t.Fatalf("expected 2 non-dependant nodes, got: %d", len(raw))
	}

	ids := map[string]bool{"a": false, "c": false}
	for _, entry := range raw {
		node, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("expected node object, got: %v", entry)
		}
		id, _ := node["ID"].(string)
		if _, ok := ids[id]; ok {
			ids[id] = true
		}
	}
	for id, seen := range ids {
		if !seen {
			t.Fatalf("expected non-dependant node %s in payload", id)
		}
	}
}
