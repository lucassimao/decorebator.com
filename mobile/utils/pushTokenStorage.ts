import AsyncStorage from "@react-native-async-storage/async-storage";
import { getApiBaseUrl } from "@/api/baseUrl";

const PUSH_TOKEN_STORAGE_KEY = "expoPushToken";
const PENDING_DEACTIVATION_KEY = "pendingExpoPushTokenDeactivation";
const DEACTIVATION_TIMEOUT_MS = 10_000;
const REGISTRATION_TIMEOUT_MS = 10_000;
let operationQueue: Promise<void> = Promise.resolve();
let retryTimer: ReturnType<typeof setTimeout> | null = null;
let retryAttempt = 0;
let registrationGeneration = 0;
let scheduledCleanupInFlight = false;

function serialize<T>(operation: () => Promise<T>): Promise<T> {
  const result = operationQueue.then(operation, operation);
  operationQueue = result.then(
    () => undefined,
    () => undefined,
  );
  return result;
}

export function getStoredPushToken(): Promise<string | null> {
  return AsyncStorage.getItem(PUSH_TOKEN_STORAGE_KEY);
}

export function beginPushTokenRegistration(): number {
  registrationGeneration += 1;
  if (retryTimer) clearTimeout(retryTimer);
  retryTimer = null;
  retryAttempt = 0;
  return registrationGeneration;
}

export function completePushTokenRegistration(
  generation: number,
  expoPushToken: string,
  registerOnServer: (signal: AbortSignal) => Promise<void>,
): Promise<boolean> {
  return serialize(async () => {
    await deactivatePendingToken();
    if (generation !== registrationGeneration) return false;
    await AsyncStorage.setItem(PENDING_DEACTIVATION_KEY, expoPushToken);

    const controller = new AbortController();
    const timeout = setTimeout(
      () => controller.abort(),
      REGISTRATION_TIMEOUT_MS,
    );
    try {
      await registerOnServer(controller.signal);
    } catch (error) {
      await AsyncStorage.setItem(PENDING_DEACTIVATION_KEY, expoPushToken);
      scheduleStoredPushTokenDeactivation();
      throw error;
    } finally {
      clearTimeout(timeout);
    }
    if (generation !== registrationGeneration) {
      await AsyncStorage.setItem(PENDING_DEACTIVATION_KEY, expoPushToken);
      await deactivatePendingToken();
      return false;
    }
    await AsyncStorage.setItem(PUSH_TOKEN_STORAGE_KEY, expoPushToken);
    await AsyncStorage.removeItem(PENDING_DEACTIVATION_KEY);
    return true;
  });
}

export function deactivateStoredPushToken(): Promise<void> {
  if (retryTimer) clearTimeout(retryTimer);
  retryTimer = null;
  registrationGeneration += 1;
  return serialize(async () => {
    const activeToken = await getStoredPushToken();
    if (activeToken) {
      await AsyncStorage.setItem(PENDING_DEACTIVATION_KEY, activeToken);
      await AsyncStorage.removeItem(PUSH_TOKEN_STORAGE_KEY);
    }
    await deactivatePendingToken();
  }).catch((error) => {
    retryAttempt += 1;
    schedulePendingCleanupRetry(registrationGeneration);
    throw error;
  });
}

async function deactivatePendingToken(): Promise<void> {
  const expoPushToken = await AsyncStorage.getItem(PENDING_DEACTIVATION_KEY);
  if (!expoPushToken) return;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), DEACTIVATION_TIMEOUT_MS);
  try {
    const response = await fetch(getApiBaseUrl() + "/push/unregister", {
      method: "POST",
      body: JSON.stringify({ expoPushToken }),
      headers: { "Content-Type": "application/json" },
      signal: controller.signal,
    });
    if (!response.ok) {
      throw new Error(`Push unregistration failed (${response.status})`);
    }
    await AsyncStorage.removeItem(PENDING_DEACTIVATION_KEY);
  } finally {
    clearTimeout(timeout);
  }
}

export function scheduleStoredPushTokenDeactivation(): void {
  if (retryTimer || scheduledCleanupInFlight) return;
  const cleanupGeneration = ++registrationGeneration;
  runScheduledCleanup(cleanupGeneration, true);
}

function runScheduledCleanup(
  cleanupGeneration: number,
  stageActiveToken: boolean,
): void {
  scheduledCleanupInFlight = true;
  serialize(async () => {
    if (stageActiveToken) {
      const activeToken = await getStoredPushToken();
      if (cleanupGeneration !== registrationGeneration) return;
      if (activeToken) {
        await AsyncStorage.setItem(PENDING_DEACTIVATION_KEY, activeToken);
        await AsyncStorage.removeItem(PUSH_TOKEN_STORAGE_KEY);
      }
    }
    if (cleanupGeneration !== registrationGeneration) return;
    await deactivatePendingToken();
  })
    .then(() => {
      retryAttempt = 0;
    })
    .catch(() => {
      retryAttempt += 1;
      schedulePendingCleanupRetry(cleanupGeneration);
    })
    .finally(() => {
      scheduledCleanupInFlight = false;
    });
}

function schedulePendingCleanupRetry(cleanupGeneration: number): void {
  if (retryTimer || cleanupGeneration !== registrationGeneration) return;
  const delay = Math.min(1000 * 2 ** retryAttempt, 5 * 60 * 1000);
  retryTimer = setTimeout(() => {
    retryTimer = null;
    if (cleanupGeneration === registrationGeneration) {
      runScheduledCleanup(cleanupGeneration, false);
    }
  }, delay);
}
