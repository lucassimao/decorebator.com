import * as SecureStore from "expo-secure-store";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { Platform } from "react-native";
import { DEFAULT_ERROR } from "./constants";
import offlineManager from "@/utils/offlineManager";
import * as Sentry from "@sentry/react-native";
import { getApiBaseUrl } from "./baseUrl";
import { requestAnalyticsIdentityReset } from "@/utils/activationEvents";
import {
  deactivateStoredPushToken,
  scheduleStoredPushTokenDeactivation,
} from "@/utils/pushTokenStorage";

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
const WEB_SESSION_MARKER = "hasAuthenticatedSession";
const WEB_SESSION_EVENT = "decorebator-session-change";
const PENDING_USER_CLEANUP_KEY = "pendingDurableUserCleanup";
const NATIVE_SESSION_CREDENTIALS_KEY = "sessionCredentials";
let cachedAuthorization: string | null | undefined;
let cachedRefreshToken: string | null | undefined;
type RefreshResult = "refreshed" | "rejected" | "stale";
let refreshInFlight: Promise<RefreshResult> | null = null;
let legacyCookieCleanupInFlight: Promise<void> | null = null;
let legacyCookieCleanupRetryTimer: ReturnType<typeof setTimeout> | null = null;
let nextLegacyCookieCleanupAttempt = 0;
let legacyCookieCleanupComplete = false;
let durableCleanupInFlight: Promise<void> | null = null;
let sessionGeneration = 0;
let nativeSessionEpoch = 0;
const nativeSessionListeners = new Set<(epoch: string | null) => void>();
let sessionCacheGeneration = 0;
const inFlightSessionCacheWrites = new Set<Promise<void>>();

function authClientHeaders(): Record<string, string> {
  return Platform.OS === "ios" || Platform.OS === "android"
    ? { "X-Auth-Client": "native" }
    : {};
}

function boundedAbortSignal(
  external: AbortSignal | undefined,
  timeoutMs: number,
) {
  const controller = new AbortController();
  const abort = () => controller.abort();
  if (external?.aborted) controller.abort();
  else external?.addEventListener("abort", abort, { once: true });
  const timeout = setTimeout(() => {
    abort();
    external?.removeEventListener("abort", abort);
  }, timeoutMs);
  return {
    signal: controller.signal,
    cleanup: () => {
      clearTimeout(timeout);
      external?.removeEventListener("abort", abort);
    },
  };
}

async function withBrowserSessionLock<T>(
  operation: () => Promise<T>,
  signal?: AbortSignal,
): Promise<T> {
  if (Platform.OS !== "web") return operation();
  if (typeof navigator === "undefined" || !navigator.locks) {
    throw new Error("Browser session coordination unavailable");
  }
  const bounded = boundedAbortSignal(signal, 15_000);
  try {
    return await navigator.locks.request(
      "decorebator-session-refresh",
      { signal: bounded.signal },
      operation,
    );
  } finally {
    bounded.cleanup();
  }
}

function ensureLegacyCookieCleanup() {
  if (
    legacyCookieCleanupComplete ||
    legacyCookieCleanupInFlight ||
    Date.now() < nextLegacyCookieCleanupAttempt
  ) {
    return;
  }
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15_000);
  legacyCookieCleanupInFlight = fetch(
    API_URL + "/session/legacy-cookie-cleanup",
    {
      method: "POST",
      credentials: "include",
      signal: controller.signal,
    },
  )
    .then((response) => {
      if (!response.ok) throw new Error("Legacy cookie cleanup unavailable");
      if (legacyCookieCleanupRetryTimer) {
        clearTimeout(legacyCookieCleanupRetryTimer);
        legacyCookieCleanupRetryTimer = null;
      }
      legacyCookieCleanupComplete = true;
      nextLegacyCookieCleanupAttempt = Number.POSITIVE_INFINITY;
      notifyWebSessionChange();
    })
    .catch(() => {
      nextLegacyCookieCleanupAttempt = Date.now() + 60_000;
      if (!legacyCookieCleanupRetryTimer) {
        legacyCookieCleanupRetryTimer = setTimeout(() => {
          legacyCookieCleanupRetryTimer = null;
          ensureLegacyCookieCleanup();
        }, 60_000);
      }
    })
    .finally(() => {
      clearTimeout(timeout);
      legacyCookieCleanupInFlight = null;
    });
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
}

