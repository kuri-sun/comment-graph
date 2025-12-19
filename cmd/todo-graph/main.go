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
		dir, output, err := parseGenerateFlags(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runGenerate(p, dir, output))
	case "check":
		dir, err := parseDirFlag(os.Args[2:], "check")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runCheck(p, dir))
	case "deps":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "deps requires a subcommand (e.g. set)")
			os.Exit(1)
		}
		sub := os.Args[2]
		switch sub {
		case "set":
			dir, child, parents, err := parseDepsSetFlags(os.Args[3:])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(runDepsSet(p, dir, child, parents))
		case "detach":
			dir, child, target, detachAll, err := parseDepsDetachFlags(os.Args[3:])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(runDepsDetach(p, dir, child, target, detachAll))
		default:
			fmt.Fprintf(os.Stderr, "unknown deps subcommand: %s\n", sub)
			os.Exit(1)
		}
	case "fix":
		dir, err := parseDirFlag(os.Args[2:], "fix")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runFix(p, dir))
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

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

func resolveRoot(dir string) (string, error) {
	if dir == "" {
		return currentRoot()
	}
	return filepath.Abs(dir)
}

func currentRoot() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(root)
}

func parseGenerateFlags(args []string) (string, string, error) {
	dir := ""
	output := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dir":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("missing value for --dir")
			}
			if dir != "" {
				return "", "", fmt.Errorf("duplicate --dir flag")
			}
			dir = args[i+1]
			i++
		case "--output":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("missing value for --output")
			}
			if output != "" {
				return "", "", fmt.Errorf("duplicate --output flag")
			}
			output = args[i+1]
			i++
		default:
			return "", "", fmt.Errorf("unknown flag for generate: %s", args[i])
		}
	}
	return dir, output, nil
}

func parseDirFlag(args []string, cmd string) (string, error) {
	dir := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dir":
			if i+1 >= len(args) {
				return "", fmt.Errorf("missing value for --dir")
			}
			dir = args[i+1]
			i++
		default:
			return "", fmt.Errorf("unknown flag for %s: %s", cmd, arg)
		}
	}
	return dir, nil
}

func parseDepsSetFlags(args []string) (string, string, []string, error) {
	dir := ""
	id := ""
	var parents []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dir":
			if i+1 >= len(args) {
				return "", "", nil, fmt.Errorf("missing value for --dir")
			}
			dir = args[i+1]
			i++
		case "--id":
			if i+1 >= len(args) {
				return "", "", nil, fmt.Errorf("missing value for --id")
			}
			id = args[i+1]
			i++
		case "--depends-on":
			if i+1 >= len(args) {
				return "", "", nil, fmt.Errorf("missing value for --depends-on")
			}
			list := strings.Split(args[i+1], ",")
			for _, p := range list {
				p = strings.TrimSpace(p)
				if p != "" {
					parents = append(parents, p)
				}
			}
			i++
		default:
			return "", "", nil, fmt.Errorf("unknown flag for deps set: %s", arg)
		}
	}
	if id == "" {
		return "", "", nil, fmt.Errorf("--id is required")
	}
	if len(parents) == 0 {
		return "", "", nil, fmt.Errorf("--depends-on requires at least one parent id")
	}
	return dir, id, parents, nil
}

func parseDepsDetachFlags(args []string) (string, string, string, bool, error) {
	dir := ""
	id := ""
	target := ""
	detachAll := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dir":
			if i+1 >= len(args) {
				return "", "", "", false, fmt.Errorf("missing value for --dir")
			}
			dir = args[i+1]
			i++
		case "--id":
			if i+1 >= len(args) {
				return "", "", "", false, fmt.Errorf("missing value for --id")
			}
			id = args[i+1]
			i++
		case "--target":
			if i+1 >= len(args) {
				return "", "", "", false, fmt.Errorf("missing value for --target")
			}
			target = args[i+1]
			i++
		case "--all":
			detachAll = true
		default:
			return "", "", "", false, fmt.Errorf("unknown flag for deps detach: %s", arg)
		}
	}
	if id == "" {
		return "", "", "", false, fmt.Errorf("--id is required")
	}
	if target == "" && !detachAll {
		return "", "", "", false, fmt.Errorf("--target is required unless --all")
	}
	return dir, id, target, detachAll, nil
}

func printHelp() {
	fmt.Printf("todo-graph CLI (version %s)\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  todo-graph generate     Scan repository and write .todo-graph")
	fmt.Println("      --dir <path>        Run against a different root (defaults to cwd; useful in scripts)")
	fmt.Println("      --output <path>     Write .todo-graph to a different path (for CI artifacts)")
	fmt.Println("  todo-graph check        Validate TODO graph consistency")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("  todo-graph deps set     Update @todo-deps for a TODO id")
	fmt.Println("      --id <id>           Target TODO id to update")
	fmt.Println("      --depends-on <ids>  Comma-separated parent TODO ids")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("  todo-graph deps detach  Remove a parent from a TODO's @todo-deps")
	fmt.Println("      --id <id>           Target TODO id to update")
	fmt.Println("      --target <id>       Parent TODO id to remove")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("  todo-graph fix          Auto-add @todo-id placeholders for missing TODO ids")
	fmt.Println("      --dir <path>        Target a different root")
}
