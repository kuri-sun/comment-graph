package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kuri-sun/todo-graph/internal/engine"
	"github.com/kuri-sun/todo-graph/internal/graph"
)

func runGenerate(p printer, dir, output string) int {
	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	graph, errs, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	report := engine.ValidateGraph(graph, errs)
	if code, failed := validateAndReport(p, "Generate completed", graph, report, nil, false); failed {
		return code
	}

	if err := engine.WriteGraph(root, output, graph); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write .todo-graph: %v\n", err)
		p.resultLine(false)
		return 1
	}

	fmt.Println()
	p.section("Generate complete")
	p.resultLine(true)
	target := filepath.Join(root, ".todo-graph")
	if output != "" {
		target = output
		if !filepath.IsAbs(target) {
			target = filepath.Join(root, output)
		}
	}
	abs, _ := filepath.Abs(target)
	p.infof("generated : %s", abs)
	return 0
}

func runCheck(p printer, dir string) int {
	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	scanned, scanErrs, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		p.resultLine(false)
		return 3
	}

	printFailureHeader := func() {
		fmt.Fprintln(os.Stderr)
		p.section("Check completed")
		p.resultLine(false)
	}

	printErrorsSection := func() {
		fmt.Fprintln(os.Stderr)
		p.section("Errors")
	}

	report := engine.ValidateGraph(scanned, scanErrs)
	if len(report.ScanErrors) > 0 || len(report.UndefinedEdges) > 0 || len(report.Cycles) > 0 || len(report.Isolated) > 0 {
		if code, failed := validateAndReport(p, "Check completed", scanned, report, nil, false); failed {
			return code
		}
	}

	fileGraph, err := engine.ReadGraph(root)
	if err != nil {
		printFailureHeader()
		printErrorsSection()
		fmt.Fprintf(os.Stderr, "  - failed to read .todo-graph (run todo-graph generate): %v\n", err)
		fmt.Fprintln(os.Stderr)
		return 3
	}

	if code, failed := validateAndReport(p, "Check completed", scanned, report, &fileGraph, true); failed {
		return code
	}

	fmt.Println()
	p.section("Check complete")
	p.resultLine(true)
	roots := findRoots(scanned)
	p.infof("root TODOs : %d", len(roots))
	p.infof("total TODOs: %d", len(scanned.Todos))
	return 0
}

func runDepsSet(p printer, dir, child string, parents []string) int {
	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	scanned, scanErrs, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	report := engine.ValidateGraph(scanned, scanErrs)
	if code, failed := validateAndReport(p, "Deps set validation", scanned, report, nil, false); failed {
		return code
	}

	if err := engine.UpdateDeps(root, scanned, child, parents); err != nil {
		fmt.Fprintf(os.Stderr, "update deps failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	updated, _, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rescan failed after update: %v\n", err)
		p.resultLine(false)
		return 1
	}
	if err := engine.WriteGraph(root, "", updated); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write .todo-graph: %v\n", err)
		p.resultLine(false)
		return 1
	}

	fmt.Println()
	p.section("Deps set complete")
	p.resultLine(true)
	p.infof("updated @todo-deps for %s", child)
	return 0
}

func runDepsDetach(p printer, dir, child, target string, detachAll bool) int {
	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	scanned, scanErrs, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	report := engine.ValidateGraph(scanned, scanErrs)
	if code, failed := validateAndReport(p, "Deps detach validation", scanned, report, nil, false); failed {
		return code
	}

	parents := engine.CurrentParents(scanned, child)
	var remaining []string
	if detachAll {
		remaining = nil
	} else {
		found := false
		for _, pID := range parents {
			if pID == target {
				found = true
				continue
			}
			remaining = append(remaining, pID)
		}
		if !found {
			fmt.Fprintf(os.Stderr, "parent %q not found on TODO %q\n", target, child)
			p.resultLine(false)
			return 1
		}
	}

	if err := engine.UpdateDepsAllowEmpty(root, scanned, child, remaining); err != nil {
		fmt.Fprintf(os.Stderr, "detach failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	updated, _, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rescan failed after detach: %v\n", err)
		p.resultLine(false)
		return 1
	}
	if err := engine.WriteGraph(root, "", updated); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write .todo-graph: %v\n", err)
		p.resultLine(false)
		return 1
	}

	fmt.Println()
	p.section("Deps detach complete")
	p.resultLine(true)
	if detachAll {
		p.infof("removed all parents from %s", child)
	} else {
		p.infof("removed parent %s from %s", target, child)
	}
	return 0
}

func runFix(p printer, dir string) int {
	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	report, err := engine.FixMissingIDs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fix failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	fmt.Println()
	p.section("Fix complete")
	p.resultLine(true)

	if report.Added == 0 {
		p.infoLine("no missing @todo-id placeholders to add")
	} else {
		p.infof("placeholders added: %d", report.Added)
	}
	if len(report.Others) > 0 {
		p.warnLine("other scan errors remain; run `todo-graph check` after fixing ids")
		for _, e := range report.Others {
			fmt.Fprintf(os.Stderr, "  - %s:%d: %s\n", e.File, e.Line, e.Msg)
		}
	}
	return 0
}

func runView(dir string, rootsOnly bool) int {
	p := newPrinter()
	if code := runGenerate(p, dir, ""); code != 0 {
		return code
	}

	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	g, err := engine.ReadGraph(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read .todo-graph (run todo-graph generate): %v\n", err)
		return 1
	}

	fmt.Println()
	p.section("TODO Graph")
	for _, line := range renderTree(g, rootsOnly) {
		fmt.Println("  " + line)
	}
	return 0
}

