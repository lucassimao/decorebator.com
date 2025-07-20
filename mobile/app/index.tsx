import { useUserSession } from "@/hooks/useUserSession";
import { prefetchWordlists } from "@/hooks/useWordlists";
import { SplashScreen, useRouter } from "expo-router";
import { useEffect } from "react";
import * as Sentry from "@sentry/react-native";
import { useQueryClient } from "@tanstack/react-query";

export default function Index() {
  const { user, isLoading, error } = useUserSession();
  const isError = !!error;
  const router = useRouter();
  const queryClient = useQueryClient();

  useEffect(() => {
    // Still waiting for the user session, do nothing.
    if (isLoading) {
      return;
    }

    // We have a result, prepare to navigate
    const navigate = async () => {
      try {
        // Hide the splash screen just before navigation
        await SplashScreen.hideAsync();
      } catch (e) {
        // Splash screen might already be hidden, ignore
        console.warn("Error hiding splash screen:", e);
      }

      if (isError || !user) {
        // User is not authenticated
        if (isError && error) {
          // Log authentication errors to Sentry for debugging
          Sentry.captureException(error);
        }
        router.replace("/signin");
      } else {
        // User is authenticated
        // Start prefetching wordlists in the background (non-blocking)
        prefetchWordlists(queryClient).catch((err) => {
          // Don't block navigation if prefetch fails
          console.warn("Failed to prefetch wordlists:", err);
        });

        router.replace("/dashboard");
      }
    };

    navigate();
  }, [isLoading, isError, user, router, queryClient, error]);

  // Render nothing while the splash screen is visible
  return null;
}