export async function sigout() {
  return withBrowserSessionLock(async () => {
    sessionGeneration += 1;
    await markDurableCleanupPending();
    const refreshToken = getRefreshToken();
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 15_000);
    let response: Response;
    try {
      response = await fetch(API_URL + "/logout", {
        method: "POST",
        credentials: "include",
        headers: {
          ...authClientHeaders(),
          ...(refreshToken ? { "X-Refresh-Token": refreshToken } : {}),
        },
        signal: controller.signal,
      });
    } finally {
      clearTimeout(timeout);
    }
    if (!response.ok) {
      if (isDefinitiveClientRejection(response)) {
        await clearDurableCleanupMarker();
      }
      throw new Error("Logout temporarily unavailable");
    }
    Sentry.setUser(null);
    await clearDurableUserState(true);
  });
}
export async function signin(data: UserSignin) {
  return withBrowserSessionLock(async () => {
    await recoverPendingDurableCleanup(true);
    // A sign-in route can be reached directly while another identity is
    // active. Remove that identity before establishing its replacement.
    await clearDurableUserState(true);
    const generation = ++sessionGeneration;
    const endpoint = API_URL + "/login";
    const bounded = boundedAbortSignal(undefined, 15_000);
    let response: Response;
    try {
      response = await fetch(endpoint, {
        method: "POST",
        body: JSON.stringify(data),
        headers: {
          "Content-Type": "application/json",
          ...authClientHeaders(),
        },
        credentials: "include",
        signal: bounded.signal,
      });
    } finally {
      bounded.cleanup();
    }
    if (!response.ok) {
      throw new Error(SIGN_IN_ERROR);
    }
    if (!saveSessionResponse(response, generation)) {
      throw new Error("Session changed during sign in");
    }
  });
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

  if (!hasAuthenticationSession()) {
    throw new Error("Authentication required");
  }

  if (updates.updatePassword) {
    sessionGeneration += 1;
    await markDurableCleanupPending();
  }
  const bounded = boundedAbortSignal(undefined, 15_000);
  let response: Response;
  let body: any;
  try {
    response = await authenticatedFetch(endpoint, {
      method: "PATCH",
      body: JSON.stringify(updates),
      headers: {
        "Content-Type": "application/json",
        ...authenticationHeaders(),
      },
      credentials: "include",
      signal: bounded.signal,
    });
    body = await response.json();
  } finally {
    bounded.cleanup();
  }

  if (!response.ok) {
    if (updates.updatePassword && isDefinitiveClientRejection(response)) {
      await clearDurableCleanupMarker();
    }
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

  if (!hasAuthenticationSession()) {
    throw new Error("Authentication required");
  }

  const response = await authenticatedFetch(endpoint, {
    method: "GET",
    headers: authenticationHeaders(),
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

  if (!hasAuthenticationSession()) {
    throw new Error("Authentication required");
  }

  sessionGeneration += 1;
  await markDurableCleanupPending();
  const bounded = boundedAbortSignal(undefined, 15_000);
  let response: Response;
  try {
    response = await authenticatedFetch(endpoint, {
      method: "DELETE",
      headers: authenticationHeaders(),
      signal: bounded.signal,
    });
  } finally {
    bounded.cleanup();
  }

  if (!response.ok) {
    if (isDefinitiveClientRejection(response)) {
      await clearDurableCleanupMarker();
    }
    const body = await response.json();
    const message =
      body?.error ||
      Object.values(body?.validationErrors)?.[0] ||
      DEFAULT_ERROR;
    throw new Error(message);
  }

  Sentry.setUser(null);
  await clearDurableUserState();
}

function saveSessionResponse(
  response: Response,
  expectedGeneration = sessionGeneration,
  identityChanged = true,
): boolean {
  if (expectedGeneration !== sessionGeneration) return false;
  if (Platform.OS === "web") {
    if (typeof localStorage !== "undefined") {
      localStorage.removeItem("authorization");
      const currentEpoch = localStorage.getItem(WEB_SESSION_MARKER);
      localStorage.setItem(
        WEB_SESSION_MARKER,
        currentEpoch ??
          (typeof crypto !== "undefined" && crypto.randomUUID
            ? crypto.randomUUID()
            : `${Date.now()}-${Math.random()}`),
      );
      notifyWebSessionChange();
    }
    cachedAuthorization = null;
    return true;
  }
  const authorization = response.headers.get("authorization");
  const refreshToken = response.headers.get("x-refresh-token");
  if (Platform.OS === "ios" || Platform.OS === "android") {
    if (!authorization || !refreshToken) {
      throw new Error(DEFAULT_ERROR);
    }
    SecureStore.setItem(
      NATIVE_SESSION_CREDENTIALS_KEY,
      JSON.stringify({ authorization, refreshToken }),
    );
    cachedAuthorization = authorization;
    cachedRefreshToken = refreshToken;
    if (identityChanged) {
      nativeSessionEpoch += 1;
      notifyNativeSessionChange();
    }
    return true;
  }
  return false;
}

function loadNativeCredentials() {
  const storedPair = SecureStore.getItem(NATIVE_SESSION_CREDENTIALS_KEY);
  if (storedPair) {
    try {
      const parsed = JSON.parse(storedPair) as {
        authorization?: unknown;
        refreshToken?: unknown;
      };
      if (
        typeof parsed.authorization === "string" &&
        typeof parsed.refreshToken === "string"
      ) {
        cachedAuthorization = parsed.authorization;
        cachedRefreshToken = parsed.refreshToken;
        return;
      }
    } catch {
      // Malformed consolidated credentials are not safe to recover.
    }
  }
  const hasLegacyCredentials =
    SecureStore.getItem("authorization") !== null ||
    SecureStore.getItem("refreshToken") !== null;
  cachedAuthorization = null;
  cachedRefreshToken = null;
  if (storedPair !== null || hasLegacyCredentials) {
    // Split legacy keys can contain tokens from different identities after a
    // crash. Fail closed and let the durable cleanup barrier purge them.
    SecureStore.setItem(PENDING_USER_CLEANUP_KEY, "true");
    retryPendingDurableCleanup();
  }
}

function isDefinitiveClientRejection(response: Response) {
  return (
    response.status >= 400 && response.status < 500 && response.status !== 408
  );
}

function guardResponseIdentity(
  response: Response,
  identityIsCurrent: () => boolean,
  releaseRequestTimeout: () => void,
): Response {
  const assertCurrent = () => {
    if (!identityIsCurrent()) {
      throw new Error("Session changed during request");
    }
  };
  let guardedHeaders: Headers | undefined;
  let guardedBody: ReadableStream<Uint8Array> | null | undefined;
  return new Proxy(response, {
    get(target, property) {
      assertCurrent();
      if (property === "clone") {
        return () => {
          assertCurrent();
          throw new Error("Cloning authenticated responses is unsupported");
        };
      }
      if (property === "headers") {
        guardedHeaders ??= new Proxy(target.headers, {
          get(headers, headerProperty) {
            assertCurrent();
            const value = Reflect.get(headers, headerProperty, headers);
            if (typeof value !== "function") return value;
            return (...args: unknown[]) => {
              assertCurrent();
              if (headerProperty === "forEach") {
                const callback = args[0] as (
                  value: string,
                  key: string,
                  parent: Headers,
                ) => void;
                const thisArg = args[1];
                return headers.forEach((headerValue, key) => {
                  assertCurrent();
                  callback.call(thisArg, headerValue, key, guardedHeaders!);
                });
              }
              const result = value.apply(headers, args);
              if (
                headerProperty === "entries" ||
                headerProperty === "keys" ||
                headerProperty === "values" ||
                headerProperty === Symbol.iterator
              ) {
                const iterator = result as Iterator<unknown>;
                return {
                  next() {
                    assertCurrent();
                    return iterator.next();
                  },
                  [Symbol.iterator]() {
                    return this;
                  },
                };
              }
              return result;
            };
          },
        });
        return guardedHeaders;
      }
      if (property === "body") {
        if (guardedBody !== undefined) return guardedBody;
        const body = target.body;
        if (!body) {
          guardedBody = null;
          return null;
        }
        const reader = body.getReader();
        const stream = new ReadableStream<Uint8Array>(
          {
            async pull(controller) {
              assertCurrent();
              const chunk = await reader.read();
              assertCurrent();
              if (chunk.done) {
                releaseRequestTimeout();
                controller.close();
              } else controller.enqueue(chunk.value);
            },
            cancel(reason) {
              assertCurrent();
              releaseRequestTimeout();
              return reader.cancel(reason);
            },
          },
          { highWaterMark: 0 },
        );
        guardedBody = new Proxy(stream, {
          get(bodyStream, streamProperty) {
            assertCurrent();
            if (streamProperty === "getReader") {
              return () => {
                assertCurrent();
                const guardedReader = bodyStream.getReader();
                return new Proxy(guardedReader, {
                  get(streamReader, readerProperty) {
                    if (readerProperty === "read") {
                      return async () => {
                        assertCurrent();
                        const chunk = await streamReader.read();
                        assertCurrent();
                        return chunk;
                      };
                    }
                    assertCurrent();
                    const value = Reflect.get(
                      streamReader,
                      readerProperty,
                      streamReader,
                    );
                    if (typeof value !== "function") return value;
                    return (...readerArgs: unknown[]) => {
                      assertCurrent();
                      return value.apply(streamReader, readerArgs);
                    };
                  },
                });
              };
            }
            if (
              streamProperty === "pipeThrough" ||
              streamProperty === "pipeTo" ||
              streamProperty === "tee" ||
              streamProperty === "values" ||
              streamProperty === Symbol.asyncIterator
            ) {
              return () => {
                throw new Error(
                  "Streaming authenticated responses requires getReader()",
                );
              };
            }
            const value = Reflect.get(bodyStream, streamProperty, bodyStream);
            if (typeof value !== "function") return value;
            return (...args: unknown[]) => {
              assertCurrent();
              return value.apply(bodyStream, args);
            };
          },
        });
        return guardedBody;
      }
      if (
        property === "json" ||
        property === "text" ||
        property === "arrayBuffer" ||
        property === "blob" ||
        property === "formData" ||
        property === "bytes"
      ) {
        return async (...args: unknown[]) => {
          assertCurrent();
          const reader = Reflect.get(target, property, target) as (
            ...readerArgs: unknown[]
          ) => Promise<unknown>;
          try {
            const value = await reader.apply(target, args);
            assertCurrent();
            return value;
          } finally {
            releaseRequestTimeout();
          }
        };
      }
      return Reflect.get(target, property, target);
    },
  });
}

function getRefreshToken() {
  if (Platform.OS === "web") return null;
  if (cachedRefreshToken === undefined) {
    loadNativeCredentials();
  }
  return cachedRefreshToken;
}

async function refreshSession(
  signal?: AbortSignal,
  browserLockHeld = false,
): Promise<RefreshResult> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    const generation = sessionGeneration;
    const bounded = boundedAbortSignal(signal, 15_000);
    try {
      const refreshToken = getRefreshToken();
      const response = await fetch(API_URL + "/session/refresh", {
        method: "POST",
        credentials: "include",
        headers: {
          ...authClientHeaders(),
          ...(refreshToken ? { "X-Refresh-Token": refreshToken } : {}),
        },
        signal: bounded.signal,
      });
      if (response.status === 401) {
        if (generation !== sessionGeneration) return "stale";
        await clearDurableUserState(browserLockHeld);
        return "rejected";
      }
      if (!response.ok) {
        throw new Error(
          `Session refresh temporarily unavailable (${response.status})`,
        );
      }
      return saveSessionResponse(response, generation, false)
        ? "refreshed"
        : "stale";
    } finally {
      bounded.cleanup();
    }
  })().finally(() => {
    refreshInFlight = null;
  });
  return refreshInFlight;
}

