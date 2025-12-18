package engine

import (
	"os"
	"path/filepath"
	"testing"
)

// First test: parsing TODO with and without colon prefix.
func TestScanParsesTodoWithAndWithoutColon(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#a]
// TODO[#b]
// depends-on: #a
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(g.Todos))
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	e := g.Edges[0]
	if e.From != "a" || e.To != "b" || e.Type != "blocks" {
		t.Fatalf("unexpected edge: %+v", e)
	}
}

func TestScanMetadataStopsAtNonComment(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#a]
// depends-on: #b
const x = 1
// depends-on: #c
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	e := g.Edges[0]
	if e.From != "b" || e.To != "a" {
		t.Fatalf("unexpected edge: %+v", e)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
