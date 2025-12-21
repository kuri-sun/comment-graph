package graph

// Node represents a comment-graph node discovered in source code.
type Node struct {
	ID   string
	File string
	Line int
}

// Edge models a dependency edge between nodes.
// Only the "blocks" type is supported in the MVP.
type Edge struct {
	From string
	To   string
	Type string
}

// Graph is the in-memory representation of .comment-graph.
type Graph struct {
	Nodes map[string]Node
	Edges []Edge
}
