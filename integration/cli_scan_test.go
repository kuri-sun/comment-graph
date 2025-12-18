package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Integration: run `todo-graph scan` end-to-end against a temp TS repo.
func TestCLIScanWritesTodoGraph(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, tmp, filepath.Join("src", "index.ts"), `// TODO:[#a] first
// depends-on: #b
// TODO:[#b] second
`)

	bin := buildCLI(t)
	runCmd(t, bin, tmp, "scan")

	got := readFile(t, filepath.Join(tmp, ".todo-graph"))
	if !strings.Contains(got, "\n  a:\n") || !strings.Contains(got, "\n  b:\n") {
		t.Fatalf("unexpected todos section:\n%s", got)
	}
	if !strings.Contains(got, "from: \"b\"\n    to: \"a\"") {
		t.Fatalf("expected edge b->a, got:\n%s", got)
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
	bin := filepath.Join(t.TempDir(), "todo-graph")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/todo-graph")
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
