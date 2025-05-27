import React, { useEffect, useState } from "react";
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  TouchableOpacity,
  Image,
  ImageBackground,
  Dimensions,
  ScrollView,
  Alert,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Avatar } from "react-native-paper";
import * as usersApi from "@/api/users";
import { router } from "expo-router";
import { Wordlist } from "@/api/wordlists";
import CreateWordlistModal from "@/components/dashboard/CreateWordlistModal";

const { width, height } = Dimensions.get("window");

const EmptyDashboard = () => {
  const user = usersApi.getUserInfo();
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    if (!user) router.push("/signin");
  }, [user]);

  const handleCreateSuccess = (wordlistName: string) => {
    // Handle successful creation
    Alert.alert(
      "Success!",
      `Your wordlist "${wordlistName}" has been created.`,
      [
        {
          text: "OK",
          onPress: () => {
            router.push("/dashboard");
          },
        },
      ],
    );
  };

  return (
    <ImageBackground
      source={require("@/assets/images/dashboard-bg.png")}
      style={styles.backgroundImage}
      resizeMode="cover"
    >
      <SafeAreaView style={styles.container}>
        <ScrollView
          contentContainerStyle={styles.scrollContent}
          showsVerticalScrollIndicator={false}
        >
          {/* Header */}
          <View style={styles.header}>
            <TouchableOpacity style={styles.settingsButton}>
              <Ionicons name="settings-outline" size={24} color="#2D3436" />
            </TouchableOpacity>

            <TouchableOpacity style={styles.profileButton}>
              <Avatar.Image
                size={36}
                source={{ uri: "https://i.pravatar.cc/100" }}
                style={styles.profileImage}
              />
            </TouchableOpacity>
          </View>

          {/* Top Content Container */}
          <View style={styles.topContent}>
            {/* Greeting */}
            <View style={styles.greetingContainer}>
              <Text style={styles.greeting}>Good Morning,</Text>
              <Text style={styles.userName}>{user?.firstName}!</Text>
              <Text style={styles.subtitle}>
                Ready to learn something new today?
              </Text>
            </View>

            {/* Stats */}
            <View style={styles.statsContainer}>
              <View style={styles.statItem}>
                <Text style={styles.statLabel}>Total Words</Text>
                <Text style={styles.statValue}>0</Text>
              </View>
              <View style={styles.statItem}>
                <Text style={styles.statLabel}>Wordlists</Text>
                <Text style={styles.statValue}>0</Text>
              </View>
              <View style={styles.statItem}>
                <Text style={styles.statLabel}>Words Learned</Text>
                <Text style={styles.statValue}>0</Text>
              </View>
            </View>
          </View>

          {/* Illustration Image */}
          <View style={styles.illustrationContainer}>
            <Image
              source={require("@/assets/images/empty-dashboard-bg.png")} // Your illustration with character, books, plants
              style={styles.illustrationImage}
              resizeMode="contain"
            />
          </View>

          {/* Bottom content */}
          <View style={styles.bottomContent}>
            {/* No wordlists message */}
            <Text style={styles.noWordlistsText}>No wordlists yet</Text>

            {/* CTA Button */}
            <TouchableOpacity
              style={styles.ctaButton}
              onPress={() => setShowCreateModal(true)}
              activeOpacity={0.8}
            >
              <Text style={styles.ctaButtonText}>
                Create your first wordlist
              </Text>
              <Ionicons
                name="add-circle"
                size={24}
                color="#FFFFFF"
                style={styles.ctaIcon}
              />
            </TouchableOpacity>
          </View>
        </ScrollView>
      </SafeAreaView>

      {/* Create Wordlist Modal */}
      <CreateWordlistModal
        visible={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={handleCreateSuccess}
      />
    </ImageBackground>
  );
};

const styles = StyleSheet.create({
  backgroundImage: {
    flex: 1,
    width: width,
    height: height,
  },
  container: {
    flex: 1,
  },
  scrollContent: {
    flexGrow: 1,
    minHeight: height - 100, // Account for safe area
  },
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
  profileImage: {
    width: "100%",
    height: "100%",
  },
  topContent: {
    paddingHorizontal: 20,
  },
  greetingContainer: {
    marginBottom: 20,
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
  statsContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    backgroundColor: "rgba(255, 255, 255, 0.7)",
    borderRadius: 16,
    paddingVertical: 20,
    paddingHorizontal: 10,
    marginBottom: 0, // Remove bottom margin to connect with illustration
    // Subtle shadow for depth
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 8,
    elevation: 2,
  },
  statItem: {
    alignItems: "center",
    flex: 1,
  },
  statLabel: {
    fontSize: 14,
    color: "#636E72",
    marginBottom: 8,
  },
  statValue: {
    fontSize: 36,
    fontWeight: "700",
    color: "#2D3436",
  },
  illustrationContainer: {
    // width: width,
    // marginLeft: -20, // Negative margin to achieve full width
    // marginTop: -10, // Slight overlap with stats for smooth transition
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  illustrationImage: {
    width: width,
    height: width * 0.7, // Adjust ratio based on your illustration
    maxHeight: 300,
  },
  bottomContent: {
    paddingHorizontal: 20,
    paddingBottom: 30,
    alignItems: "center",
  },
  noWordlistsText: {
    fontSize: 20,
    fontWeight: "500",
    color: "#2D3436",
    textAlign: "center",
    marginBottom: 30,
  },
  ctaButton: {
    backgroundColor: "#FF7B54",
    borderRadius: 12,
    paddingVertical: 16,
    paddingHorizontal: 24,
    flexDirection: "row",
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 5,
    width: "100%",
  },
  ctaButtonText: {
    color: "#FFFFFF",
    fontSize: 18,
    fontWeight: "600",
    marginRight: 8,
  },
  ctaIcon: {
    marginLeft: 4,
  },
});

export default EmptyDashboard;
