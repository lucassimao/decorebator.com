import { getProfile, type UserProfile } from "@/api/users";
import {
  getSubscriptionStatus,
  type SubscriptionStatus,
} from "@/api/subscriptions";
import offlineManager from "@/utils/offlineManager";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import * as Sentry from "@sentry/react-native";
import AsyncStorage from "@react-native-async-storage/async-storage";

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
  const [cachedUser, setCachedUser] = useState<UserProfile | null>(null);
  const [cachedSubscription, setCachedSubscription] =
    useState<SubscriptionStatus | null>(null);

  useEffect(() => {
    let cancelled = false;

    const loadCachedSession = async () => {
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
  }, []);

  // Single user profile query
  const {
    data: user,
    isLoading: userLoading,
    error: userError,
  } = useQuery({
    queryKey: ["userProfile"],
    queryFn: getProfile,
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
    staleTime: 5 * 60 * 1000, // 5 minutes - allows optimistic data to persist
  });

  // Derive essential data
  const effectiveUser = user ?? cachedUser;
  const effectiveSubscription = subscription ?? cachedSubscription;
  const userId = effectiveUser?.id || null;
  const country = effectiveUser?.country || null;
  const subscriptionPlan = effectiveUser?.subscriptionPlan || "free";
  const isPremium = subscriptionPlan !== "free";
  const hasOptimisticSubscription =
    effectiveSubscription?.hasOptimisticSubscription ?? false;
  const isLoading = (userLoading || subscriptionLoading) && !effectiveUser;
  const isOnline = offlineManager.getNetworkStatus();
  const error =
    !isOnline && effectiveUser ? null : userError || subscriptionError;

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
      AsyncStorage.multiSet([
        ["cachedUserProfile", JSON.stringify(effectiveUser)],
        ["cachedSubscription", JSON.stringify(effectiveSubscription)],
      ]).catch((error) => {
        console.warn("Failed to persist cached session:", error);
      });

      // Essential: Update offline manager for premium features
      offlineManager.setUserPremiumStatus(isPremium);

      // Essential: Set Sentry context for error tracking
      Sentry.setUser({
        id: effectiveUser.id.toString(),
        email: effectiveUser.email,
        username: `${effectiveUser.firstName} ${effectiveUser.lastName}`,
      });
    }
  }, [effectiveUser, effectiveSubscription, isPremium]);

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
