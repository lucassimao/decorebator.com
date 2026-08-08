import * as SecureStore from "expo-secure-store";
import AsyncStorage from "@react-native-async-storage/async-storage";
import offlineManager from "@/utils/offlineManager";

jest.mock("expo-secure-store", () => ({
  getItem: jest.fn((key: string) =>
    key === "sessionCredentials"
      ? JSON.stringify({
          authorization: "stored-token",
          refreshToken: "stored-refresh-token",
        })
      : null,
  ),
  setItem: jest.fn(),
  deleteItemAsync: jest.fn(() => Promise.resolve()),
}));

jest.mock("@/api/baseUrl", () => ({
  getApiBaseUrl: () => "https://api.example.test",
}));

jest.mock("@/utils/offlineManager", () => ({
  __esModule: true,
  default: {
    clearCache: jest.fn(() => Promise.resolve()),
    setUserPremiumStatus: jest.fn(),
  },
}));

const {
  authenticatedFetch,
  clearSessionCredentials,
  deleteProfile,
  getAuthorization,
  signin,
  sigout,
  update,
} = jest.requireActual<typeof import("@/api/users")>("@/api/users");

describe("authorization cache", () => {
  const freshToken = "e30.eyJzdWJzY3JpcHRpb25QbGFuIjoiZnJlZSJ9.signature";
  const freshRefresh = "refresh-token-value";
  beforeEach(() => {
    jest.clearAllMocks();
    (SecureStore.deleteItemAsync as jest.Mock).mockResolvedValue(undefined);
  });

  it("reads secure storage only once across repeated consumers", () => {
    expect(getAuthorization()).toBe("stored-token");
    expect(getAuthorization()).toBe("stored-token");
    expect(SecureStore.getItem).toHaveBeenCalledTimes(1);
  });

  it("updates the cache when sign in stores a refreshed credential", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization" ? freshToken : freshRefresh,
        },
      } as unknown as Response),
    );

    await signin({ email: "test@example.com", password: "secret" });

    expect(SecureStore.setItem).toHaveBeenCalledWith(
      "sessionCredentials",
      JSON.stringify({
        authorization: freshToken,
        refreshToken: freshRefresh,
      }),
    );
    expect(getAuthorization()).toBe(freshToken);
  });

  it("clears cached access even when sign-out storage cleanup completes later", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: true, status: 204 } as Response),
    );

    await sigout();

    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("authorization");
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("refreshToken");
    expect(getAuthorization()).toBeNull();
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith(
      "pendingDurableUserCleanup",
    );
  });

  it("clears native credentials immediately after a password security action", async () => {
    await AsyncStorage.multiSet([
      ["cachedUserProfile", "profile"],
      ["cachedSubscription", "subscription"],
    ]);
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization" ? freshToken : freshRefresh,
        },
      } as unknown as Response),
    );
    await signin({ email: "test@example.com", password: "secret" });

    await clearSessionCredentials();

    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("authorization");
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("refreshToken");
    expect(offlineManager.clearCache).toHaveBeenCalled();
    expect(offlineManager.setUserPremiumStatus).toHaveBeenCalledWith(false);
    expect(AsyncStorage.multiRemove).toHaveBeenCalledWith([
      "cachedUserProfile",
      "cachedSubscription",
    ]);
    expect(getAuthorization()).toBeNull();
  });

  it("clears native credentials after account deletion", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization" ? freshToken : freshRefresh,
        },
      } as unknown as Response),
    );
    await signin({ email: "test@example.com", password: "secret" });
    await AsyncStorage.multiSet([
      ["cachedUserProfile", "profile"],
      ["cachedSubscription", "subscription"],
      ["flashcard_position_42", "3"],
      ["wordlistFabHintSeen_42", "true"],
      ["pushPromptedAfterFirstQuiz", "true"],
      ["theme", "dark"],
    ]);
    jest.clearAllMocks();
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: true, status: 204 } as Response),
    );

    await deleteProfile();

    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("authorization");
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("refreshToken");
    expect(offlineManager.setUserPremiumStatus).toHaveBeenCalledWith(false);
    expect(AsyncStorage.multiRemove).toHaveBeenCalledWith([
      "cachedUserProfile",
      "cachedSubscription",
      "flashcard_position_42",
      "wordlistFabHintSeen_42",
      "pushPromptedAfterFirstQuiz",
    ]);
    expect(await AsyncStorage.getItem("theme")).toBe("dark");
    expect(getAuthorization()).toBeNull();
  });

  it("still clears storage and credentials when offline asset cleanup fails", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization" ? freshToken : freshRefresh,
        },
      } as unknown as Response),
    );
    await signin({ email: "test@example.com", password: "secret" });
    await AsyncStorage.setItem("cachedUserProfile", "profile");
    (offlineManager.clearCache as jest.Mock).mockRejectedValueOnce(
      new Error("asset cleanup failed"),
    );
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: true, status: 204 } as Response),
    );

    await expect(deleteProfile()).rejects.toThrow(
      "Durable user cleanup incomplete",
    );

    expect(AsyncStorage.multiRemove).toHaveBeenCalledWith([
      "cachedUserProfile",
    ]);
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("authorization");
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("refreshToken");
  });

  it("attempts both native credential deletions when one secure-store call fails", async () => {
    (SecureStore.deleteItemAsync as jest.Mock).mockImplementation(
      async (name) => {
        if (name === "authorization")
          throw new Error("secure store unavailable");
      },
    );

    await expect(clearSessionCredentials()).rejects.toThrow(
      "Durable user cleanup incomplete",
    );

    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("authorization");
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("refreshToken");
    expect(getAuthorization()).toBeNull();
    expect(SecureStore.setItem).toHaveBeenCalledWith(
      "pendingDurableUserCleanup",
      "true",
    );
  });

  it("coalesces concurrent 401 refreshes and retries with the rotated access token", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization" ? freshToken : freshRefresh,
        },
      } as unknown as Response),
    );
    await signin({ email: "test@example.com", password: "secret" });
    const rotatedAccess = "rotated-access";
    const rotatedRefresh = "rotated-refresh";
    let protectedCalls = 0;
    let refreshCalls = 0;
    global.fetch = jest.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/session/refresh")) {
        refreshCalls += 1;
        return Promise.resolve({
          ok: true,
          status: 200,
          headers: {
            get: (name: string) =>
              name.toLowerCase() === "authorization"
                ? rotatedAccess
                : rotatedRefresh,
          },
        } as unknown as Response);
      }
      protectedCalls += 1;
      return Promise.resolve({
        ok: protectedCalls > 2,
        status: protectedCalls > 2 ? 200 : 401,
        headers: { get: () => null },
      } as unknown as Response);
    }) as jest.Mock;

    const [first, second] = await Promise.all([
      authenticatedFetch("https://api.example.test/one"),
      authenticatedFetch("https://api.example.test/two"),
    ]);

    expect(first.status).toBe(200);
    expect(second.status).toBe(200);
    expect(refreshCalls).toBe(1);
    expect(SecureStore.setItem).toHaveBeenCalledWith(
      "sessionCredentials",
      JSON.stringify({
        authorization: rotatedAccess,
        refreshToken: rotatedRefresh,
      }),
    );
  });

  it("does not let an older refresh overwrite a replacement identity", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization"
              ? "identity-a-access"
              : "identity-a-refresh",
        },
      } as unknown as Response),
    );
    await signin({ email: "a@example.com", password: "secret" });

    let resolveStaleRefresh!: (response: Response) => void;
    let refreshStarted!: () => void;
    const refreshDidStart = new Promise<void>((resolve) => {
      refreshStarted = resolve;
    });
    global.fetch = jest.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/session/refresh")) {
        refreshStarted();
        return new Promise<Response>((resolve) => {
          resolveStaleRefresh = resolve;
        });
      }
      if (url.endsWith("/login")) {
        return Promise.resolve({
          ok: true,
          status: 200,
          headers: {
            get: (name: string) =>
              name.toLowerCase() === "authorization"
                ? "identity-b-access"
                : "identity-b-refresh",
          },
        } as unknown as Response);
      }
      return Promise.resolve({ ok: false, status: 401 } as Response);
    }) as jest.Mock;

    const staleRequest = authenticatedFetch(
      "https://api.example.test/protected",
    );
    await refreshDidStart;
    await signin({ email: "b@example.com", password: "secret" });
    resolveStaleRefresh({
      ok: true,
      status: 200,
      headers: {
        get: (name: string) =>
          name.toLowerCase() === "authorization"
            ? "stale-identity-a-access"
            : "stale-identity-a-refresh",
      },
    } as unknown as Response);

    await expect(staleRequest).rejects.toThrow(
      "Session changed during request",
    );
    expect(getAuthorization()).toBe("identity-b-access");
    expect(SecureStore.setItem).not.toHaveBeenCalledWith(
      "sessionCredentials",
      expect.stringContaining("stale-identity-a"),
    );
  });

  it("does not let an older refresh rejection erase a replacement identity", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization"
              ? "identity-a-access"
              : "identity-a-refresh",
        },
      } as unknown as Response),
    );
    await signin({ email: "a@example.com", password: "secret" });

    let resolveStaleRefresh!: (response: Response) => void;
    let refreshStarted!: () => void;
    const refreshDidStart = new Promise<void>((resolve) => {
      refreshStarted = resolve;
    });
    global.fetch = jest.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/session/refresh")) {
        refreshStarted();
        return new Promise<Response>((resolve) => {
          resolveStaleRefresh = resolve;
        });
      }
      if (url.endsWith("/login")) {
        return Promise.resolve({
          ok: true,
          status: 200,
          headers: {
            get: (name: string) =>
              name.toLowerCase() === "authorization"
                ? "identity-b-access"
                : "identity-b-refresh",
          },
        } as unknown as Response);
      }
      return Promise.resolve({ ok: false, status: 401 } as Response);
    }) as jest.Mock;

    const staleRequest = authenticatedFetch(
      "https://api.example.test/protected",
    );
    await refreshDidStart;
    await signin({ email: "b@example.com", password: "secret" });
    jest.clearAllMocks();
    resolveStaleRefresh({ ok: false, status: 401 } as Response);

    await expect(staleRequest).rejects.toThrow(
      "Session changed during request",
    );
    expect(getAuthorization()).toBe("identity-b-access");
    expect(SecureStore.deleteItemAsync).not.toHaveBeenCalledWith(
      "sessionCredentials",
    );
  });

  it("clears rejected session credentials without retrying the protected request", async () => {
    await AsyncStorage.setItem("cachedUserProfile", "profile");
    await AsyncStorage.setItem("expoPushToken", "ExponentPushToken[device]");
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization" ? freshToken : freshRefresh,
        },
      } as unknown as Response),
    );
    await signin({ email: "test@example.com", password: "secret" });

    let protectedCalls = 0;
    global.fetch = jest.fn((input: RequestInfo | URL) => {
      if (String(input).endsWith("/push/unregister")) {
        return Promise.resolve({ ok: true, status: 204 } as Response);
      }
      if (String(input).endsWith("/session/refresh")) {
        return Promise.resolve({ ok: false, status: 401 } as Response);
      }
      protectedCalls += 1;
      return Promise.resolve({ ok: false, status: 401 } as Response);
    }) as jest.Mock;

    await expect(
      authenticatedFetch("https://api.example.test/users"),
    ).rejects.toThrow("Authentication session expired");
    expect(protectedCalls).toBe(1);
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("authorization");
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("refreshToken");
    expect(offlineManager.clearCache).toHaveBeenCalled();
    expect(offlineManager.setUserPremiumStatus).toHaveBeenCalledWith(false);
    expect(AsyncStorage.multiRemove).toHaveBeenCalledWith([
      "cachedUserProfile",
    ]);
    expect(AsyncStorage.removeItem).toHaveBeenCalledWith("expoPushToken");
    expect(await AsyncStorage.getItem("expoPushToken")).toBeNull();
    expect(getAuthorization()).toBeNull();
  });

  it("treats a refresh 403 as transient and preserves the local session", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization" ? freshToken : freshRefresh,
        },
      } as unknown as Response),
    );
    await signin({ email: "test@example.com", password: "secret" });
    jest.clearAllMocks();

    global.fetch = jest.fn((input: RequestInfo | URL) =>
      Promise.resolve({
        ok: false,
        status: String(input).endsWith("/session/refresh") ? 403 : 401,
      } as Response),
    ) as jest.Mock;

    await expect(
      authenticatedFetch("https://api.example.test/users"),
    ).rejects.toThrow("Session refresh temporarily unavailable (403)");
    expect(SecureStore.deleteItemAsync).not.toHaveBeenCalledWith(
      "authorization",
    );
    expect(SecureStore.deleteItemAsync).not.toHaveBeenCalledWith(
      "refreshToken",
    );
    expect(getAuthorization()).toBe(freshToken);
  });

  it("preserves durable credentials when refresh infrastructure is unavailable", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization" ? freshToken : freshRefresh,
        },
      } as unknown as Response),
    );
    await signin({ email: "test@example.com", password: "secret" });
    jest.clearAllMocks();

    global.fetch = jest.fn((input: RequestInfo | URL) =>
      Promise.resolve({
        ok: false,
        status: String(input).endsWith("/session/refresh") ? 503 : 401,
      } as Response),
    ) as jest.Mock;

    await expect(
      authenticatedFetch("https://api.example.test/users"),
    ).rejects.toThrow("Session refresh temporarily unavailable (503)");
    expect(SecureStore.deleteItemAsync).not.toHaveBeenCalledWith(
      "authorization",
    );
    expect(SecureStore.deleteItemAsync).not.toHaveBeenCalledWith(
      "refreshToken",
    );
    expect(getAuthorization()).toBe(freshToken);
  });

  it.each([408, 503])(
    "retains the durable cleanup barrier when logout returns ambiguous HTTP %s",
    async (logoutStatus) => {
      global.fetch = jest.fn(() =>
        Promise.resolve({
          ok: true,
          status: 200,
          headers: {
            get: (name: string) =>
              name.toLowerCase() === "authorization"
                ? freshToken
                : freshRefresh,
          },
        } as unknown as Response),
      );
      await signin({ email: "test@example.com", password: "secret" });
      await AsyncStorage.setItem(
        "expoPushToken",
        "ExponentPushToken[logout-device]",
      );
      jest.clearAllMocks();
      global.fetch = jest.fn((input: RequestInfo | URL) =>
        Promise.resolve({
          ok: String(input).endsWith("/push/unregister"),
          status: String(input).endsWith("/push/unregister")
            ? 204
            : logoutStatus,
        } as Response),
      ) as jest.Mock;

      await expect(sigout()).rejects.toThrow("Logout temporarily unavailable");

      expect(SecureStore.setItem).toHaveBeenCalledWith(
        "pendingDurableUserCleanup",
        "true",
      );
      expect(SecureStore.deleteItemAsync).not.toHaveBeenCalledWith(
        "pendingDurableUserCleanup",
      );
      expect(offlineManager.clearCache).not.toHaveBeenCalled();
      expect(AsyncStorage.removeItem).not.toHaveBeenCalledWith("expoPushToken");
      expect(await AsyncStorage.getItem("expoPushToken")).toBe(
        "ExponentPushToken[logout-device]",
      );
      expect(getAuthorization()).toBe(freshToken);
    },
  );

  it("cleans identity state before replacing a native session", async () => {
    await AsyncStorage.multiSet([
      ["cachedUserProfile", "old-profile"],
      ["offline_cache_wordlists", "old-wordlists"],
      ["expoPushToken", "ExponentPushToken[old-device]"],
    ]);
    global.fetch = jest.fn((input: RequestInfo | URL) => {
      if (String(input).endsWith("/push/unregister")) {
        return Promise.resolve({ ok: true, status: 204 } as Response);
      }
      return Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization" ? freshToken : freshRefresh,
        },
      } as unknown as Response);
    }) as jest.Mock;

    await signin({ email: "new@example.com", password: "secret" });

    expect(AsyncStorage.removeItem).toHaveBeenCalledWith("expoPushToken");
    expect(offlineManager.clearCache).toHaveBeenCalled();
    expect(AsyncStorage.multiRemove).toHaveBeenCalledWith([
      "cachedUserProfile",
    ]);
    expect(SecureStore.setItem).toHaveBeenLastCalledWith(
      "sessionCredentials",
      JSON.stringify({
        authorization: freshToken,
        refreshToken: freshRefresh,
      }),
    );
  });

  it.each([
    [
      "password update",
      () =>
        update({
          updatePassword: {
            currentPassword: "old",
            newPassword: "new-password",
          },
        }),
    ],
    ["account deletion", () => deleteProfile()],
  ])(
    "retains the cleanup barrier on an ambiguous 5xx %s",
    async (_label, operation) => {
      global.fetch = jest.fn((input: RequestInfo | URL) => {
        if (String(input).endsWith("/push/unregister")) {
          return Promise.resolve({ ok: true, status: 204 } as Response);
        }
        if (String(input).endsWith("/login")) {
          return Promise.resolve({
            ok: true,
            status: 200,
            headers: {
              get: (name: string) =>
                name.toLowerCase() === "authorization"
                  ? freshToken
                  : freshRefresh,
            },
          } as unknown as Response);
        }
        return Promise.resolve({
          ok: false,
          status: 503,
          json: () => Promise.resolve({ error: "temporarily unavailable" }),
        } as unknown as Response);
      }) as jest.Mock;
      await signin({ email: "test@example.com", password: "secret" });
      jest.clearAllMocks();

      await expect(operation()).rejects.toThrow("temporarily unavailable");

      expect(SecureStore.setItem).toHaveBeenCalledWith(
        "pendingDurableUserCleanup",
        "true",
      );
      expect(SecureStore.deleteItemAsync).not.toHaveBeenCalledWith(
        "pendingDurableUserCleanup",
      );
    },
  );
});
