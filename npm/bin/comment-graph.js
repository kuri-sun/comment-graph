#!/usr/bin/env node
const { spawnSync } = require("node:child_process");
const { existsSync } = require("node:fs");
const path = require("node:path");

const binDir = __dirname;
const exe = process.platform === "win32" ? "comment-graph.exe" : "comment-graph";
const binPath = path.join(binDir, exe);

if (!existsSync(binPath)) {
  console.error(`comment-graph binary not found at ${binPath}. Did postinstall succeed?`);
  process.exit(1);
}

const args = process.argv.slice(2);
const result = spawnSync(binPath, args, { stdio: "inherit" });
process.exit(result.status ?? 1);
