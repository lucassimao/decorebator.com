import * as SecureStore from "expo-secure-store";

jest.mock("expo-secure-store", () => ({
  getItem: jest.fn((key: string) => {
    if (key === "authorization") return "identity-b-access";
    if (key === "refreshToken") return "identity-a-refresh";
    return null;
  }),
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

jest.mock("@/utils/pushTokenStorage", () => ({
  deactivateStoredPushToken: jest.fn(() => Promise.resolve()),
  scheduleStoredPushTokenDeactivation: jest.fn(),
}));

describe("legacy native credential quarantine", () => {
  it("never combines split legacy access and refresh credentials", () => {
    const { getAuthorization } =
      jest.requireActual<typeof import("@/api/users")>("@/api/users");

    expect(getAuthorization()).toBeNull();
    expect(SecureStore.setItem).toHaveBeenCalledWith(
      "pendingDurableUserCleanup",
      "true",
    );
    expect(SecureStore.setItem).not.toHaveBeenCalledWith(
      "sessionCredentials",
      expect.anything(),
    );
  });
});
