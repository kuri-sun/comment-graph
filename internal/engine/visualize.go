package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kuri-sun/todo-graph/internal/graph"
)

// RenderMermaid renders the graph in Mermaid syntax.
func RenderMermaid(g graph.Graph) string {
	var b strings.Builder
	b.WriteString("graph TD\n")

	edges := sortEdges(g.Edges)
	for _, e := range edges {
		b.WriteString(fmt.Sprintf("  %s --> %s\n", e.From, e.To))
	}
	if len(edges) == 0 {
		// ensure empty graph still outputs something
		b.WriteString("  %% no edges\n")
	}
	return b.String()
}

func sortEdges(edges []graph.Edge) []graph.Edge {
	out := append([]graph.Edge{}, edges...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].From == out[j].From {
			return out[i].To < out[j].To
		}
		return out[i].From < out[j].From
	})
	return out
}
