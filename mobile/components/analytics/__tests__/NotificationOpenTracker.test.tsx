import { act, render } from "@testing-library/react-native";
import * as Notifications from "expo-notifications";
import type { NotificationResponse } from "expo-notifications";

import { NotificationOpenTracker } from "../NotificationOpenTracker";

const mockCapture = jest.fn();
const mockPosthog = { capture: mockCapture };
const mockRemove = jest.fn();
const mockPush = jest.fn();
const mockRouter = { push: mockPush };
let mockResponseListener: (response: NotificationResponse) => void;
let mockLastResponse: NotificationResponse | null = null;
let mockPathname = "/dashboard";
let mockCurrentWordlistId: string | undefined;

jest.mock("posthog-react-native", () => ({
  usePostHog: () => mockPosthog,
}));

jest.mock("expo-router", () => ({
  usePathname: () => mockPathname,
  useGlobalSearchParams: () => ({ wordlistId: mockCurrentWordlistId }),
  useRouter: () => mockRouter,
}));

jest.mock("expo-notifications", () => ({
  getLastNotificationResponseAsync: jest.fn(() =>
    Promise.resolve(mockLastResponse),
  ),
  clearLastNotificationResponse: jest.fn(),
  addNotificationResponseReceivedListener: jest.fn(
    (listener: (response: NotificationResponse) => void) => {
      mockResponseListener = listener;
      return { remove: mockRemove };
    },
  ),
}));

function notificationResponse(
  identifier: string,
  type: string,
  data: Record<string, unknown> = {},
): NotificationResponse {
  return {
    actionIdentifier: "default",
    notification: {
      date: 0,
      request: {
        identifier,
        trigger: null,
        content: {
          title: "Private title",
          body: "Private body",
          data: { type, wordlistName: "Private list", ...data },
          sound: null,
        },
      },
    },
  } as unknown as NotificationResponse;
}

describe("NotificationOpenTracker", () => {
  beforeEach(() => {
    mockCapture.mockClear();
    mockRemove.mockClear();
    mockPush.mockClear();
    jest.mocked(Notifications.clearLastNotificationResponse).mockClear();
    mockLastResponse = null;
    mockPathname = "/dashboard";
    mockCurrentWordlistId = undefined;
  });

  it("captures a bounded notification kind once without notification content", async () => {
    const view = render(<NotificationOpenTracker />);
    await act(async () => {});

    const response = notificationResponse("request-1", "due_items_reminder", {
      wordlistId: 12,
    });
    act(() => mockResponseListener(response));
    act(() => mockResponseListener(response));

    expect(mockCapture).toHaveBeenCalledTimes(1);
    expect(mockCapture).toHaveBeenCalledWith("notification_opened", {
      eventVersion: 1,
      source: "push_notification",
      notificationType: "due_items_reminder",
    });
    expect(JSON.stringify(mockCapture.mock.calls)).not.toContain("Private");
    expect(mockPush).toHaveBeenCalledTimes(1);
    expect(mockPush).toHaveBeenCalledWith({
      pathname: "/quiz",
      params: { wordlistId: "12" },
    });
    view.unmount();
    expect(mockRemove).toHaveBeenCalledTimes(1);
  });

  it("collapses unknown notification data to a safe value", async () => {
    render(<NotificationOpenTracker />);
    await act(async () => {});

    act(() =>
      mockResponseListener(notificationResponse("request-2", "private-value")),
    );

    expect(mockCapture).toHaveBeenCalledWith("notification_opened", {
      eventVersion: 1,
      source: "push_notification",
      notificationType: "unknown",
    });
    expect(mockPush).not.toHaveBeenCalled();
  });

  it("defers a cold-start route until the index gate finishes", async () => {
    mockPathname = "/";
    const response = notificationResponse(
      "cold-request",
      "daily_practice_reminder",
    );
    mockLastResponse = response;

    const view = render(<NotificationOpenTracker />);
    await act(async () => {});
    act(() => mockResponseListener(response));

    expect(mockPush).not.toHaveBeenCalled();
    mockPathname = "/quiz";
    view.rerender(<NotificationOpenTracker />);
    await act(async () => {});

    expect(mockCapture).toHaveBeenCalledTimes(1);
    expect(mockCapture).toHaveBeenCalledWith("notification_opened", {
      eventVersion: 1,
      source: "push_notification",
      notificationType: "daily_practice_reminder",
    });
    expect(Notifications.clearLastNotificationResponse).toHaveBeenCalledTimes(
      1,
    );
    expect(mockPush).toHaveBeenCalledWith("/dashboard");
  });

  it("does not enter a protected route while authentication is visible", async () => {
    mockPathname = "/";
    const view = render(<NotificationOpenTracker />);
    await act(async () => {});

    act(() =>
      mockResponseListener(
        notificationResponse("signed-out", "due_items_reminder", {
          wordlistId: 12,
        }),
      ),
    );

    expect(mockPush).not.toHaveBeenCalled();
    mockPathname = "/signin";
    view.rerender(<NotificationOpenTracker />);
    expect(mockPush).not.toHaveBeenCalled();
    mockPathname = "/dashboard";
    view.rerender(<NotificationOpenTracker />);
    await act(async () => {});
    expect(mockPush).toHaveBeenCalledWith({
      pathname: "/quiz",
      params: { wordlistId: "12" },
    });
  });

  it("does not navigate for a malformed due-items payload", async () => {
    render(<NotificationOpenTracker />);
    await act(async () => {});

    act(() =>
      mockResponseListener(
        notificationResponse("invalid", "due_items_reminder", {
          wordlistId: "1/quiz",
        }),
      ),
    );

    expect(mockPush).not.toHaveBeenCalled();
  });

  it("does not remount the matching quiz for a repeated destination", async () => {
    mockPathname = "/quiz";
    mockCurrentWordlistId = "12";
    render(<NotificationOpenTracker />);
    await act(async () => {});

    act(() =>
      mockResponseListener(
        notificationResponse("same-quiz", "due_items_reminder", {
          wordlistId: 12,
        }),
      ),
    );

    expect(mockPush).not.toHaveBeenCalled();
  });
});
