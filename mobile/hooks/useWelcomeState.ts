import { useState, useEffect, useCallback } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";

interface UseWelcomeStateProps {
  hasNoWordlist: boolean;
  isLoading: boolean;
}

export const useWelcomeState = ({
  hasNoWordlist,
  isLoading,
}: UseWelcomeStateProps) => {
  const [showWelcomeOverlay, setShowWelcomeOverlay] = useState(false);
  const [isNewUser, setIsNewUser] = useState(false);
  const [hasSeenDashboardFlag, setHasSeenDashboardFlag] = useState(false);
  const [hasCheckedWelcomeState, setHasCheckedWelcomeState] = useState(false);

  // Check if user is first-time user immediately on mount
  useEffect(() => {
    let cancelled = false;
    const checkFirstTimeUser = async () => {
      try {
        const entries = await AsyncStorage.multiGet([
          "justSignedUp",
          "hasSeenDashboard",
        ]);
        const isJustSignedUp = entries.find(
          ([key]) => key === "justSignedUp",
        )?.[1];
        const hasSeenDashboard = entries.find(
          ([key]) => key === "hasSeenDashboard",
        )?.[1];

        if (!cancelled) {
          setHasSeenDashboardFlag(hasSeenDashboard === "true");
        }

        if (__DEV__) {
          console.log("Dashboard welcome check (immediate):", {
            isJustSignedUp,
            hasSeenDashboard,
          });
        }

        if (isJustSignedUp && !cancelled) {
          setIsNewUser(true);
          setShowWelcomeOverlay(true);

          await AsyncStorage.removeItem("justSignedUp");
          await AsyncStorage.setItem("hasSeenDashboard", "true");

          if (!cancelled) {
            setHasSeenDashboardFlag(true);
          }
        }
      } catch (error) {
        console.warn("Error checking first-time user status:", error);
      } finally {
        if (!cancelled) {
          setHasCheckedWelcomeState(true);
        }
      }
    };

    checkFirstTimeUser();

    return () => {
      cancelled = true;
    };
  }, []);

  const markDashboardSeen = useCallback(() => {
    setHasSeenDashboardFlag(true);
    AsyncStorage.setItem("hasSeenDashboard", "true").catch((error) => {
      console.warn("Failed to persist dashboard seen flag:", error);
    });
  }, []);

  // Fallback: Show welcome for empty wordlists if no welcome was shown yet
  useEffect(() => {
    if (
      !hasCheckedWelcomeState ||
      isLoading ||
      !hasNoWordlist ||
      showWelcomeOverlay ||
      isNewUser ||
      hasSeenDashboardFlag
    ) {
      return;
    }

    const timer = setTimeout(() => {
      if (__DEV__) {
        console.log(
          "Fallback: Showing welcome for empty wordlist (persisting dismissal after first display)",
        );
      }
      setShowWelcomeOverlay(true);
      markDashboardSeen();
    }, 1000);

    return () => clearTimeout(timer);
  }, [
    hasCheckedWelcomeState,
    isLoading,
    hasNoWordlist,
    showWelcomeOverlay,
    isNewUser,
    hasSeenDashboardFlag,
    markDashboardSeen,
  ]);

  const handleWelcomeDismiss = useCallback(() => {
    setShowWelcomeOverlay(false);
    markDashboardSeen();
  }, [markDashboardSeen]);

  const handleWelcomeGetStarted = useCallback(
    (onGetStarted: () => void) => {
      handleWelcomeDismiss();
      onGetStarted();
    },
    [handleWelcomeDismiss],
  );

  return {
    showWelcomeOverlay,
    isNewUser,
    setIsNewUser,
    hasSeenDashboardFlag,
    handleWelcomeDismiss,
    handleWelcomeGetStarted,
  };
};
