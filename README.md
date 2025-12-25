# comment-graph

A lightweight CLI that reads your code comments and turns them into a dependency graph, making your codebase traceable at a glance.

[![Release](https://img.shields.io/github/v/release/kuri-sun/comment-graph)](https://github.com/kuri-sun/comment-graph/releases)
[![npm](https://img.shields.io/npm/v/@comment-graph/comment-graph)](https://www.npmjs.com/package/@comment-graph/comment-graph)
[![CI](https://github.com/kuri-sun/comment-graph/actions/workflows/ci.yml/badge.svg)](https://github.com/kuri-sun/comment-graph/actions/workflows/ci.yml)

## Installation

- Go: `go install github.com/kuri-sun/comment-graph/cmd/comment-graph@latest`.
- Node: `npm install --save-dev @comment-graph/comment-graph` or `npx @comment-graph/comment-graph`.

## Usage

- [CLI](cmd/comment-graph/README.md)
- [Node](npm/README.md)

## Quick start

```ts
// user.ts
// @cgraph-id db-migration
// @cgraph-label Database migration

// cache.ts
// @cgraph-id cache-user
// @cgraph-deps db-migration
```

Then run:

```
 comment-graph generate   # writes comment-graph.yml
```

Will generate a yaml file:

```yaml
version: 1

nodes:
  db-migration:
    file: backend/db/migrate.go
    line: 12
    label: "Database migration"

  cache-user:
    file: backend/cache/user.go
    line: 34

edges:
  - from: "db-migration"
    to: "cache-user"
    type: "blocks"
```

## Integration

Nvim plugin: [comment-graph.nvim](../comment-graph.nvim)
