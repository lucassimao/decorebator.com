import FontAwesome from "@expo/vector-icons/FontAwesome";
import { useFonts } from "expo-font";
import { Stack } from "expo-router";
import * as SplashScreen from "expo-splash-screen";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect } from "react";
import { PaperProvider } from "react-native-paper";

import {
  DarkTheme as NavigationDarkTheme,
  DefaultTheme as NavigationDefaultTheme,
} from "@react-navigation/native";
import { adaptNavigationTheme } from "react-native-paper";

import * as colorSchemes from "@/constants/colorSchemes";

const { LightTheme, DarkTheme } = adaptNavigationTheme({
  reactNavigationLight: NavigationDefaultTheme,
  reactNavigationDark: NavigationDarkTheme,
});

import { MD3DarkTheme, MD3LightTheme } from "react-native-paper";

const CombinedDefaultTheme = {
  ...MD3LightTheme,
  ...LightTheme,
  colors: {
    ...MD3LightTheme.colors,
    ...LightTheme.colors,
    ...colorSchemes.lightThemeColors,

    // ✅ Primary brand colors
    primary: "#2A5D4D",
    background: "#FFF9F1",
    surface: "#FFFFFF",
    text: "#1D3B29",

    error: "#c60000", // Darker red for text, icon, and borders
    onError: "#c60000",
    errorContainer: "#C60000", // Light red-ish for container if needed
    onErrorContainer: "#C60000",

    // // ✅ Outline colors (borders)
    outline: "#E0E0E0", // Default border color when not focused
  },
  fonts: {
    ...MD3LightTheme.fonts,
    ...LightTheme.fonts,
  },
};
const CombinedDarkTheme = {
  ...MD3DarkTheme,
  ...DarkTheme,
  colors: {
    ...MD3DarkTheme.colors,
    ...DarkTheme.colors,
    ...colorSchemes.darkThemeColors,
  },
  fonts: {
    ...MD3DarkTheme.fonts,
    ...DarkTheme.fonts,
  },
};

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
    <PaperProvider theme={CombinedDefaultTheme}>
      <QueryClientProvider client={queryClient}>
        <Stack>
          <Stack.Screen name="signup" options={{ headerShown: false }} />
          <Stack.Screen name="signin" options={{ headerShown: false }} />
          <Stack.Screen name="resetPassword" options={{ headerShown: false }} />
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
            name="subscription"
            options={{
              headerShown: true,
              headerTitle: "Subscription",
            }}
          />
        </Stack>
      </QueryClientProvider>
    </PaperProvider>
  );
}
