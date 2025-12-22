package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kuri-sun/comment-graph/internal/engine"
	"github.com/kuri-sun/comment-graph/internal/graph"
)

// validateAndReport renders validation errors consistently. Returns (exitCode, failed).
func validateAndReport(p printer, header string, scanned graph.Graph, report engine.CheckReport, fileGraph *graph.Graph, checkDrift bool) (int, bool) {
	printFailureHeader := func() {
		fmt.Fprintln(os.Stderr)
		p.section(header)
		p.resultLine(false)
	}
	printErrorsSection := func() {
		fmt.Fprintln(os.Stderr)
		p.section("Errors")
	}

	ensureHeader := func(done *bool) {
		if !*done {
			printFailureHeader()
			printErrorsSection()
			*done = true
		}
	}

	headerPrinted := false

	if len(report.ScanErrors) > 0 {
		ensureHeader(&headerPrinted)
		for _, e := range report.ScanErrors {
			fmt.Fprintf(os.Stderr, "  - %s:%d: %s\n", e.File, e.Line, e.Msg)
		}
		fmt.Fprintln(os.Stderr)
		p.warnLine("Fix scan issues and re-run `comment-graph check`.")
		fmt.Fprintln(os.Stderr)
		return 3, true
	}

	if len(report.UndefinedEdges) > 0 {
		ensureHeader(&headerPrinted)
		for _, e := range report.UndefinedEdges {
			fromNode, fromOK := scanned.Nodes[e.From]
			toNode, toOK := scanned.Nodes[e.To]
			switch {
			case !fromOK && toOK:
				fmt.Fprintf(os.Stderr, "  - missing %q (at %s:%d)\n", e.From, toNode.File, toNode.Line)
			case fromOK && !toOK:
				fmt.Fprintf(os.Stderr, "  - missing %q (at %s:%d)\n", e.To, fromNode.File, fromNode.Line)
			case !fromOK && !toOK:
				fmt.Fprintf(os.Stderr, "  - missing nodes %q and %q (edge present but ids undefined)\n", e.From, e.To)
			default:
				fmt.Fprintf(os.Stderr, "  - undefined node reference: %s -> %s\n", e.From, e.To)
			}
		}
		fmt.Fprintln(os.Stderr)
		return 1, true
	}

	if len(report.Cycles) > 0 {
		ensureHeader(&headerPrinted)
		fmt.Fprintln(os.Stderr, "  - cycles detected:")
		for _, c := range report.Cycles {
			fmt.Fprintf(os.Stderr, "    cycle: %s\n", strings.Join(c, " -> "))
		}
		fmt.Fprintln(os.Stderr)
		return 2, true
	}

	mismatch := false
	if len(report.Isolated) > 0 {
		ensureHeader(&headerPrinted)
		fmt.Fprintf(os.Stderr, "  - isolated nodes: %s\n", strings.Join(report.Isolated, ", "))
		mismatch = true
	}

	if checkDrift && fileGraph != nil && !engine.GraphsEqual(scanned, *fileGraph) {
		ensureHeader(&headerPrinted)
		fmt.Fprintln(os.Stderr, "  - comment-graph.yml is out of date (run comment-graph generate)")
		mismatch = true
	}

	if mismatch {
		fmt.Fprintln(os.Stderr)
		return 3, true
	}

	return 0, false
}

// validationStatus mirrors validateAndReport's exit codes without rendering.
func validationStatus(scanned graph.Graph, report engine.CheckReport, fileGraph *graph.Graph, checkDrift bool) (int, bool) {
	if len(report.ScanErrors) > 0 {
		return 3, true
	}
	if len(report.UndefinedEdges) > 0 {
		return 1, true
	}
	if len(report.Cycles) > 0 {
		return 2, true
	}
	mismatch := len(report.Isolated) > 0
	if checkDrift && fileGraph != nil && !engine.GraphsEqual(scanned, *fileGraph) {
		mismatch = true
	}
	if mismatch {
		return 3, true
	}
	return 0, false
}

func findRoots(g graph.Graph) []string {
	indegree := make(map[string]int, len(g.Nodes))
	for id := range g.Nodes {
		indegree[id] = 0
	}
	for _, e := range g.Edges {
		indegree[e.To]++
	}
	var roots []string
	for id, d := range indegree {
		if d == 0 {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	return roots
}

func renderTree(g graph.Graph, rootsOnly bool) []string {
	adj := make(map[string][]string)
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	for k := range adj {
		sort.Strings(adj[k])
	}

	roots := findRoots(g)
	if len(roots) == 0 {
		for id := range g.Nodes {
			roots = append(roots, id)
		}
		sort.Strings(roots)
	}

	if rootsOnly {
		lines := make([]string, 0, len(roots))
		for _, r := range roots {
			n, ok := g.Nodes[r]
			location := ""
			if ok {
				location = fmt.Sprintf(" (%s:%d)", n.File, n.Line)
			}
			lines = append(lines, fmt.Sprintf("- [] %s%s", r, location))
		}
		return lines
	}

	var lines []string
	visited := make(map[string]bool)

	var dfs func(id string, depth int, stack map[string]bool)
	dfs = func(id string, depth int, stack map[string]bool) {
		n, ok := g.Nodes[id]
		location := ""
		if ok {
			location = fmt.Sprintf(" (%s:%d)", n.File, n.Line)
		}
		prefix := strings.Repeat("    ", depth)
		if stack[id] {
			lines = append(lines, fmt.Sprintf("%s- [] %s%s [cycle]", prefix, id, location))
			return
		}
		if visited[id] && depth > 0 {
			lines = append(lines, fmt.Sprintf("%s- [] %s%s [seen]", prefix, id, location))
			return
		}
		lines = append(lines, fmt.Sprintf("%s- [] %s%s", prefix, id, location))
		stack[id] = true
		visited[id] = true
		for _, next := range adj[id] {
			dfs(next, depth+1, stack)
		}
		delete(stack, id)
	}

	for _, r := range roots {
		dfs(r, 0, make(map[string]bool))
	}
	return lines
}