async function clearLocalCredentials() {
  try {
    if (Platform.OS === "web") {
      localStorage.removeItem("authorization");
      localStorage.removeItem(WEB_SESSION_MARKER);
      notifyWebSessionChange();
    } else if (Platform.OS === "ios" || Platform.OS === "android") {
      const results = await Promise.allSettled([
        SecureStore.deleteItemAsync(NATIVE_SESSION_CREDENTIALS_KEY),
        SecureStore.deleteItemAsync("authorization"),
        SecureStore.deleteItemAsync("refreshToken"),
      ]);
      const failures: unknown[] = [];
      for (const result of results) {
        if (result.status === "rejected") {
          failures.push(result.reason);
          console.error(
            "Error deleting secure session credential:",
            result.reason,
          );
        }
      }
      if (failures.length > 0) {
        throw new AggregateError(failures, "Secure session cleanup incomplete");
      }
    } else {
      throw new Error("Unknown platform: " + Platform.OS);
    }
  } finally {
    cachedAuthorization = null;
    cachedRefreshToken = null;
    if (Platform.OS === "ios" || Platform.OS === "android") {
      nativeSessionEpoch += 1;
      notifyNativeSessionChange();
    }
    requestAnalyticsIdentityReset();
  }
}

async function clearDurableUserState(browserLockHeld = false): Promise<void> {
  if (Platform.OS === "web" && !browserLockHeld) {
    return withBrowserSessionLock(() => clearDurableUserState(true));
  }
  sessionGeneration += 1;
  if (durableCleanupInFlight) return durableCleanupInFlight;
  durableCleanupInFlight = (async () => {
    await markDurableCleanupPending();
    Sentry.setUser(null);
    let pushFailure: unknown;
    try {
      await deactivateStoredPushToken();
    } catch (error) {
      pushFailure = error;
      scheduleStoredPushTokenDeactivation();
      console.error("Error clearing durable user state:", error);
    }
    const results = await Promise.allSettled([
      clearSessionScopedCaches(),
      clearLocalCredentials(),
    ]);
    const failures = results.filter((result) => result.status === "rejected");
    if (pushFailure !== undefined) {
      failures.push({ status: "rejected", reason: pushFailure });
    }
    for (const failure of failures) {
      console.error("Error clearing durable user state:", failure.reason);
    }
    if (failures.length === 0) {
      await clearDurableCleanupMarker();
      return;
    }
    throw new AggregateError(
      failures.map((failure) => failure.reason),
      "Durable user cleanup incomplete",
    );
  })().finally(() => {
    durableCleanupInFlight = null;
  });
  return durableCleanupInFlight;
}

