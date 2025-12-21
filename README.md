# comment-graph

A small CLI that scans your codebase for comment metadata and builds a dependency graph.

## Installation

- Binary download: Download the appropriate archive from the GitHub releases page, unpack, and put comment-graph on your PATH.
- Go: `go install github.com/kuri-sun/todo-graph/cmd/todo-graph@latest` (binary name prints as comment-graph).
- Node: `npm install --save-dev todo-graph` or `npx todo-graph` (wrapper still named todo-graph for now).

## Usage

- [CLI](cmd/todo-graph/README.md)
- [Node](npm/README.md)

## Quick start

```ts
// user.ts
// @cgraph-id: db-migration

// cache.ts
// @cgraph-id: cache-user
// @cgraph-deps: db-migration
```

Then run:

```
comment-graph generate   # writes .comment-graph file.
```

Will generate a yaml file:

```yaml
nodes:
  db-migration:
    file: backend/db/migrate.go
    line: 12

  cache-user:
    file: backend/cache/user.go
    line: 34

edges:
  - from: "db-migration"
    to: "cache-user"
    type: "blocks"
```
