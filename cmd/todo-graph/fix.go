package main

import (
	"fmt"
	"os"

	"github.com/kuri-sun/todo-graph/internal/engine"
)

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
