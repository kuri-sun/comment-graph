# comment-graph CLI

## Usage

- `comment-graph check` — validate references, detect cycles/isolated nodes, and ensure `comment-graph.yml` matches source.
- `comment-graph graph` — stream JSON (graph + validation report) to stdout without writing repo files.
- `comment-graph generate` — scan for comment graph nodes and write `comment-graph.yml`.

### Flags and behavior

- `generate --format <yaml|json>` — choose output format (json writes `comment-graph.json`).
- `generate --allow-errors` — write output even when validation fails (errors included in JSON output).
- `--dir <path>` — run commands against a different repository root.
- `--help`, `-h` — show usage.

### Comment graph syntax

```ts
// @cgraph-id some-id
// @cgraph-label Optional human label
// @cgraph-deps dep-a, dep-b
```

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

Rules:

- Comment metadata must start on a comment line (not inline after code).
- Metadata must immediately follow the comment line; only `@cgraph-id` (required), `@cgraph-label` (optional), and `@cgraph-deps` are allowed.
- IDs must match the regex `^[a-z0-9_-]+$`.
- `@cgraph-deps` is comma-separated; spaces are allowed after commas.

## Supported comment styles

The scanner understands whole-line or block comments that start with one of:

- `//` — C/C++/C#/Java/Go/JS/TS/Swift
- `#` — Python, Shell, Ruby, YAML
- `--` — SQL, Lua
- `/* ... */` and `{/* ... */}` — C-family block comments (JSX/TSX friendly)
- `<!-- ... -->` — HTML/Markdown
- `""" ... """` / `''' ... '''` — Python-style docstrings

Inline trailing comments (`code(); // @cgraph-id ...`) are not picked up; place metadata on comment lines.
