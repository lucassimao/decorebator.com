import { useEffect, useRef, useState } from "react";
import * as Notifications from "expo-notifications";
import { useGlobalSearchParams, usePathname, useRouter } from "expo-router";
import { usePostHog } from "posthog-react-native";

import {
  ACTIVATION_EVENT_NAMES,
  captureActivationEvent,
  normalizeNotificationType,
} from "@/utils/activationEvents";
import {
  getNotificationDestination,
  isAuthenticationPath,
  type NotificationDestination,
} from "@/utils/notificationRouting";

export function NotificationOpenTracker() {
  const posthog = usePostHog();
  const pathname = usePathname();
  const { wordlistId: currentWordlistId } = useGlobalSearchParams<{
    wordlistId?: string;
  }>();
  const router = useRouter();
  const handled = useRef(new Set<string>());
  const [pendingDestination, setPendingDestination] =
    useState<NotificationDestination | null>(null);

  useEffect(() => {
    if (
      !pendingDestination ||
      pathname === "/" ||
      isAuthenticationPath(pathname)
    ) {
      return;
    }
    setPendingDestination(null);
    if (pendingDestination.kind === "dashboard") {
      if (pathname !== "/dashboard") router.push("/dashboard");
      return;
    }

    if (
      pathname === "/quiz" &&
      currentWordlistId === pendingDestination.wordlistId
    ) {
      return;
    }
    router.push({
      pathname: "/quiz",
      params: { wordlistId: pendingDestination.wordlistId },
    });
  }, [currentWordlistId, pathname, pendingDestination, router]);

  useEffect(() => {
    let mounted = true;
    const capture = (response: Notifications.NotificationResponse | null) => {
      if (!mounted || !response) return;
      const request = response.notification.request;
      if (handled.current.has(request.identifier)) return;
      handled.current.add(request.identifier);
      setPendingDestination(getNotificationDestination(request.content.data));
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
