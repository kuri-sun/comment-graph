# comment-graph CLI

## Usage

- `comment-graph check` — validate TODO references, detect cycles/isolated nodes, and ensure `.comment-graph` matches source.
- `comment-graph graph` — stream JSON (graph + validation report) to stdout without writing repo files.
- `comment-graph generate` — scan for TODOs and write `.comment-graph`.

### Flags and behavior

- `--keywords <list>` — comma-separated keywords to scan (default: `TODO,FIXME,NOTE,WARNING,HACK,CHANGED,REVIEW`).
- `generate --output <path>` — write `.comment-graph` to a custom path.
- `generate --format <yaml|json>` — choose output format (json writes `.comment-graph.json`).
- `generate --allow-errors` — write output even when validation fails (errors included in JSON output).
- `--dir <path>` — run commands against a different repository root.
- `--help`, `-h` — show usage.

### TODO syntax

```ts
// TODO: short description
// @todo-id some-id
// @todo-deps dep-a, dep-b
```

Rules:

- TODO must start on a comment line (not inline after code).
- Metadata must immediately follow the TODO; only `@todo-id` (required) and `@todo-deps` are allowed.
- IDs must match the regex `^[a-z0-9_-]+$`.
- `@todo-deps` is comma-separated;
