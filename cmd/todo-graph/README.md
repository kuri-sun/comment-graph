# todo-graph CLI

## Usage

- `todo-graph generate` — scan for TODOs and write `.todo-graph` (runs validation first).
- `todo-graph check` — validate TODO references, detect cycles/isolated nodes, and ensure `.todo-graph` matches source.
- `todo-graph visualize` — read `.todo-graph` and print an indented tree of the TODO graph (invokes generate first).