async function markDurableCleanupPending() {
  if (Platform.OS === "web") {
    localStorage.setItem(PENDING_USER_CLEANUP_KEY, "true");
    notifyWebSessionChange();
    return;
  }
  SecureStore.setItem(PENDING_USER_CLEANUP_KEY, "true");
  notifyNativeSessionChange();
}

function hasPendingDurableCleanup() {
  if (Platform.OS === "web") {
    return (
      typeof localStorage !== "undefined" &&
      localStorage.getItem(PENDING_USER_CLEANUP_KEY) === "true"
    );
  }
  return SecureStore.getItem(PENDING_USER_CLEANUP_KEY) === "true";
}

async function clearDurableCleanupMarker() {
  if (Platform.OS === "web") {
    localStorage.removeItem(PENDING_USER_CLEANUP_KEY);
    notifyWebSessionChange();
    return;
  }
  await SecureStore.deleteItemAsync(PENDING_USER_CLEANUP_KEY);
  notifyNativeSessionChange();
}

async function recoverPendingDurableCleanup(browserLockHeld = false) {
  if (hasPendingDurableCleanup()) {
    await clearDurableUserState(browserLockHeld);
  }
}

function retryPendingDurableCleanup() {
  recoverPendingDurableCleanup().catch((error) =>
    console.error("Error retrying durable cleanup:", error),
  );
}

