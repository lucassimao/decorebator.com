const fs = require("node:fs");
const path = require("node:path");
const { execSync } = require("node:child_process");
const { parseProjectEnv } = require("@expo/env");

const projectRoot = process.cwd();
const environment = process.env.EAS_ENVIRONMENT || "production";

function loadFromEasEnv() {
  try {
    const json = execSync(`eas env:list --environment ${environment} --json`, {
      cwd: projectRoot,
      stdio: ["ignore", "pipe", "pipe"],
      encoding: "utf8",
    });
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
      const text = execSync(`eas env:list --environment ${environment}`, {
        cwd: projectRoot,
        stdio: ["ignore", "pipe", "pipe"],
        encoding: "utf8",
      });
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
