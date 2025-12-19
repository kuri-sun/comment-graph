package main

import (
	"fmt"
	"os"

	"github.com/kuri-sun/todo-graph/internal/engine"
)

func runDepsSet(p printer, dir, child string, parents []string) int {
	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	scanned, scanErrs, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	report := engine.ValidateGraph(scanned, scanErrs)
	if code, failed := validateAndReport(p, "Deps set validation", scanned, report, nil, false); failed {
		return code
	}

	if err := engine.UpdateDeps(root, scanned, child, parents); err != nil {
		fmt.Fprintf(os.Stderr, "update deps failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	updated, _, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rescan failed after update: %v\n", err)
		p.resultLine(false)
		return 1
	}
	if err := engine.WriteGraph(root, "", updated); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write .todo-graph: %v\n", err)
		p.resultLine(false)
		return 1
	}

	fmt.Println()
	p.section("Deps set complete")
	p.resultLine(true)
	p.infof("updated @todo-deps for %s", child)
	return 0
}

func runDepsDetach(p printer, dir, child, target string, detachAll bool) int {
	root, err := resolveRoot(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve working directory: %v\n", err)
		return 1
	}

	scanned, scanErrs, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	report := engine.ValidateGraph(scanned, scanErrs)
	if code, failed := validateAndReport(p, "Deps detach validation", scanned, report, nil, false); failed {
		return code
	}

	parents := engine.CurrentParents(scanned, child)
	var remaining []string
	if detachAll {
		remaining = nil
	} else {
		found := false
		for _, pID := range parents {
			if pID == target {
				found = true
				continue
			}
			remaining = append(remaining, pID)
		}
		if !found {
			fmt.Fprintf(os.Stderr, "parent %q not found on TODO %q\n", target, child)
			p.resultLine(false)
			return 1
		}
	}

	if err := engine.UpdateDepsAllowEmpty(root, scanned, child, remaining); err != nil {
		fmt.Fprintf(os.Stderr, "detach failed: %v\n", err)
		p.resultLine(false)
		return 1
	}

	updated, _, err := engine.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rescan failed after detach: %v\n", err)
		p.resultLine(false)
		return 1
	}
	if err := engine.WriteGraph(root, "", updated); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write .todo-graph: %v\n", err)
		p.resultLine(false)
		return 1
	}

	fmt.Println()
	p.section("Deps detach complete")
	p.resultLine(true)
	if detachAll {
		p.infof("removed all parents from %s", child)
	} else {
		p.infof("removed parent %s from %s", target, child)
	}
	return 0
}
