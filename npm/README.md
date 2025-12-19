## Installation

```bash
npm install --save-dev todo-graph
# or
yarn add -D todo-graph
# or
pnpm add -D todo-graph
```

This package includes a Node wrapper that downloads the matching `todo-graph` binary on install. No other dependencies are needed.

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

### Rules:

- TODO must start on a comment line (not inline after code).
- Metadata must immediately follow the TODO (no blank/non-comment lines); only `@todo-id` and `@todo-deps` are allowed.
- IDs must use lowercase letters/digits/hyphens/underscores. Missing `@todo-id` is an error.
- `@todo-deps` is comma-separated; `#` is optional.
