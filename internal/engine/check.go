package engine

import (
	"fmt"
	"sort"

	"github.com/kuri-sun/comment-graph/internal/graph"
)

// CheckReport contains the results of validation.
type CheckReport struct {
	UndefinedEdges []graph.Edge `json:"undefinedEdges"`
	Cycles         [][]string   `json:"cycles"`
	Isolated       []string     `json:"isolated"`
	ScanErrors     []ScanError  `json:"scanErrors"`
	Mismatch       bool         `json:"mismatch"`
}

// ValidateGraph runs dependency checks on a scanned graph.
func ValidateGraph(g graph.Graph, scanErrs []ScanError) CheckReport {
	undefined := findUndefined(g)
	cycles := findCycles(g)
	isolated := findIsolated(g)

	sort.Strings(isolated)

	return CheckReport{
		UndefinedEdges: undefined,
		Cycles:         cycles,
		Isolated:       isolated,
		ScanErrors:     scanErrs,
	}
}

func findUndefined(g graph.Graph) []graph.Edge {
	var out []graph.Edge
	for _, e := range g.Edges {
		if _, ok := g.Todos[e.From]; !ok {
			out = append(out, e)
			continue
		}
		if _, ok := g.Todos[e.To]; !ok {
			out = append(out, e)
			continue
		}
	}
	return out
}

func findIsolated(g graph.Graph) []string {
	degree := make(map[string]int)
	for id := range g.Todos {
		degree[id] = 0
	}
	for _, e := range g.Edges {
		degree[e.From]++
		degree[e.To]++
	}
	var isolated []string
	for id, d := range degree {
		if d == 0 {
			isolated = append(isolated, id)
		}
	}
	return isolated
}

func findCycles(g graph.Graph) [][]string {
	adj := make(map[string][]string)
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}

	var cycles [][]string
	visited := make(map[string]bool)
	var stack []string
	onStack := make(map[string]bool)

	var dfs func(string)
	dfs = func(node string) {
		visited[node] = true
		onStack[node] = true
		stack = append(stack, node)

		for _, next := range adj[node] {
			if !visited[next] {
				dfs(next)
			} else if onStack[next] {
				cycle := extractCycle(stack, next)
				if len(cycle) > 0 {
					cycles = append(cycles, cycle)
				}
			}
		}

		onStack[node] = false
		stack = stack[:len(stack)-1]
	}

	for id := range g.Todos {
		if !visited[id] {
			dfs(id)
		}
	}

	return dedupeCycles(cycles)
}

func extractCycle(stack []string, target string) []string {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == target {
			cycle := append([]string{}, stack[i:]...)
			cycle = append(cycle, target)
			return cycle
		}
	}
	return nil
}

func dedupeCycles(cycles [][]string) [][]string {
	type key struct {
		start string
		path  string
	}
	seen := make(map[key]bool)
	var out [][]string
	for _, cycle := range cycles {
		if len(cycle) == 0 {
			continue
		}
		normalized := normalizeCycle(cycle)
		k := key{start: normalized[0], path: fmt.Sprint(normalized)}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, normalized)
	}
	return out
}

func normalizeCycle(c []string) []string {
	if len(c) == 0 {
		return c
	}
	// find lexicographically smallest starting point to make comparison stable
	minIdx := 0
	for i := 1; i < len(c)-1; i++ { // last element duplicates start
		if c[i] < c[minIdx] {
			minIdx = i
		}
	}
	var out []string
	for i := 0; i < len(c)-1; i++ {
		out = append(out, c[(minIdx+i)%(len(c)-1)])
	}
	out = append(out, out[0])
	return out
}
