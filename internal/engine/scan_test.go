package engine

import (
	"os"
	"path/filepath"
	"strings"
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

func TestScanRejectsEmptyID(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#]
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 scan error, got %d: %+v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Msg, "must not be empty") {
		t.Fatalf("expected empty id error, got %+v", errs)
	}
}

func TestScanRejectsUppercaseID(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#Bad]
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 scan error, got %d: %+v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Msg, "lowercase letters, digits, hyphens, or underscores") {
		t.Fatalf("expected charset error, got %+v", errs)
	}
}

func TestScanDetectsDuplicateIDs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#dup]
`)
	writeFile(t, dir, filepath.Join("sub", "b.go"), `// TODO:[#dup]
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 scan error, got %d: %+v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Msg, "duplicate TODO id") {
		t.Fatalf("unexpected error message: %v", errs[0])
	}
}

func TestScanParsesDependsAndBlocksLists(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#a]
// depends-on: #b, #c
// blocks: #d #e
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(g.Edges))
	}
	want := map[string]bool{
		"b->a": true,
		"c->a": true,
		"a->d": true,
		"a->e": true,
	}
	for _, e := range g.Edges {
		key := e.From + "->" + e.To
		if !want[key] {
			t.Fatalf("unexpected edge: %+v", e)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing edges: %+v", want)
	}
}

func TestScanMetadataStopsAtNextTODO(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#a]
// depends-on: #b
// TODO:[#c]
// depends-on: #d
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	want := map[string]bool{
		"b->a": true,
		"d->c": true,
	}
	if len(g.Edges) != len(want) {
		t.Fatalf("expected %d edges, got %d", len(want), len(g.Edges))
	}
	for _, e := range g.Edges {
		key := e.From + "->" + e.To
		if !want[key] {
			t.Fatalf("unexpected edge: %+v", e)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing edges: %+v", want)
	}
}

func TestScanRejectsMetadataWithoutHash(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#a]
// depends-on: b
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 scan error, got %d: %+v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Msg, "must start with #") {
		t.Fatalf("expected missing hash error, got %+v", errs)
	}
}

func TestScanUnknownMetadataKeyIgnored(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#a]
// owner: someone
// status: todo
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Edges) != 0 {
		t.Fatalf("expected 0 edges, got %d", len(g.Edges))
	}
	if len(g.Todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(g.Todos))
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
