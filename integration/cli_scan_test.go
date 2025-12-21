package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Integration: run `comment-graph generate` end-to-end against a temp TS repo with cross-file deps.
func TestCLIGenerateWritesTodoGraph(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("sample", "index.ts"), tmp)
	copyFixtureFile(t, filepath.Join("sample", "users.ts"), tmp)

	bin := buildCLI(t)
	runCmd(t, bin, tmp, "generate")

	got := readFile(t, filepath.Join(tmp, ".comment-graph"))
	if !strings.Contains(got, "\n  cache-sample:\n") || !strings.Contains(got, "\n  db-sample:\n") || !strings.Contains(got, "\n  cleanup-sample:\n") {
		t.Fatalf("unexpected todos section:\n%s", got)
	}
	if !strings.Contains(got, "from: \"db-sample\"\n    to: \"cache-sample\"") {
		t.Fatalf("expected edge db-sample->cache-sample, got:\n%s", got)
	}
	if !strings.Contains(got, "from: \"cache-sample\"\n    to: \"cleanup-sample\"") {
		t.Fatalf("expected edge cache-sample->cleanup-sample, got:\n%s", got)
	}
}

// Integration: `check` should fail with exit 1 on undefined references.
func TestCLICheckFailsOnUndefinedReference(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("undefined", "index.ts"), tmp)

	bin := buildCLI(t)
	code, out := runCmdExpectExit(t, bin, tmp, 1, "generate")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d\nout:\n%s", code, out)
	}
	if !strings.Contains(out, "missing \"missing-id\"") {
		t.Fatalf("expected missing id in output, got:\n%s", out)
	}
}

// Integration: `check` should fail with exit 2 on cycles.
func TestCLICheckDetectsCycle(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("cycle", "a.ts"), tmp)
	copyFixtureFile(t, filepath.Join("cycle", "b.ts"), tmp)

	bin := buildCLI(t)
	code, out := runCmdExpectExit(t, bin, tmp, 2, "generate")
	if code != 2 {
		t.Fatalf("expected exit 2, got %d\nout:\n%s", code, out)
	}
	if !strings.Contains(out, "cycle") {
		t.Fatalf("expected cycle output, got:\n%s", out)
	}
}

// Integration: `check` should ignore .comment-graph drift and rely on source scan only.
func TestCLICheckIgnoresDrift(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("sample", "index.ts"), tmp)
	copyFixtureFile(t, filepath.Join("sample", "users.ts"), tmp)

	bin := buildCLI(t)
	runCmd(t, bin, tmp, "generate")

	// mutate .comment-graph to introduce drift
	path := filepath.Join(tmp, ".comment-graph")
	contents := readFile(t, path)
	contents = strings.Replace(contents, "cleanup-sample", "cleanup-sample-changed", 1)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("mutate .comment-graph: %v", err)
	}

	code, out := runCmdExpectExit(t, bin, tmp, 0, "check")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nout:\n%s", code, out)
	}
	if strings.Contains(out, ".comment-graph is out of date") {
		t.Fatalf("did not expect drift output, got:\n%s", out)
	}
}

// Integration: `check` should surface isolated TODOs via exit 3.
func TestCLICheckDetectsIsolated(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("isolated", "index.ts"), tmp)

	bin := buildCLI(t)
	code, out := runCmdExpectExit(t, bin, tmp, 3, "generate")
	if code != 3 {
		t.Fatalf("expected exit 3, got %d\nout:\n%s", code, out)
	}
	if !strings.Contains(out, "isolated nodes") {
		t.Fatalf("expected isolated nodes output, got:\n%s", out)
	}
}

// Integration: version command/flag should print the CLI version and exit 0.
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

// Integration: check should surface scan/undefined errors even without .comment-graph.
func TestCLICheckReportsScanErrorsWithoutGraph(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("undefined", "index.ts"), tmp)

	bin := buildCLI(t)
	code, out := runCmdExpectExit(t, bin, tmp, 1, "check")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d\nout:\n%s", code, out)
	}
	if !strings.Contains(out, "missing \"missing-id\"") {
		t.Fatalf("expected missing id error, got:\n%s", out)
	}
	if strings.Contains(out, "failed to read .comment-graph") {
		t.Fatalf("expected scan error before .comment-graph read, got:\n%s", out)
	}
}

// Integration: --dir should run commands against a different working directory.
func TestCLIDirFlagTargetsRoot(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("sample", "index.ts"), tmp)
	copyFixtureFile(t, filepath.Join("sample", "users.ts"), tmp)

	bin := buildCLI(t)
	otherDir := t.TempDir()

	runCmd(t, bin, otherDir, "generate", "--dir", tmp)

	got := readFile(t, filepath.Join(tmp, ".comment-graph"))
	if !strings.Contains(got, "\n  cache-sample:\n") || !strings.Contains(got, "\n  db-sample:\n") || !strings.Contains(got, "\n  cleanup-sample:\n") {
		t.Fatalf("unexpected todos section:\n%s", got)
	}

	code, out := runCmdExpectExit(t, bin, otherDir, 0, "check", "--dir", tmp)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nout:\n%s", code, out)
	}
}

// Integration: generate should support writing .comment-graph to a custom path.
func TestCLIGenerateSupportsOutputFlag(t *testing.T) {
	tmp := t.TempDir()
	copyFixtureFile(t, filepath.Join("sample", "index.ts"), tmp)
	copyFixtureFile(t, filepath.Join("sample", "users.ts"), tmp)

	bin := buildCLI(t)
	runCmd(t, bin, tmp, "generate")

	got := readFile(t, filepath.Join(tmp, ".comment-graph"))
	if !strings.Contains(got, "\n  cache-sample:\n") || !strings.Contains(got, "\n  db-sample:\n") {
		t.Fatalf("unexpected todos section:\n%s", got)
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
