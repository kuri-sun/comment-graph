package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"todo-graph/internal/engine"
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
	case "scan":
		os.Exit(runScan(p))
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

func runScan(p printer) int {
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
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr)
		p.sectionErrRed("Errors")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s:%d: %s\n", e.File, e.Line, e.Msg)
		}
		fmt.Fprintln(os.Stderr)
		p.warnLine("Fix the TODO metadata above and re-run `todo-graph scan`.")
		p.resultLine(false)
		return 1
	}

	if err := engine.WriteGraph(root, graph); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write .todo-graph: %v\n", err)
		p.resultLine(false)
		return 1
	}

	fmt.Println()
	p.section("Scan complete")
	p.resultLine(true)
	p.infof("todos : %d", len(graph.Todos))
	p.infof("genereated: %s", filepath.Join(root, ".todo-graph"))
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

	report := engine.ValidateGraph(scanned, scanErrs)
	if len(report.ScanErrors) > 0 {
		fmt.Fprintln(os.Stderr)
		p.sectionErrRed("Errors")
		for _, e := range report.ScanErrors {
			fmt.Fprintf(os.Stderr, "  - %s:%d: %s\n", e.File, e.Line, e.Msg)
		}
		fmt.Fprintln(os.Stderr)
		p.warnLine("Fix scan issues and re-run `todo-graph check`.")
		p.resultLine(false)
		return 3
	}
	if len(report.UndefinedEdges) > 0 {
		p.sectionErr("Undefined TODO references")
		for _, e := range report.UndefinedEdges {
			fmt.Fprintf(os.Stderr, "undefined TODO reference: %s -> %s\n", e.From, e.To)
		}
		p.resultLine(false)
		return 1
	}
	if len(report.Cycles) > 0 {
		p.sectionErr("Cycles detected")
		for _, c := range report.Cycles {
			fmt.Fprintf(os.Stderr, "cycle: %s\n", strings.Join(c, " -> "))
		}
		p.resultLine(false)
		return 2
	}
	fileGraph, err := engine.ReadGraph(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read .todo-graph (run todo-graph scan): %v\n", err)
		p.resultLine(false)
		return 3
	}

	mismatch := false
	if len(report.Isolated) > 0 {
		p.sectionErr("Isolated TODOs")
		fmt.Fprintf(os.Stderr, "  isolated TODOs: %s\n", strings.Join(report.Isolated, ", "))
		mismatch = true
	}
	if !engine.GraphsEqual(scanned, fileGraph) {
		p.sectionErr("Out of date graph")
		fmt.Fprintln(os.Stderr, "  .todo-graph is out of date (run todo-graph scan)")
		mismatch = true
	}
	if mismatch {
		p.resultLine(false)
		return 3
	}

	fmt.Println()
	p.section("Check complete")
	p.resultLine(true)
	return 0
}

func runVisualize(args []string) int {
	format := "mermaid"
	if len(args) >= 2 && args[0] == "--format" {
		format = args[1]
	}
	if format != "mermaid" {
		fmt.Fprintf(os.Stderr, "unsupported format: %s\n", format)
		return 1
	}

	root, err := currentRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	g, err := engine.ReadGraph(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read .todo-graph (run todo-graph scan): %v\n", err)
		return 1
	}

	out := engine.RenderMermaid(g)
	fmt.Println(out)
	return 0
}

func currentRoot() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(root)
}

func printHelp() {
	fmt.Printf("todo-graph CLI (version %s)\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  todo-graph scan       Scan repository and update .todo-graph")
	fmt.Println("  todo-graph check      Validate TODO graph consistency")
	fmt.Println("  todo-graph visualize  Output graph in a given format")
}
