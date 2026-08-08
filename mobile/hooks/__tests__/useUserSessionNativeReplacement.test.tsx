import AsyncStorage from "@react-native-async-storage/async-storage";
import { act, renderHook, waitFor } from "@testing-library/react-native";
import { useQuery } from "@tanstack/react-query";
import { getAuthenticationSessionEpoch, signin } from "@/api/users";
import { useUserSession } from "@/hooks/useUserSession";

const mockSecureValues = new Map<string, string>();
const mockQueryClient = {
  clear: jest.fn(),
  getQueryCache: () => ({ clear: jest.fn() }),
  getMutationCache: () => ({ clear: jest.fn() }),
  invalidateQueries: jest.fn(),
};

jest.mock("expo-secure-store", () => ({
  getItem: jest.fn((key: string) => mockSecureValues.get(key) ?? null),
  setItem: jest.fn((key: string, value: string) =>
    mockSecureValues.set(key, value),
  ),
  deleteItemAsync: jest.fn(async (key: string) => {
    mockSecureValues.delete(key);
  }),
}));

jest.mock("@tanstack/react-query", () => ({
  useQuery: jest.fn(),
  useQueryClient: () => mockQueryClient,
}));

jest.mock("@/api/baseUrl", () => ({
  getApiBaseUrl: () => "https://api.example.test",
}));

jest.mock("@/utils/offlineManager", () => ({
  __esModule: true,
  default: {
    clearCache: jest.fn(() => Promise.resolve()),
    getNetworkStatus: jest.fn(() => true),
    setUserPremiumStatus: jest.fn(),
  },
}));

jest.mock("@/utils/pushNotifications", () => ({
  registerDevicePushToken: jest.fn(() => Promise.resolve(null)),
  resumeStoredPushTokenDeactivation: jest.fn(),
}));

jest.unmock("@/api/users");

describe("native mounted session replacement", () => {
  beforeEach(async () => {
    jest.clearAllMocks();
    mockSecureValues.clear();
    await AsyncStorage.clear();
    jest.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useQuery>);
  });

  it("drops identity A memory before exposing identity B's epoch", async () => {
    mockSecureValues.set(
      "sessionCredentials",
      JSON.stringify({
        authorization: "identity-a-access",
        refreshToken: "identity-a-refresh",
      }),
    );
    await AsyncStorage.setItem(
      "cachedUserProfile",
      JSON.stringify({
        id: 1,
        email: "a@example.com",
        subscriptionPlan: "free",
        notificationsEnabled: false,
      }),
    );
    const originalMultiSet = (
      AsyncStorage.multiSet as jest.Mock
    ).getMockImplementation();
    let releaseIdentityAWrite!: () => void;
    let identityAWriteStarted!: () => void;
    const identityAWrite = new Promise<void>((resolve) => {
      identityAWriteStarted = resolve;
    });
    const identityAWriteGate = new Promise<void>((resolve) => {
      releaseIdentityAWrite = resolve;
    });
    (AsyncStorage.multiSet as jest.Mock).mockImplementationOnce(
      async (entries: [string, string][]) => {
        identityAWriteStarted();
        await identityAWriteGate;
        return originalMultiSet?.(entries);
      },
    );
    const epochA = getAuthenticationSessionEpoch();
    const { result } = renderHook(() => useUserSession());
    await waitFor(() => expect(result.current.user?.id).toBe(1));
    await identityAWrite;

    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        headers: {
          get: (name: string) =>
            name.toLowerCase() === "authorization"
              ? "identity-b-access"
              : "identity-b-refresh",
        },
      } as unknown as Response),
    );
    let replacement!: Promise<void>;
    replacement = signin({ email: "b@example.com", password: "secret" });
    await waitFor(() => expect(getAuthenticationSessionEpoch()).toBeNull());
    releaseIdentityAWrite();
    await act(async () => replacement);

    expect(getAuthenticationSessionEpoch()).not.toBe(epochA);
    expect(result.current.user).toBeNull();
    expect(await AsyncStorage.getItem("cachedUserProfile")).toBeNull();
  });
});
