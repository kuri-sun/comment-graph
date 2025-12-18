# todo-graph

> **TODO is not a comment. It’s a dependency.**

`todo-graph` parses TODO comments, tracks dependencies between them, and keeps a lightweight `.todo-graph` file you can validate and visualize.

## Why

- TODOs vanish without trace.
- Preconditions hide in comments.
- Order between TODOs is tribal knowledge.

So `todo-graph` only tracks relationships—no status, no owner, no progress.

## TODO syntax

```txt
TODO:[#id]   // or TODO[#id]   // or TODO: cache-user implement cache
```

- ID regex: `[a-z0-9_-]+`.
- If you omit `[#id]`, an ID is derived from the description (slugged, lowercase).
- Example: `// TODO:[#cache-user] Implement user cache` or `// TODO: cache-user implement user cache`

Optional metadata (only dependencies are used):

```txt
// TODO: cache-user implement user cache
// depends-on: #db-migration
```

- Unknown keys are ignored.
- `status` / `owner` are out of scope.
  `blocks` is not parsed in MVP (only `depends-on` is supported).

### TODO block boundary

- Starts at a `TODO:[#id]` line.
- Extends until the next `TODO:` line, or until a non-comment line is seen.

## `.todo-graph` (SSOT)

```yaml
version: 1

todos:
  cache-user:
    file: "src/userService.ts"
    line: 12

edges:
  - from: db-migration
    to: cache-user
    type: blocks
```

Only dependencies are stored; TODO state or owners are never recorded.

## Commands

- `todo-graph scan` — scan the repo and (re)write `.todo-graph`.
- `todo-graph check` — validate undefined references, cycles, isolated TODOs, and `.todo-graph` drift.
- `todo-graph visualize --format mermaid` — render the graph as Mermaid.

### Exit codes (`check`)

| code | meaning             |
| ---- | ------------------- |
| 0    | clean               |
| 1    | undefined reference |
| 2    | cycle detected      |
| 3    | graph/code mismatch |

## Edge cases.

- `TODO:` without `[#id]` → ID is derived from description.
- Empty ID (`TODO:[#]`) → error.
- Duplicate IDs → error.
- Unknown metadata keys → ignored.
- Self-dependency and cycles are detected.
- Comments ending immediately after a TODO block stop metadata parsing.

## Non Yet.

- Status/progress tracking.
- Ownership.
- Issue tracker integration.
- Auto-fixing TODOs.

## Usage (example)

```bash
todo-graph scan
todo-graph check
todo-graph visualize --format mermaid
```

> **Your TODOs already form a graph. You just don’t track it yet.**
