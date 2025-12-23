# comment-graph

[![Release](https://img.shields.io/github/v/release/kuri-sun/comment-graph)](https://github.com/kuri-sun/comment-graph/releases)
[![npm](https://img.shields.io/npm/v/comment-graph)](https://www.npmjs.com/package/comment-graph)
[![CI](https://github.com/kuri-sun/comment-graph/actions/workflows/ci.yml/badge.svg)](https://github.com/kuri-sun/comment-graph/actions/workflows/ci.yml)

A small CLI that scans your codebase for comment metadata and builds a dependency graph.

## Installation

- Binary download: Download the appropriate archive from the GitHub releases page, unpack, and put comment-graph on your PATH.
- Go: `go install github.com/kuri-sun/comment-graph/cmd/comment-graph@latest`.
- Node: `npm install --save-dev comment-graph` or `npx comment-graph`.

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

Nvim plugin [comment-graph.nvim](../comment-graph.nvim) lets you browse graphs via a tree/preview UI without writing repo files.
