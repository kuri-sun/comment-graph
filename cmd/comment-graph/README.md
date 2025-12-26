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

- `@cgraph-id` — required unique ID for the node (lowercase letters, digits, hyphens, underscores).
- `@cgraph-label` — optional human-friendly label shown in outputs/preview.
- `@cgraph-deps` — comma-separated list of IDs that block this item.

## Rules:

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
