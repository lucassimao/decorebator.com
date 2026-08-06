import * as Updates from "expo-updates";

const PROD_API_URL = "https://api.decorebator.com";
const LOCAL_HOSTS = new Set([
  "0.0.0.0",
  "10.0.2.2",
  "127.0.0.1",
  "[::1]",
  "localhost",
]);

function currentChannel() {
  const legacyChannel = (Updates as { releaseChannel?: string }).releaseChannel;
  return Updates.channel ?? legacyChannel;
}

function canonicalHostname(url: URL) {
  return url.hostname.toLowerCase().replace(/\.+$/, "");
}

function isLocalUrl(url: string) {
  try {
    const parsed = new URL(url);
    return (
      parsed.protocol === "http:" || LOCAL_HOSTS.has(canonicalHostname(parsed))
    );
  } catch {
    return false;
  }
}

function storeTestApiUrl(raw: string | undefined) {
  if (!raw) {
    throw new Error("store-test must define EXPO_PUBLIC_API_URL");
  }

  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    throw new Error("store-test EXPO_PUBLIC_API_URL must be an absolute URL");
  }
  if (parsed.protocol !== "https:" || isLocalUrl(raw)) {
    throw new Error("store-test must use an HTTPS API URL");
  }
  if (canonicalHostname(parsed) === canonicalHostname(new URL(PROD_API_URL))) {
    throw new Error("store-test must not use the production API");
  }
  return raw;
}

export function getApiBaseUrl() {
  const raw = process.env.EXPO_PUBLIC_API_URL;
  const channel = currentChannel();

  if (channel === "store-test") {
    return storeTestApiUrl(raw);
  }

  if (channel === "production") {
    if (!raw) {
      return PROD_API_URL;
    }

    if (isLocalUrl(raw)) {
      console.error(
        `EXPO_PUBLIC_API_URL=${raw} is not allowed on production channel; using ${PROD_API_URL} instead.`,
      );
      return PROD_API_URL;
    }
  }

  if (!raw) {
    if (__DEV__) {
      throw new Error("EXPO_PUBLIC_API_URL is not set.");
    }
    return PROD_API_URL;
  }

  return raw;
}
