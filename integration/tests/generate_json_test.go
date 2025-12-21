package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Integration: generate should support writing JSON output without touching YAML.
func TestCLIGenerateWritesJSON(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("sample", "index.ts"), tmp)
	copyFixtureFile(t, filepath.Join("sample", "users.ts"), tmp)

	bin := buildCLI(t)
	runCmd(t, bin, tmp, "generate", "--format", "json")

	path := filepath.Join(tmp, ".comment-graph.json")
	data := readFile(t, path)
	var decoded map[string]any
	if err := json.Unmarshal([]byte(data), &decoded); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if decoded["version"] != float64(1) {
		t.Fatalf("expected version 1, got %v", decoded["version"])
	}
	if _, err := os.Stat(filepath.Join(tmp, ".comment-graph")); err != nil && !os.IsNotExist(err) {
		t.Fatalf("stat .comment-graph: %v", err)
	}
}
