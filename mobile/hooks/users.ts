import { getProfile } from "@/api/users";
import offlineManager from "@/utils/offlineManager";
import { setUserContext } from "@/utils/sentry";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";

export const useUserInfo = () => {
  const {
    data: user,
    isLoading,
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
      offlineManager.setUserPremiumStatus(isPremium);
      // Set Sentry user context
      setUserContext({
        id: user.id.toString(),
        email: user.email,
        name: `${user.firstName} ${user.lastName}`,
      });
    }
  }, [user, isPremium]);

  return {
    userInfo: user,
    loading: isLoading,
    error,
    isPremium,
    cacheConfig,
  };
};
