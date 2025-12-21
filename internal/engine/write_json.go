package engine

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/kuri-sun/comment-graph/internal/graph"
)

// RenderGraphJSON renders the graph to JSON in the same shape as .todo-graph.json.
func RenderGraphJSON(g graph.Graph) ([]byte, error) {
	payload := map[string]any{
		"version": 1,
		"todos":   g.Todos,
		"edges":   g.Edges,
	}
	return json.MarshalIndent(payload, "", "  ")
}

// WriteGraphJSON renders the graph to .todo-graph.json in JSON format.
// If outputPath is empty, it writes to root/.todo-graph.json. Relative paths are resolved against root.
func WriteGraphJSON(root, outputPath string, g graph.Graph, report *CheckReport) error {
	path := outputPath
	if path == "" {
		path = filepath.Join(root, ".todo-graph.json")
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

// RenderGraphPayloadJSON renders a tooling-friendly payload containing the graph and a validation report.
// This is not the on-disk .todo-graph format; it's intended for editor integrations.
func RenderGraphPayloadJSON(g graph.Graph, report *CheckReport) ([]byte, error) {
	payload := map[string]any{
		"graph": map[string]any{
			"version": 1,
			"todos":   g.Todos,
			"edges":   g.Edges,
		},
	}
	if report != nil {
		payload["report"] = report
	}
	return json.MarshalIndent(payload, "", "  ")
}
