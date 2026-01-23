/**
 * ota-release.js
 *
 * Why this exists:
 * - With runtimeVersion policy "fingerprint", some JS-only changes still create
 *   new runtimes. OTAs published to a new runtime won't reach devices running
 *   the currently deployed store builds.
 * - This script guards production OTA releases by:
 *   1) validating production env vars, and
 *   2) verifying the local runtime fingerprint matches the deployed runtimes.
 * - If a mismatch is detected, it offers to publish targeted OTAs for the
 *   deployed runtimes (iOS + Android), then restores the original config.
 *
 * Planned migration:
 * - Switch runtimeVersion policy from "fingerprint" to "appVersion" once the
 *   next store release is ready. That makes OTAs stable for a given store
 *   version and avoids runtime mismatches for JS-only updates.
 */

const fs = require("node:fs");
const path = require("node:path");
const { execSync } = require("node:child_process");
const readline = require("node:readline");
const { parseProjectEnv } = require("@expo/env");

const projectRoot = process.cwd();
const environment = process.env.EAS_ENVIRONMENT || "production";
const allowRuntimeMismatch = process.env.ALLOW_RUNTIME_MISMATCH === "1";
const otaMessage = process.env.OTA_MESSAGE || "";

function loadFromEasEnv() {
  try {
    const json = execSync(
      `eas env:list --environment ${environment} --json --non-interactive`,
      {
        cwd: projectRoot,
        stdio: ["ignore", "pipe", "pipe"],
        encoding: "utf8",
        env: { ...process.env, CI: "1" },
      },
    );
    const cleaned = json
      .split("\n")
      .filter(
        (line) =>
          line.trim().startsWith("{") ||
          line.trim().startsWith("[") ||
          line.trim().startsWith("]"),
      )
      .join("\n")
      .trim();
    const parsed = JSON.parse(cleaned);
    const env = {};
    for (const item of parsed) {
      if (item?.name) {
        env[item.name] = item.value;
      }
    }
    return { env, source: "eas" };
  } catch (error) {
    try {
      const text = execSync(
        `eas env:list --environment ${environment} --non-interactive`,
        {
          cwd: projectRoot,
          stdio: ["ignore", "pipe", "pipe"],
          encoding: "utf8",
          env: { ...process.env, CI: "1" },
        },
      );
      const env = {};
      const lines = text.split("\n");
      for (const line of lines) {
        const match = line.match(/^([A-Z0-9_]+)=(.*)$/);
        if (match) {
          env[match[1]] = match[2];
        }
      }
      if (Object.keys(env).length) {
        return { env, source: "eas-text" };
      }
    } catch (fallbackError) {
      return { env: {}, source: "eas-error", error: fallbackError };
    }
    return { env: {}, source: "eas-error", error };
  }
}

function extractRuntimeFromOutput(output) {
  if (!output) {
    return null;
  }

  try {
    const parsed = JSON.parse(output);
    if (Array.isArray(parsed) && parsed.length > 0) {
      const candidate = parsed[0];
      return (
        candidate?.runtimeVersion ||
        candidate?.runtime ||
        candidate?.fingerprint ||
        candidate?.hash ||
        null
      );
    }
    if (typeof parsed === "object" && parsed) {
      return (
        parsed.runtimeVersion ||
        parsed.runtime ||
        parsed.fingerprint ||
        parsed.hash ||
        null
      );
    }
  } catch {
    // Fall through to regex extraction.
  }

  const match = output.match(/\b[a-f0-9]{40}\b/i);
  return match ? match[0] : null;
}

