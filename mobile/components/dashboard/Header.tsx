import * as usersApi from "@/api/users";
import { Ionicons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { router } from "expo-router";
import React from "react";
import { Image, StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useTranslation } from "react-i18next";

export const Header = () => {
  const { t } = useTranslation();

  // Fetch user profile
  const { data: user, isLoading } = useQuery({
    queryKey: ["userProfile"],
    queryFn: usersApi.getProfile,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

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
          <Ionicons name="settings-outline" size={24} color="#2D3436" />
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

const styles = StyleSheet.create({
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
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  profileButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    overflow: "hidden",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
    backgroundColor: "#FFFFFF",
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
    backgroundColor: "#FF7B54",
    justifyContent: "center",
    alignItems: "center",
    borderRadius: 20,
  },
  avatarText: {
    fontSize: 18,
    fontWeight: "600",
    color: "#FFFFFF",
  },
  greetingContainer: {
    marginBottom: 20,
    paddingHorizontal: 20,
  },
  greeting: {
    fontSize: 28,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 4,
  },
  userName: {
    fontSize: 28,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: "#636E72",
    marginTop: 4,
  },
});
