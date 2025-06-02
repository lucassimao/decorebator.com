import * as SecureStore from "expo-secure-store";
import * as jwt from "./jwt";
import { Platform } from "react-native";
import { AUTH_REQUIRED_ERROR, DEFAULT_ERROR } from "./constants";

type UserProfile = {
  id: number;
  email: string;
  firstName: string;
  lastName: string;
  profilePictureUrl?: string;
  country?: string;
  dateOfBirth?: string; // YYYY-MM-DD
  createdAt: string;
};
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

export type UpdateInput = {
  firstName?: string;
  lastName?: string;
  profilePicture?: string;
  profilePictureFileExtension?: string;
  country?: string;
  dateOfBirth?: string;

  // if set, triggers a password update
  updatePassword?: {
    currentPassword: string;
    newPassword: String;
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

export async function update(updates: UpdateInput) {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/users";
  const authorization = getAuthorization();

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
  const authorization = getAuthorization();

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

  return body;
}

export async function deleteProfile(): Promise<void> {
  const endpoint = process.env.EXPO_PUBLIC_API_URL + "/users";
  const authorization = getAuthorization();

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
