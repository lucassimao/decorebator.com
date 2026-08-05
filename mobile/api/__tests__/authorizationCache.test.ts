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

const { getAuthorization, signin, sigout } =
  jest.requireActual<typeof import("@/api/users")>("@/api/users");

describe("authorization cache", () => {
  const freshToken = "e30.eyJzdWJzY3JpcHRpb25QbGFuIjoiZnJlZSJ9.signature";
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
        headers: { get: () => freshToken },
      } as unknown as Response),
    );

    await signin({ email: "test@example.com", password: "secret" });

    expect(SecureStore.setItem).toHaveBeenCalledWith(
      "authorization",
      freshToken,
    );
    expect(getAuthorization()).toBe(freshToken);
  });

  it("clears cached access even when sign-out storage cleanup completes later", async () => {
    await sigout();

    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith("authorization");
    expect(getAuthorization()).toBeNull();
  });
});
