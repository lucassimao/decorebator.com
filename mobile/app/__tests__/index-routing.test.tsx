import AsyncStorage from "@react-native-async-storage/async-storage";
import { render, waitFor } from "@testing-library/react-native";
import * as Sentry from "@sentry/react-native";

import Index from "@/app/index";

const mockReplace = jest.fn();
const mockHideSplash = jest.fn(() => Promise.resolve());
const mockPrefetchWordlists = jest.fn(() => Promise.resolve());
let mockSession: {
  user: { id: number } | null;
  isLoading: boolean;
  error: Error | null;
};

jest.mock("expo-router", () => ({
  SplashScreen: { hideAsync: () => mockHideSplash() },
  useRouter: () => ({ replace: mockReplace }),
}));

jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({ id: "query-client" }),
}));

jest.mock("@/hooks/useUserSession", () => ({
  useUserSession: () => mockSession,
}));

jest.mock("@/hooks/useWordlists", () => ({
  prefetchWordlists: () => mockPrefetchWordlists(),
}));

describe("initial app routing", () => {
  beforeEach(async () => {
    jest.clearAllMocks();
    await AsyncStorage.clear();
    mockSession = { user: null, isLoading: false, error: null };
  });

  it("sends a new unauthenticated install to onboarding", async () => {
    render(<Index />);

    await waitFor(() =>
      expect(mockReplace).toHaveBeenCalledWith("/onboarding"),
    );
    expect(AsyncStorage.getItem).toHaveBeenCalledWith(
      "@decorebator/onboarding:v1:seen",
    );
    expect(mockHideSplash).toHaveBeenCalledTimes(1);
  });

  it("sends an onboarding-seen unauthenticated install to sign in", async () => {
    await AsyncStorage.setItem("@decorebator/onboarding:v1:seen", "true");

    render(<Index />);

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith("/signin"));
  });

  it("fails safely to onboarding when first-launch storage cannot be read", async () => {
    const warning = jest.spyOn(console, "warn").mockImplementation(() => {});
    jest
      .mocked(AsyncStorage.getItem)
      .mockRejectedValueOnce(new Error("unavailable"));

    render(<Index />);

    await waitFor(() =>
      expect(mockReplace).toHaveBeenCalledWith("/onboarding"),
    );
    expect(warning).toHaveBeenCalledWith(
      "Failed to read onboarding state:",
      expect.any(Error),
    );
    warning.mockRestore();
  });

  it("prefetches and routes an authenticated user to the dashboard", async () => {
    mockSession = { user: { id: 42 }, isLoading: false, error: null };

    render(<Index />);

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith("/dashboard"));
    expect(mockPrefetchWordlists).toHaveBeenCalledTimes(1);
    expect(AsyncStorage.getItem).not.toHaveBeenCalled();
  });

  it("keeps cached/authenticated access while reporting a partial session error", async () => {
    const error = new Error("subscription unavailable");
    mockSession = { user: { id: 42 }, isLoading: false, error };

    render(<Index />);

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith("/dashboard"));
    expect(Sentry.captureException).toHaveBeenCalledWith(error);
  });

  it("routes an explicit session error to sign in and reports it", async () => {
    const error = new Error("session rejected");
    mockSession = { user: null, isLoading: false, error };

    render(<Index />);

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith("/signin"));
    expect(Sentry.captureException).toHaveBeenCalledWith(error);
    expect(AsyncStorage.getItem).not.toHaveBeenCalled();
  });

  it("waits for session hydration and navigates only once", async () => {
    mockSession = { user: null, isLoading: true, error: null };
    const view = render(<Index />);
    expect(mockReplace).not.toHaveBeenCalled();

    mockSession = { user: null, isLoading: false, error: null };
    view.rerender(<Index />);

    await waitFor(() => expect(mockReplace).toHaveBeenCalledTimes(1));

    mockSession = { user: { id: 7 }, isLoading: false, error: null };
    view.rerender(<Index />);
    await waitFor(() => expect(mockReplace).toHaveBeenCalledTimes(1));
    expect(mockPrefetchWordlists).not.toHaveBeenCalled();
  });
});
