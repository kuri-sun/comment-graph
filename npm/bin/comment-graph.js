#!/usr/bin/env node
const { spawn } = require("node:child_process");
const { existsSync } = require("node:fs");
const path = require("node:path");

const isWin = process.platform === "win32";
const bin = path.join(__dirname, isWin ? "comment-graph.exe" : "comment-graph");

if (!existsSync(bin)) {
  console.error("comment-graph binary is missing. Try reinstalling the package.");
  process.exit(1);
}

const child = spawn(bin, process.argv.slice(2), { stdio: "inherit" });
child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code == null ? 1 : code);
  }
});
