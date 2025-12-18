package engine

import (
	"testing"

	"todo-graph/internal/graph"
)

// First check test: undefined references should be reported.
func TestValidateGraphReportsUndefinedEdges(t *testing.T) {
	g := graph.Graph{
		Todos: map[string]graph.Todo{
			"a": {ID: "a"},
		},
		Edges: []graph.Edge{
			{From: "a", To: "missing", Type: "blocks"},
		},
	}

	report := ValidateGraph(g, nil)

	if len(report.UndefinedEdges) != 1 {
		t.Fatalf("expected 1 undefined edge, got %d", len(report.UndefinedEdges))
	}
	edge := report.UndefinedEdges[0]
	if edge.From != "a" || edge.To != "missing" {
		t.Fatalf("unexpected edge: %+v", edge)
	}
}

func TestValidateGraphDetectsCycles(t *testing.T) {
	g := graph.Graph{
		Todos: map[string]graph.Todo{
			"a": {ID: "a"},
			"b": {ID: "b"},
		},
		Edges: []graph.Edge{
			{From: "a", To: "b", Type: "blocks"},
			{From: "b", To: "a", Type: "blocks"},
		},
	}

	report := ValidateGraph(g, nil)

	if len(report.Cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d", len(report.Cycles))
	}
	want := []string{"a", "b", "a"}
	got := report.Cycles[0]
	if len(got) != len(want) {
		t.Fatalf("unexpected cycle length: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected cycle: %v", got)
		}
	}
}
