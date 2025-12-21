# todo-graph

A small CLI that scans your codebase for `TODO` comments, builds a dependency graph.

## Installation

- Binary download: Download the appropriate archive from the GitHub releases page, unpack, and put todo-graph on your PATH.
- Go: `go install github.com/kuri-sun/todo-graph/cmd/todo-graph@latest`.
- Node: `npm install --save-dev todo-graph` or `npx todo-graph`.

## Usage

- [CLI](cmd/todo-graph/README.md)
- [Node](npm/README.md)

### Flags and behavior (CLI)

- `--keywords <list>` — comma-separated keywords to scan (default: TODO,FIXME,NOTE,WARNING,HACK,CHANGED,REVIEW).
- `generate --output <path>` — write `.todo-graph` to a custom path.
- `generate --format <yaml|json>` — choose output format (json writes `.todo-graph.json`).
- `--dir <path>` — run commands against a different repository root.
- `--help`, `-h` — show usage.
- Colors auto-enable on TTY; set `NO_COLOR=1` to disable.

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
