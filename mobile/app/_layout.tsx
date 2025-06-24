import FontAwesome from "@expo/vector-icons/FontAwesome";
import { useFonts } from "expo-font";
import { Stack } from "expo-router";

import { useI18n } from "@/hooks/useI18n";
import { SnackbarProvider } from "@/hooks/useSnackbar";
import { UpgradePromptDialogProvider } from "@/hooks/useUpgradePromptDialog";
import { ThemeProvider } from "@/contexts/ThemeContext";
import i18n from "@/i18n";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { PostHogProvider } from "posthog-react-native";
import { useEffect } from "react";
import { I18nextProvider } from "react-i18next";
import * as Sentry from "@sentry/react-native";

Sentry.init({
  dsn: "https://1905051f98d938186e63c776dec05a68@o4509430877257728.ingest.us.sentry.io/4509553206099968",

  // Adds more context data to events (IP address, cookies, user, etc.)
  // For more information, visit: https://docs.sentry.io/platforms/react-native/data-management/data-collected/
  sendDefaultPii: true,

  // Configure Session Replay
  replaysSessionSampleRate: 0.1,
  replaysOnErrorSampleRate: 1,
  integrations: [Sentry.mobileReplayIntegration()],

  // uncomment the line below to enable Spotlight (https://spotlightjs.com)
  // spotlight: __DEV__,
});
// Import offline test utility in development
if (__DEV__) {
  import("@/utils/offlineTest");
}

export {
  // Catch any errors thrown by the Layout component.
  ErrorBoundary,
} from "expo-router";

export const unstable_settings = {
  // Ensure that reloading on `/modal` keeps a back button present.
  initialRouteName: "index",
};

export default Sentry.wrap(function RootLayout() {
  const [loaded, error] = useFonts({
    SpaceMono: require("../assets/fonts/SpaceMono-Regular.ttf"),
    ...FontAwesome.font,
  });

  // Expo Router uses Error Boundaries to catch errors in the navigation tree.
  useEffect(() => {
    if (error) throw error;
  }, [error]);

  if (!loaded) {
    return null;
  }

  return <RootLayoutNav />;
});

const queryClient = new QueryClient();

function RootLayoutNav() {
  return (
    <I18nextProvider i18n={i18n}>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          <LanguageInitializer />
          <SnackbarProvider>
            <UpgradePromptDialogProvider>
              <PostHogProvider
                apiKey={process.env.EXPO_PUBLIC_POSTHOG_KEY}
                options={{
                  host: "https://us.i.posthog.com",
                  disabled: __DEV__,
                }}
              >
                <Stack>
                  <Stack.Screen name="index" options={{ headerShown: false }} />
                  <Stack.Screen
                    name="analytics"
                    options={{ headerShown: false }}
                  />
                  <Stack.Screen
                    name="practice"
                    options={{ headerShown: false }}
                  />
                  <Stack.Screen
                    name="signup"
                    options={{ headerShown: false }}
                  />
                  <Stack.Screen name="quiz" options={{ headerShown: false }} />
                  <Stack.Screen
                    name="signin"
                    options={{ headerShown: false }}
                  />
                  <Stack.Screen
                    name="forgotPassword"
                    options={{ headerShown: false }}
                  />
                  <Stack.Screen
                    name="dashboard/index"
                    options={{
                      headerShown: false,
                      headerTitle: "Dashboard",
                    }}
                  />
                  <Stack.Screen
                    name="dashboard/welcome"
                    options={{
                      headerShown: false,
                    }}
                  />
                  <Stack.Screen
                    name="settings"
                    options={{ headerShown: false }}
                  />
                  <Stack.Screen
                    name="profileSettings"
                    options={{ headerShown: false }}
                  />
                </Stack>
              </PostHogProvider>
            </UpgradePromptDialogProvider>
          </SnackbarProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </I18nextProvider>
  );
}

// Component to initialize language from user profile
function LanguageInitializer() {
  useI18n(); // This hook will handle language initialization
  return null;
}
