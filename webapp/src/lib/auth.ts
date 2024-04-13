"use client";

import { ApplicationError, ValidationError } from "./common";

export async function authenticate(username: string, password: string) {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/login`, {
      method: "POST",
      body: JSON.stringify({ Email: username, Password: password }),
      headers: {
        "Content-Type": "application/json",
      },
    });
    if (!response.ok) {
      throw new ApplicationError(
        "Invalid username or password.",
        "AuthenticationError",
      );
    }
    const result = await response.json();
    localStorage.setItem("token", result.token);
  } catch (error) {
    if (error instanceof ApplicationError) {
      throw error;
    } else {
      console.error(error);
      throw new Error("Failed to authenticate.");
    }
  }
}

export function isAuthenticated(): boolean {
  const token = localStorage.getItem("token");
  if (!token) {
    return false;
  }

  const payloadBase64 = token.split(".")[1];
  const decodedJson = atob(payloadBase64);
  const decoded = JSON.parse(decodedJson);
  const exp = decoded.exp; // Get the expiration time

  // No expiration time, considered expired for safety
  if (!exp) {
    localStorage.removeItem("token");
    return false;
  }

  const now = Math.floor(Date.now() / 1000); // Current time in Unix timestamp
  return exp > now;
}

export function getAuthToken(): string {
  return localStorage.getItem("token") || "";
}

type UserRegistrationDTO = {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
};

export async function signup(dto: UserRegistrationDTO) {
  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/users`, {
    method: "POST",
    body: JSON.stringify({
      email: dto.email,
      password: dto.password,
      firstName: dto.firstName,
      lastName: dto.lastName,
    }),
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!response.ok) {
    const result = await response.json();
    if ("validationErrors" in result) {
      throw new ValidationError(result.validationErrors);
    } else {
      throw new ApplicationError(result.error, "SignUpError");
    }
  }

  const token = response.headers.get("Authorization");
  if (token) {
    localStorage.setItem("token", token);
  }
}
