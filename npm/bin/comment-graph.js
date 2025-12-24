#!/usr/bin/env node
const fs = require("node:fs");
const path = require("node:path");

const wasmDir = path.join(__dirname, "..", "wasm");
const wasmPath = path.join(wasmDir, "comment-graph.wasm");
const wasmExecPath = path.join(wasmDir, "wasm_exec.js");

if (!fs.existsSync(wasmPath)) {
  console.error(`comment-graph wasm missing at ${wasmPath}`);
  process.exit(1);
}
if (!fs.existsSync(wasmExecPath)) {
  console.error(`wasm_exec.js missing at ${wasmExecPath}`);
  process.exit(1);
}

// Go runtime shim from the Go toolchain; defines global Go.
require(wasmExecPath);

async function main() {
  const go = new Go();
  go.argv = ["comment-graph", ...process.argv.slice(2)];
  go.env = process.env;
  go.exit = (code) => {
    process.exit(code);
  };

  const bytes = fs.readFileSync(wasmPath);
  const { instance } = await WebAssembly.instantiate(bytes, go.importObject);
  await go.run(instance);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
