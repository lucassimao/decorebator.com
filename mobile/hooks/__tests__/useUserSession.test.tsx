import AsyncStorage from "@react-native-async-storage/async-storage";
import { act, renderHook, waitFor } from "@testing-library/react-native";
import { useQuery } from "@tanstack/react-query";

import { useUserSession } from "@/hooks/useUserSession";
import { getAuthorization } from "@/api/users";

jest.mock("@/api/users", () => ({
  getAuthorization: jest.fn(),
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
}));

describe("useUserSession credential gate", () => {
  beforeEach(async () => {
    jest.clearAllMocks();
    await AsyncStorage.clear();
    jest.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useQuery>);
  });

  it("treats a missing credential as a normal empty session", async () => {
    jest.mocked(getAuthorization).mockReturnValue(null);
    await AsyncStorage.setItem(
      "cachedUserProfile",
      JSON.stringify({ id: 99, subscriptionPlan: "premium" }),
    );

    const { result } = renderHook(() => useUserSession());

    await act(async () => {});
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.user).toBeNull();
    expect(result.current.error).toBeNull();
    expect(AsyncStorage.multiGet).not.toHaveBeenCalled();
    for (const [options] of jest.mocked(useQuery).mock.calls) {
      expect(options).toEqual(expect.objectContaining({ enabled: false }));
    }
  });
});
