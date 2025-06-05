import FontAwesome from "@expo/vector-icons/FontAwesome";
import { useFonts } from "expo-font";
import { Stack } from "expo-router";
import * as SplashScreen from "expo-splash-screen";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect } from "react";
import { SnackbarProvider } from "@/hooks/useSnackbar";
import { UpgradePromptDialogProvider } from "@/hooks/useUpgradePromptDialog";
import { I18nextProvider } from "react-i18next";
import i18n from "@/i18n";
import { useI18n } from "@/hooks/useI18n";
import { usePostHog, PostHogProvider } from "posthog-react-native";
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
  initialRouteName: "signin",
};

// Prevent the splash screen from auto-hiding before asset loading is complete.
SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const [loaded, error] = useFonts({
    SpaceMono: require("../assets/fonts/SpaceMono-Regular.ttf"),
    ...FontAwesome.font,
  });

  // Expo Router uses Error Boundaries to catch errors in the navigation tree.
  useEffect(() => {
    if (error) throw error;
  }, [error]);

  useEffect(() => {
    if (loaded) {
      SplashScreen.hideAsync();
    }
  }, [loaded]);

  if (!loaded) {
    return null;
  }

  return <RootLayoutNav />;
}

const queryClient = new QueryClient();

function RootLayoutNav() {
  return (
    <I18nextProvider i18n={i18n}>
      <QueryClientProvider client={queryClient}>
        <LanguageInitializer />
        <SnackbarProvider>
          <UpgradePromptDialogProvider>
            <PostHogProvider
              apiKey={process.env.EXPO_PUBLIC_POSTHOG_KEY}
              options={{ host: "https://us.i.posthog.com", disabled: __DEV__ }}
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
                <Stack.Screen name="signup" options={{ headerShown: false }} />
                <Stack.Screen name="quiz" options={{ headerShown: false }} />
                <Stack.Screen name="signin" options={{ headerShown: false }} />
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
      </QueryClientProvider>
    </I18nextProvider>
  );
}

// Component to initialize language from user profile
function LanguageInitializer() {
  useI18n(); // This hook will handle language initialization
  return null;
}
