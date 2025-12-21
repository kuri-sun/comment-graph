package graph

// Todo represents a TODO comment discovered in source code.
type Todo struct {
	ID   string
	File string
	Line int
}

// Edge models a dependency edge between TODOs.
// Only the "blocks" type is supported in the MVP.
type Edge struct {
	From string
	To   string
	Type string
}

// Graph is the in-memory representation of .comment-graph.
type Graph struct {
	Todos map[string]Todo
	Edges []Edge
}
