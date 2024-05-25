import * as SecureStore from "expo-secure-store";
import * as jwt from "./jwt";

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
};

export const AUTH_REQUIRED_ERROR = "Authentication required.";
export const DEFAULT_ERROR = "Could not process your request.";
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
    await SecureStore.setItemAsync("authorization", authorization);
  } else {
    throw new Error(DEFAULT_ERROR);
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
  });
  if (!response.ok) {
    throw new Error(SIGN_IN_ERROR);
  }
  const authorization = response.headers.get("authorization");
  if (authorization) {
    await SecureStore.setItemAsync("authorization", authorization);
  } else {
    throw new Error(DEFAULT_ERROR);
  }
}

export async function getUserInfo(): Promise<UserInfo | null> {
  const authorization = await SecureStore.getItemAsync("authorization");

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const decoded = jwt.decode(authorization);
  return {
    firstName: decoded.payload?.firstName,
    lastName: decoded.payload?.lastName,
    id: +decoded.payload?.sub,
  };
}
