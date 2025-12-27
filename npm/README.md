## Installation

```bash
npm install --save-dev comment-graph
# or
yarn add -D comment-graph
# or
pnpm add -D comment-graph
```

## Usage

You can visualize dependencies graph between comments with [comment-graph.nvim](https://github.com/kuri-sun/comment-graph.nvim).

## Syntax

```ts
// user.ts
// @cgraph-id db-migration

// cache.ts
// @cgraph-id cache-user
// @cgraph-label Cache user reads
// @cgraph-deps db-migration
```

### Rules:

- Comment metadata must start on a comment line (not inline after code).
- Metadata must immediately follow the comment (no blank/non-comment lines); only `@cgraph-id` (required), `@cgraph-label` (optional), and `@cgraph-deps` are allowed.
- IDs must use lowercase letters/digits/hyphens/underscores. Missing `@cgraph-id` is an error.
- `@cgraph-deps` is comma-separated; spaces after commas are allowed.

## Supported comment styles

- `//` — C/C++/C#/Java/Go/JS/TS/Swift
- `#` — Python, Shell, Ruby, YAML
- `--` — SQL, Lua
- `/* ... */` and `{/* ... */}` — C-family block comments (JSX/TSX friendly)
- `<!-- ... -->` — HTML/Markdown
- `""" ... """` / `''' ... '''` — Python-style docstrings

Inline trailing comments (`code(); // @cgraph-id ...`) are not picked up; place metadata on comment lines.
