package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixMissingIDsAddsPlaceholder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	write := []string{
		"// TODO: add caching",
		"// @todo-deps db-migration",
		"func main() {}",
	}
	if err := os.WriteFile(path, []byte(strings.Join(write, "\n")), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	report, err := FixMissingIDs(dir)
	if err != nil {
		t.Fatalf("fix: %v", err)
	}
	if report.Added != 1 {
		t.Fatalf("expected 1 placeholder added, got %d", report.Added)
	}
	if len(report.Missing) != 1 {
		t.Fatalf("expected 1 missing error captured, got %d", len(report.Missing))
	}

	after := mustReadFile(t, path)
	if !strings.Contains(after, "@todo-id todo-test-go-1") {
		t.Fatalf("expected generated id in file, got:\n%s", after)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return string(data)
}
