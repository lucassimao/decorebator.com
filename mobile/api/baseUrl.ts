import * as Updates from "expo-updates";

const PROD_API_URL = "https://api.decorebator.com";
const LOCAL_HOSTS = ["10.0.2.2", "localhost", "127.0.0.1"];

function isProductionChannel() {
  const legacyChannel = (Updates as { releaseChannel?: string }).releaseChannel;
  const channel = Updates.channel ?? legacyChannel;
  return channel === "production";
}

function isLocalUrl(url: string) {
  const normalized = url.toLowerCase();
  if (LOCAL_HOSTS.some((host) => normalized.includes(host))) {
    return true;
  }
  return normalized.startsWith("http://");
}

export function getApiBaseUrl() {
  const raw = process.env.EXPO_PUBLIC_API_URL;

  if (isProductionChannel()) {
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
