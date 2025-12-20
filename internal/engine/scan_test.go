package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanParsesTodoWithDeps(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO: first task
// @todo-id a
// @todo-deps b, c

// TODO: second task
// @todo-id b
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
	writeFile(t, dir, "a.go", `// TODO: first
// @todo-id a
// @todo-deps b
const x = 1
// @todo-deps c
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "metadata without a preceding TODO") {
		t.Fatalf("expected orphan metadata error, got %+v", errs)
	}
	if len(g.Edges) != 1 || g.Edges[0].From != "b" || g.Edges[0].To != "a" {
		t.Fatalf("unexpected edges: %+v", g.Edges)
	}
}

func TestMissingIdErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO: no id
// @todo-deps a
`)

	g, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %+v", errs)
	}
	if len(g.Todos) != 0 {
		t.Fatalf("expected missing-id TODO to be skipped, got %+v", g.Todos)
	}
}

func TestInvalidIdErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO: bad id
// @todo-id BadID
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
	writeFile(t, dir, "a.go", `// TODO: dup
// @todo-id dup
`)
	writeFile(t, dir, filepath.Join("sub", "b.go"), `// TODO: dup2
// @todo-id dup
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "duplicate") {
		t.Fatalf("expected duplicate error, got %+v", errs)
	}
}

func TestInlineTodoErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `const x = 1 // TODO: bad
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "comment line") {
		t.Fatalf("expected inline error, got %+v", errs)
	}
}

func TestOrphanMetadataErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// @todo-id orphan
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "metadata without a preceding TODO") {
		t.Fatalf("expected orphan error, got %+v", errs)
	}
}

func TestUnknownMetadataErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `// TODO: something
// @todo-id a
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
	writeFile(t, dir, "a.go", `// TODO: deps bad
// @todo-id a
// @todo-deps b c
`)

	_, errs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "comma-separated") {
		t.Fatalf("expected comma-separated error, got %+v", errs)
	}
}

func TestBlockCommentTodos(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.js", `/*
TODO: first
@todo-id a
@todo-deps b
*/
/* TODO: second
@todo-id b
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

func TestJSXBlockComments(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.tsx", `{/*
TODO: hero headline
@todo-id hero-copy
*/}
{/*
TODO: hook CTA
@todo-id cta-wireup
@todo-deps hero-copy
*/}
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
	if len(g.Edges) != 1 || g.Edges[0].From != "hero-copy" || g.Edges[0].To != "cta-wireup" {
		t.Fatalf("unexpected edges: %+v", g.Edges)
	}
}

func TestJSXSingleLineBlocks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.tsx", `{/* TODO: hero headline */ }
{/* @todo-id hero-copy */}
{/* @todo-deps cta-wireup */}
{/* TODO: hook CTA */ }
{/* @todo-id cta-wireup */}
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
	if len(g.Edges) != 1 || g.Edges[0].From != "cta-wireup" || g.Edges[0].To != "hero-copy" {
		t.Fatalf("unexpected edges: %+v", g.Edges)
	}
}

func TestHtmlComments(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.html", `<!-- TODO: first -->
<!-- @todo-id a -->
<!-- @todo-deps b -->
<!-- TODO: second -->
<!-- @todo-id b -->
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

func TestUnterminatedBlockDoesNotLeak(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.js", `/* TODO: first
@todo-id a
// TODO: second
// @todo-id b
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
}

func TestMixedLanguages(t *testing.T) {
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

func TestIgnoresTodoInCode(t *testing.T) {
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
