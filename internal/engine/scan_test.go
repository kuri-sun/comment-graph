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
// DEPS: #a
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
// DEPS: #b
const x = 1
// DEPS: #c
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

func TestScanParsesDependsList(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#a]
// DEPS: #b, #c
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(g.Edges))
	}
	want := map[string]bool{
		"b->a": true,
		"c->a": true,
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
// DEPS: #b
// TODO:[#c]
// DEPS: #d
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
// DEPS: b
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

func TestScanDerivesIDWhenMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO: cache-user add cache layer
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(g.Todos))
	}
	if _, ok := g.Todos["todo-1"]; !ok {
		t.Fatalf("expected derived id todo-1, got %+v", g.Todos)
	}
}

func TestScanDerivesIDFromFirstToken(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO: Remove legacy endpoints!
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if _, ok := g.Todos["todo-1"]; !ok {
		t.Fatalf("expected derived id todo-1, got %+v", g.Todos)
	}
}

func TestScanIgnoresBinaryFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.bin", string([]byte{0x00, 0x01, 0x02}))

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Todos) != 0 || len(g.Edges) != 0 {
		t.Fatalf("expected empty graph, got %+v", g)
	}
}

func TestScanParsesBlockCommentTODOs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.js", `/*
TODO:[#a]
DEPS: #b
*/
/* TODO:[#b] */
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
	if len(g.Edges) != 1 || g.Edges[0].From != "b" || g.Edges[0].To != "a" {
		t.Fatalf("unexpected edges: %+v", g.Edges)
	}
}

func TestScanParsesHtmlStyleComments(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.html", `<!-- TODO:[#a] -->
<!-- DEPS: #b -->
<!-- TODO:[#b] -->
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
}

func TestScanParsesTripleQuoteComments(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.py", `"""
TODO:[#a]
DEPS: #b
"""
""" TODO:[#b] """
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
	if len(g.Edges) != 1 || g.Edges[0].From != "b" || g.Edges[0].To != "a" {
		t.Fatalf("unexpected edges: %+v", g.Edges)
	}
}

func TestScanIgnoresTodoInCode(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `package main

var x = "TODO: not a todo"
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Todos) != 0 {
		t.Fatalf("expected no todos, got %+v", g.Todos)
	}
}

func TestScanParsesInlineComments(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.ts", `const x = 1; // TODO:[#a]
function foo() { /* TODO:[#b] */ }
SELECT 1 -- TODO:[#c]
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Todos) != 3 {
		t.Fatalf("expected 3 todos, got %d", len(g.Todos))
	}
}

func TestScanBlockStartEndSameLine(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.js", `/* TODO:[#a] DEPS: #b */
/* TODO:[#b] */
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

func TestScanMultipleTodosInBlock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.js", `/*
TODO:[#a]
TODO:[#b]
DEPS: #a
*/
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Todos) != 2 || len(g.Edges) != 1 {
		t.Fatalf("unexpected graph: %+v", g)
	}
	if g.Edges[0].From != "a" || g.Edges[0].To != "b" {
		t.Fatalf("unexpected edge: %+v", g.Edges[0])
	}
}

func TestScanUnterminatedBlockDoesNotLeak(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.js", `/* TODO:[#a]
// TODO:[#b]
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Todos) != 2 {
		t.Fatalf("expected 2 todos, got %+v", g.Todos)
	}
	if len(g.Edges) != 0 {
		t.Fatalf("expected 0 edges, got %+v", g.Edges)
	}
}

func TestScanRejectsSpaceSeparatedDeps(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO:[#a]
// DEPS: #b #c
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "comma-separated") {
		t.Fatalf("expected comma-separated error, got %+v", errs)
	}
}

func TestScanMixedLanguages(t *testing.T) {
	dir := t.TempDir()
	copyFixture(t, filepath.Join("mixed", "go"), filepath.Join(dir, "go"))
	copyFixture(t, filepath.Join("mixed", "ts"), filepath.Join(dir, "ts"))
	copyFixture(t, filepath.Join("mixed", "js"), filepath.Join(dir, "js"))

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %+v", errs)
	}
	if len(g.Todos) != 6 {
		t.Fatalf("expected 6 todos, got %d", len(g.Todos))
	}

	wantEdges := map[string]bool{
		"go-root->ts-root":   true,
		"js-root->ts-root":   true,
		"js-root->js-helper": true,
		"ts-root->go-root":   true,
	}

	for _, e := range g.Edges {
		key := e.From + "->" + e.To
		if wantEdges[key] {
			delete(wantEdges, key)
		}
	}

	if len(wantEdges) != 0 {
		t.Fatalf("missing edges: %+v", wantEdges)
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

func copyFixture(t *testing.T, srcRel, dst string) {
	t.Helper()
	src := filepath.Join(findModuleRoot(t), "integration", "testdata", srcRel)
	err := filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found")
		}
		dir = parent
	}
}
