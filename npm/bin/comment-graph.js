#!/usr/bin/env node
const { spawnSync } = require("node:child_process");
const { existsSync } = require("node:fs");
const path = require("node:path");

const binPath = path.join(__dirname, "comment-graph" + (process.platform === "win32" ? ".exe" : ""));

if (!existsSync(binPath)) {
  console.error(`comment-graph binary missing at ${binPath}`);
  process.exit(1);
}

const result = spawnSync(binPath, process.argv.slice(2), { stdio: "inherit" });
process.exit(result.status ?? 1);
