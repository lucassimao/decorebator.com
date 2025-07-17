import { getProfile } from "@/api/users";
import offlineManager from "@/utils/offlineManager";
import * as Sentry from "@sentry/react-native";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef } from "react";

export const useUserInfo = () => {
  const queryClient = useQueryClient();
  const previousPlanRef = useRef<string | null>(null);

  const {
    data: user,
    isLoading,
    isSuccess,
    isError,
    error,
  } = useQuery({
    queryKey: ["userProfile"],
    queryFn: getProfile,
  });

  // derive isPremium whenever `user` changes
  const isPremium =
    !!user &&
    (user.subscriptionPlan === "monthly" || user.subscriptionPlan === "annual");

  // Cache timing strategy based on subscription tier
  // Premium users get fresher data for better UX after quiz sessions
  const cacheConfig = {
    dataFreshnessDuration: isPremium ? 10 * 1000 : 15 * 60 * 1000, // 10s vs 15min
    memoryRetentionTime: isPremium ? 2 * 60 * 1000 : 60 * 60 * 1000, // 2min vs 1hr
    autoRefreshOnFocus: isPremium, // Premium users get automatic refresh when returning to screen
    alwaysFetchOnMount: isPremium ? ("always" as const) : false, // Premium users always get fresh data on mount
  };

  useEffect(() => {
    if (user) {
      const currentPlan = user.subscriptionPlan;
      const previousPlan = previousPlanRef.current;

      // Check if subscription plan changed
      if (previousPlan !== null && previousPlan !== currentPlan) {
        const wasPremium =
          previousPlan === "monthly" || previousPlan === "annual";
        const isNowPremium =
          currentPlan === "monthly" || currentPlan === "annual";

        // If premium status changed, invalidate analytics caches
        if (wasPremium !== isNowPremium) {
          console.log(
            `Subscription plan changed from ${previousPlan} to ${currentPlan}, invalidating caches`,
          );

          // Invalidate all analytics queries (they all include isPremium in their key)
          queryClient.invalidateQueries({
            predicate: (query) => {
              const key = query.queryKey;
              return (
                Array.isArray(key) && key.length >= 3 && key[0] === "analytics"
                // All analytics queries will be invalidated since they depend on premium status
              );
            },
          });

          // Also invalidate wordlists cache as free users have limits
          queryClient.invalidateQueries({
            queryKey: ["wordlists"],
          });

          console.log(
            "Invalidated analytics and wordlists caches for premium status change",
          );
        }
      }

      // Update the previous plan reference
      previousPlanRef.current = currentPlan;

      offlineManager.setUserPremiumStatus(isPremium);
      // Set Sentry user context
      Sentry.setUser({
        id: user.id.toString(),
        email: user.email,
        username: `${user.firstName} ${user.lastName}`,
      });
    } else {
      // Handle logout: reset previous plan tracking
      previousPlanRef.current = null;
    }
  }, [user, queryClient, isPremium]);

  return {
    userInfo: user,
    data: user, // Alias for consistency with useQuery pattern
    loading: isLoading,
    isLoading,
    isSuccess,
    isError,
    error,
    isPremium,
    cacheConfig,
  };
};
