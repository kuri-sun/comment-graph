# todo-graph

A small CLI that scans your codebase for `TODO` comments, builds a dependency graph.

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
todo-graph generate   # writes .todo-graph file.
todo-graph deps set --id <todo> --depends-on <ids>   # update @todo-deps in source
todo-graph fix        # auto-add @todo-id placeholders for missing TODOs
todo-graph view       # prints an indented tree of TODOs
```

Will generate a yaml file:

```yaml
todos:
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
  - from: "cache-user"
    to: "cleanup-sessions"
    type: "blocks"
```

### Rules:

- TODO must start on a comment line (not inline after code).
- Metadata must immediately follow the TODO (no blank/non-comment lines); only `@todo-id` and `@todo-deps` are allowed.
- IDs must use lowercase letters/digits/hyphens/underscores. Missing `@todo-id` is an error.
- `@todo-deps` is comma-separated; `#` is optional.
