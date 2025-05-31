import {
  AUTH_REQUIRED_ERROR,
  DEFAULT_ERROR,
  TOKEN_VALIDATION_ERROR,
} from "./constants";
import { getAuthorization, sigout } from "./users";
import { router } from "expo-router";

export async function callAPI<T>(
  method: "GET" | "POST" | "DELETE" | "PATCH" | "PUT",
  endpoint: string,
  body?: string,
): Promise<T> {
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const response = await fetch(endpoint, {
    method,
    headers: {
      authorization,
    },
    body,
    credentials: "include",
  });

  let responseBody: any;

  if (response.status != 204) {
    responseBody = await response.json();
  }
  if (!response.ok) {
    const message =
      responseBody?.error ||
      Object.values(responseBody?.validationErrors)?.[0] ||
      DEFAULT_ERROR;

    if (message == TOKEN_VALIDATION_ERROR) {
      await sigout();
      router.dismissAll();
      router.replace("/signin");
    } else {
      throw new Error(message);
    }
  }
  return responseBody;
}
