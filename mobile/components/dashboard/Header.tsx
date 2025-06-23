import * as usersApi from "@/api/users";
import { Ionicons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { router } from "expo-router";
import React, { useEffect } from "react";
import { Image, StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useTranslation } from "react-i18next";
import { usePostHog } from "posthog-react-native";
import { useTheme } from "@/contexts/ThemeContext";

export const Header = () => {
  const { t } = useTranslation();
  const posthog = usePostHog();
  const { theme } = useTheme();
  const styles = createStyles(theme);

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
        >
          <Ionicons
            name="settings-outline"
            size={24}
            color={theme.colors.text.primary}
          />
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.profileButton}
          onPress={handleProfilePress}
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
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    profileButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      overflow: "hidden",
      backgroundColor: theme.colors.background.surface,
      ...theme.shadows.sm,
    },
    avatarContainer: {
      width: "100%",
      height: "100%",
    },
    profileImage: {
      width: "100%",
      height: "100%",
      borderRadius: 20,
    },
    avatarPlaceholder: {
      width: "100%",
      height: "100%",
      backgroundColor: theme.colors.primary,
      justifyContent: "center",
      alignItems: "center",
      borderRadius: 20,
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
  });
