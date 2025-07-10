import { useUserInfo } from "@/hooks/users";
import { useTheme } from "@/contexts/ThemeContext";
import React, { useEffect } from "react";
import {
  ActivityIndicator,
  Dimensions,
  ImageBackground,
  SafeAreaView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import * as usersApi from "@/api/users";
import * as wordlistsApi from "@/api/wordlists";
import { useRouter } from "expo-router";
import { Redirect } from "expo-router";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { SkeletonDashboard } from "@/components/ui/SkeletonDashboard";

export default function Index() {
  const { userInfo, error, loading } = useUserInfo();
  const router = useRouter();
  const { theme } = useTheme();
  const { t } = useTranslation();

  // Track if we should show skeleton for authenticated users
  const [showDashboardSkeleton, setShowDashboardSkeleton] =
    React.useState(false);

  React.useEffect(() => {
    const checkSkeletonNeed = async () => {
      if (!loading && userInfo && !error) {
        // Check if user just signed up - if so, skip skeleton and go directly to dashboard
        try {
          const isJustSignedUp = await AsyncStorage.getItem("justSignedUp");
          if (isJustSignedUp) {
            // Just signed up users go directly to dashboard with welcome overlay
            setShowDashboardSkeleton(false);
            return;
          }
        } catch (error) {
          console.warn("Error checking signup status:", error);
        }

        // Show skeleton for existing users (they have data to load)
        setShowDashboardSkeleton(true);
        const timer = setTimeout(() => {
          setShowDashboardSkeleton(false);
        }, 1500);

        return () => clearTimeout(timer);
      }
    };

    checkSkeletonNeed();
  }, [loading, userInfo, error]);

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

  // Show simple loading for initial authentication check
  if (loading) {
    return (
      <SafeAreaView style={styles.simpleLoadingContainer}>
        <ActivityIndicator size="large" color="#FF7B54" />
      </SafeAreaView>
    );
  }

  // Show loading with skeleton UI only for authenticated users transitioning to dashboard
  if (showDashboardSkeleton) {
    return (
      <>
        {theme.mode === "dark" ? (
          <SafeAreaView style={styles.containerDark}>
            <View style={styles.loadingContainer}>
              <Text style={styles.loadingText}>{t("common.loading")}</Text>
              <SkeletonDashboard />
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
                <Text style={styles.loadingText}>{t("common.loading")}</Text>
                <SkeletonDashboard />
              </View>
            </SafeAreaView>
          </ImageBackground>
        )}
      </>
    );
  }

  // If user is authenticated and skeleton is not being shown, redirect to dashboard
  if (userInfo && !showDashboardSkeleton) {
    return <Redirect href="/dashboard" />;
  }

  // If not authenticated, redirect to signin
  usersApi.sigout();
  return <Redirect href="/signin" />;
}

const { width, height } = Dimensions.get("window");

const styles = StyleSheet.create({
  simpleLoadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: "#FFF9F0",
  },
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
    justifyContent: "flex-start",
    alignItems: "stretch",
    paddingTop: 20,
  },
  loadingText: {
    textAlign: "center",
    fontSize: 16,
    color: "#FF7B54",
    marginBottom: 20,
    fontWeight: "500",
  },
});
