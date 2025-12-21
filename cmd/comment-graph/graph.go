package main

import (
	"fmt"
	"os"

	"github.com/kuri-sun/comment-graph/internal/engine"
)

func runGraph(p printer, dir string, allowErrors bool) int {
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
	code, failed := validationStatus(graph, report, nil, false)
	exitCode := code
	if failed && allowErrors {
		exitCode = 0
	}

	payload, err := engine.RenderGraphPayloadJSON(graph, &report, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to render graph json: %v\n", err)
		return 1
	}
	fmt.Println(string(payload))
	if failed && !allowErrors {
		return code
	}
	return exitCode
}
