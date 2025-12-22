# comment-graph CLI

## Usage

- `comment-graph check` — validate references, detect cycles/isolated nodes, and ensure `.comment-graph` matches source.
- `comment-graph graph` — stream JSON (graph + validation report) to stdout without writing repo files.
- `comment-graph generate` — scan for comment graph nodes and write `.comment-graph`.

### Flags and behavior

- `generate --format <yaml|json>` — choose output format (json writes `.comment-graph.json`).
- `generate --allow-errors` — write output even when validation fails (errors included in JSON output).
- `--dir <path>` — run commands against a different repository root.
- `--help`, `-h` — show usage.

### Comment graph syntax

```ts
// @cgraph-id some-id
// @cgraph-deps dep-a, dep-b
```

Rules:

- Comment metadata must start on a comment line (not inline after code).
- Metadata must immediately follow the comment line; only `@cgraph-id` (required) and `@cgraph-deps` are allowed.
- IDs must match the regex `^[a-z0-9_-]+$`.
- `@cgraph-deps` is comma-separated; spaces are allowed after commas.
