package engine

import (
	"sort"

	"github.com/kuri-sun/comment-graph/internal/graph"
)

// GraphsEqual returns true if nodes and edges match, ignoring ordering.
func GraphsEqual(a, b graph.Graph) bool {
	if len(a.Nodes) != len(b.Nodes) {
		return false
	}
	for id, n := range a.Nodes {
		bn, ok := b.Nodes[id]
		if !ok {
			return false
		}
		if n.File != bn.File || n.Line != bn.Line {
			return false
		}
	}

	ae := normalizeEdges(a.Edges)
	be := normalizeEdges(b.Edges)
	if len(ae) != len(be) {
		return false
	}
	for i := range ae {
		if ae[i] != be[i] {
			return false
		}
	}
	return true
}

func normalizeEdges(edges []graph.Edge) []graph.Edge {
	out := append([]graph.Edge{}, edges...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].From == out[j].From {
			if out[i].To == out[j].To {
				return out[i].Type < out[j].Type
			}
			return out[i].To < out[j].To
		}
		return out[i].From < out[j].From
	})
	return out
}
