import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import offlineManager from "@/utils/offlineManager";
import * as userApi from "@/api/users";

export function useOffline() {
  const [isOnline, setIsOnline] = useState(true);
  const [isOfflineAvailable, setIsOfflineAvailable] = useState(false);

  // Get user profile to check premium status
  const { data: profile } = useQuery({
    queryKey: ["userProfile"],
    queryFn: userApi.getProfile,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  useEffect(() => {
    // Set initial network status
    setIsOnline(offlineManager.getNetworkStatus());

    // Subscribe to network changes
    const unsubscribe = offlineManager.subscribeToNetworkChanges((online) => {
      setIsOnline(online);
    });

    return unsubscribe;
  }, []);

  useEffect(() => {
    // Update premium status in offline manager
    const isPremium =
      profile?.subscriptionPlan === "monthly" ||
      profile?.subscriptionPlan === "annual";
    offlineManager.setUserPremiumStatus(isPremium);
    setIsOfflineAvailable(isPremium && !isOnline);
  }, [profile?.subscriptionPlan, isOnline]);

  return {
    isOnline,
    isOfflineAvailable,
    isPremium:
      profile?.subscriptionPlan === "monthly" ||
      profile?.subscriptionPlan === "annual",
  };
}
