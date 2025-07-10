import * as usersApi from "@/api/users";
import { Ionicons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { router } from "expo-router";
import React, { useEffect } from "react";
import {
  Animated,
  Image,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { useTranslation } from "react-i18next";
import { usePostHog } from "posthog-react-native";
import { useTheme } from "@/contexts/ThemeContext";

// Shimmer Element for Header Loading States
interface HeaderShimmerElementProps {
  style: any;
  skeletonStyles: {
    elementColor: string;
    shimmerHighlight: string;
  };
  shimmerAnim: Animated.Value;
}

const HeaderShimmerElement: React.FC<HeaderShimmerElementProps> = ({
  style,
  skeletonStyles,
  shimmerAnim,
}) => {
  const shimmerOpacity = shimmerAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [0.4, 1.0],
  });

  const backgroundColor = shimmerAnim.interpolate({
    inputRange: [0, 0.5, 1],
    outputRange: [
      skeletonStyles.elementColor,
      skeletonStyles.shimmerHighlight,
      skeletonStyles.elementColor,
    ],
  });

  return (
    <Animated.View
      style={[
        style,
        {
          backgroundColor,
          opacity: shimmerOpacity,
        },
      ]}
    />
  );
};

export const Header = () => {
  const { t } = useTranslation();
  const posthog = usePostHog();
  const { theme } = useTheme();
  const styles = createStyles(theme);

  // Shimmer animation for loading state
  const shimmerAnim = React.useRef(new Animated.Value(0)).current;

  // Fetch user profile
  const { data: user, isLoading } = useQuery({
    queryKey: ["userProfile"],
    queryFn: usersApi.getProfile,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Shimmer animation loop
  React.useEffect(() => {
    if (isLoading) {
      const shimmerLoop = Animated.loop(
        Animated.sequence([
          Animated.timing(shimmerAnim, {
            toValue: 1,
            duration: 1000,
            useNativeDriver: false,
          }),
          Animated.timing(shimmerAnim, {
            toValue: 0,
            duration: 1000,
            useNativeDriver: false,
          }),
        ]),
      );
      shimmerLoop.start();

      return () => shimmerLoop.stop();
    }
  }, [shimmerAnim, isLoading]);

  useEffect(() => {
    if (!user) return;

    posthog.identify(String(user.id), {
      $set: {
        email: user.email,
        name: user.firstName + ` ` + user.lastName,
      },
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  // Refresh user session when screen comes into focus
  // useFocusEffect(
  //   React.useCallback(() => {
  //     refetch();
  //   }, []),
  // );

  const handleSettingsPress = () => {
    router.push("/settings");
  };

  const handleProfilePress = () => {
    router.push("/profileSettings");
  };

  // Get time-based greeting
  const getGreeting = () => {
    const hour = new Date().getHours();
    if (hour < 12) return t("dashboard.greetings.morning");
    if (hour < 18) return t("dashboard.greetings.afternoon");
    return t("dashboard.greetings.evening");
  };

  // Get theme-aware skeleton styles
  const getSkeletonStyles = () => {
    const isDark = theme.mode === "dark";
    return {
      elementColor: isDark
        ? "rgba(255, 255, 255, 0.08)"
        : "rgba(0, 0, 0, 0.08)",
      shimmerHighlight: isDark
        ? "rgba(255, 255, 255, 0.15)"
        : "rgba(0, 0, 0, 0.15)",
    };
  };

  const skeletonStyles = getSkeletonStyles();

  // Always render header to prevent layout shift
  // if (isLoading) return null;

  const profilePicture = user?.profilePictureUrl;

  return (
    <>
      <View style={styles.header}>
        <TouchableOpacity
          style={styles.settingsButton}
          onPress={handleSettingsPress}
          activeOpacity={0.7}
        >
          <Ionicons
            name="settings-outline"
            size={26}
            color={theme.colors.text.primary}
          />
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.profileButton}
          onPress={handleProfilePress}
          activeOpacity={0.7}
        >
          <View style={styles.avatarContainer}>
            {isLoading ? (
              <HeaderShimmerElement
                style={styles.avatarSkeleton}
                skeletonStyles={skeletonStyles}
                shimmerAnim={shimmerAnim}
              />
            ) : profilePicture ? (
              <Image
                source={{ uri: profilePicture }}
                style={styles.profileImage}
              />
            ) : (
              <View style={styles.avatarPlaceholder}>
                <Text style={styles.avatarText}>
                  {user?.firstName?.[0]?.toUpperCase() || "U"}
                </Text>
              </View>
            )}
          </View>
        </TouchableOpacity>
      </View>
      {/* Greeting */}
      <View style={styles.greetingContainer}>
        <Text style={styles.greeting}>{getGreeting()},</Text>
        {isLoading ? (
          <HeaderShimmerElement
            style={styles.userNameSkeleton}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        ) : (
          <Text style={styles.userName}>
            {user?.firstName || t("common.user")}!
          </Text>
        )}
        <Text style={styles.subtitle}>{t("dashboard.header.subtitle")}</Text>
      </View>
    </>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    header: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      paddingTop: 10,
      paddingBottom: 20,
      paddingHorizontal: 20,
    },
    settingsButton: {
      width: 44,
      height: 44,
      borderRadius: 22,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      ...theme.shadows.md,
      shadowColor: theme.colors.text.primary,
      shadowOpacity: theme.mode === "light" ? 0.08 : 0.2,
    },
    profileButton: {
      width: 44,
      height: 44,
      borderRadius: 22,
      overflow: "hidden",
      backgroundColor: theme.colors.background.surface,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      ...theme.shadows.md,
      shadowColor: theme.colors.text.primary,
      shadowOpacity: theme.mode === "light" ? 0.08 : 0.2,
    },
    avatarContainer: {
      width: "100%",
      height: "100%",
    },
    profileImage: {
      width: "100%",
      height: "100%",
      borderRadius: 22,
    },
    avatarPlaceholder: {
      width: "100%",
      height: "100%",
      backgroundColor: theme.colors.primary,
      justifyContent: "center",
      alignItems: "center",
      borderRadius: 22,
    },
    avatarText: {
      fontSize: 18,
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    greetingContainer: {
      marginBottom: 20,
      paddingHorizontal: 20,
    },
    greeting: {
      fontSize: 28,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 4,
    },
    userName: {
      fontSize: 28,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 8,
    },
    subtitle: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      marginTop: 4,
    },
    // Skeleton loading styles
    avatarSkeleton: {
      width: 44,
      height: 44,
      borderRadius: 22,
    },
    userNameSkeleton: {
      width: 120,
      height: 28,
      borderRadius: 14,
      marginBottom: 8,
    },
  });
