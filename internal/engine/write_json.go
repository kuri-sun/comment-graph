package engine

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/kuri-sun/todo-graph/internal/graph"
)

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

	payload := map[string]any{
		"version": 1,
		"todos":   g.Todos,
		"edges":   g.Edges,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(abs, data, 0o644)
}
