import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";
import { DEFAULT_ERROR } from "./constants";
import offlineManager from "@/utils/offlineManager";
import * as Sentry from "@sentry/react-native";
import { getApiBaseUrl } from "./baseUrl";
import { requestAnalyticsIdentityReset } from "@/utils/activationEvents";

export type UserProfile = {
  id: number;
  email: string;
  firstName: string;
  lastName: string;
  profilePictureUrl?: string;
  country?: string;
  dateOfBirth?: string; // YYYY-MM-DD
  createdAt: string;
  preferredLanguage: string;
  subscriptionPlan: "free" | "monthly" | "annual";
  notificationsEnabled: boolean;
};
export type UserSignup = {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  country?: string; // Optional ISO 3166-1 alpha-2 country code
  preferredLanguage?: string; // Optional language code
};

export type UserSignin = {
  email: string;
  password: string;
};

export type UpdateInput = {
  firstName?: string;
  lastName?: string;
  country?: string;
  dateOfBirth?: string;
  preferredLanguage?: string;
  notificationsEnabled?: boolean;

  // if set, triggers a password update
  updatePassword?: {
    currentPassword: string;
    newPassword: string;
  };

  // if set, triggers profile pic update
  updateProfilePicture?: {
    base64Data: string;
    extension: string;
  };
};

export type UpdatePasswordPayload = NonNullable<UpdateInput["updatePassword"]>;

export const SIGN_IN_ERROR =
  "Invalid credentials. Are you using the correct email and password?";

const API_URL = getApiBaseUrl();
let cachedAuthorization: string | null | undefined;
let cachedRefreshToken: string | null | undefined;
let refreshInFlight: Promise<boolean> | null = null;

function authClientHeaders(): Record<string, string> {
  return Platform.OS === "ios" || Platform.OS === "android"
    ? { "X-Auth-Client": "native" }
    : {};
}

