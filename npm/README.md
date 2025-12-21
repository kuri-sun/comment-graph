## Installation

```bash
npm install --save-dev comment-graph
# or
yarn add -D comment-graph
# or
pnpm add -D comment-graph
```

This package includes a Node wrapper that downloads the matching CLI binary on install.

## Usage

After install, the `comment-graph` binary is available via `npx`/`yarn dlx` or from `node_modules/.bin`:

```bash
npx comment-graph generate --dir .
npx comment-graph check --dir .
```

Or add a script:

```json
{
  "scripts": {
    "comment-graph:generate": "comment-graph generate --dir .",
    "comment-graph:check": "comment-graph check --dir ."
  }
}
```

Run it with `npm run comment-graph:check` (or `yarn`/`pnpm`).

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

### Rules:

- Metadata must be adjacent in the same comment block; only `@cgraph-id` and `@cgraph-deps` are allowed.
- IDs must use lowercase letters/digits/hyphens/underscores. Missing `@cgraph-id` is an error.
- `@cgraph-deps` is comma-separated.
