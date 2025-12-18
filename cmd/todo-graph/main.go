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

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	p := newPrinter()

	cmd := os.Args[1]
	switch cmd {
	case "generate":
		output, err := parseGenerateFlags(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runGenerate(p, output))
	case "check":
		os.Exit(runCheck(p))
	case "visualize":
		os.Exit(runVisualize(os.Args[2:]))
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func runGenerate(p printer, output string) int {
	root, err := currentRoot()
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
	if code, failed := validateAndReport(p, graph, report, nil, false); failed {
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

func runCheck(p printer) int {
	root, err := currentRoot()
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

	fileGraph, err := engine.ReadGraph(root)
	if err != nil {
		printFailureHeader()
		printErrorsSection()
		fmt.Fprintf(os.Stderr, "  - failed to read .todo-graph (run todo-graph generate): %v\n", err)
		fmt.Fprintln(os.Stderr)
		return 3
	}

	report := engine.ValidateGraph(scanned, scanErrs)
	if code, failed := validateAndReport(p, scanned, report, &fileGraph, true); failed {
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

func runVisualize(args []string) int {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "visualize no longer accepts format flags; mermaid output was removed")
		return 1
	}

	p := newPrinter()
	if code := runGenerate(p, ""); code != 0 {
		return code
	}

	root, err := currentRoot()
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
	for _, line := range renderTree(g) {
		fmt.Println("  " + line)
	}
	return 0
}

// validateAndReport renders validation errors consistently. Returns (exitCode, failed).
func validateAndReport(p printer, scanned graph.Graph, report engine.CheckReport, fileGraph *graph.Graph, checkDrift bool) (int, bool) {
	printFailureHeader := func() {
		fmt.Fprintln(os.Stderr)
		p.section("Check completed")
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

func renderTree(g graph.Graph) []string {
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

func currentRoot() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(root)
}

func parseGenerateFlags(args []string) (string, error) {
	output := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output":
			if i+1 >= len(args) {
				return "", fmt.Errorf("missing value for --output")
			}
			if output != "" {
				return "", fmt.Errorf("duplicate --output flag")
			}
			output = args[i+1]
			i++
		default:
			return "", fmt.Errorf("unknown flag for generate: %s", args[i])
		}
	}
	return output, nil
}

func printHelp() {
	fmt.Printf("todo-graph CLI (version %s)\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  todo-graph generate     Scan repository and write .todo-graph")
	fmt.Println("      --output <path>     Write .todo-graph to a different path (for CI artifacts)")
	fmt.Println("  todo-graph check        Validate TODO graph consistency")
	fmt.Println("  todo-graph visualize    Show the graph as an indented tree")
}
