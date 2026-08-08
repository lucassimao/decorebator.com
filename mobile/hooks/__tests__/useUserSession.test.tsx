import AsyncStorage from "@react-native-async-storage/async-storage";
import { act, renderHook, waitFor } from "@testing-library/react-native";
import { useQuery } from "@tanstack/react-query";

import { useUserSession } from "@/hooks/useUserSession";
import {
  clearSessionScopedCaches,
  getAuthenticationSessionEpoch,
  hasAuthenticationSession,
  isAuthenticationSessionReady,
  subscribeAuthenticationSessionChanges,
} from "@/api/users";
import { resumeStoredPushTokenDeactivation } from "@/utils/pushNotifications";

const mockQueryClient = {
  clear: jest.fn(),
  getQueryCache: () => ({ clear: jest.fn() }),
  getMutationCache: () => ({ clear: jest.fn() }),
  invalidateQueries: jest.fn(),
};

jest.mock("@tanstack/react-query", () => ({
  useQuery: jest.fn(),
  useQueryClient: () => mockQueryClient,
}));

jest.mock("@/api/users", () => ({
  getAuthorization: jest.fn(),
  hasAuthenticationSession: jest.fn(),
  isAuthenticationSessionReady: jest.fn(() => true),
  getAuthenticationSessionEpoch: jest.fn(() => "native"),
  subscribeAuthenticationSessionChanges: jest.fn(() => jest.fn()),
  clearSessionScopedCaches: jest.fn(() => Promise.resolve()),
  persistSessionSnapshot: jest.fn(() => Promise.resolve()),
  getProfile: jest.fn(),
}));

jest.mock("@/api/subscriptions", () => ({
  getSubscriptionStatus: jest.fn(),
}));

jest.mock("@/utils/offlineManager", () => ({
  __esModule: true,
  default: {
    getNetworkStatus: jest.fn(() => true),
    setUserPremiumStatus: jest.fn(),
  },
}));

jest.mock("@/utils/pushNotifications", () => ({
  registerDevicePushToken: jest.fn(() => Promise.resolve(null)),
  resumeStoredPushTokenDeactivation: jest.fn(),
}));