async function clearUserScopedStorage() {
  const allKeys = await AsyncStorage.getAllKeys();
  const userKeys = allKeys.filter(
    (key) =>
      key === "cachedUserProfile" ||
      key === "cachedSubscription" ||
      key === "hasSeenDashboard" ||
      key === "justSignedUp" ||
      key === "pushPromptedAfterFirstQuiz" ||
      key.startsWith("flashcard_position_") ||
      key.startsWith("flashcard_save_position_") ||
      key.startsWith("wordlistFabHintSeen_"),
  );
  if (userKeys.length > 0) {
    await AsyncStorage.multiRemove(userKeys);
  }
}

export function persistSessionSnapshot(
  user: UserProfile,
  subscription: unknown,
  epoch: string,
): Promise<void> {
  const generation = sessionCacheGeneration;
  const write = (async () => {
    if (getAuthenticationSessionEpoch() !== epoch) return;
    await AsyncStorage.multiSet([
      ["cachedUserProfile", JSON.stringify(user)],
      ["cachedSubscription", JSON.stringify(subscription)],
    ]);
    if (
      generation !== sessionCacheGeneration ||
      getAuthenticationSessionEpoch() !== epoch
    ) {
      throw new Error("Session changed during cache persistence");
    }
  })().finally(() => inFlightSessionCacheWrites.delete(write));
  inFlightSessionCacheWrites.add(write);
  return write;
}

