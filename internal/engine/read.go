package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kuri-sun/todo-graph/internal/graph"
)

// ReadGraph parses the .comment-graph file from the repository root.
func ReadGraph(root string) (graph.Graph, error) {
	path := filepath.Join(root, ".comment-graph")
	data, err := os.ReadFile(path)
	if err != nil {
		return graph.Graph{}, err
	}

	lines := strings.Split(string(data), "\n")

	var g graph.Graph
	g.Nodes = make(map[string]graph.Node)

	mode := ""
	currentID := ""
	var currentEdge *graph.Edge

	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "#"):
			continue
		case strings.HasPrefix(line, "version:"):
			continue
		case line == "nodes:":
			mode = "nodes"
			currentID = ""
			currentEdge = nil
			continue
		case line == "edges:":
			mode = "edges"
			currentID = ""
			currentEdge = nil
			continue
		}

		switch mode {
		case "nodes":
			if strings.HasSuffix(line, ":") {
				currentID = strings.TrimSuffix(line, ":")
				continue
			}
			if currentID == "" {
				continue
			}
			if strings.HasPrefix(line, "file:") {
				pathVal := strings.TrimSpace(strings.TrimPrefix(line, "file:"))
				if unquoted, err := strconv.Unquote(pathVal); err == nil {
					pathVal = unquoted
				}
				n := g.Nodes[currentID]
				n.ID = currentID
				n.File = pathVal
				g.Nodes[currentID] = n
				continue
			}
			if strings.HasPrefix(line, "line:") {
				lineVal := strings.TrimSpace(strings.TrimPrefix(line, "line:"))
				n, err := strconv.Atoi(lineVal)
				if err != nil {
					return graph.Graph{}, fmt.Errorf(".comment-graph:%d: invalid line number %q", i+1, lineVal)
				}
				node := g.Nodes[currentID]
				node.ID = currentID
				node.Line = n
				g.Nodes[currentID] = node
				continue
			}
		case "edges":
			if line == "[]" {
				continue
			}
			if strings.HasPrefix(line, "- ") {
				if currentEdge != nil {
					if currentEdge.From != "" && currentEdge.To != "" {
						g.Edges = append(g.Edges, *currentEdge)
					}
				}
				currentEdge = &graph.Edge{}
				line = strings.TrimPrefix(line, "- ")
			}
			if currentEdge == nil {
				continue
			}
			if strings.HasPrefix(line, "from:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "from:"))
				if unquoted, err := strconv.Unquote(val); err == nil {
					val = unquoted
				}
				currentEdge.From = val
				continue
			}
			if strings.HasPrefix(line, "to:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "to:"))
				if unquoted, err := strconv.Unquote(val); err == nil {
					val = unquoted
				}
				currentEdge.To = val
				continue
			}
			if strings.HasPrefix(line, "type:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "type:"))
				if unquoted, err := strconv.Unquote(val); err == nil {
					val = unquoted
				}
				currentEdge.Type = val
				continue
			}
		}
	}

	if currentEdge != nil && currentEdge.From != "" && currentEdge.To != "" {
		g.Edges = append(g.Edges, *currentEdge)
	}

	return g, nil
}
