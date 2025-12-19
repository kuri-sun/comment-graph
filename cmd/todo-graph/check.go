package main

import (
	"fmt"
	"os"

	"github.com/kuri-sun/todo-graph/internal/engine"
)

func runCheck(p printer, dir string, keywords []string) int {
	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	scanned, scanErrs, err := engine.ScanWithKeywords(root, keywords)
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
