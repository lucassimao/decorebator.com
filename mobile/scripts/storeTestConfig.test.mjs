import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { URL } from "node:url";

const root = new URL("../", import.meta.url);
const [appConfig, easConfig] = await Promise.all([
  readFile(new URL("app.json", root), "utf8").then(JSON.parse),
  readFile(new URL("eas.json", root), "utf8").then(JSON.parse),
]);

test("store-test builds are store-distributed through the preview environment", () => {
  const profile = easConfig.build["store-test"];

  assert.equal(profile.channel, "store-test");
  assert.equal(profile.environment, "preview");
  assert.notEqual(profile.distribution, "internal");
  assert.equal(profile.ios.autoIncrement, "buildNumber");
  assert.equal(profile.android.autoIncrement, "versionCode");
});

test("store-test submission cannot target Google production", () => {
  const storeTest = easConfig.submit["store-test"];
  const production = easConfig.submit.production;

  assert.equal(storeTest.android.track, "internal");
  assert.equal(storeTest.android.releaseStatus, "completed");
  assert.equal(production.android.track, "production");
  assert.notEqual(storeTest.android.track, production.android.track);
});

test("store identifiers agree across app and submit configuration", () => {
  assert.equal(
    appConfig.expo.ios.bundleIdentifier,
    "com.lsimaocosta.decorebator",
  );
  assert.equal(appConfig.expo.android.package, "com.lsimaocosta.decorebator");
  assert.equal(easConfig.submit["store-test"].ios.ascAppId, "6749329064");
  assert.equal(easConfig.submit.production.ios.ascAppId, "6749329064");
});

test("native IAP plugin is present in the store build source config", () => {
  const pluginNames = appConfig.expo.plugins.map((plugin) =>
    Array.isArray(plugin) ? plugin[0] : plugin,
  );

  assert.ok(pluginNames.includes("expo-iap"));
  assert.ok(!pluginNames.includes("react-native-purchases"));
});
