import { getAuthorization } from "./users";
import { DEFAULT_ERROR } from "./constants";
import { Platform } from "react-native";

const API_URL = process.env.EXPO_PUBLIC_API_URL;

export type PaymentProvider = "stripe" | "revenuecat";

interface RestorePurchasesRequest {
  appUserId: string;
  platform: "ios" | "android";
}

export async function restorePurchases(appUserId: string): Promise<void> {
  const endpoint = `${API_URL}/subscription/revenuecat/restore`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const platform =
    Platform.OS === "ios"
      ? "ios"
      : Platform.OS === "android"
        ? "android"
        : "web";

  if (platform === "web") {
    throw new Error("RevenueCat is not available on web");
  }

  const request: RestorePurchasesRequest = {
    appUserId,
    platform: platform as "ios" | "android",
  };

  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: authorization,
    },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const body = await response.json();
    throw new Error(body?.error || DEFAULT_ERROR);
  }
}
