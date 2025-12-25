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

func TestSupportedCommentStyles(t *testing.T) {
	cases := []struct {
		name      string
		filename  string
		content   string
		wantNodes []string
		wantEdges []string
	}{
		{
			name:     "hash comments (py/sh)",
			filename: "main.py",
			content: `# @cgraph-id a
# @cgraph-deps b
# @cgraph-id b
`,
			wantNodes: []string{"a", "b"},
			wantEdges: []string{"b->a"},
		},
		{
			name:     "sql dash dash",
			filename: "query.sql",
			content: `-- @cgraph-id a
-- @cgraph-deps b
-- @cgraph-id b
`,
			wantNodes: []string{"a", "b"},
			wantEdges: []string{"b->a"},
		},
		{
			name:     "html comments",
			filename: "index.html",
			content: `<!--
@cgraph-id a
@cgraph-deps b
-->
<!-- @cgraph-id b -->
`,
			wantNodes: []string{"a", "b"},
			wantEdges: []string{"b->a"},
		},
		{
			name:     "python triple quotes",
			filename: "doc.py",
			content: `"""
@cgraph-id a
@cgraph-deps b
"""
"""
@cgraph-id b
"""
`,
			wantNodes: []string{"a", "b"},
			wantEdges: []string{"b->a"},
		},
		{
			name:     "jsx style block",
			filename: "component.jsx",
			content: `{/*
@cgraph-id a
@cgraph-deps b
*/}
{/* @cgraph-id b */}
`,
			wantNodes: []string{"a", "b"},
			wantEdges: []string{"b->a"},
		},
		{
			name:     "inline trailing comment ignored",
			filename: "inline.go",
			content: `package main

func main() { fmt.Println("hi") } // @cgraph-id a
`,
			wantNodes: []string{},
			wantEdges: []string{},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, tt.filename, tt.content)

			g, errs, err := Scan(dir)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}
			if len(errs) != 0 {
				t.Fatalf("unexpected scan errors: %+v", errs)
			}
			if len(g.Nodes) != len(tt.wantNodes) {
				t.Fatalf("expected %d nodes, got %d", len(tt.wantNodes), len(g.Nodes))
			}
			for _, id := range tt.wantNodes {
				if _, ok := g.Nodes[id]; !ok {
					t.Fatalf("missing node %s", id)
				}
			}
			if len(g.Edges) != len(tt.wantEdges) {
				t.Fatalf("expected %d edges, got %d", len(tt.wantEdges), len(g.Edges))
			}
			edgeSet := make(map[string]bool)
			for _, e := range g.Edges {
				edgeSet[e.From+"->"+e.To] = true
			}
			for _, e := range tt.wantEdges {
				if !edgeSet[e] {
					t.Fatalf("missing edge %s", e)
				}
			}
		})
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
