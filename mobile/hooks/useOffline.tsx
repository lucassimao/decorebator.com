import { useEffect, useState } from "react";
import offlineManager from "@/utils/offlineManager";
import { useUserSession } from "@/hooks/useUserSession";

export function useOffline() {
  const [isOnline, setIsOnline] = useState(true);
  const [isOfflineAvailable, setIsOfflineAvailable] = useState(false);

  // Get premium status from centralized user session
  const { isPremium } = useUserSession();

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
    setIsOfflineAvailable(isPremium && !isOnline);
  }, [isPremium, isOnline]);

  return {
    isOnline,
    isOfflineAvailable,
    isPremium,
  };
}
