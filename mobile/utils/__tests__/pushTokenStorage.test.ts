import AsyncStorage from "@react-native-async-storage/async-storage";
import {
  beginPushTokenRegistration,
  completePushTokenRegistration,
  deactivateStoredPushToken,
  scheduleStoredPushTokenDeactivation,
} from "@/utils/pushTokenStorage";

jest.mock("@/api/baseUrl", () => ({
  getApiBaseUrl: () => "https://api.example.test",
}));

describe("push token storage", () => {
  beforeEach(async () => {
    jest.clearAllMocks();
    await AsyncStorage.multiRemove([
      "expoPushToken",
      "pendingExpoPushTokenDeactivation",
    ]);
  });

  it("deactivates a registration that completes after session invalidation", async () => {
    const generation = beginPushTokenRegistration();
    let finishRegistration!: () => void;
    const serverRegistration = jest.fn(
      () =>
        new Promise<void>((resolve) => {
          finishRegistration = resolve;
        }),
    );
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: true, status: 204 } as Response),
    );

    const registration = completePushTokenRegistration(
      generation,
      "ExponentPushToken[new-device]",
      serverRegistration,
    );
    while (!finishRegistration) {
      await new Promise((resolve) => setTimeout(resolve, 0));
    }
    const deactivation = deactivateStoredPushToken();
    finishRegistration();

    await expect(registration).resolves.toBe(false);
    await expect(deactivation).resolves.toBeUndefined();
    expect(global.fetch).toHaveBeenCalledTimes(1);
    expect(await AsyncStorage.getItem("expoPushToken")).toBeNull();
    expect(
      await AsyncStorage.getItem("pendingExpoPushTokenDeactivation"),
    ).toBeNull();
  });

  it("retains the exact failed capability and clears it before a new registration", async () => {
    await AsyncStorage.setItem(
      "expoPushToken",
      "ExponentPushToken[old-device]",
    );
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: false, status: 503 } as Response),
    );

    await expect(deactivateStoredPushToken()).rejects.toThrow(
      "Push unregistration failed (503)",
    );
    expect(await AsyncStorage.getItem("expoPushToken")).toBeNull();
    expect(await AsyncStorage.getItem("pendingExpoPushTokenDeactivation")).toBe(
      "ExponentPushToken[old-device]",
    );

    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      status: 204,
    } as Response);
    const registerOnServer = jest.fn(() => Promise.resolve());
    const generation = beginPushTokenRegistration();

    await expect(
      completePushTokenRegistration(
        generation,
        "ExponentPushToken[new-device]",
        registerOnServer,
      ),
    ).resolves.toBe(true);

    expect(registerOnServer).toHaveBeenCalledTimes(1);
    expect(await AsyncStorage.getItem("expoPushToken")).toBe(
      "ExponentPushToken[new-device]",
    );
    expect(
      await AsyncStorage.getItem("pendingExpoPushTokenDeactivation"),
    ).toBeNull();
  });

  it("stages and deactivates an active token discovered without credentials", async () => {
    jest.useFakeTimers();
    await AsyncStorage.setItem(
      "expoPushToken",
      "ExponentPushToken[orphaned-device]",
    );
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: true, status: 204 } as Response),
    );

    scheduleStoredPushTokenDeactivation();
    await jest.runOnlyPendingTimersAsync();

    expect(global.fetch).toHaveBeenCalledWith(
      "https://api.example.test/push/unregister",
      expect.objectContaining({
        body: JSON.stringify({
          expoPushToken: "ExponentPushToken[orphaned-device]",
        }),
      }),
    );
    expect(await AsyncStorage.getItem("expoPushToken")).toBeNull();
    expect(
      await AsyncStorage.getItem("pendingExpoPushTokenDeactivation"),
    ).toBeNull();
    jest.useRealTimers();
  });

  it("cancels an old retry before registering a replacement identity", async () => {
    jest.useFakeTimers();
    await AsyncStorage.setItem(
      "expoPushToken",
      "ExponentPushToken[shared-device]",
    );
    global.fetch = jest
      .fn()
      .mockResolvedValueOnce({ ok: false, status: 503 } as Response)
      .mockResolvedValue({ ok: true, status: 204 } as Response);

    scheduleStoredPushTokenDeactivation();
    await jest.advanceTimersByTimeAsync(0);
    for (let attempt = 0; attempt < 10 && !jest.getTimerCount(); attempt += 1) {
      await Promise.resolve();
    }
    expect(jest.getTimerCount()).toBe(1);

    const registerOnServer = jest.fn(() => Promise.resolve());
    const generation = beginPushTokenRegistration();
    await expect(
      completePushTokenRegistration(
        generation,
        "ExponentPushToken[shared-device]",
        registerOnServer,
      ),
    ).resolves.toBe(true);
    await jest.runOnlyPendingTimersAsync();

    expect(registerOnServer).toHaveBeenCalledTimes(1);
    expect(global.fetch).toHaveBeenCalledTimes(2);
    expect(await AsyncStorage.getItem("expoPushToken")).toBe(
      "ExponentPushToken[shared-device]",
    );
    jest.useRealTimers();
  });

  it("re-arms pending cleanup after explicit deactivation supersedes an old retry", async () => {
    jest.useFakeTimers();
    await AsyncStorage.setItem(
      "expoPushToken",
      "ExponentPushToken[retry-device]",
    );
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: false, status: 503 } as Response),
    );

    scheduleStoredPushTokenDeactivation();
    for (let attempt = 0; attempt < 20 && !jest.getTimerCount(); attempt += 1) {
      await Promise.resolve();
    }
    expect(jest.getTimerCount()).toBe(1);

    await expect(deactivateStoredPushToken()).rejects.toThrow(
      "Push unregistration failed (503)",
    );
    for (let attempt = 0; attempt < 20 && !jest.getTimerCount(); attempt += 1) {
      await Promise.resolve();
    }

    expect(jest.getTimerCount()).toBe(1);
    expect(await AsyncStorage.getItem("expoPushToken")).toBeNull();
    expect(await AsyncStorage.getItem("pendingExpoPushTokenDeactivation")).toBe(
      "ExponentPushToken[retry-device]",
    );
    jest.clearAllTimers();
    jest.useRealTimers();
  });
});