function runEasBuildList(platform) {
  try {
    const json = execSync(
      `eas build:list --platform ${platform} --status finished --limit 5 --json --non-interactive`,
      {
        cwd: projectRoot,
        stdio: ["ignore", "pipe", "pipe"],
        encoding: "utf8",
        env: { ...process.env, CI: "1" },
      },
    );
    const cleaned = json
      .split("\n")
      .filter(
        (line) =>
          line.trim().startsWith("{") ||
          line.trim().startsWith("[") ||
          line.trim().startsWith("]"),
      )
      .join("\n")
      .trim();
    const parsed = JSON.parse(cleaned);
    const builds = Array.isArray(parsed) ? parsed : [];
    const prodBuild = builds.find((build) => {
      const status = (build?.status || "").toString().toLowerCase();
      const channel = (build?.channel || "").toString().toLowerCase();
      const distribution = (build?.distribution || "").toString().toLowerCase();
      return (
        status === "finished" &&
        channel === "production" &&
        distribution === "store"
      );
    });
    return (
      prodBuild?.runtimeVersion ||
      prodBuild?.fingerprint?.hash ||
      prodBuild?.fingerprint ||
      null
    );
  } catch (error) {
    try {
      const text = execSync(
        `eas build:list --platform ${platform} --status finished --limit 5 --non-interactive`,
        {
          cwd: projectRoot,
          stdio: ["ignore", "pipe", "pipe"],
          encoding: "utf8",
          env: { ...process.env, CI: "1" },
        },
      );
      const match = text.match(/Runtime Version\s+([a-f0-9]{40})/i);
      return match ? match[1] : null;
    } catch {
      return null;
    }
  }
}

function getLocalRuntime() {
  try {
    const json = execSync(`npx --yes @expo/fingerprint ${projectRoot} --json`, {
      cwd: projectRoot,
      stdio: ["ignore", "pipe", "pipe"],
      encoding: "utf8",
    });
    return extractRuntimeFromOutput(json);
  } catch (error) {
    try {
      const text = execSync(`npx --yes @expo/fingerprint ${projectRoot}`, {
        cwd: projectRoot,
        stdio: ["ignore", "pipe", "pipe"],
        encoding: "utf8",
      });
      return extractRuntimeFromOutput(text);
    } catch {
      return null;
    }
  }
}

function readAppConfig() {
  const appJsonPath = path.join(projectRoot, "app.json");
  const raw = fs.readFileSync(appJsonPath, "utf8");
  return { appJsonPath, raw, data: JSON.parse(raw) };
}

function writeAppConfig(appJsonPath, data) {
  fs.writeFileSync(appJsonPath, JSON.stringify(data, null, 2) + "\n");
}

function publishForRuntime(platform, runtimeVersion) {
  const messageBase = otaMessage || "OTA update";
  const message = `${messageBase} (${platform} runtime ${runtimeVersion.slice(
    0,
    8,
  )})`;
  execSync(
    `eas update --branch production --platform ${platform} --environment ${environment} --message "${message}" --clear-cache`,
    {
      cwd: projectRoot,
      stdio: "inherit",
      env: {
        ...process.env,
        CI: "1",
        EXPO_NO_DOTENV: "1",
      },
    },
  );
}

function publishStandardUpdate() {
  const messageArg = otaMessage ? `--message "${otaMessage}"` : "";
  execSync(
    `eas update --branch production --environment ${environment} --clear-cache ${messageArg}`,
    {
      cwd: projectRoot,
      stdio: "inherit",
      env: {
        ...process.env,
        CI: "1",
        EXPO_NO_DOTENV: "1",
      },
    },
  );
}

function promptYesNo(question) {
  if (!process.stdin.isTTY) {
    return null;
  }

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question(question, (answer) => {
      rl.close();
      resolve(answer.trim().toLowerCase());
    });
  });
}

async function maybePublishDeployedRuntimes(iosRuntime, androidRuntime) {
  if (!iosRuntime || !androidRuntime) {
    throw new Error(
      "[ota-check] Missing deployed runtime versions; cannot publish targeted OTAs.",
    );
  }

  const { appJsonPath, data } = readAppConfig();
  const originalRuntime = data.expo.runtimeVersion;

  try {
    data.expo.runtimeVersion = iosRuntime;
    writeAppConfig(appJsonPath, data);
    publishForRuntime("ios", iosRuntime);

    data.expo.runtimeVersion = androidRuntime;
    writeAppConfig(appJsonPath, data);
    publishForRuntime("android", androidRuntime);
  } finally {
    data.expo.runtimeVersion = originalRuntime;
    writeAppConfig(appJsonPath, data);
  }
}

