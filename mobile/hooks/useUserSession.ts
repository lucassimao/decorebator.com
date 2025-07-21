import { getProfile, type UserProfile } from "@/api/users";
import {
  getSubscriptionStatus,
  type SubscriptionStatus,
} from "@/api/subscriptions";
import offlineManager from "@/utils/offlineManager";
import * as Sentry from "@sentry/react-native";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";

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
  const userId = user?.id || null;
  const country = user?.country || null;
  const subscriptionPlan = user?.subscriptionPlan || "free";
  const isPremium = subscriptionPlan !== "free";
  const hasOptimisticSubscription =
    subscription?.hasOptimisticSubscription ?? false;
  const isLoading = userLoading || subscriptionLoading;
  const error = userError || subscriptionError;

  // Cache timing strategy based on subscription tier
  const cacheConfig = {
    dataFreshnessDuration: isPremium ? 10 * 1000 : 15 * 60 * 1000, // 10s vs 15min
    memoryRetentionTime: isPremium ? 2 * 60 * 1000 : 60 * 60 * 1000, // 2min vs 1hr
    autoRefreshOnFocus: isPremium, // Premium users get automatic refresh when returning to screen
    alwaysFetchOnMount: (isPremium ? "always" : false) as "always" | false, // Premium users always get fresh data on mount
  };

  // Essential side effects only
  useEffect(() => {
    if (user) {
      // Essential: Update offline manager for premium features
      offlineManager.setUserPremiumStatus(isPremium);

      // Essential: Set Sentry context for error tracking
      Sentry.setUser({
        id: user.id.toString(),
        email: user.email,
        username: `${user.firstName} ${user.lastName}`,
      });
    }
  }, [user, isPremium]);

  return {
    // User essentials
    user: user || null,
    userId,
    country,

    // Subscription essentials
    isPremium,
    subscriptionPlan,
    subscription,
    hasOptimisticSubscription,

    // Cache configuration for analytics
    cacheConfig,

    // Loading states
    isLoading,
    error,
  };
}
