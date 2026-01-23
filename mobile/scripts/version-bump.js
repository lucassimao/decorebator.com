#!/usr/bin/env node
const fs = require("node:fs");
const path = require("node:path");
const readline = require("node:readline");
const semver = require("semver");

const projectRoot = process.cwd();
const packageJsonPath = path.join(projectRoot, "package.json");
const appJsonPath = path.join(projectRoot, "app.json");

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

function writeJson(filePath, data) {
  fs.writeFileSync(filePath, JSON.stringify(data, null, 2) + "\n");
}

function prompt(question) {
  if (!process.stdin.isTTY) {
    return Promise.resolve(null);
  }
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });
  return new Promise((resolve) => {
    rl.question(question, (answer) => {
      rl.close();
      resolve(answer.trim());
    });
  });
}

async function main() {
  const pkg = readJson(packageJsonPath);
  const app = readJson(appJsonPath);

  const currentPkgVersion = pkg.version || "unknown";
  const currentAppVersion = app?.expo?.version || "unknown";

  console.log(`[version] package.json: ${currentPkgVersion}`);
  console.log(`[version] app.json: ${currentAppVersion}`);

  const target = await prompt("New version (e.g. 1.1.1): ");
  if (!target) {
    console.error("[version] No version provided. Aborting.");
    process.exit(1);
  }

  const valid = semver.valid(target);
  if (!valid) {
    console.error(`[version] Invalid semver: "${target}"`);
    process.exit(1);
  }

  if (target === currentPkgVersion && target === currentAppVersion) {
    console.log("[version] Already at that version. No changes made.");
    return;
  }

  pkg.version = target;
  if (!app.expo) {
    app.expo = {};
  }
  app.expo.version = target;

  writeJson(packageJsonPath, pkg);
  writeJson(appJsonPath, app);

  console.log(`[version] Updated to ${target}.`);
  console.log("");
  console.log("Next steps for a store release:");
  console.log("- Build production binaries:");
  console.log("  - npm run build:ios");
  console.log("  - npm run build:android");
  console.log("- Submit to stores:");
  console.log("  - npm run submit:ios");
  console.log("  - npm run submit:android");
}

main().catch((error) => {
  console.error(error?.message || error);
  process.exit(1);
});
