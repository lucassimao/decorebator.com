import { useUserSession } from "@/hooks/useUserSession";
import { prefetchWordlists } from "@/hooks/useWordlists";
import { SplashScreen, useRouter } from "expo-router";
import { useEffect, useRef } from "react";
import * as Sentry from "@sentry/react-native";
import { useQueryClient } from "@tanstack/react-query";
import { getInitialRoute, readOnboardingSeen } from "@/utils/launchRouting";

export default function Index() {
  const { user, isLoading, error } = useUserSession();
  const isError = !!error;
  const router = useRouter();
  const queryClient = useQueryClient();
  const navigationStarted = useRef(false);

  useEffect(() => {
    // Still waiting for the user session, do nothing.
    if (isLoading || navigationStarted.current) {
      return;
    }
    navigationStarted.current = true;

    // We have a result, prepare to navigate
    const navigate = async () => {
      let onboardingSeen = false;
      if (!user && !isError) {
        try {
          onboardingSeen = await readOnboardingSeen();
        } catch (storageError) {
          // A storage failure must not hide onboarding from a new install.
          console.warn("Failed to read onboarding state:", storageError);
        }
      }

      const route = getInitialRoute({
        isAuthenticated: Boolean(user),
        hasAuthError: isError,
        onboardingSeen,
      });

      try {
        // Hide the splash screen just before navigation
        await SplashScreen.hideAsync();
      } catch (e) {
        // Splash screen might already be hidden, ignore
        console.warn("Error hiding splash screen:", e);
      }

      if (error) {
        Sentry.captureException(error);
      }

      if (route === "/dashboard") {
        // User is authenticated
        // Start prefetching wordlists in the background (non-blocking)
        prefetchWordlists(queryClient).catch((err) => {
          // Don't block navigation if prefetch fails
          console.warn("Failed to prefetch wordlists:", err);
        });
      }

      router.replace(route);
    };

    navigate();
  }, [isLoading, isError, user, router, queryClient, error]);

  // Render nothing while the splash screen is visible
  return null;
}
