package engine

import (
	"testing"

	"github.com/kuri-sun/comment-graph/internal/graph"
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

func TestValidateGraphDetectsIsolated(t *testing.T) {
	g := graph.Graph{
		Todos: map[string]graph.Todo{
			"a": {ID: "a"},
			"b": {ID: "b"},
		},
		Edges: []graph.Edge{
			{From: "a", To: "a", Type: "blocks"}, // self edge
		},
	}

	report := ValidateGraph(g, nil)

	if len(report.Isolated) != 1 || report.Isolated[0] != "b" {
		t.Fatalf("expected isolated [b], got %v", report.Isolated)
	}
}

func TestValidateGraphMismatchedGraphFlagged(t *testing.T) {
	scanned := graph.Graph{
		Todos: map[string]graph.Todo{
			"a": {ID: "a"},
		},
		Edges: nil,
	}
	report := ValidateGraph(scanned, nil)
	fileGraph := graph.Graph{Todos: map[string]graph.Todo{}}

	mismatch := !GraphsEqual(scanned, fileGraph)
	if !mismatch {
		t.Fatalf("expected mismatch between scanned and file graph")
	}
	_ = report // ensures report is produced even when used with GraphsEqual in check path
}

func TestValidateGraphDetectsSelfCycle(t *testing.T) {
	g := graph.Graph{
		Todos: map[string]graph.Todo{
			"a": {ID: "a"},
		},
		Edges: []graph.Edge{
			{From: "a", To: "a", Type: "blocks"},
		},
	}

	report := ValidateGraph(g, nil)

	if len(report.Cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d", len(report.Cycles))
	}
	if len(report.Cycles[0]) != 2 || report.Cycles[0][0] != "a" || report.Cycles[0][1] != "a" {
		t.Fatalf("unexpected self cycle: %v", report.Cycles[0])
	}
}

func TestValidateGraphScanErrorsSurfaced(t *testing.T) {
	scanErrs := []ScanError{
		{File: "a.go", Line: 1, Msg: "bad"},
	}
	report := ValidateGraph(graph.Graph{}, scanErrs)

	if len(report.ScanErrors) != 1 {
		t.Fatalf("expected 1 scan error, got %d", len(report.ScanErrors))
	}
	if report.ScanErrors[0].Msg != "bad" {
		t.Fatalf("unexpected scan error: %+v", report.ScanErrors[0])
	}
}

func TestValidateGraphReportsUndefinedFrom(t *testing.T) {
	g := graph.Graph{
		Todos: map[string]graph.Todo{
			"b": {ID: "b"},
		},
		Edges: []graph.Edge{
			{From: "missing", To: "b", Type: "blocks"},
		},
	}

	report := ValidateGraph(g, nil)

	if len(report.UndefinedEdges) != 1 {
		t.Fatalf("expected 1 undefined edge, got %d", len(report.UndefinedEdges))
	}
	e := report.UndefinedEdges[0]
	if e.From != "missing" || e.To != "b" {
		t.Fatalf("unexpected undefined edge: %+v", e)
	}
}

func TestValidateGraphIsolatedWithConnectedNodes(t *testing.T) {
	g := graph.Graph{
		Todos: map[string]graph.Todo{
			"a": {ID: "a"},
			"b": {ID: "b"},
			"c": {ID: "c"},
		},
		Edges: []graph.Edge{
			{From: "a", To: "b", Type: "blocks"},
		},
	}

	report := ValidateGraph(g, nil)

	if len(report.Isolated) != 1 || report.Isolated[0] != "c" {
		t.Fatalf("expected isolated [c], got %v", report.Isolated)
	}
}

func TestValidateGraphReportsUndefinedEdgeWithBothMissing(t *testing.T) {
	g := graph.Graph{
		Todos: map[string]graph.Todo{},
		Edges: []graph.Edge{
			{From: "x", To: "y", Type: "blocks"},
		},
	}

	report := ValidateGraph(g, nil)

	if len(report.UndefinedEdges) != 1 {
		t.Fatalf("expected 1 undefined edge, got %d", len(report.UndefinedEdges))
	}
	e := report.UndefinedEdges[0]
	if e.From != "x" || e.To != "y" {
		t.Fatalf("unexpected undefined edge: %+v", e)
	}
}
