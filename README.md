# todo-graph

A small CLI that scans your codebase for `TODO` comments, builds a dependency graph, and makes it easy to track and review them.

## Installation

- Binary download: Download the appropriate archive from the GitHub releases page, unpack, and put todo-graph on your PATH.
- Go: `go install github.com/kuri-sun/todo-graph/cmd/todo-graph@latest`.
- Node: `npm install --save-dev todo-graph` or `npx todo-graph`.

## Quick start

```ts
// user.ts

// TODO:[#db-migration] database migration
function getUser() {
```

```ts
// cache.ts

// TODO: cache-user add cache layer for user reads
// DEPS: #db-migration
function cacheUser() {
```

Then run:

```
todo-graph visualize  # shows a tree of the graph
```

IDs must use lowercase letters/digits/hyphens/underscores. If no `[#id]` is provided, an ID like `todo-<line>` is generated.
Only `DEPS` metadata is parsed (one or more IDs, comma-separated, each prefixed with `#`).
