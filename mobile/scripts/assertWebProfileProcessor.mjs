import { readdir, readFile } from "node:fs/promises";
import path from "node:path";

const outputRoot = path.resolve(".expo/profile-upload-web-smoke");
const javascript = [];

async function collect(directory) {
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      await collect(target);
    } else if (entry.isFile() && entry.name.endsWith(".js")) {
      javascript.push(await readFile(target, "utf8"));
    }
  }
}

await collect(outputRoot);
const bundle = javascript.join("\n");
const webFailure =
  "Profile image uploads are available in the iOS and Android apps.";

if (!bundle.includes(webFailure)) {
  throw new Error(
    "Metro web export did not resolve the profile-image processor web entry",
  );
}
if (bundle.includes('requireNativeModule("ProfileImageProcessor")')) {
  throw new Error("Metro web export retained the native profile-image module");
}