export async function signup(data: UserSignup) {
  const endpoint = API_URL + "/users";

  const response = await fetch(endpoint, {
    method: "POST",
    body: JSON.stringify(data),
    headers: {
      "Content-Type": "application/json",
      ...authClientHeaders(),
    },
    credentials: "include",
  });
  if (!response.ok) {
    const body = await response.json();
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
  saveSessionResponse(response);
}

export async function sigout() {
  const refreshToken = getRefreshToken();
  try {
    await fetch(API_URL + "/logout", {
      method: "POST",
      credentials: "include",
      headers: {
        ...authClientHeaders(),
        ...(refreshToken ? { "X-Refresh-Token": refreshToken } : {}),
      },
    });
  } catch {
    // Local sign-out must still complete while offline.
  }
  // Clear offline cache when signing out
  try {
    await offlineManager.clearCache();
  } catch (error) {
    console.error("Error clearing offline cache:", error);
  }

  Sentry.setUser(null);

  await clearLocalCredentials();
}
export async function signin(data: UserSignin) {
  const endpoint = API_URL + "/login";

  const response = await fetch(endpoint, {
    method: "POST",
    body: JSON.stringify(data),
    headers: {
      "Content-Type": "application/json",
      ...authClientHeaders(),
    },
    credentials: "include",
  });
  if (!response.ok) {
    throw new Error(SIGN_IN_ERROR);
  }
  saveSessionResponse(response);
}

export async function requestResetEmailPassword(email: string) {
  const endpoint = API_URL + "/password/send-reset-email";

  const response = await fetch(endpoint, {
    method: "POST",
    body: JSON.stringify({
      email,
    }),
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
  });

  if (!response.ok) {
    const body = await response.json();
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }
}

export async function update(updates: UpdateInput) {
  const endpoint = API_URL + "/users";
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await authenticatedFetch(endpoint, {
    method: "PATCH",
    body: JSON.stringify(updates),
    headers: {
      "Content-Type": "application/json",
      Authorization: authorization,
    },
    credentials: "include",
  });

  const body = await response.json();

  if (!response.ok) {
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }

  return body;
}

export async function getProfile(): Promise<UserProfile> {
  const endpoint = API_URL + "/users";
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await authenticatedFetch(endpoint, {
    method: "GET",
    headers: {
      Authorization: authorization,
    },
  });

  const body = await response.json();
  if (!response.ok) {
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }

  offlineManager.setUserPremiumStatus(body.subscriptionPlan !== "free");

  return body;
}

export async function deleteProfile(): Promise<void> {
  const endpoint = API_URL + "/users";
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await authenticatedFetch(endpoint, {
    method: "DELETE",
    headers: {
      Authorization: authorization,
    },
  });

  if (!response.ok) {
    const body = await response.json();
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }

  Sentry.setUser(null);
}

function saveAuthorization(authorization: string) {
  // Save the authorization token
  if (Platform.OS === "web") {
    localStorage.setItem("authorization", authorization);
  } else if (Platform.OS === "ios" || Platform.OS === "android") {
    SecureStore.setItem("authorization", authorization);
  } else {
    throw new Error("Unknown platform: " + Platform.OS);
  }
  cachedAuthorization = authorization;
}

function saveSessionResponse(response: Response) {
  const authorization = response.headers.get("authorization");
  if (!authorization) {
    throw new Error(DEFAULT_ERROR);
  }
  saveAuthorization(authorization);
  const refreshToken = response.headers.get("x-refresh-token");
  if (Platform.OS === "ios" || Platform.OS === "android") {
    if (!refreshToken) {
      throw new Error(DEFAULT_ERROR);
    }
    SecureStore.setItem("refreshToken", refreshToken);
    cachedRefreshToken = refreshToken;
  }
}

function getRefreshToken() {
  if (Platform.OS === "web") return null;
  if (cachedRefreshToken === undefined) {
    cachedRefreshToken = SecureStore.getItem("refreshToken");
  }
  return cachedRefreshToken;
}

async function refreshSession(): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    const refreshToken = getRefreshToken();
    const response = await fetch(API_URL + "/session/refresh", {
      method: "POST",
      credentials: "include",
      headers: {
        ...authClientHeaders(),
        ...(refreshToken ? { "X-Refresh-Token": refreshToken } : {}),
      },
    });
    if (response.status === 401 || response.status === 403) {
      await clearLocalCredentials();
      return false;
    }
    if (!response.ok) return false;
    saveSessionResponse(response);
    return true;
  })()
    .catch(async () => {
      // Preserve credentials on transient network failures so offline users can
      // retry later; only an explicit server rejection invalidates the session.
      return false;
    })
    .finally(() => {
      refreshInFlight = null;
    });
  return refreshInFlight;
}

async function clearLocalCredentials() {
  try {
    if (Platform.OS === "web") {
      localStorage.removeItem("authorization");
    } else if (Platform.OS === "ios" || Platform.OS === "android") {
      await SecureStore.deleteItemAsync("authorization");
      await SecureStore.deleteItemAsync("refreshToken");
    } else {
      throw new Error("Unknown platform: " + Platform.OS);
    }
  } finally {
    cachedAuthorization = null;
    cachedRefreshToken = null;
    requestAnalyticsIdentityReset();
  }
}

export async function authenticatedFetch(
  input: RequestInfo | URL,
  init: RequestInit = {},
): Promise<Response> {
  const execute = () => {
    const headers = new Headers(init.headers);
    const authorization = getAuthorization();
    if (authorization) headers.set("Authorization", authorization);
    for (const [key, value] of Object.entries(authClientHeaders())) {
      headers.set(key, value);
    }
    return fetch(input, { ...init, headers, credentials: "include" });
  };
  let response = await execute();
  if (response.status === 401 && (await refreshSession())) {
    response = await execute();
  }
  return response;
}

export function getAuthorization() {
  if (cachedAuthorization === undefined) {
    if (Platform.OS === "web") {
      cachedAuthorization =
        typeof localStorage === "undefined"
          ? null
          : localStorage.getItem("authorization");
    } else if (Platform.OS === "ios" || Platform.OS === "android") {
      cachedAuthorization = SecureStore.getItem("authorization");
    } else {
      throw new Error("Unsupported platform: " + Platform.OS);
    }
  }
  return cachedAuthorization;
}
