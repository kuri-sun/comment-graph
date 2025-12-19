package main

import (
	"fmt"
	"os"

	"github.com/kuri-sun/todo-graph/internal/engine"
)

func runView(dir string, rootsOnly bool) int {
	p := newPrinter()
	if code := runGenerate(p, dir, ""); code != 0 {
		return code
	}

	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	g, err := engine.ReadGraph(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read .todo-graph (run todo-graph generate): %v\n", err)
		return 1
	}

	fmt.Println()
	p.section("TODO Graph")
	for _, line := range renderTree(g, rootsOnly) {
		fmt.Println("  " + line)
	}
	return 0
}
