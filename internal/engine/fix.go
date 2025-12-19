package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type FixReport struct {
	Added   int
	Missing []ScanError
	Others  []ScanError
}

// FixMissingIDs scans using default keywords.
func FixMissingIDs(root string) (FixReport, error) {
	return FixMissingIDsWithKeywords(root, nil)
}

// FixMissingIDsWithKeywords scans for TODOs without @todo-id and inserts a generated id placeholder.
// Placeholders are derived from the relative file path and line number to ensure stability.
func FixMissingIDsWithKeywords(root string, keywords []string) (FixReport, error) {
	graph, errs, err := ScanWithKeywords(root, keywords)
	if err != nil {
		return FixReport{}, err
	}

	existing := make(map[string]struct{}, len(graph.Todos))
	for id := range graph.Todos {
		existing[id] = struct{}{}
	}

	report := FixReport{}
	missingByFile := make(map[string][]int)
	for _, e := range errs {
		if isMissingIDError(e.Msg) {
			missingByFile[e.File] = append(missingByFile[e.File], e.Line)
			report.Missing = append(report.Missing, e)
		} else {
			report.Others = append(report.Others, e)
		}
	}

	if len(missingByFile) == 0 {
		return report, nil
	}

	for file, lines := range missingByFile {
		added, err := addPlaceholders(root, file, lines, existing)
		if err != nil {
			return report, err
		}
		report.Added += added
	}

	return report, nil
}

func isMissingIDError(msg string) bool {
	return strings.Contains(msg, "TODO id is required") || strings.Contains(msg, "TODO id must not be empty")
}

func addPlaceholders(root, rel string, lines []int, existing map[string]struct{}) (int, error) {
	path := filepath.Join(root, rel)
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	rows := strings.Split(string(content), "\n")
	sort.Ints(lines)

	added := 0
	offset := 0
	for _, line := range lines {
		idx := line - 1 + offset
		if idx < 0 || idx >= len(rows) {
			continue
		}
		id := generatePlaceholder(rel, line, existing)
		prefix, suffix := commentDelimiters(rows[idx])
		insert := fmt.Sprintf("%s @todo-id %s%s", prefix, id, suffix)
		rows = insertLine(rows, idx+1, insert)
		offset++
		added++
	}

	return added, os.WriteFile(path, []byte(strings.Join(rows, "\n")), 0o644)
}

var commentPrefix = regexp.MustCompile(`^(\s*)(//|#|--|/\*|<!--|\*|"""|''')`)

func commentDelimiters(line string) (string, string) {
	m := commentPrefix.FindStringSubmatch(line)
	if len(m) < 3 {
		return "//", ""
	}
	prefix := m[1] + m[2]
	switch m[2] {
	case "<!--":
		return prefix, " -->"
	case "/*":
		return prefix, " */"
	default:
		return prefix, ""
	}
}

var nonAllowed = regexp.MustCompile(`[^a-z0-9]+`)

func generatePlaceholder(rel string, line int, existing map[string]struct{}) string {
	base := strings.ToLower(rel)
	base = nonAllowed.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "todo"
	}
	candidateBase := fmt.Sprintf("todo-%s-%d", base, line)
	candidate := candidateBase
	i := 1
	for {
		if _, ok := existing[candidate]; !ok {
			existing[candidate] = struct{}{}
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", candidateBase, i)
		i++
	}
}

func insertLine(rows []string, idx int, line string) []string {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(rows) {
		return append(rows, line)
	}
	rows = append(rows[:idx+1], rows[idx:]...)
	rows[idx] = line
	return rows
}
