package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type graphPayload struct {
	Graph struct {
		Version int `json:"version"`
		Nodes   map[string]struct {
			File string `json:"file"`
			Line int    `json:"line"`
		} `json:"nodes"`
		Edges []struct {
			From string `json:"from"`
			To   string `json:"to"`
			Type string `json:"type"`
		} `json:"edges"`
	} `json:"graph"`
	Report struct {
		ScanErrors     []struct{ Msg string }      `json:"scanErrors"`
		UndefinedEdges []struct{ From, To string } `json:"undefinedEdges"`
		Cycles         [][]string                  `json:"cycles"`
		Isolated       []string                    `json:"isolated"`
	} `json:"report"`
}

func TestCLIGraphStreamsJSON(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("sample", "index.ts"), tmp)
	copyFixtureFile(t, filepath.Join("sample", "users.ts"), tmp)

	bin := buildCLI(t)
	_, out := runCmdExpectExit(t, bin, tmp, 0, "graph")

	payload := decodeGraph(t, out)
	if payload.Graph.Version != 1 {
		t.Fatalf("expected graph version 1, got %d", payload.Graph.Version)
	}
	for _, id := range []string{"cache-sample", "db-sample", "cleanup-sample"} {
		if _, ok := payload.Graph.Nodes[id]; !ok {
			t.Fatalf("expected node %s in graph output", id)
		}
	}
	if !hasEdge(payload.Graph.Edges, "db-sample", "cache-sample") || !hasEdge(payload.Graph.Edges, "cache-sample", "cleanup-sample") {
		t.Fatalf("missing expected edges: %+v", payload.Graph.Edges)
	}
}

func TestCLIGraphSupportsDirFlag(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("sample", "index.ts"), tmp)
	copyFixtureFile(t, filepath.Join("sample", "users.ts"), tmp)

	bin := buildCLI(t)
	otherDir := t.TempDir()

	_, out := runCmdExpectExit(t, bin, otherDir, 0, "graph", "--dir", tmp)
	payload := decodeGraph(t, out)
	if len(payload.Graph.Nodes) == 0 {
		t.Fatalf("expected nodes when using --dir")
	}
}

func TestCLIGraphSupportsCommentStyles(t *testing.T) {
	tmp := t.TempDir()
	files := []string{
		filepath.Join("comment-styles", "py_doc.py"),
		filepath.Join("comment-styles", "sql.sql"),
		filepath.Join("comment-styles", "html.html"),
		filepath.Join("comment-styles", "block.js"),
		filepath.Join("comment-styles", "hash.sh"),
		filepath.Join("comment-styles", "inline.go"),
	}
	for _, f := range files {
		copyFixtureFile(t, f, tmp)
	}

	bin := buildCLI(t)
	_, out := runCmdExpectExit(t, bin, tmp, 0, "graph")

	payload := decodeGraph(t, out)
	wantNodes := []string{"hash-root", "block-root", "html-root", "sql-root", "py-root"}
	for _, n := range wantNodes {
		if _, ok := payload.Graph.Nodes[n]; !ok {
			t.Fatalf("missing node %s in graph output", n)
		}
	}
	if _, ok := payload.Graph.Nodes["inline-ignored"]; ok {
		t.Fatalf("inline trailing comment was unexpectedly parsed")
	}
	wantEdges := []struct{ from, to string }{
		{"hash-root", "block-root"},
		{"block-root", "html-root"},
		{"html-root", "sql-root"},
		{"sql-root", "py-root"},
	}
	for _, e := range wantEdges {
		if !hasEdge(payload.Graph.Edges, e.from, e.to) {
			t.Fatalf("missing edge %s->%s in output", e.from, e.to)
		}
	}
}

