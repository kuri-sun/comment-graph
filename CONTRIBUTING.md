# Contributing

Thanks for your interest in improving todo-graph! Please follow these guidelines to help us move quickly.

## Prerequisites

- Go 1.25
- Node 20+ (only if working on the npm wrapper)
- `gofmt`, `go vet` (run via CI)

## Setup

```bash
git clone https://github.com/kuri-sun/todo-graph.git
cd todo-graph
go mod download
```

## Development workflow

- Format and lint:
  ```bash
  gofmt -w .
  go vet ./...
  ```
- Run tests:
  ```bash
  go test ./...
  # For integration tests, ensure GOCACHE/GOMODCACHE are writable
  ```

## Pull requests

- Keep PRs focused; describe the change and rationale.
- Add or update tests for any behavior changes.
- If adding CLI flags or user-facing behavior, update the README/CLI docs.
- CI runs gofmt, go vet, and go test; please ensure they pass locally first.

## Reporting issues

When filing an issue, include:
- What you expected vs. what happened
- Repro steps (ideally minimal)
- Environment (OS, Go version)
