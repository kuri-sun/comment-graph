package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kuri-sun/todo-graph/internal/engine"
)

func runGenerate(p printer, dir, output, errorsOutput, format string, keywords []string) int {
	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	graph, errs, err := engine.ScanWithKeywords(root, keywords)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	report := engine.ValidateGraph(graph, errs)
	code, failed := validateAndReport(p, "Generate completed", graph, report, nil, false)

	switch format {
	case "yaml":
		if err := engine.WriteGraph(root, output, graph); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write .todo-graph: %v\n", err)
			p.resultLine(false)
			return 1
		}
	case "json":
		if err := engine.WriteGraphJSON(root, output, graph, nil); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write .todo-graph.json: %v\n", err)
			p.resultLine(false)
			return 1
		}
	default:
		fmt.Fprintf(os.Stderr, "unsupported format: %s\n", format)
		p.resultLine(false)
		return 1
	}

	fmt.Println()
	p.section("Generate complete")
	p.resultLine(true)
	target := targetPath(root, output, format)
	abs, _ := filepath.Abs(target)
	p.infof("generated : %s", abs)
	if errorsOutput != "" {
		if err := engine.WriteErrorsJSON(root, errorsOutput, report); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write errors json: %v\n", err)
		}
	}
	if failed {
		return code
	}
	return 0
}

func targetPath(root, output, format string) string {
	if output != "" {
		if filepath.IsAbs(output) {
			return output
		}
		return filepath.Join(root, output)
	}
	filename := ".todo-graph"
	if format == "json" {
		filename = ".todo-graph.json"
	}
	return filepath.Join(root, filename)
}
