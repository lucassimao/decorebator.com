import * as SecureStore from "expo-secure-store";
import * as jwt from "./jwt";
import { Platform } from "react-native";
import { AUTH_REQUIRED_ERROR, DEFAULT_ERROR } from "./constants";

export type UserSignup = {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
};

export type UserSignin = {
  email: string;
  password: string;
};

export type UserInfo = {
  firstName: string;
  lastName: string;
  id: number;
  subscriptionPlan?: "free" | "monthly" | "annual";
};

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
    saveAuthorization(authorization);
  } else {
    throw new Error(DEFAULT_ERROR);
  }
}

export async function sigout() {
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
    saveAuthorization(authorization);
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

export function getUserInfo(): UserInfo | null {
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const decoded = jwt.decode(authorization);
  return {
    firstName: decoded.payload?.firstName,
    lastName: decoded.payload?.lastName,
    id: +decoded.payload?.sub,
    subscriptionPlan: decoded.payload?.subscriptionPlan || "free",
  };
}

function saveAuthorization(authorization: string) {
  if (Platform.OS === "web") {
    localStorage.setItem("authorization", authorization);
  } else if (Platform.OS === "ios" || Platform.OS === "android") {
    SecureStore.setItem("authorization", authorization);
  } else {
    throw new Error("Unknown platform: " + Platform.OS);
  }
}

export function getAuthorization() {
  if (Platform.OS === "web") {
    return localStorage.getItem("authorization");
  } else if (Platform.OS === "ios" || Platform.OS === "android") {
    return SecureStore.getItem("authorization");
  } else {
    throw new Error("Unsupported platform: " + Platform.OS);
  }
}

export async function refreshToken(): Promise<UserInfo> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/auth/refresh";
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      Authorization: authorization,
    },
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error("Failed to refresh token");
  }

  const data = await response.json();

  // Save the new token
  if (data.token) {
    saveAuthorization(data.token);
  }

  // Return updated user info
  return data.user as UserInfo;
}
