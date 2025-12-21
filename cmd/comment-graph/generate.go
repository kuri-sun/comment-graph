package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kuri-sun/comment-graph/internal/engine"
)

func runGenerate(p printer, dir, output, format string, allowErrors bool) int {
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
	stdoutOutput := output == "-"

	if stdoutOutput {
		code, failed := validationStatus(graph, report, nil, false)
		exitCode := code
		if failed && allowErrors {
			exitCode = 0
		}
		switch format {
		case "yaml":
			fmt.Println(engine.RenderGraphYAML(graph))
		case "json":
			data, err := engine.RenderGraphPayloadJSON(graph, &report, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to render .comment-graph json: %v\n", err)
				return 1
			}
			fmt.Println(string(data))
		default:
			fmt.Fprintf(os.Stderr, "unsupported format: %s\n", format)
			return 1
		}
		if failed && !allowErrors {
			return code
		}
		return exitCode
	}

	code, failed := validateAndReport(p, "Generate completed", graph, report, nil, false)
	exitCode := code
	if failed && allowErrors {
		exitCode = 0
	}

	switch format {
	case "yaml":
		if err := engine.WriteGraph(root, output, graph); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write .comment-graph: %v\n", err)
			p.resultLine(false)
			return 1
		}
	case "json":
		if err := engine.WriteGraphJSON(root, output, graph, nil); err != nil {
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
	target := targetPath(root, output, format)
	abs, _ := filepath.Abs(target)
	p.infof("generated : %s", abs)
	if failed {
		return exitCode
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
	filename := ".comment-graph"
	if format == "json" {
		filename = ".comment-graph.json"
	}
	return filepath.Join(root, filename)
}
