import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";
import { DEFAULT_ERROR } from "./constants";
import { decode } from "./jwt";
import offlineManager from "@/utils/offlineManager";
import * as Sentry from "@sentry/react-native";

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

export async function signup(data: UserSignup) {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/users";

  const response = await fetch(endpoint, {
    method: "POST",
    body: JSON.stringify(data),
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
  const authorization = response.headers.get("authorization");
  if (authorization) {
    await saveAuthorization(authorization);
  } else {
    throw new Error(DEFAULT_ERROR);
  }
}

export async function sigout() {
  // Clear offline cache when signing out
  try {
    await offlineManager.clearCache();
  } catch (error) {
    console.error("Error clearing offline cache:", error);
  }

  Sentry.setUser(null);

  if (Platform.OS === "web") {
    localStorage.removeItem("authorization");
  } else if (Platform.OS === "ios" || Platform.OS === "android") {
    await SecureStore.deleteItemAsync("authorization");
  } else {
    throw new Error("Unknown platform: " + Platform.OS);
  }
}
export async function signin(data: UserSignin) {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/login";

  const response = await fetch(endpoint, {
    method: "POST",
    body: JSON.stringify(data),
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
  });
  if (!response.ok) {
    throw new Error(SIGN_IN_ERROR);
  }
  const authorization = response.headers.get("authorization");
  if (authorization) {
    await saveAuthorization(authorization);
  } else {
    throw new Error(DEFAULT_ERROR);
  }
}

export async function requestResetEmailPassword(email: string) {
  const endpoint =
    process.env.EXPO_PUBLIC_API_URL + "/password/send-reset-email";

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
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/users";
  const authorization = await getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
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
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/users";
  const authorization = await getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
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

  // Check for JWT refresh (same pattern as login/signup)
  const newAuthorization = response.headers.get("authorization");
  if (newAuthorization) {
    await saveAuthorization(newAuthorization);
  }

  return body;
}

export async function deleteProfile(): Promise<void> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/users";
  const authorization = await getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
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

async function saveAuthorization(authorization: string) {
  // Save the authorization token
  if (Platform.OS === "web") {
    localStorage.setItem("authorization", authorization);
  } else if (Platform.OS === "ios" || Platform.OS === "android") {
    await SecureStore.setItemAsync("authorization", authorization);
  } else {
    throw new Error("Unknown platform: " + Platform.OS);
  }

  // Update offline manager with premium status from JWT
  try {
    const decoded = decode(authorization);
    const isPremium =
      decoded.payload.subscriptionPlan === "monthly" ||
      decoded.payload.subscriptionPlan === "annual";
    offlineManager.setUserPremiumStatus(isPremium);
  } catch (error) {
    console.error("Error updating offline manager:", error);
  }
}

export async function getAuthorization() {
  if (Platform.OS === "web") {
    return localStorage.getItem("authorization");
  } else if (Platform.OS === "ios" || Platform.OS === "android") {
    return SecureStore.getItemAsync("authorization");
  } else {
    throw new Error("Unsupported platform: " + Platform.OS);
  }
}
