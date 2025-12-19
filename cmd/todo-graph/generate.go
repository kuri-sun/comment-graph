package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kuri-sun/todo-graph/internal/engine"
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
