import {
  hasAuthenticationSession,
  isAuthenticationSessionReady,
  clearSessionScopedCaches,
  getAuthenticationSessionEpoch,
  getProfile,
  persistSessionSnapshot,
  subscribeAuthenticationSessionChanges,
  type UserProfile,
} from "@/api/users";
import {
  getSubscriptionStatus,
  type SubscriptionStatus,
} from "@/api/subscriptions";
import offlineManager from "@/utils/offlineManager";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import * as Sentry from "@sentry/react-native";
import AsyncStorage from "@react-native-async-storage/async-storage";
import {
  registerDevicePushToken,
  resumeStoredPushTokenDeactivation,
} from "@/utils/pushNotifications";

interface UserSessionData {
  // User essentials
  user: UserProfile | null;
  userId: number | null;
  country: string | null;

  // Subscription essentials
  isPremium: boolean;
  subscriptionPlan: "free" | "monthly" | "annual";
  subscription: SubscriptionStatus | null | undefined;
  hasOptimisticSubscription: boolean;

  // Cache configuration for analytics
  cacheConfig: {
    dataFreshnessDuration: number;
    memoryRetentionTime: number;
    autoRefreshOnFocus: boolean;
    alwaysFetchOnMount: "always" | false;
  };

  // Loading states
  isLoading: boolean;
  error: Error | null;
}

