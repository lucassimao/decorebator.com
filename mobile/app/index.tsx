import { useUserInfo } from "@/hooks/users";
import { useTheme } from "@/contexts/ThemeContext";
import React, { useEffect } from "react";
import {
  ActivityIndicator,
  Dimensions,
  ImageBackground,
  SafeAreaView,
  StyleSheet,
  View,
} from "react-native";
import * as usersApi from "@/api/users";
import * as wordlistsApi from "@/api/wordlists";
import { useRouter } from "expo-router";
import { Redirect } from "expo-router";
import { useQuery } from "@tanstack/react-query";

export default function Index() {
  const { userInfo, error, loading } = useUserInfo();
  const router = useRouter();
  const { theme } = useTheme();

  // Prefetch wordlists when user is authenticated
  useQuery({
    queryKey: ["wordlists"],
    queryFn: () => wordlistsApi.getUserWordlists(),
    enabled: !!userInfo && !error,
    staleTime: 0,
  });

  useEffect(() => {
    if (error) {
      console.error(error);
      usersApi.sigout();
      router.replace("/signin");
    }
  }, [error, router]);

  // Show loading only during initial authentication check
  if (loading) {
    return (
      <>
        {theme.mode === "dark" ? (
          <SafeAreaView style={styles.containerDark}>
            <View style={styles.loadingContainer}>
              <ActivityIndicator size="large" color="#FF7B54" />
            </View>
          </SafeAreaView>
        ) : (
          <ImageBackground
            source={require("@/assets/images/signup-bg3.png")}
            style={styles.backgroundImage}
            imageStyle={styles.backgroundImageStyle}
            resizeMode="cover"
          >
            <SafeAreaView style={styles.container}>
              <View style={styles.loadingContainer}>
                <ActivityIndicator size="large" color="#FF7B54" />
              </View>
            </SafeAreaView>
          </ImageBackground>
        )}
      </>
    );
  }

  // If user is authenticated, redirect to dashboard without showing loading
  if (userInfo) {
    return <Redirect href="/dashboard" />;
  }

  // If not authenticated, redirect to signin
  usersApi.sigout();
  return <Redirect href="/signin" />;
}

const { width, height } = Dimensions.get("window");

const styles = StyleSheet.create({
  backgroundImage: {
    flex: 1,
    width: width,
    height: height,
    backgroundColor: "#FFF9F0", // Fallback warm background color
  },
  backgroundImageStyle: {
    opacity: 0.7, // Same opacity as dashboard
    width: "100%",
    height: "100%",
  },
  container: {
    flex: 1,
    backgroundColor: "transparent", // Let background image show through
  },
  containerDark: {
    flex: 1,
    backgroundColor: "#0F0F0F", // Dark theme background
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
});
