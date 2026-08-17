import {
  AUTH_REQUIRED_ERROR,
  DEFAULT_ERROR,
  TOKEN_VALIDATION_ERROR,
} from "./constants";
import {
  authenticatedFetch,
  authenticationHeaders,
  clearSessionCredentials,
  hasAuthenticationSession,
} from "./users";
import { router } from "expo-router";

export type APIResponseWithMetadata<T> = {
  body: T;
  nextCursor: string | null;
  definitionsContinuation?: string | null;
};

export async function callAPIWithMetadata<T>(
  method: "GET" | "POST" | "DELETE" | "PATCH" | "PUT",
  endpoint: string,
  body?: string,
  timeoutMs = 15000, // 15 second default timeout
): Promise<APIResponseWithMetadata<T>> {
  if (!hasAuthenticationSession()) {
    throw new Error(AUTH_REQUIRED_ERROR);
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);
  let response: Response;
  let responseBody: any;
  try {
    response = await authenticatedFetch(endpoint, {
      method,
      headers: {
        ...authenticationHeaders(),
      },
      body,
      credentials: "include",
      signal: controller.signal,
    });
    if (response.status !== 204) {
      responseBody = await response.json();
    }
  } catch (error) {
    if (controller.signal.aborted) {
      throw new Error(`Request timeout after ${timeoutMs}ms`);
    }
    throw error;
  } finally {
    clearTimeout(timeout);
  }
  if (!response.ok) {
    const message =
      responseBody?.error ||
      Object.values(responseBody?.validationErrors)?.[0] ||
      DEFAULT_ERROR;

    if (message === TOKEN_VALIDATION_ERROR) {
      await clearSessionCredentials();
      router.dismissAll();
      router.replace("/signin");
    } else {
      const error: any = new Error(message);
      error.status = response.status;
      error.data = responseBody;
      throw error;
    }
  }
  const nextCursor = response.headers.get("X-Next-Cursor")?.trim() || null;
  const definitionsContinuation =
    response.headers.get("X-Definitions-Continuation")?.trim() || null;
  return { body: responseBody, nextCursor, definitionsContinuation };
}

export async function callAPI<T>(
  method: "GET" | "POST" | "DELETE" | "PATCH" | "PUT",
  endpoint: string,
  body?: string,
  timeoutMs = 15000,
): Promise<T> {
  const response = await callAPIWithMetadata<T>(
    method,
    endpoint,
    body,
    timeoutMs,
  );
  return response.body;
}
