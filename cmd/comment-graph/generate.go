package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kuri-sun/comment-graph/internal/engine"
)

func runGenerate(p printer, dir, format string, allowErrors bool) int {
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

	code, failed := validateAndReport(p, "Generate completed", graph, report, nil, false)
	exitCode := code
	if failed && allowErrors {
		exitCode = 0
	}

	switch format {
	case "yaml":
		if err := engine.WriteGraph(root, "", graph); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write .comment-graph: %v\n", err)
			p.resultLine(false)
			return 1
		}
	case "json":
		if err := engine.WriteGraphJSON(root, "", graph, nil); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write .comment-graph.json: %v\n", err)
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
	ok := !failed || allowErrors
	p.resultLine(ok)
	if failed && allowErrors {
		p.warnLine("validation failed; output written due to --allow-errors")
	}
	filename := ".comment-graph"
	if format == "json" {
		filename = ".comment-graph.json"
	}
	target := filepath.Join(root, filename)
	abs, _ := filepath.Abs(target)
	p.infof("generated : %s", abs)
	if failed {
		return exitCode
	}
	return 0
}
