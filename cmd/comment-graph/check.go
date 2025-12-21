package main

import (
	"fmt"
	"os"

	"github.com/kuri-sun/todo-graph/internal/engine"
)

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

	report := engine.ValidateGraph(scanned, scanErrs)
	if len(report.ScanErrors) > 0 || len(report.UndefinedEdges) > 0 || len(report.Cycles) > 0 || len(report.Isolated) > 0 {
		if code, failed := validateAndReport(p, "Check completed", scanned, report, nil, false); failed {
			return code
		}
	}

	fmt.Println()
	p.section("Check complete")
	p.resultLine(true)
	roots := findRoots(scanned)
	p.infof("root nodes : %d", len(roots))
	p.infof("total nodes: %d", len(scanned.Nodes))
	return 0
}
