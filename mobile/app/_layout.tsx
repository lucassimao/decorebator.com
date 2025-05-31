import FontAwesome from "@expo/vector-icons/FontAwesome";
import { useFonts } from "expo-font";
import { Stack } from "expo-router";
import * as SplashScreen from "expo-splash-screen";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect } from "react";
import { SnackbarProvider } from "@/hooks/useSnackbar";
import { UpgradePromptDialogProvider } from "@/hooks/useUpgradePromptDialog";

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
    <QueryClientProvider client={queryClient}>
      <SnackbarProvider>
        <UpgradePromptDialogProvider>
          <Stack>
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
            <Stack.Screen name="settings" options={{ headerShown: false }} />
            <Stack.Screen
              name="profileSettings"
              options={{ headerShown: false }}
            />
          </Stack>
        </UpgradePromptDialogProvider>
      </SnackbarProvider>
    </QueryClientProvider>
  );
}
