# todo-graph (npm wrapper)

Thin npm wrapper that downloads the `todo-graph` binary for your platform and exposes it as `todo-graph` on the PATH.

## Usage

```bash
npm install todo-graph
npx todo-graph scan
```

## Notes

- Requires Node.js 16+.
- On install, downloads release artifacts from `github.com/todo-graph/todo-graph` based on the package version.
- Override source via env vars:
  - `TODO_GRAPH_REPO` (default `todo-graph/todo-graph`)
  - `TODO_GRAPH_HOST` (default `https://github.com`)