// validateAndReport renders validation errors consistently. Returns (exitCode, failed).
func validateAndReport(p printer, header string, scanned graph.Graph, report engine.CheckReport, fileGraph *graph.Graph, checkDrift bool) (int, bool) {
	printFailureHeader := func() {
		fmt.Fprintln(os.Stderr)
		p.section(header)
		p.resultLine(false)
	}
	printErrorsSection := func() {
		fmt.Fprintln(os.Stderr)
		p.section("Errors")
	}

	ensureHeader := func(done *bool) {
		if !*done {
			printFailureHeader()
			printErrorsSection()
			*done = true
		}
	}

	headerPrinted := false

	if len(report.ScanErrors) > 0 {
		ensureHeader(&headerPrinted)
		for _, e := range report.ScanErrors {
			fmt.Fprintf(os.Stderr, "  - %s:%d: %s\n", e.File, e.Line, e.Msg)
		}
		fmt.Fprintln(os.Stderr)
		p.warnLine("Fix scan issues and re-run `todo-graph check`.")
		fmt.Fprintln(os.Stderr)
		return 3, true
	}

	if len(report.UndefinedEdges) > 0 {
		ensureHeader(&headerPrinted)
		for _, e := range report.UndefinedEdges {
			fromTodo, fromOK := scanned.Todos[e.From]
			toTodo, toOK := scanned.Todos[e.To]
			switch {
			case !fromOK && toOK:
				fmt.Fprintf(os.Stderr, "  - missing %q (at %s:%d)\n", e.From, toTodo.File, toTodo.Line)
			case fromOK && !toOK:
				fmt.Fprintf(os.Stderr, "  - missing %q (at %s:%d)\n", e.To, fromTodo.File, fromTodo.Line)
			case !fromOK && !toOK:
				fmt.Fprintf(os.Stderr, "  - missing TODOs %q and %q (edge present but ids undefined)\n", e.From, e.To)
			default:
				fmt.Fprintf(os.Stderr, "  - undefined TODO reference: %s -> %s\n", e.From, e.To)
			}
		}
		fmt.Fprintln(os.Stderr)
		return 1, true
	}

	if len(report.Cycles) > 0 {
		ensureHeader(&headerPrinted)
		fmt.Fprintln(os.Stderr, "  - cycles detected:")
		for _, c := range report.Cycles {
			fmt.Fprintf(os.Stderr, "    cycle: %s\n", strings.Join(c, " -> "))
		}
		fmt.Fprintln(os.Stderr)
		return 2, true
	}

	mismatch := false
	if len(report.Isolated) > 0 {
		ensureHeader(&headerPrinted)
		fmt.Fprintf(os.Stderr, "  - isolated TODOs: %s\n", strings.Join(report.Isolated, ", "))
		mismatch = true
	}

	if checkDrift && fileGraph != nil && !engine.GraphsEqual(scanned, *fileGraph) {
		ensureHeader(&headerPrinted)
		fmt.Fprintln(os.Stderr, "  - .todo-graph is out of date (run todo-graph generate)")
		mismatch = true
	}

	if mismatch {
		fmt.Fprintln(os.Stderr)
		return 3, true
	}

	return 0, false
}

func findRoots(g graph.Graph) []string {
	indegree := make(map[string]int, len(g.Todos))
	for id := range g.Todos {
		indegree[id] = 0
	}
	for _, e := range g.Edges {
		indegree[e.To]++
	}
	var roots []string
	for id, d := range indegree {
		if d == 0 {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	return roots
}

func renderTree(g graph.Graph, rootsOnly bool) []string {
	adj := make(map[string][]string)
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	for k := range adj {
		sort.Strings(adj[k])
	}

	roots := findRoots(g)
	if len(roots) == 0 {
		for id := range g.Todos {
			roots = append(roots, id)
		}
		sort.Strings(roots)
	}

	if rootsOnly {
		lines := make([]string, 0, len(roots))
		for _, r := range roots {
			t, ok := g.Todos[r]
			location := ""
			if ok {
				location = fmt.Sprintf(" (%s:%d)", t.File, t.Line)
			}
			lines = append(lines, fmt.Sprintf("- [] %s%s", r, location))
		}
		return lines
	}

	var lines []string
	visited := make(map[string]bool)

	var dfs func(id string, depth int, stack map[string]bool)
	dfs = func(id string, depth int, stack map[string]bool) {
		t, ok := g.Todos[id]
		location := ""
		if ok {
			location = fmt.Sprintf(" (%s:%d)", t.File, t.Line)
		}
		prefix := strings.Repeat("    ", depth)
		if stack[id] {
			lines = append(lines, fmt.Sprintf("%s- [] %s%s [cycle]", prefix, id, location))
			return
		}
		if visited[id] && depth > 0 {
			lines = append(lines, fmt.Sprintf("%s- [] %s%s [seen]", prefix, id, location))
			return
		}
		lines = append(lines, fmt.Sprintf("%s- [] %s%s", prefix, id, location))
		stack[id] = true
		visited[id] = true
		for _, next := range adj[id] {
			dfs(next, depth+1, stack)
		}
		delete(stack, id)
	}

	for _, r := range roots {
		dfs(r, 0, make(map[string]bool))
	}
	return lines
}
