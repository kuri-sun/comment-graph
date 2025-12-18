package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"todo-graph/internal/engine"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "scan":
		os.Exit(runScan())
	case "check":
		os.Exit(runCheck())
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

func runScan() int {
	root, err := currentRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	graph, errs, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		return 1
	}
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s:%d: %s\n", e.File, e.Line, e.Msg)
		}
		return 1
	}

	if err := engine.WriteGraph(root, graph); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write .todo-graph: %v\n", err)
		return 1
	}

	fmt.Printf("Updated .todo-graph with %d TODO(s)\n", len(graph.Todos))
	return 0
}

func runCheck() int {
	root, err := currentRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	scanned, scanErrs, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		return 3
	}

	report := engine.ValidateGraph(scanned, scanErrs)
	if len(report.ScanErrors) > 0 {
		for _, e := range report.ScanErrors {
			fmt.Fprintf(os.Stderr, "%s:%d: %s\n", e.File, e.Line, e.Msg)
		}
		return 3
	}
	if len(report.UndefinedEdges) > 0 {
		for _, e := range report.UndefinedEdges {
			fmt.Fprintf(os.Stderr, "undefined TODO reference: %s -> %s\n", e.From, e.To)
		}
		return 1
	}
	if len(report.Cycles) > 0 {
		for _, c := range report.Cycles {
			fmt.Fprintf(os.Stderr, "cycle: %s\n", strings.Join(c, " -> "))
		}
		return 2
	}
	fileGraph, err := engine.ReadGraph(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read .todo-graph (run todo-graph scan): %v\n", err)
		return 3
	}

	mismatch := false
	if len(report.Isolated) > 0 {
		fmt.Fprintf(os.Stderr, "isolated TODOs: %s\n", strings.Join(report.Isolated, ", "))
		mismatch = true
	}
	if !engine.GraphsEqual(scanned, fileGraph) {
		fmt.Fprintln(os.Stderr, ".todo-graph is out of date (run todo-graph scan)")
		mismatch = true
	}
	if mismatch {
		return 3
	}

	fmt.Println("Graph is consistent")
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
	fmt.Println("todo-graph CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  todo-graph scan       Scan repository and update .todo-graph")
	fmt.Println("  todo-graph check      Validate TODO graph consistency")
	fmt.Println("  todo-graph visualize  Output graph in a given format")
}
