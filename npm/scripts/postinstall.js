#!/usr/bin/env node
const { spawnSync } = require("node:child_process");
const {
  createWriteStream,
  chmodSync,
  existsSync,
  mkdirSync,
  renameSync,
} = require("node:fs");
const { pipeline } = require("node:stream");
const { promisify } = require("node:util");
const https = require("node:https");
const path = require("node:path");

const pkg = require("../package.json");

const pipe = promisify(pipeline);

async function main() {
  const destDir = path.join(__dirname, "..", "bin");
  mkdirSync(destDir, { recursive: true });

  const platform = mapPlatform(process.platform);
  const arch = mapArch(process.arch);
  const version = pkg.version;
  const repo = "kuri-sun/comment-graph";
  const host = "https://github.com";

  const ext = platform === "windows" ? "zip" : "tar.gz";
  const archiveName = `comment-graph_${version}_${platform}_${arch}.${ext}`;
  const url = `${host}/${repo}/releases/download/v${version}/${archiveName}`;

  const archivePath = path.join(destDir, archiveName);
  console.log(
    `Downloading comment-graph ${version} (${platform}/${arch}) from ${url}`,
  );
  await download(url, archivePath);

  const extractDir = path.join(destDir, "tmp-extract");
  mkdirSync(extractDir, { recursive: true });
  extract(archivePath, extractDir, ext);

  const binaryName =
    platform === "windows" ? "comment-graph.exe" : "comment-graph";
  const extractedPath = path.join(extractDir, binaryName);
  if (!existsSync(extractedPath)) {
    console.error("Failed to locate extracted binary at", extractedPath);
    process.exit(1);
  }

  const finalPath = path.join(destDir, binaryName);
  renameSync(extractedPath, finalPath);
  chmodSync(finalPath, 0o755);
  console.log("Installed comment-graph binary to", finalPath);
}

function mapPlatform(p) {
  switch (p) {
    case "darwin":
      return "darwin";
    case "linux":
      return "linux";
    case "win32":
      return "windows";
    default:
      console.error(`Unsupported platform: ${p}`);
      process.exit(1);
  }
}

function mapArch(a) {
  switch (a) {
    case "x64":
      return "amd64";
    case "arm64":
      return "arm64";
    default:
      console.error(`Unsupported architecture: ${a}`);
      process.exit(1);
  }
}

async function download(url, dest) {
  const write = createWriteStream(dest);
  await new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode !== 200) {
          reject(new Error(`download failed: ${res.statusCode}`));
          return;
        }
        pipe(res, write).then(resolve).catch(reject);
      })
      .on("error", reject);
  });
}

function extract(archive, dest, ext) {
  if (ext === "zip") {
    const unzip = spawnSync(
      "powershell",
      [
        "-Command",
        `Expand-Archive -Path "${archive}" -DestinationPath "${dest}" -Force`,
      ],
      {
        stdio: "inherit",
      },
    );
    if (unzip.status !== 0) {
      console.error("Failed to unzip archive");
      process.exit(1);
    }
    return;
  }

  const tar = spawnSync("tar", ["-xzf", archive, "-C", dest], {
    stdio: "inherit",
  });
  if (tar.status !== 0) {
    console.error("Failed to untar archive (tar command required)");
    process.exit(1);
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
