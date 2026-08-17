import fs from "node:fs";
import path from "node:path";

export const FIRST_PROFILE_PROCESSOR_RUNTIME = "1.1.2";

export function readNativeRuntimeContract(projectRoot = process.cwd()) {
  const app = JSON.parse(
    fs.readFileSync(path.join(projectRoot, "app.json"), "utf8"),
  );
  const pkg = JSON.parse(
    fs.readFileSync(path.join(projectRoot, "package.json"), "utf8"),
  );
  return { app, pkg };
}

export function assertNativeRuntimeContract({ app, pkg }) {
  const appVersion = app?.expo?.version;
  if (appVersion !== pkg?.version) {
    throw new Error("app.json and package.json versions must match");
  }
  if (compareVersions(appVersion, FIRST_PROFILE_PROCESSOR_RUNTIME) < 0) {
    throw new Error(
      `profile image processor requires runtime ${FIRST_PROFILE_PROCESSOR_RUNTIME} or newer`,
    );
  }
  if (app?.expo?.runtimeVersion?.policy !== "appVersion") {
    throw new Error("runtimeVersion policy must remain appVersion");
  }
  if (app?.expo?.updates?.enabled !== true) {
    throw new Error("Expo updates must be explicitly enabled");
  }
  return appVersion;
}

function compareVersions(left, right) {
  const parse = (value) => {
    if (typeof value !== "string" || !/^\d+\.\d+\.\d+$/.test(value)) {
      throw new Error("app version must be a three-part numeric version");
    }
    return value.split(".").map(Number);
  };
  const leftParts = parse(left);
  const rightParts = parse(right);
  for (let index = 0; index < leftParts.length; index += 1) {
    if (leftParts[index] !== rightParts[index]) {
      return leftParts[index] - rightParts[index];
    }
  }
  return 0;
}

export function requireNativeRuntimeReady(environment, appVersion) {
  if (environment.NATIVE_RUNTIME_READY_VERSION !== appVersion) {
    throw new Error(
      `refusing OTA: set NATIVE_RUNTIME_READY_VERSION=${appVersion} only after matching iOS and Android binaries are available`,
    );
  }
}