func TestCLIGraphAllowErrorsFlag(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("undefined", "index.ts"), tmp)

	bin := buildCLI(t)
	code, out := runCmdExpectExit(t, bin, tmp, 0, "graph", "--allow-errors")
	if code != 0 {
		t.Fatalf("expected exit 0 with --allow-errors, got %d", code)
	}
	payload := decodeGraph(t, out)
	if len(payload.Report.UndefinedEdges) == 0 {
		t.Fatalf("expected report to include undefined edges")
	}
}

func TestCLICheckFailsOnUndefinedReference(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("undefined", "index.ts"), tmp)

	bin := buildCLI(t)
	code, out := runCmdExpectExit(t, bin, tmp, 1, "check")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d\nout:\n%s", code, out)
	}
	if !strings.Contains(out, "missing \"missing-id\"") {
		t.Fatalf("expected missing id in output, got:\n%s", out)
	}
}

func TestCLICheckDetectsCycle(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("cycle", "a.ts"), tmp)
	copyFixtureFile(t, filepath.Join("cycle", "b.ts"), tmp)

	bin := buildCLI(t)
	code, out := runCmdExpectExit(t, bin, tmp, 2, "check")
	if code != 2 {
		t.Fatalf("expected exit 2, got %d\nout:\n%s", code, out)
	}
	if !strings.Contains(out, "cycle") {
		t.Fatalf("expected cycle output, got:\n%s", out)
	}
}

func TestCLICheckDetectsIsolated(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("isolated", "index.ts"), tmp)

	bin := buildCLI(t)
	code, out := runCmdExpectExit(t, bin, tmp, 3, "check")
	if code != 3 {
		t.Fatalf("expected exit 3, got %d\nout:\n%s", code, out)
	}
	if !strings.Contains(out, "isolated nodes") {
		t.Fatalf("expected isolated nodes output, got:\n%s", out)
	}
}

func TestCLIVersionCommand(t *testing.T) {
	bin := buildCLI(t)
	dir := t.TempDir()

	_, out := runCmdExpectExit(t, bin, dir, 0, "--version")
	versionOutput := strings.TrimSpace(out)
	if versionOutput == "" {
		t.Fatalf("expected version output, got empty string")
	}

	_, out = runCmdExpectExit(t, bin, dir, 0, "version")
	if strings.TrimSpace(out) != versionOutput {
		t.Fatalf("expected version output %q, got %q", versionOutput, strings.TrimSpace(out))
	}
}

func decodeGraph(t *testing.T, out string) graphPayload {
	t.Helper()
	var payload graphPayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode graph output: %v\nout:\n%s", err, out)
	}
	return payload
}

func hasEdge(edges []struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}, from, to string) bool {
	for _, e := range edges {
		if e.From == from && e.To == to {
			return true
		}
	}
	return false
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return string(data)
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	return abs
}

func buildCLI(t *testing.T) string {
	t.Helper()
	root := findModuleRoot(t)
	bin := filepath.Join(t.TempDir(), "comment-graph")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/comment-graph")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOCACHE="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\nout:\n%s", err, string(out))
	}
	return bin
}

func runCmd(t *testing.T, bin, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOCACHE="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\nout:\n%s", err, string(out))
	}
}

func runCmdExpectExit(t *testing.T, bin, dir string, expect int, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOCACHE="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if err == nil {
		if expect == 0 {
			return 0, string(out)
		}
		t.Fatalf("expected exit %d, got 0\nout:\n%s", expect, string(out))
	}
	var exitCode int
	if ee, ok := err.(*exec.ExitError); ok {
		exitCode = ee.ExitCode()
	} else {
		t.Fatalf("command failed: %v\nout:\n%s", err, string(out))
	}
	if expect != exitCode {
		t.Fatalf("expected exit %d, got %d\nout:\n%s", expect, exitCode, string(out))
	}
	return exitCode, string(out)
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir := absPath(t, ".")
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

func copyFixtureFile(t *testing.T, name, dst string) {
	t.Helper()
	src := filepath.Join(findModuleRoot(t), "integration", "testdata", name)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	target := filepath.Join(dst, name)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