describe("useUserSession credential gate", () => {
  beforeEach(async () => {
    jest.clearAllMocks();
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue("native");
    jest.mocked(isAuthenticationSessionReady).mockReturnValue(true);
    jest.mocked(clearSessionScopedCaches).mockResolvedValue(undefined);
    await AsyncStorage.clear();
    jest.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useQuery>);
  });

  it("keeps launch loading until browser session initialization completes", async () => {
    let notifySessionChange!: (epoch: string | null) => void;
    jest.mocked(hasAuthenticationSession).mockReturnValue(false);
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue(null);
    jest.mocked(isAuthenticationSessionReady).mockReturnValue(false);
    jest
      .mocked(subscribeAuthenticationSessionChanges)
      .mockImplementationOnce((listener) => {
        notifySessionChange = listener;
        return jest.fn();
      });

    const { result } = renderHook(() => useUserSession());
    expect(result.current.isLoading).toBe(true);

    jest.mocked(isAuthenticationSessionReady).mockReturnValue(true);
    act(() => notifySessionChange(null));

    await waitFor(() => expect(result.current.isLoading).toBe(false));
  });

  it("resumes durable token cleanup when notifications are disabled", async () => {
    jest.mocked(hasAuthenticationSession).mockReturnValue(true);
    jest.mocked(useQuery).mockReturnValue({
      data: {
        id: 7,
        subscriptionPlan: "free",
        notificationsEnabled: false,
      },
      isLoading: false,
      error: null,
    } as ReturnType<typeof useQuery>);

    renderHook(() => useUserSession());

    await waitFor(() =>
      expect(resumeStoredPushTokenDeactivation).toHaveBeenCalled(),
    );
  });

  it("fails a mounted session closed when durable cleanup becomes pending", async () => {
    let notifySessionChange!: (epoch: string | null) => void;
    jest.mocked(hasAuthenticationSession).mockReturnValue(true);
    jest.mocked(useQuery).mockReturnValue({
      data: {
        id: 8,
        subscriptionPlan: "free",
        notificationsEnabled: true,
      },
      isLoading: false,
      error: null,
    } as ReturnType<typeof useQuery>);
    jest
      .mocked(subscribeAuthenticationSessionChanges)
      .mockImplementationOnce((listener) => {
        notifySessionChange = listener;
        return jest.fn();
      });
    const { result } = renderHook(() => useUserSession());
    await waitFor(() => expect(result.current.user?.id).toBe(8));

    jest.mocked(isAuthenticationSessionReady).mockReturnValue(false);
    jest.mocked(getAuthenticationSessionEpoch).mockReturnValue(null);
    jest.mocked(hasAuthenticationSession).mockReturnValue(false);
    act(() => notifySessionChange(null));

    await waitFor(() => expect(result.current.isLoading).toBe(true));
    expect(result.current.user).toBeNull();
  });

  it("treats a missing credential as a normal empty session", async () => {
    jest.mocked(hasAuthenticationSession).mockReturnValue(false);
    await AsyncStorage.setItem(
      "cachedUserProfile",
      JSON.stringify({ id: 99, subscriptionPlan: "premium" }),
    );

    const { result } = renderHook(() => useUserSession());

    await act(async () => {});
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.user).toBeNull();
    expect(result.current.error).toBeNull();
    expect(hasAuthenticationSession).toHaveBeenCalled();
    expect(AsyncStorage.multiGet).not.toHaveBeenCalled();
    for (const [options] of jest.mocked(useQuery).mock.calls) {
      expect(options).toEqual(expect.objectContaining({ enabled: false }));
    }
  });

  it("reconciles an epoch change missed between render and subscription", async () => {
    jest.mocked(hasAuthenticationSession).mockReturnValue(true);
    jest
      .mocked(getAuthenticationSessionEpoch)
      .mockReturnValueOnce("identity-a")
      .mockReturnValue("identity-b");
    let resolveCleanup!: () => void;
    jest.mocked(clearSessionScopedCaches).mockImplementationOnce(
      () =>
        new Promise<void>((resolve) => {
          resolveCleanup = resolve;
        }),
    );

    const { result } = renderHook(() => useUserSession());

    expect(clearSessionScopedCaches).toHaveBeenCalledTimes(1);
    expect(result.current.isLoading).toBe(true);
    expect(jest.mocked(useQuery).mock.calls.slice(-2)).toEqual([
      [expect.objectContaining({ enabled: false })],
      [expect.objectContaining({ enabled: false })],
    ]);

    await act(async () => resolveCleanup());
    await waitFor(() =>
      expect(jest.mocked(useQuery).mock.calls.slice(-2)).toEqual([
        [expect.objectContaining({ enabled: true })],
        [expect.objectContaining({ enabled: true })],
      ]),
    );
    expect(result.current.isLoading).toBe(false);
  });

  it("remains loading when changed-session cache cleanup fails", async () => {
    const warning = jest.spyOn(console, "error").mockImplementation(() => {});
    jest.mocked(hasAuthenticationSession).mockReturnValue(true);
    jest
      .mocked(getAuthenticationSessionEpoch)
      .mockReturnValueOnce("identity-a")
      .mockReturnValue("identity-b");
    jest
      .mocked(clearSessionScopedCaches)
      .mockRejectedValueOnce(new Error("cleanup failed"));

    const { result } = renderHook(() => useUserSession());

    await waitFor(() => expect(clearSessionScopedCaches).toHaveBeenCalled());
    expect(result.current.isLoading).toBe(true);
    expect(result.current.user).toBeNull();
    warning.mockRestore();
  });
});
