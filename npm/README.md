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
// TODO: database migration
// @todo-id db-migration

// cache.ts
// TODO: add cache layer for user reads
// @todo-id cache-user
// @todo-deps db-migration
```

Then run:

```
comment-graph generate   # writes .comment-graph file.
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
