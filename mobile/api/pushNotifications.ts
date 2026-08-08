import { DEFAULT_ERROR } from "./constants";
import {
  authenticatedFetch,
  authenticationHeaders,
  hasAuthenticationSession,
} from "./users";
import * as Sentry from "@sentry/react-native";
import { getApiBaseUrl } from "./baseUrl";

export type RegisterPushTokenInput = {
  expoPushToken: string;
  platform: "ios" | "android";
  deviceId?: string;
  timezone: string;
  locale?: string;
};

export async function registerPushToken(
  input: RegisterPushTokenInput,
  signal?: AbortSignal,
) {
  const endpoint = getApiBaseUrl() + "/push/register";
  if (!hasAuthenticationSession()) {
    throw new Error("Authentication required");
  }

  const response = await authenticatedFetch(endpoint, {
    method: "POST",
    body: JSON.stringify(input),
    headers: {
      "Content-Type": "application/json",
      ...authenticationHeaders(),
    },
    credentials: "include",
    signal,
  });

  if (!response.ok) {
    let body: any = null;
    try {
      body = await response.json();
    } catch {
      // Ignore JSON parsing errors to surface status-based failure instead.
    }
    Sentry.addBreadcrumb({
      message: "Push register request failed",
      category: "notifications",
      level: "error",
      data: {
        endpoint,
        status: response.status,
      },
    });
    const message =
      body?.error ||
      Object.values(body?.validationErrors || {})?.[0] ||
      `${DEFAULT_ERROR} (status ${response.status})`;
    throw new Error(message);
  }
}
