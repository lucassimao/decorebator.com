import assert from "node:assert/strict";
import test from "node:test";
import {
  assertNativeRuntimeContract,
  readNativeRuntimeContract,
  requireNativeRuntimeReady,
} from "./nativeRuntimeContract.mjs";

test("the profile processor is isolated to a new app-version runtime", () => {
  const appVersion = assertNativeRuntimeContract(readNativeRuntimeContract());
  assert.equal(appVersion, "1.1.2");
});

test("OTA publishing requires matching native binaries", () => {
  assert.throws(() => requireNativeRuntimeReady({}, "1.1.2"), /refusing OTA/);
  assert.throws(
    () =>
      requireNativeRuntimeReady(
        { NATIVE_RUNTIME_READY_VERSION: "1.1.1" },
        "1.1.2",
      ),
    /refusing OTA/,
  );
  assert.doesNotThrow(() =>
    requireNativeRuntimeReady(
      { NATIVE_RUNTIME_READY_VERSION: "1.1.2" },
      "1.1.2",
    ),
  );
});
