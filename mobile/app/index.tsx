import { useUserInfo } from "@/hooks/users";
import { useTheme } from "@/contexts/ThemeContext";
import React, { useEffect, useRef } from "react";
import {
  ActivityIndicator,
  Animated,
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

export default function Index() {
  const { userInfo, error, loading } = useUserInfo();
  const router = useRouter();
  const { theme } = useTheme();
  const { t } = useTranslation();

  // Track if we should show skeleton for authenticated users
  const [showDashboardSkeleton, setShowDashboardSkeleton] =
    React.useState(false);

  // Shimmer animation for skeleton
  const shimmerAnim = useRef(new Animated.Value(0)).current;

  React.useEffect(() => {
    if (!loading && userInfo && !error) {
      // Show skeleton when we have an authenticated user ready to go to dashboard
      setShowDashboardSkeleton(true);
      const timer = setTimeout(() => {
        setShowDashboardSkeleton(false);
      }, 1500);

      return () => clearTimeout(timer);
    }
  }, [loading, userInfo, error]);

  // Shimmer animation loop
  React.useEffect(() => {
    if (showDashboardSkeleton) {
      const shimmerLoop = Animated.loop(
        Animated.sequence([
          Animated.timing(shimmerAnim, {
            toValue: 1,
            duration: 1200,
            useNativeDriver: true,
          }),
          Animated.timing(shimmerAnim, {
            toValue: 0,
            duration: 1200,
            useNativeDriver: true,
          }),
        ]),
      );
      shimmerLoop.start();

      return () => shimmerLoop.stop();
    }
  }, [showDashboardSkeleton, shimmerAnim]);

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

  // Get theme-aware skeleton styles
  const getSkeletonStyles = () => {
    const isDark = theme.mode === "dark";
    return {
      elementColor: isDark ? "rgba(255, 255, 255, 0.1)" : "rgba(0, 0, 0, 0.1)",
      containerColor: isDark
        ? "rgba(255, 255, 255, 0.05)"
        : "rgba(255, 255, 255, 0.9)",
      shadowColor: isDark ? "#FFFFFF" : "#000000",
    };
  };

  // Shimmer component for skeleton elements
  const ShimmerElement = ({ style }: { style: any }) => {
    const skeletonStyles = getSkeletonStyles();
    const shimmerOpacity = shimmerAnim.interpolate({
      inputRange: [0, 1],
      outputRange: [0.3, 0.8],
    });

    return (
      <Animated.View
        style={[
          style,
          {
            backgroundColor: skeletonStyles.elementColor,
            opacity: shimmerOpacity,
          },
        ]}
      />
    );
  };

  // Skeleton UI component for dashboard loading
  const renderSkeletonDashboard = () => {
    const skeletonStyles = getSkeletonStyles();

    return (
      <View style={styles.skeletonContainer}>
        {/* Header skeleton */}
        <View style={styles.skeletonHeader}>
          <ShimmerElement style={styles.skeletonCircle} />
          <ShimmerElement style={styles.skeletonTitle} />
          <ShimmerElement style={styles.skeletonCircle} />
        </View>

        {/* Stats skeleton */}
        <View
          style={[
            styles.skeletonStats,
            { backgroundColor: skeletonStyles.containerColor },
          ]}
        >
          <View style={styles.skeletonStatItem}>
            <ShimmerElement style={styles.skeletonStatValue} />
            <ShimmerElement style={styles.skeletonStatLabel} />
          </View>
          <View style={styles.skeletonStatItem}>
            <ShimmerElement style={styles.skeletonStatValue} />
            <ShimmerElement style={styles.skeletonStatLabel} />
          </View>
          <View style={styles.skeletonStatItem}>
            <ShimmerElement style={styles.skeletonStatValue} />
            <ShimmerElement style={styles.skeletonStatLabel} />
          </View>
        </View>

        {/* Section header skeleton */}
        <View style={styles.skeletonSectionHeader}>
          <ShimmerElement style={styles.skeletonSectionTitle} />
          <ShimmerElement style={styles.skeletonAddButton} />
        </View>

        {/* Wordlist items skeleton */}
        <View style={styles.skeletonWordlists}>
          {[1, 2, 3].map((item) => (
            <View
              key={item}
              style={[
                styles.skeletonWordlistItem,
                { backgroundColor: skeletonStyles.containerColor },
              ]}
            >
              <View style={styles.skeletonWordlistInfo}>
                <ShimmerElement style={styles.skeletonWordlistTitle} />
                <ShimmerElement style={styles.skeletonWordlistSubtitle} />
              </View>
              <ShimmerElement style={styles.skeletonWordlistProgress} />
            </View>
          ))}
        </View>
      </View>
    );
  };

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
              {renderSkeletonDashboard()}
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
                {renderSkeletonDashboard()}
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
  // Skeleton styles - theme-aware with better contrast
  skeletonContainer: {
    flex: 1,
    paddingHorizontal: 20,
  },
  skeletonHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingVertical: 20,
  },
  skeletonCircle: {
    width: 40,
    height: 40,
    borderRadius: 20,
  },
  skeletonTitle: {
    width: 120,
    height: 24,
    borderRadius: 12,
  },
  skeletonStats: {
    flexDirection: "row",
    justifyContent: "space-between",
    borderRadius: 16,
    paddingVertical: 20,
    paddingHorizontal: 10,
    marginBottom: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  skeletonStatItem: {
    alignItems: "center",
    flex: 1,
  },
  skeletonStatValue: {
    width: 40,
    height: 36,
    borderRadius: 18,
    marginBottom: 8,
  },
  skeletonStatLabel: {
    width: 60,
    height: 14,
    borderRadius: 7,
  },
  skeletonSectionHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 16,
  },
  skeletonSectionTitle: {
    width: 150,
    height: 20,
    borderRadius: 10,
  },
  skeletonAddButton: {
    width: 34,
    height: 34,
    borderRadius: 17,
  },
  skeletonWordlists: {
    gap: 16,
  },
  skeletonWordlistItem: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    borderRadius: 12,
    padding: 16,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  skeletonWordlistInfo: {
    flex: 1,
  },
  skeletonWordlistTitle: {
    width: "80%",
    height: 18,
    borderRadius: 9,
    marginBottom: 8,
  },
  skeletonWordlistSubtitle: {
    width: "60%",
    height: 14,
    borderRadius: 7,
  },
  skeletonWordlistProgress: {
    width: 60,
    height: 60,
    borderRadius: 30,
  },
});
