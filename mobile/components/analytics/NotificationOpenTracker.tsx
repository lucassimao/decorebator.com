import { useEffect, useRef } from "react";
import * as Notifications from "expo-notifications";
import { usePostHog } from "posthog-react-native";

import {
  ACTIVATION_EVENT_NAMES,
  captureActivationEvent,
  normalizeNotificationType,
} from "@/utils/activationEvents";

export function NotificationOpenTracker() {
  const posthog = usePostHog();
  const handled = useRef(new Set<string>());

  useEffect(() => {
    let mounted = true;
    const capture = (response: Notifications.NotificationResponse | null) => {
      if (!mounted || !response) return;
      const request = response.notification.request;
      if (handled.current.has(request.identifier)) return;
      handled.current.add(request.identifier);
      captureActivationEvent(
        posthog,
        ACTIVATION_EVENT_NAMES.NOTIFICATION_OPENED,
        {
          source: "push_notification",
          notificationType: normalizeNotificationType(
            request.content.data?.type,
          ),
        },
      );
      try {
        Notifications.clearLastNotificationResponse();
      } catch {
        // Some development runtimes do not expose the native clear operation.
      }
    };

    void Notifications.getLastNotificationResponseAsync()
      .then(capture)
      .catch(() => undefined);
    const subscription =
      Notifications.addNotificationResponseReceivedListener(capture);

    return () => {
      mounted = false;
      subscription.remove();
    };
  }, [posthog]);

  return null;
}
