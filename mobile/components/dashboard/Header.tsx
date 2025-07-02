import * as usersApi from "@/api/users";
import { Ionicons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { router } from "expo-router";
import React, { useEffect } from "react";
import { Image, StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useTranslation } from "react-i18next";
import { usePostHog } from "posthog-react-native";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive } from "@/hooks/useResponsive";

export const Header = () => {
  const { t } = useTranslation();
  const posthog = usePostHog();
  const { theme } = useTheme();
  const { isTablet, type: deviceType } = useResponsive();
  const styles = createStyles(theme, isTablet, deviceType);

  // Fetch user profile
  const { data: user, isLoading } = useQuery({
    queryKey: ["userProfile"],
    queryFn: usersApi.getProfile,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

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

  // TODO improve this
  if (isLoading) return null;

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
            {profilePicture ? (
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
        <Text style={styles.userName}>
          {user?.firstName || t("common.user")}!
        </Text>
        <Text style={styles.subtitle}>{t("dashboard.header.subtitle")}</Text>
      </View>
    </>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  deviceType: string,
) =>
  StyleSheet.create({
    header: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      paddingTop: theme.spacing.sm,
      paddingBottom: theme.spacing.md,
      paddingHorizontal: theme.spacing.md,
    },
    settingsButton: {
      width: theme.touchTargets.comfortable,
      height: theme.touchTargets.comfortable,
      borderRadius: theme.touchTargets.comfortable / 2,
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
      width: theme.touchTargets.comfortable,
      height: theme.touchTargets.comfortable,
      borderRadius: theme.touchTargets.comfortable / 2,
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
      borderRadius: theme.touchTargets.comfortable / 2,
    },
    avatarPlaceholder: {
      width: "100%",
      height: "100%",
      backgroundColor: theme.colors.primary,
      justifyContent: "center",
      alignItems: "center",
      borderRadius: theme.touchTargets.comfortable / 2,
    },
    avatarText: {
      fontSize: theme.typography.sizes.body,
      fontWeight: theme.typography.weights.semibold,
      color: theme.colors.text.inverse,
    },
    greetingContainer: {
      marginBottom: theme.spacing.md,
      paddingHorizontal: theme.spacing.md,
    },
    greeting: {
      fontSize: isTablet
        ? theme.typography.sizes.display
        : theme.typography.sizes.title,
      fontWeight: theme.typography.weights.semibold,
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.xs,
    },
    userName: {
      fontSize: isTablet
        ? theme.typography.sizes.display
        : theme.typography.sizes.title,
      fontWeight: theme.typography.weights.semibold,
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.xs,
    },
    subtitle: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.secondary,
      marginTop: theme.spacing.xs,
    },
  });