async function drainSessionCacheWriters() {
  sessionCacheGeneration += 1;
  const pendingWrites = Array.from(inFlightSessionCacheWrites);
  if (pendingWrites.length > 0) await Promise.allSettled(pendingWrites);
}

export async function clearSessionScopedCaches(): Promise<void> {
  await drainSessionCacheWriters();
  offlineManager.setUserPremiumStatus(false);
  const results = await Promise.allSettled([
    offlineManager.clearCache(),
    clearUserScopedStorage(),
  ]);
  const failures = results.filter((result) => result.status === "rejected");
  if (failures.length > 0) {
    throw new AggregateError(
      failures.map((failure) => failure.reason),
      "Session cache cleanup incomplete",
    );
  }
}

export async function clearSessionCredentials() {
  await clearDurableUserState();
}

export async function authenticatedFetch(
  input: RequestInfo | URL,
  init: RequestInit = {},
): Promise<Response> {
  const bounded = boundedAbortSignal(init.signal ?? undefined, 15_000);
  const signal = bounded.signal;
  const requestGeneration = sessionGeneration;
  const requestCleanupPending = hasPendingDurableCleanup();
  const browserRequestEpoch =
    Platform.OS === "web" ? getAuthenticationSessionEpoch() : null;
  const requestIdentityIsCurrent = () =>
    Platform.OS === "web"
      ? getAuthenticationSessionEpoch() === browserRequestEpoch &&
        hasPendingDurableCleanup() === requestCleanupPending
      : sessionGeneration === requestGeneration;
  const execute = () => {
    if (!requestIdentityIsCurrent()) {
      throw new Error("Session changed during request");
    }
    const headers = new Headers(init.headers);
    const authorization = getAuthorization();
    if (authorization) headers.set("Authorization", authorization);
    for (const [key, value] of Object.entries(authClientHeaders())) {
      headers.set(key, value);
    }
    return fetch(input, {
      ...init,
      headers,
      credentials: "include",
      signal,
    }).then((nextResponse) => {
      if (!requestIdentityIsCurrent()) {
        throw new Error("Session changed during request");
      }
      const guardedResponse = guardResponseIdentity(
        nextResponse,
        requestIdentityIsCurrent,
        bounded.cleanup,
      );
      if (nextResponse.body == null) bounded.cleanup();
      return guardedResponse;
    });
  };
  const performRequest = async (browserLockHeld: boolean) => {
    let response = await execute();
    if (response.status !== 401) {
      if (!requestIdentityIsCurrent()) {
        throw new Error("Session changed during request");
      }
      return response;
    }
    if (!requestIdentityIsCurrent()) {
      throw new Error("Session changed during request");
    }
    const refreshResult = await refreshSession(signal, browserLockHeld);
    if (refreshResult === "stale") {
      throw new Error("Session changed during request");
    }
    if (refreshResult === "rejected") {
      throw new Error("Authentication session expired");
    }
    if (refreshResult === "refreshed") {
      response = await execute();
    }
    return response;
  };
  try {
    if (Platform.OS === "web") {
      const response = await withBrowserSessionLock(
        () => performRequest(true),
        signal,
      );
      return response;
    }
    const response = await performRequest(false);
    return response;
  } catch (error) {
    bounded.cleanup();
    throw error;
  }
}

