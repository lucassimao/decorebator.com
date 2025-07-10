import { CreateWordlistModal } from "@/components/dashboard/CreateWordlistModal";
import { Header } from "@/components/dashboard/Header";
import { useTheme } from "@/contexts/ThemeContext";
import { Ionicons } from "@expo/vector-icons";
import { router } from "expo-router";
import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Dimensions,
  Image,
  ImageBackground,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";

const { width, height } = Dimensions.get("window");

const EmptyDashboard = () => {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const { t } = useTranslation();
  const { theme } = useTheme();

  const handleCreateSuccess = () => {
    setShowCreateModal(false);
    router.dismissAll();
    router.replace("/dashboard");
  };

  return (
    <>
      {theme.mode === "dark" ? (
        <SafeAreaView style={styles.containerDark}>
          <ScrollView
            contentContainerStyle={styles.scrollContent}
            showsVerticalScrollIndicator={false}
          >
            <Header />

            {/* Illustration Image */}
            <View style={styles.illustrationContainer}>
              <Image
                source={require("@/assets/images/empty-dashboard-bg.png")}
                style={styles.illustrationImage}
                resizeMode="contain"
              />
            </View>

            {/* Bottom content */}
            <View style={styles.bottomContent}>
              {/* No wordlists message */}
              <Text style={styles.noWordlistsTextDark}>
                {t("dashboard.wordlists.noWordlistsYet")}
              </Text>

              {/* CTA Button */}
              <TouchableOpacity
                style={styles.ctaButton}
                onPress={() => setShowCreateModal(true)}
                activeOpacity={0.8}
              >
                <Text style={styles.ctaButtonText}>
                  {t("dashboard.wordlists.createFirstWordlist")}
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

          <CreateWordlistModal
            visible={showCreateModal}
            onClose={() => setShowCreateModal(false)}
            onSuccess={handleCreateSuccess}
          />
        </SafeAreaView>
      ) : (
        <ImageBackground
          source={require("@/assets/images/signup-bg3.png")}
          style={styles.backgroundImage}
          imageStyle={styles.backgroundImageStyle}
          resizeMode="cover"
        >
          <SafeAreaView style={styles.container}>
            <ScrollView
              contentContainerStyle={styles.scrollContent}
              showsVerticalScrollIndicator={false}
            >
              <Header />

              {/* Illustration Image */}
              <View style={styles.illustrationContainer}>
                <Image
                  source={require("@/assets/images/empty-dashboard-bg.png")}
                  style={styles.illustrationImage}
                  resizeMode="contain"
                />
              </View>

              {/* Bottom content */}
              <View style={styles.bottomContent}>
                {/* No wordlists message */}
                <Text style={styles.noWordlistsText}>
                  {t("dashboard.wordlists.noWordlistsYet")}
                </Text>

                {/* CTA Button */}
                <TouchableOpacity
                  style={styles.ctaButton}
                  onPress={() => setShowCreateModal(true)}
                  activeOpacity={0.8}
                >
                  <Text style={styles.ctaButtonText}>
                    {t("dashboard.wordlists.createFirstWordlist")}
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

            <CreateWordlistModal
              visible={showCreateModal}
              onClose={() => setShowCreateModal(false)}
              onSuccess={handleCreateSuccess}
            />
          </SafeAreaView>
        </ImageBackground>
      )}
    </>
  );
};

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
  noWordlistsTextDark: {
    fontSize: 20,
    fontWeight: "500",
    color: "#FFFFFF",
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
