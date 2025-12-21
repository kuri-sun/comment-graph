package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuri-sun/todo-graph/internal/graph"
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
	path := filepath.Join(dir, ".todo-graph.errors.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read errors: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := decoded["UndefinedEdges"]; !ok {
		t.Fatalf("expected undefined edges in errors json, got: %v", decoded)
	}
	if _, ok := decoded["Isolated"]; !ok {
		t.Fatalf("expected isolated in errors json, got: %v", decoded)
	}
}
