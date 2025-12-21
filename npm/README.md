## Installation

```bash
npm install --save-dev todo-graph
# or
yarn add -D todo-graph
# or
pnpm add -D todo-graph
```

This package includes a Node wrapper that downloads the matching CLI binary on install. The binary prints as `comment-graph` but the wrapper command remains `todo-graph` for now.

## Usage

After install, the `todo-graph` binary is available via `npx`/`yarn dlx` or from `node_modules/.bin`:

```bash
npx todo-graph generate --dir .
npx todo-graph check --dir .
```

Or add a script:

```json
{
  "scripts": {
    "todo-graph:generate": "todo-graph generate --dir .",
    "todo-graph:check": "todo-graph check --dir ."
  }
}
```

Run it with `npm run todo-graph:check` (or `yarn`/`pnpm`).

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
todo-graph generate   # writes .comment-graph file.
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

### Rules:

- Metadata must be adjacent in the same comment block; only `@cgraph-id` and `@cgraph-deps` are allowed.
- IDs must use lowercase letters/digits/hyphens/underscores. Missing `@cgraph-id` is an error.
- `@cgraph-deps` is comma-separated.
