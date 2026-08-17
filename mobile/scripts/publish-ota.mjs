#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import {
  assertNativeRuntimeContract,
  readNativeRuntimeContract,
  requireNativeRuntimeReady,
} from "./nativeRuntimeContract.mjs";
import { buildOTAArguments } from "./otaArguments.mjs";

const target = process.argv[2];
let easArguments;

try {
  easArguments = buildOTAArguments(target, process.argv.slice(3), process.env);
  const appVersion = assertNativeRuntimeContract(readNativeRuntimeContract());
  requireNativeRuntimeReady(process.env, appVersion);
} catch (error) {
  console.error(error instanceof Error ? error.message : error);
  process.exit(1);
}

const result = spawnSync("eas", easArguments, {
  env: process.env,
  stdio: "inherit",
});
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
process.exit(result.status ?? 1);
