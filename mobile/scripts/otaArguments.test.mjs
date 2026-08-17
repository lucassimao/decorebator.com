import assert from "node:assert/strict";
import test from "node:test";
import { buildOTAArguments } from "./otaArguments.mjs";

test("builds every documented OTA target", () => {
  assert.deepEqual(buildOTAArguments("development").slice(0, 5), [
    "update",
    "--environment",
    "development",
    "--channel",
    "development",
  ]);
  assert.ok(buildOTAArguments("preview").includes("preview"));
  assert.ok(buildOTAArguments("production").includes("production"));
});

test("forwards one validated explicit message", () => {
  assert.deepEqual(buildOTAArguments("preview", ["--message", "Fix upload"]), [
    "update",
    "--environment",
    "preview",
    "--channel",
    "preview",
    "--clear-cache",
    "--message",
    "Fix upload",
  ]);
  assert.throws(() => buildOTAArguments("preview", ["--branch", "other"]));
});

test("translates OTA_MESSAGE and rejects conflicting message sources", () => {
  assert.deepEqual(
    buildOTAArguments("production", [], { OTA_MESSAGE: "Fix upload" }).slice(
      -2,
    ),
    ["--message", "Fix upload"],
  );
  assert.throws(() =>
    buildOTAArguments("production", ["--message", "one"], {
      OTA_MESSAGE: "two",
    }),
  );
});
