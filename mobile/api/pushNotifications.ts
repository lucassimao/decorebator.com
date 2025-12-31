import { DEFAULT_ERROR } from "./constants";
import { getAuthorization } from "./users";
import * as Sentry from "@sentry/react-native";

export type RegisterPushTokenInput = {
  expoPushToken: string;
  platform: "ios" | "android";
  deviceId?: string;
  timezone: string;
  locale?: string;
};

export async function registerPushToken(input: RegisterPushTokenInput) {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/push/register";
  const authorization = await getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
    method: "POST",
    body: JSON.stringify(input),
    headers: {
      "Content-Type": "application/json",
      Authorization: authorization,
    },
    credentials: "include",
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

export async function unregisterPushToken(expoPushToken: string) {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/push/unregister";
  const authorization = await getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
    method: "POST",
    body: JSON.stringify({ expoPushToken }),
    headers: {
      "Content-Type": "application/json",
      Authorization: authorization,
    },
    credentials: "include",
  });

  if (!response.ok) {
    let body: any = null;
    try {
      body = await response.json();
    } catch {
      // Ignore JSON parsing errors to surface status-based failure instead.
    }
    Sentry.addBreadcrumb({
      message: "Push unregister request failed",
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
