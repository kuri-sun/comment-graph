# todo-graph

A small CLI that scans your codebase for `TODO` comments, builds a dependency graph, and lets you validate or visualize it.

## Installation

- Binary download: Download the appropriate archive from the GitHub releases page, unpack, and put todo-graph on your PATH.
- Go: `go install github.com/kuri-sun/todo-graph/cmd/todo-graph@latest`.
- Node: `npm install --save-dev todo-graph` or `npx todo-graph`.

## Usage

Run commands in the repo you want to track TODOs:

- `todo-graph generate` — scan for TODOs and write `.todo-graph`.
- `todo-graph check` — validate TODO references, detect cycles/isolated nodes, and ensure `.todo-graph` matches source.
- `todo-graph visualize` — read `.todo-graph` and print an indented tree of the TODO graph.

### Quick start

```
// TODO[#id] short description
// deps: #a, #b
```

IDs must use lowercase letters/digits/hyphens/underscores. If no `[#id]` is provided, an ID is derived from the description.
Only `deps` metadata is parsed (one or more IDs, comma-separated, each prefixed with `#`).
