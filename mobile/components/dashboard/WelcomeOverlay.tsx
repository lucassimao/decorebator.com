import React, { useEffect, useRef } from "react";
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Animated,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";

interface WelcomeOverlayProps {
  visible: boolean;
  onGetStarted: () => void;
  onSkip: () => void;
}

export const WelcomeOverlay: React.FC<WelcomeOverlayProps> = ({
  visible,
  onGetStarted,
  onSkip,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const welcomeOpacity = useRef(new Animated.Value(0)).current;

  // Animate welcome overlay when shown
  useEffect(() => {
    if (visible) {
      welcomeOpacity.setValue(0);
      Animated.timing(welcomeOpacity, {
        toValue: 1,
        duration: 300,
        useNativeDriver: true,
      }).start();
    }
  }, [visible, welcomeOpacity]);

  if (!visible) return null;

  return (
    <View style={styles.welcomeOverlay}>
      <Animated.View style={[styles.welcomeModal, { opacity: welcomeOpacity }]}>
        <View style={styles.welcomeHeader}>
          <Text style={styles.welcomeTitle}>
            {t("welcome.title", "Welcome to Decorebator! 🎉")}
          </Text>
          <Text style={styles.welcomeSubtitle}>
            {t(
              "welcome.subtitle",
              "Your AI-powered language learning journey starts here",
            )}
          </Text>
        </View>

        <View style={styles.welcomeFeatures}>
          <View style={styles.welcomeFeature}>
            <View style={styles.welcomeFeatureIcon}>
              <Ionicons name="book" size={24} color={theme.colors.primary} />
            </View>
            <Text style={styles.welcomeFeatureText}>
              {t("welcome.feature1", "Create custom wordlists")}
            </Text>
          </View>

          <View style={styles.welcomeFeature}>
            <View style={styles.welcomeFeatureIcon}>
              <Ionicons name="bulb" size={24} color={theme.colors.primary} />
            </View>
            <Text style={styles.welcomeFeatureText}>
              {t("welcome.feature2", "AI-powered definitions & images")}
            </Text>
          </View>

          <View style={styles.welcomeFeature}>
            <View style={styles.welcomeFeatureIcon}>
              <Ionicons name="repeat" size={24} color={theme.colors.primary} />
            </View>
            <Text style={styles.welcomeFeatureText}>
              {t("welcome.feature3", "Smart spaced repetition")}
            </Text>
          </View>
        </View>

        <TouchableOpacity
          style={styles.welcomeButton}
          onPress={onGetStarted}
          activeOpacity={0.8}
        >
          <Text style={styles.welcomeButtonText}>
            {t("welcome.getStarted", "Create Your First Wordlist")}
          </Text>
          <Ionicons name="arrow-forward" size={20} color="#FFFFFF" />
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.welcomeSkip}
          onPress={onSkip}
          activeOpacity={0.7}
        >
          <Text style={styles.welcomeSkipText}>
            {t("welcome.skip", "Skip for now")}
          </Text>
        </TouchableOpacity>
      </Animated.View>
    </View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    welcomeOverlay: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: "rgba(0, 0, 0, 0.6)",
      justifyContent: "center",
      alignItems: "center",
      zIndex: 1000,
    },
    welcomeModal: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: 20,
      padding: 24,
      marginHorizontal: 20,
      maxWidth: 400,
      width: "100%",
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.3,
      shadowRadius: 12,
      elevation: 8,
    },
    welcomeHeader: {
      alignItems: "center",
      marginBottom: 24,
    },
    welcomeTitle: {
      fontSize: 24,
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: 8,
    },
    welcomeSubtitle: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: 22,
    },
    welcomeFeatures: {
      marginBottom: 24,
    },
    welcomeFeature: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: 16,
    },
    welcomeFeatureIcon: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: `${theme.colors.primary}15`,
      justifyContent: "center",
      alignItems: "center",
      marginRight: 16,
    },
    welcomeFeatureText: {
      flex: 1,
      fontSize: 16,
      color: theme.colors.text.primary,
      fontWeight: "500",
    },
    welcomeButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: 12,
      paddingVertical: 16,
      paddingHorizontal: 24,
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      marginBottom: 12,
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.3,
      shadowRadius: 8,
      elevation: 5,
    },
    welcomeButtonText: {
      color: "#FFFFFF",
      fontSize: 18,
      fontWeight: "600",
      marginRight: 8,
    },
    welcomeSkip: {
      paddingVertical: 12,
      alignItems: "center",
    },
    welcomeSkipText: {
      color: theme.colors.text.secondary,
      fontSize: 16,
      fontWeight: "500",
    },
  });
