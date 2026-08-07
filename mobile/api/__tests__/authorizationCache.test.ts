import * as SecureStore from "expo-secure-store";

jest.mock("expo-secure-store", () => ({
  getItem: jest.fn(() => "stored-token"),
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

const { authenticatedFetch, getAuthorization, signin, sigout } =
  jest.requireActual<typeof import("@/api/users")>("@/api/users");

describe("authorization cache", () => {
  const freshToken = "e30.eyJzdWJzY3JpcHRpb25QbGFuIjoiZnJlZSJ9.signature";
  const freshRefresh = "refresh-token-value";
  beforeEach(() => {
    jest.clearAllMocks();
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
      "authorization",
      freshToken,
    );
    expect(SecureStore.setItem).toHaveBeenCalledWith(
      "refreshToken",
      freshRefresh,
    );
    expect(getAuthorization()).toBe(freshToken);
  });

  it("clears cached access even when sign-out storage cleanup completes later", async () => {
    await sigout();

    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("authorization");
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("refreshToken");
    expect(getAuthorization()).toBeNull();
  });

  it("coalesces concurrent 401 refreshes and retries with the rotated access token", async () => {
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
      "authorization",
      rotatedAccess,
    );
  });

  it("clears rejected session credentials without retrying the protected request", async () => {
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
      if (String(input).endsWith("/session/refresh")) {
        return Promise.resolve({ ok: false, status: 401 } as Response);
      }
      protectedCalls += 1;
      return Promise.resolve({ ok: false, status: 401 } as Response);
    }) as jest.Mock;

    const response = await authenticatedFetch("https://api.example.test/users");

    expect(response.status).toBe(401);
    expect(protectedCalls).toBe(1);
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("authorization");
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("refreshToken");
    expect(getAuthorization()).toBeNull();
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

    global.fetch = jest.fn((input: RequestInfo | URL) =>
      Promise.resolve({
        ok: false,
        status: String(input).endsWith("/session/refresh") ? 503 : 401,
      } as Response),
    ) as jest.Mock;

    const response = await authenticatedFetch("https://api.example.test/users");

    expect(response.status).toBe(401);
    expect(SecureStore.deleteItemAsync).not.toHaveBeenCalled();
    expect(getAuthorization()).toBe(freshToken);
  });
});
