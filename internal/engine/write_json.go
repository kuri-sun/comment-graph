package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/kuri-sun/todo-graph/internal/graph"
)

// RenderGraphJSON renders the graph to JSON in the same shape as .comment-graph.json.
func RenderGraphJSON(g graph.Graph) ([]byte, error) {
	payload := map[string]any{
		"version": 1,
		"nodes":   g.Nodes,
		"edges":   g.Edges,
	}
	return json.MarshalIndent(payload, "", "  ")
}

// WriteGraphJSON renders the graph to .comment-graph.json in JSON format.
// If outputPath is empty, it writes to root/.comment-graph.json. Relative paths are resolved against root.
func WriteGraphJSON(root, outputPath string, g graph.Graph, report *CheckReport) error {
	path := outputPath
	if path == "" {
		path = filepath.Join(root, ".comment-graph.json")
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}

	data, err := RenderGraphJSON(g)
	if err != nil {
		return err
	}

	return os.WriteFile(abs, data, 0o644)
}

// NonDependantNodes returns nodes that do not depend on any other node.
// A node depends on another when it appears as the "To" in an edge.
// The result is sorted by ID for stable output.
func NonDependantNodes(g graph.Graph) []graph.Node {
	indegree := make(map[string]int, len(g.Nodes))
	for id := range g.Nodes {
		indegree[id] = 0
	}
	for _, e := range g.Edges {
		if _, ok := indegree[e.To]; ok {
			indegree[e.To]++
		}
	}
	var ids []string
	for id, count := range indegree {
		if count == 0 {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)

	out := make([]graph.Node, 0, len(ids))
	for _, id := range ids {
		out = append(out, g.Nodes[id])
	}
	return out
}

// RenderGraphPayloadJSON renders a tooling-friendly payload containing the graph and a validation report.
// This is not the on-disk .comment-graph format; it's intended for editor integrations.
func RenderGraphPayloadJSON(g graph.Graph, report *CheckReport, includeNonDependants bool) ([]byte, error) {
	payload := map[string]any{
		"graph": map[string]any{
			"version": 1,
			"nodes":   g.Nodes,
			"edges":   g.Edges,
		},
	}
	if report != nil {
		payload["report"] = report
	}
	if includeNonDependants {
		payload["nonDependantNodes"] = NonDependantNodes(g)
	}
	return json.MarshalIndent(payload, "", "  ")
}
