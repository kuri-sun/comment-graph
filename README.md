# todo-graph

A small CLI that scans your codebase for `TODO` comments, builds a dependency graph, and lets you validate or visualize it.

## Installation

- With Go (requires Go 1.25): from the repo root run `go install ./cmd/todo-graph` (or `go build ./cmd/todo-graph` and use the resulting binary).
- Node users can install the npm wrapper (`npm install` inside `npm/`); it downloads the platform binary and exposes `npx todo-graph`.

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