export function useUserSession(): UserSessionData {
  const queryClient = useQueryClient();
  const [sessionEpoch, setSessionEpoch] = useState(
    getAuthenticationSessionEpoch,
  );
  const [sessionReady, setSessionReady] = useState(
    isAuthenticationSessionReady,
  );
  const sessionEpochRef = useRef(sessionEpoch);
  const sessionTransitionRef = useRef(0);
  const hasCurrentAuthentication = hasAuthenticationSession();
  const hasAuthorization = Boolean(sessionEpoch) && hasCurrentAuthentication;
  const [cachedUser, setCachedUser] = useState<UserProfile | null>(null);
  const [cachedSubscription, setCachedSubscription] =
    useState<SubscriptionStatus | null>(null);

  useEffect(() => {
    let active = true;
    const handleSessionChange = (nextEpoch: string | null) => {
      const authenticationReady = isAuthenticationSessionReady();
      if (nextEpoch === sessionEpochRef.current) {
        setSessionReady(authenticationReady);
        return;
      }
      const transition = ++sessionTransitionRef.current;
      setSessionReady(false);
      sessionEpochRef.current = null;
      setSessionEpoch(null);
      queryClient.clear();
      setCachedUser(null);
      setCachedSubscription(null);
      offlineManager.setUserPremiumStatus(false);
      Sentry.setUser(null);
      clearSessionScopedCaches()
        .then(() => {
          if (
            active &&
            transition === sessionTransitionRef.current &&
            getAuthenticationSessionEpoch() === nextEpoch
          ) {
            sessionEpochRef.current = nextEpoch;
            setSessionEpoch(nextEpoch);
            setSessionReady(authenticationReady);
          }
        })
        .catch((error) => {
          // Fail closed: never enable a new identity while data from the
          // previous identity may still be durable on the device.
          console.error("Error clearing changed browser session cache:", error);
        });
    };
    const unsubscribe =
      subscribeAuthenticationSessionChanges(handleSessionChange);
    // Close the render-to-effect race: an epoch may have changed before the
    // listener was installed, in which case no storage event can be replayed.
    handleSessionChange(getAuthenticationSessionEpoch());
    return () => {
      active = false;
      unsubscribe();
    };
  }, [queryClient]);

  useEffect(() => {
    let cancelled = false;

    const loadCachedSession = async () => {
      if (!hasAuthorization) {
        setCachedUser(null);
        setCachedSubscription(null);
        return;
      }
      try {
        const [userEntry, subscriptionEntry] = await AsyncStorage.multiGet([
          "cachedUserProfile",
          "cachedSubscription",
        ]);
        const userValue = userEntry?.[1];
        const subscriptionValue = subscriptionEntry?.[1];

        if (!cancelled) {
          setCachedUser(userValue ? JSON.parse(userValue) : null);
          setCachedSubscription(
            subscriptionValue ? JSON.parse(subscriptionValue) : null,
          );
        }
      } catch (error) {
        console.warn("Failed to load cached session:", error);
      }
    };

    loadCachedSession();

    return () => {
      cancelled = true;
    };
  }, [hasAuthorization]);

  // Single user profile query
  const {
    data: user,
    isLoading: userLoading,
    error: userError,
  } = useQuery({
    queryKey: ["userProfile"],
    queryFn: getProfile,
    enabled: hasAuthorization,
    staleTime: 5 * 60 * 1000, // 5 minutes - matches subscription query
  });

  // Single subscription query
  const {
    data: subscription,
    isLoading: subscriptionLoading,
    error: subscriptionError,
  } = useQuery({
    queryKey: ["subscription"],
    queryFn: getSubscriptionStatus,
    enabled: hasAuthorization,
    staleTime: 5 * 60 * 1000, // 5 minutes - allows optimistic data to persist
  });

  // Derive essential data
  const effectiveUser = hasAuthorization ? user ?? cachedUser : null;
  const effectiveSubscription = hasAuthorization
    ? subscription ?? cachedSubscription
    : null;
  const userId = effectiveUser?.id || null;
  const country = effectiveUser?.country || null;
  const subscriptionPlan = effectiveUser?.subscriptionPlan || "free";
  const isPremium = subscriptionPlan !== "free";
  const hasOptimisticSubscription =
    effectiveSubscription?.hasOptimisticSubscription ?? false;
  const isLoading =
    !sessionReady ||
    (hasAuthorization &&
      (userLoading || subscriptionLoading) &&
      !effectiveUser);
  const isOnline = offlineManager.getNetworkStatus();
  const error = hasAuthorization
    ? !isOnline && effectiveUser
      ? null
      : userError || subscriptionError
    : null;

  // Cache timing strategy based on subscription tier
  const cacheConfig = {
    dataFreshnessDuration: isPremium ? 10 * 1000 : 15 * 60 * 1000, // 10s vs 15min
    memoryRetentionTime: isPremium ? 2 * 60 * 1000 : 60 * 60 * 1000, // 2min vs 1hr
    autoRefreshOnFocus: isPremium, // Premium users get automatic refresh when returning to screen
    alwaysFetchOnMount: (isPremium ? "always" : false) as "always" | false, // Premium users always get fresh data on mount
  };

  // Essential side effects only
  useEffect(() => {
    if (effectiveUser) {
      persistSessionSnapshot(
        effectiveUser,
        effectiveSubscription,
        sessionEpoch ?? "",
      ).catch((error) => {
        console.warn("Failed to persist cached session:", error);
      });

      // Essential: Update offline manager for premium features
      offlineManager.setUserPremiumStatus(isPremium);

      // Essential: Set Sentry context for error tracking
      Sentry.setUser({
        id: effectiveUser.id.toString(),
      });
    }
  }, [effectiveUser, effectiveSubscription, isPremium, sessionEpoch]);

  useEffect(() => {
    if (!effectiveUser) {
      return;
    }
    if (!effectiveUser.notificationsEnabled) {
      resumeStoredPushTokenDeactivation();
      return;
    }
    registerDevicePushToken({ prompt: false }).catch((error) => {
      console.warn("Failed to auto-register push token:", error);
      Sentry.captureException(error, {
        data: {
          action: "auto_register_push_token",
          userId: effectiveUser?.id,
          platform: "mobile",
        },
      });
    });
  }, [effectiveUser]);

  return {
    // User essentials
    user: effectiveUser || null,
    userId,
    country,

    // Subscription essentials
    isPremium,
    subscriptionPlan,
    subscription: effectiveSubscription,
    hasOptimisticSubscription,

    // Cache configuration for analytics
    cacheConfig,

    // Loading states
    isLoading,
    error,
  };
}
