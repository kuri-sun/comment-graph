## Installation

```bash
npm install --save-dev comment-graph
# or
yarn add -D comment-graph
# or
pnpm add -D comment-graph
```

This package includes a Node wrapper that downloads the matching `comment-graph` binary on install. No other dependencies are needed.

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
// @cgraph-id db-migration

// cache.ts
// @cgraph-id cache-user
// @cgraph-label Cache user reads
// @cgraph-deps db-migration
```

Then run:

```
comment-graph generate   # writes .comment-graph file.
```

Will generate a yaml file:

```yaml
version: 1

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

- Comment metadata must start on a comment line (not inline after code).
- Metadata must immediately follow the comment (no blank/non-comment lines); only `@cgraph-id` (required), `@cgraph-label` (optional), and `@cgraph-deps` are allowed.
- IDs must use lowercase letters/digits/hyphens/underscores. Missing `@cgraph-id` is an error.
- `@cgraph-deps` is comma-separated; spaces after commas are allowed.