const { env: fileEnv, files } = parseProjectEnv(projectRoot, {
  mode: "production",
  silent: true,
});

const fromProcess = process.env.EXPO_PUBLIC_API_URL;
let apiUrl = fromProcess || fileEnv.EXPO_PUBLIC_API_URL;
let source = fromProcess
  ? "process"
  : fileEnv.EXPO_PUBLIC_API_URL
    ? "env-files"
    : "none";

if (!apiUrl) {
  const { env: easEnv } = loadFromEasEnv();
  if (easEnv.EXPO_PUBLIC_API_URL) {
    apiUrl = easEnv.EXPO_PUBLIC_API_URL;
    source = "eas";
  }
}

const loadedFiles = files.length
  ? files.map((file) => path.basename(file)).join(", ")
  : "none";

console.log(`[ota-check] Loaded env files: ${loadedFiles}`);
console.log(
  `[ota-check] EXPO_PUBLIC_API_URL=${apiUrl || "(missing)"} (source: ${source})`,
);

const localEnvPath = path.join(projectRoot, ".env.local");
if (fs.existsSync(localEnvPath)) {
  const localContents = fs.readFileSync(localEnvPath, "utf8");
  if (/^\s*EXPO_PUBLIC_API_URL\s*=/m.test(localContents)) {
    console.warn(
      "[ota-check] .env.local defines EXPO_PUBLIC_API_URL; this can override production updates if EAS env is missing.",
    );
  }
}

if (!apiUrl) {
  console.error(
    "[ota-check] EXPO_PUBLIC_API_URL is missing for production OTA.",
  );
  process.exit(1);
}

const normalized = apiUrl.toLowerCase();
const isLocalHost =
  normalized.includes("10.0.2.2") ||
  normalized.includes("localhost") ||
  normalized.includes("127.0.0.1");
const isHttp = normalized.startsWith("http://");

if (isLocalHost || isHttp) {
  console.error(
    `[ota-check] EXPO_PUBLIC_API_URL is not valid for production: ${apiUrl}`,
  );
  process.exit(1);
}

console.log("[ota-check] OK.");

const localRuntime = getLocalRuntime();
const iosRuntime = runEasBuildList("ios");
const androidRuntime = runEasBuildList("android");

console.log(`[ota-check] Local runtime fingerprint: ${localRuntime || "N/A"}`);
console.log(
  `[ota-check] Latest production iOS runtime: ${iosRuntime || "N/A"}`,
);
console.log(
  `[ota-check] Latest production Android runtime: ${androidRuntime || "N/A"}`,
);

async function runRuntimeChecks() {
  if (!localRuntime || !iosRuntime || !androidRuntime) {
    console.warn(
      "[ota-check] Could not fully verify runtime compatibility. Continue with caution.",
    );
    const answer = await promptYesNo(
      "Proceed with publishing the OTA anyway? (y/N): ",
    );
    if (answer === "y" || answer === "yes") {
      publishStandardUpdate();
      return;
    }
    process.exit(1);
  }

  if (localRuntime === iosRuntime && localRuntime === androidRuntime) {
    publishStandardUpdate();
    return;
  }

  const message =
    "[ota-check] Local runtime fingerprint does not match latest production builds. " +
    "This OTA will not reach existing users unless you target the deployed runtimes or ship new store builds.";

  if (allowRuntimeMismatch) {
    console.warn(`${message} (ALLOW_RUNTIME_MISMATCH=1 set)`);
    publishStandardUpdate();
    return;
  }

  const answer = await promptYesNo(
    `${message}\nPublish OTAs targeting deployed runtimes now? (y/N): `,
  );

  if (answer === "y" || answer === "yes") {
    await maybePublishDeployedRuntimes(iosRuntime, androidRuntime);
    return;
  }

  console.error(message);
  process.exit(1);
}

async function main() {
  await runRuntimeChecks();
}

main().catch((error) => {
  console.error(error?.message || error);
  process.exit(1);
});
