package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuri-sun/comment-graph/internal/graph"
)

func TestWriteErrorsJSON(t *testing.T) {
	dir := t.TempDir()
	report := CheckReport{
		UndefinedEdges: []graph.Edge{{From: "missing", To: "a", Type: "blocks"}},
		Isolated:       []string{"lonely"},
	}

	if err := WriteErrorsJSON(dir, "", report); err != nil {
		t.Fatalf("write errors: %v", err)
	}
	path := filepath.Join(dir, ".comment-graph.errors.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read errors: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := decoded["undefinedEdges"]; !ok {
		t.Fatalf("expected undefined edges in errors json, got: %v", decoded)
	}
	if _, ok := decoded["isolated"]; !ok {
		t.Fatalf("expected isolated in errors json, got: %v", decoded)
	}
}
