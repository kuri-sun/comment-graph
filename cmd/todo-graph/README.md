# todo-graph CLI

## Usage

- `todo-graph generate` — scan for TODOs and write `.todo-graph` (runs validation first).
- `todo-graph check` — validate TODO references, detect cycles/isolated nodes, and ensure `.todo-graph` matches source.
- `todo-graph visualize` — read `.todo-graph` and print an indented tree of the TODO graph (invokes generate first).

### Flags and behavior

- `--output <path>` — write `.todo-graph` to a custom path (useful for CI artifacts).
- `--help`, `-h` — show usage.
- Colors auto-enable on TTY; set `NO_COLOR=1` to disable.
- Exit codes:
  - `generate`: 0 on success; 1–3 on validation/write errors.
  - `check`: 0 success, 1 undefined refs, 2 cycles, 3 isolated/out-of-date/scan errors.
  - `visualize`: 0 success; propagates `generate` failures.

### TODO syntax

```ts
// TODO: short description
// @todo-id some-id
// @todo-deps dep-a, dep-b
```

Rules:
- TODO must start on a comment line (not inline after code).
- Metadata must immediately follow the TODO; only `@todo-id` (required) and `@todo-deps` are allowed.
- IDs must be lowercase letters/digits/hyphens/underscores. Missing `@todo-id` is an error.
- `@todo-deps` is comma-separated; `#` is optional. Space-separated deps error.
- TODOs and metadata are recognized in line-start comments and block comments that start at the beginning of a line (Go/TS/JS/HTML/triple-quote blocks).
