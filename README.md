# todo-graph

A small CLI that scans your codebase for `TODO` comments, builds a dependency graph, and lets you validate or visualize it.

## Installation

- Binary download: Download the appropriate archive from the GitHub releases page, unpack, and put `todo-graph` on your PATH.
- Go users: `go install github.com/kuri-sun/todo-graph/cmd/todo-graph@v0.0.1` (or `@latest` once newer tags exist).
- Node users: `npm install -g todo-graph` or `npx todo-graph`.

## Usage

Run commands in the repository you want to scan:

- `todo-graph generate` — scan for TODOs and write `.todo-graph`.
- `todo-graph check` — validate TODO references, detect cycles/isolated nodes, and ensure `.todo-graph` matches source.
- `todo-graph visualize` — read `.todo-graph` and print an indented tree of the TODO graph.

### TODO syntax (quick start)

```
// TODO[#id] short description
// depends-on: #other-id
```

IDs must use lowercase letters/digits/hyphens/underscores. If no `[#id]` is provided, an ID is derived from the description. Only `depends-on` metadata is parsed (one or more IDs, comma or space separated, each prefixed with `#`).
