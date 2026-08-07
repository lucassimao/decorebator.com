import {
  AUTH_REQUIRED_ERROR,
  DEFAULT_ERROR,
  TOKEN_VALIDATION_ERROR,
} from "./constants";
import { authenticatedFetch, getAuthorization, sigout } from "./users";
import { router } from "expo-router";

export async function callAPI<T>(
  method: "GET" | "POST" | "DELETE" | "PATCH" | "PUT",
  endpoint: string,
  body?: string,
  timeoutMs = 15000, // 15 second default timeout
): Promise<T> {
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  // Create timeout promise
  const timeoutPromise = new Promise<never>((_, reject) => {
    setTimeout(() => {
      reject(new Error(`Request timeout after ${timeoutMs}ms`));
    }, timeoutMs);
  });

  // Race between fetch and timeout
  const response = await Promise.race([
    authenticatedFetch(endpoint, {
      method,
      headers: {
        authorization,
      },
      body,
      credentials: "include",
    }),
    timeoutPromise,
  ]);

  let responseBody: any;

  if (response.status !== 204) {
    responseBody = await response.json();
  }
  if (!response.ok) {
    const message =
      responseBody?.error ||
      Object.values(responseBody?.validationErrors)?.[0] ||
      DEFAULT_ERROR;

    if (message === TOKEN_VALIDATION_ERROR) {
      await sigout();
      router.dismissAll();
      router.replace("/signin");
    } else {
      const error: any = new Error(message);
      error.status = response.status;
      error.data = responseBody;
      throw error;
    }
  }
  return responseBody;
}
