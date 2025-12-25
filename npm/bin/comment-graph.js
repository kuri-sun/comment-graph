#!/usr/bin/env node
const { spawnSync } = require("node:child_process");
const { existsSync, readFileSync } = require("node:fs");
const path = require("node:path");

function resolvePlatformPackage() {
  const platform = process.platform;
  const arch = process.arch;
  let pkg;
  if (platform === "linux" && arch === "x64") pkg = "@comment-graph/comment-graph-linux-64";
  else if (platform === "linux" && arch === "arm64") pkg = "@comment-graph/comment-graph-linux-arm64";
  else if (platform === "darwin" && arch === "x64") pkg = "@comment-graph/comment-graph-darwin-64";
  else if (platform === "darwin" && arch === "arm64") pkg = "@comment-graph/comment-graph-darwin-arm64";
  else if (platform === "win32" && arch === "x64") pkg = "@comment-graph/comment-graph-win32-64";
  else {
    console.error(`comment-graph: unsupported platform/arch ${platform}/${arch}`);
    process.exit(1);
  }
  let pkgJsonPath;
  try {
    pkgJsonPath = require.resolve(`${pkg}/package.json`);
  } catch (err) {
    console.error(`comment-graph: platform package ${pkg} not installed`);
    process.exit(1);
  }
  const pkgRoot = path.dirname(pkgJsonPath);
  const pkgJson = JSON.parse(readFileSync(pkgJsonPath, "utf8"));
  const binRel = (pkgJson.bin && pkgJson.bin["comment-graph"]) || "bin/comment-graph";
  const binPath = path.join(pkgRoot, binRel);
  if (!existsSync(binPath)) {
    console.error(`comment-graph: binary not found at ${binPath}`);
    process.exit(1);
  }
  return binPath;
}

const binPath = resolvePlatformPackage();
const result = spawnSync(binPath, process.argv.slice(2), { stdio: "inherit" });
process.exit(result.status ?? 1);
