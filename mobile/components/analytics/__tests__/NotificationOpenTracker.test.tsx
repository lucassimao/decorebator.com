import { act, render } from "@testing-library/react-native";
import * as Notifications from "expo-notifications";
import type { NotificationResponse } from "expo-notifications";

import { NotificationOpenTracker } from "../NotificationOpenTracker";

const mockCapture = jest.fn();
const mockRemove = jest.fn();
let mockResponseListener: (response: NotificationResponse) => void;
let mockLastResponse: NotificationResponse | null = null;

jest.mock("posthog-react-native", () => ({
  usePostHog: () => ({ capture: mockCapture }),
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
          data: { type, wordlistName: "Private list" },
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
    jest.mocked(Notifications.clearLastNotificationResponse).mockClear();
    mockLastResponse = null;
  });

  it("captures a bounded notification kind once without notification content", async () => {
    const view = render(<NotificationOpenTracker />);
    await act(async () => {});

    const response = notificationResponse("request-1", "due_items_reminder");
    act(() => mockResponseListener(response));
    act(() => mockResponseListener(response));

    expect(mockCapture).toHaveBeenCalledTimes(1);
    expect(mockCapture).toHaveBeenCalledWith("notification_opened", {
      eventVersion: 1,
      source: "push_notification",
      notificationType: "due_items_reminder",
    });
    expect(JSON.stringify(mockCapture.mock.calls)).not.toContain("Private");
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
  });

  it("captures a cold-start response once when the listener replays it", async () => {
    const response = notificationResponse(
      "cold-request",
      "daily_practice_reminder",
    );
    mockLastResponse = response;

    render(<NotificationOpenTracker />);
    await act(async () => {});
    act(() => mockResponseListener(response));

    expect(mockCapture).toHaveBeenCalledTimes(1);
    expect(mockCapture).toHaveBeenCalledWith("notification_opened", {
      eventVersion: 1,
      source: "push_notification",
      notificationType: "daily_practice_reminder",
    });
    expect(Notifications.clearLastNotificationResponse).toHaveBeenCalledTimes(
      1,
    );
  });
});
