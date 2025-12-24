#!/usr/bin/env node
// Bump versions for meta + platform npm packages in one go.

const fs = require("fs");
const path = require("path");

const newVersion = process.argv[2];
if (!newVersion) {
  console.error("Usage: node scripts/bump-version.js <version>");
  process.exit(1);
}

const pkgPaths = [
  "npm/package.json",
  "npm/comment-graph-linux-64/package.json",
  "npm/comment-graph-darwin-arm64/package.json",
  "npm/comment-graph-win32-64/package.json",
];

for (const rel of pkgPaths) {
  const full = path.join(__dirname, "..", rel);
  const pkg = JSON.parse(fs.readFileSync(full, "utf8"));
  pkg.version = newVersion;
  // Keep optionalDependencies in sync on the meta package.
  if (pkg.optionalDependencies) {
    Object.keys(pkg.optionalDependencies).forEach((key) => {
      pkg.optionalDependencies[key] = newVersion;
    });
  }
  fs.writeFileSync(full, JSON.stringify(pkg, null, 2) + "\n");
  console.log(`Updated ${rel} to ${newVersion}`);
}

console.log("Done. Remember to commit and tag manually.");
