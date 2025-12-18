# todo-graph CLI

## Usage

- `todo-graph generate` — scan for TODOs and write `.todo-graph` (runs validation first).
- `todo-graph check` — validate TODO references, detect cycles/isolated nodes, and ensure `.todo-graph` matches source.
- `todo-graph visualize` — read `.todo-graph` and print an indented tree of the TODO graph (invokes generate first).

### Flags and behavior

- `--help`, `-h` — show usage.
- Colors auto-enable on TTY; set `NO_COLOR=1` to disable.
- Exit codes:
  - `generate`: 0 on success; 1–3 on validation/write errors.
  - `check`: 0 success, 1 undefined refs, 2 cycles, 3 isolated/out-of-date/scan errors.
  - `visualize`: 0 success; propagates `generate` failures.

### TODO syntax

```ts
// TODO:[#id] short description
// DEPS: #other-id, #another
```

IDs must be lowercase letters/digits/hyphens/underscores. If no `[#id]` is provided, an ID like `todo-<line>` is generated. Only `DEPS` is parsed; values must be comma-separated, each prefixed with `#`. Multiple TODOs can appear in block or inline comments across languages (Go/TS/JS/HTML/triple-quote blocks).
