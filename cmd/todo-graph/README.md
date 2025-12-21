# todo-graph CLI

## Usage

- `todo-graph generate` — scan for TODOs and write `.todo-graph`.
- `todo-graph check` — validate TODO references, detect cycles/isolated nodes, and ensure `.todo-graph` matches source.

### Flags and behavior

- `generate --output <path>` — write `.todo-graph` to a custom path.
- `--dir <path>` — run commands against a different repository root.
- `--help`, `-h` — show usage.
- Colors auto-enable on TTY; set `NO_COLOR=1` to disable.

### TODO syntax

```ts
// TODO: short description
// @todo-id some-id
// @todo-deps dep-a, dep-b
```

Rules:

- TODO must start on a comment line (not inline after code).
- Metadata must immediately follow the TODO; only `@todo-id` (required) and `@todo-deps` are allowed.
- IDs must match the regex `^[a-z0-9_-]+$`. Missing `@todo-id` is an error.
- `@todo-deps` is comma-separated; `#` is optional. Space-separated deps error.
- TODOs and metadata are recognized in line-start comments and block comments that start at the beginning of a line (Go/TS/JS/HTML/triple-quote blocks).
