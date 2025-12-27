# comment-graph CLI

## Usage

- `comment-graph check` — validate references, detect cycles/isolated nodes.
- `comment-graph graph` — stream JSON (graph + validation report) to stdout without writing repo files (redirect to save).

### Flags and behavior

- `--dir <path>` — run commands against a different repository root.
- `--help`, `-h` — show usage.

### Syntax

```ts
// user.ts
// @cgraph-id db-migration

// cache.ts
// @cgraph-id cache-user
// @cgraph-label Cache user reads
// @cgraph-deps db-migration
```

- `@cgraph-id` — required unique ID for the node (lowercase letters, digits, hyphens, underscores).
- `@cgraph-label` — optional human-friendly label shown in outputs/preview.
- `@cgraph-deps` — comma-separated list of IDs that block this item.

## Rules:

- Comment metadata must start on a comment line (not inline after code).
- Metadata must immediately follow the comment line; only `@cgraph-id` (required), `@cgraph-label` (optional), and `@cgraph-deps` are allowed.
- IDs must match the regex `^[a-z0-9_-]+$`.
- `@cgraph-deps` is comma-separated; spaces are allowed after commas.

## Supported comment styles

- `//` — C/C++/C#/Java/Go/JS/TS/Swift
- `#` — Python, Shell, Ruby, YAML
- `--` — SQL, Lua
- `/* ... */` and `{/* ... */}` — C-family block comments (JSX/TSX friendly)
- `<!-- ... -->` — HTML/Markdown
- `""" ... """` / `''' ... '''` — Python-style docstrings

Inline trailing comments (`code(); // @cgraph-id ...`) are not picked up; place metadata on comment lines.
