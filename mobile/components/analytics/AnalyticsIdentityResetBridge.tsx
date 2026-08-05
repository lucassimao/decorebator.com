import { useEffect } from "react";
import { usePostHog } from "posthog-react-native";

import {
  resetAnalyticsIdentity,
  subscribeAnalyticsIdentityReset,
} from "@/utils/activationEvents";

export function AnalyticsIdentityResetBridge() {
  const posthog = usePostHog();

  useEffect(
    () =>
      subscribeAnalyticsIdentityReset(() => resetAnalyticsIdentity(posthog)),
    [posthog],
  );

  return null;
}
