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
