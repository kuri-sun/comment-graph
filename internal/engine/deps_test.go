package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuri-sun/comment-graph/internal/graph"
)

func TestUpdateDepsInsertsLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.go")
	content := `// @cgraph-id child
// some comment
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	g := graph.Graph{
		Nodes: map[string]graph.Node{
			"child":  {ID: "child", File: "file.go", Line: 1},
			"parent": {ID: "parent", File: "file.go", Line: 4},
		},
	}

	if err := UpdateDeps(dir, g, "child", []string{"parent"}); err != nil {
		t.Fatalf("update deps: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !strings.Contains(string(data), "@cgraph-deps parent") {
		t.Fatalf("expected deps line inserted, got:\n%s", data)
	}
}

func TestUpdateDepsRejectsMultipleDepsLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.go")
	content := `// @cgraph-id child
// @cgraph-deps a
// @cgraph-deps b
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	g := graph.Graph{
		Nodes: map[string]graph.Node{
			"child": {ID: "child", File: "file.go", Line: 1},
			"a":     {ID: "a", File: "file.go", Line: 5},
			"b":     {ID: "b", File: "file.go", Line: 6},
		},
	}

	if err := UpdateDeps(dir, g, "child", []string{"a"}); err == nil {
		t.Fatalf("expected error for multiple deps lines")
	}
}