export function getAuthorization() {
  if (cachedAuthorization === undefined) {
    if (Platform.OS === "web") {
      if (typeof localStorage !== "undefined") {
        localStorage.removeItem("authorization");
      }
      cachedAuthorization = null;
    } else if (Platform.OS === "ios" || Platform.OS === "android") {
      loadNativeCredentials();
    } else {
      throw new Error("Unsupported platform: " + Platform.OS);
    }
  }
  return cachedAuthorization;
}

export function hasAuthenticationSession() {
  if (hasPendingDurableCleanup()) {
    retryPendingDurableCleanup();
    return false;
  }
  if (Platform.OS === "web") {
    if (typeof localStorage !== "undefined") {
      localStorage.removeItem("authorization");
    }
    ensureLegacyCookieCleanup();
    if (!legacyCookieCleanupComplete) return false;
    return (
      typeof localStorage !== "undefined" &&
      localStorage.getItem(WEB_SESSION_MARKER) !== null
    );
  }
  const hasSession = Boolean(getAuthorization());
  if (!hasSession) {
    scheduleStoredPushTokenDeactivation();
  }
  return hasSession;
}

export function isAuthenticationSessionReady(): boolean {
  if (hasPendingDurableCleanup()) return false;
  if (Platform.OS !== "web") return true;
  ensureLegacyCookieCleanup();
  return legacyCookieCleanupComplete;
}

export function getAuthenticationSessionEpoch(): string | null {
  if (hasPendingDurableCleanup()) return null;
  if (Platform.OS === "web") {
    ensureLegacyCookieCleanup();
    if (!legacyCookieCleanupComplete) return null;
    return typeof localStorage === "undefined"
      ? null
      : localStorage.getItem(WEB_SESSION_MARKER);
  }
  return getAuthorization() ? `native:${nativeSessionEpoch}` : null;
}

function notifyNativeSessionChange() {
  const epoch = getAuthenticationSessionEpoch();
  for (const listener of nativeSessionListeners) listener(epoch);
}

function notifyWebSessionChange() {
  if (
    Platform.OS === "web" &&
    typeof window !== "undefined" &&
    typeof window.dispatchEvent === "function"
  ) {
    window.dispatchEvent(new Event(WEB_SESSION_EVENT));
  }
}

export function subscribeAuthenticationSessionChanges(
  listener: (epoch: string | null) => void,
): () => void {
  if (Platform.OS === "ios" || Platform.OS === "android") {
    nativeSessionListeners.add(listener);
    return () => nativeSessionListeners.delete(listener);
  }
  if (
    Platform.OS !== "web" ||
    typeof window === "undefined" ||
    typeof window.addEventListener !== "function"
  )
    return () => {};
  const notify = () => listener(getAuthenticationSessionEpoch());
  const onStorage = (event: StorageEvent) => {
    if (
      event.key === WEB_SESSION_MARKER ||
      event.key === PENDING_USER_CLEANUP_KEY
    )
      notify();
  };
  window.addEventListener("storage", onStorage);
  window.addEventListener(WEB_SESSION_EVENT, notify);
  return () => {
    window.removeEventListener("storage", onStorage);
    window.removeEventListener(WEB_SESSION_EVENT, notify);
  };
}

export function authenticationHeaders(): Record<string, string> {
  const authorization = getAuthorization();
  return authorization ? { Authorization: authorization } : {};
}
