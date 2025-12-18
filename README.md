# todo-graph

A small CLI that scans your codebase for `TODO` comments, builds a dependency graph, and makes it easy to track and review them.

## Installation

- Binary download: Download the appropriate archive from the GitHub releases page, unpack, and put todo-graph on your PATH.
- Go: `go install github.com/kuri-sun/todo-graph/cmd/todo-graph@latest`.
- Node: `npm install --save-dev todo-graph` or `npx todo-graph`.

## Quick start

```ts
// user.ts
// TODO: database migration
// @todo-id db-migration

// cache.ts
// TODO: add cache layer for user reads
// @todo-id cache-user
// @todo-deps db-migration
```

Then run:

```
todo-graph generate   # validates + writes .todo-graph
todo-graph visualize  # shows a tree of the graph
```

Pass `--dir <path>` to target a different repo root (helpful in scripts/CI).
Pass `--output <path>` to `generate` to write the graph somewhere else (handy for CI artifacts).
Use `visualize --roots-only` to list only root TODOs.

Rules:
- TODO must start on a comment line (not inline after code).
- Metadata must immediately follow the TODO (no blank/non-comment lines); only `@todo-id` and `@todo-deps` are allowed.
- IDs must use lowercase letters/digits/hyphens/underscores. Missing `@todo-id` is an error.
- `@todo-deps` is comma-separated; `#` is optional.
