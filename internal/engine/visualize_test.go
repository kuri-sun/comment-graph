package engine

import (
	"strings"
	"testing"

	"todo-graph/internal/graph"
)

// First visualize test: mermaid ordering.
func TestRenderMermaidOrdersEdges(t *testing.T) {
	g := graph.Graph{
		Edges: []graph.Edge{
			{From: "b", To: "a", Type: "blocks"},
			{From: "a", To: "b", Type: "blocks"},
		},
	}

	out := RenderMermaid(g)

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	// graph TD
	//   a --> b
	//   b --> a
	if !strings.Contains(lines[1], "a --> b") || !strings.Contains(lines[2], "b --> a") {
		t.Fatalf("unexpected ordering: %v", lines)
	}
}

func TestRenderMermaidEmptyGraph(t *testing.T) {
	out := RenderMermaid(graph.Graph{})
	if !strings.Contains(out, "%% no edges") {
		t.Fatalf("expected placeholder for empty graph, got: %s", out)
	}
}
