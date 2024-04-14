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
      credentials: "include",
    });
    if (!response.ok) {
      throw new ApplicationError(
        "Invalid username or password.",
        "AuthenticationError",
      );
    }
    localStorage.setItem("authenticated", "true");
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
  return localStorage.getItem("authenticated") == "true";
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
    credentials: "include",
  });
  if (!response.ok) {
    const result = await response.json();
    if ("validationErrors" in result) {
      throw new ValidationError(result.validationErrors);
    } else {
      throw new ApplicationError(result.error, "SignUpError");
    }
  }
}

export async function logout() {
  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/logout`, {
    credentials: "include",
  });
  if (!response.ok) {
    throw new Error("Failed to logout user");
  }
}
