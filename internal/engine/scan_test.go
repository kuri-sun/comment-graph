package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanParsesNodesWithDeps(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// @cgraph-id a
// @cgraph-deps b, c
// some comment
// @cgraph-id b
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(g.Edges))
	}
	want := map[string]bool{"b->a": true, "c->a": true}
	for _, e := range g.Edges {
		k := e.From + "->" + e.To
		if !want[k] {
			t.Fatalf("unexpected edge: %+v", e)
		}
		delete(want, k)
	}
	if len(want) != 0 {
		t.Fatalf("missing edges: %+v", want)
	}
}

func TestMetadataStopsAtNonComment(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// @cgraph-id a
// @cgraph-deps b
const x = 1
// @cgraph-deps c
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "metadata without @cgraph-id") {
		t.Fatalf("expected orphan metadata error, got %+v", errs)
	}
	if len(g.Edges) != 1 || g.Edges[0].From != "b" || g.Edges[0].To != "a" {
		t.Fatalf("unexpected edges: %+v", g.Edges)
	}
}

func TestMissingIdErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// @cgraph-deps a
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "metadata without @cgraph-id") {
		t.Fatalf("expected missing-id error, got %+v", errs)
	}
	if len(g.Nodes) != 0 {
		t.Fatalf("expected no nodes, got %+v", g.Nodes)
	}
}

func TestInvalidIdErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// @cgraph-id BadID
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "lowercase") {
		t.Fatalf("expected lowercase error, got %+v", errs)
	}
}

func TestDuplicateIds(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// @cgraph-id dup
`)
	writeFile(t, dir, filepath.Join("sub", "b.go"), `// @cgraph-id dup
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "duplicate") {
		t.Fatalf("expected duplicate error, got %+v", errs)
	}
}

func TestUnknownMetadataErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// @cgraph-id a
// @owner alice
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "unknown metadata") {
		t.Fatalf("expected unknown metadata error, got %+v", errs)
	}
}

func TestSpaceSeparatedDepsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// @cgraph-id a
// @cgraph-deps b c
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "comma-separated") {
		t.Fatalf("expected comma-separated error, got %+v", errs)
	}
}

func TestBlockComments(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.js", `/*
@cgraph-id a
@cgraph-deps b
*/
/* 
@cgraph-id b
*/
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Edges) != 1 || g.Edges[0].From != "b" || g.Edges[0].To != "a" {
		t.Fatalf("unexpected edges: %+v", g.Edges)
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
