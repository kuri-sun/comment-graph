package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// WriteErrorsJSON writes the validation report to a JSON file.
// If outputPath is empty, it writes to root/comment-graph.errors.json. Relative paths are resolved against root.
func WriteErrorsJSON(root, outputPath string, report CheckReport) error {
	path := outputPath
	if path == "" {
		path = filepath.Join(root, "comment-graph.errors.json")
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

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(abs, data, 0o644)
}
