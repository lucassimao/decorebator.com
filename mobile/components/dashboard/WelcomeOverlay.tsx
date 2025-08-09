import React, { useEffect, useRef } from "react";
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Animated,
  Modal,
  Platform,
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
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);
  const welcomeOpacity = useRef(new Animated.Value(0)).current;
  const scale = useRef(new Animated.Value(0.98)).current;

  // Animate welcome overlay when shown
  useEffect(() => {
    if (visible) {
      welcomeOpacity.setValue(0);
      scale.setValue(0.98);
      Animated.parallel([
        Animated.timing(welcomeOpacity, {
          toValue: 1,
          duration: 220,
          useNativeDriver: true,
        }),
        Animated.spring(scale, {
          toValue: 1,
          useNativeDriver: true,
          friction: 7,
          tension: 80,
        }),
      ]).start();
    }
  }, [visible, welcomeOpacity, scale]);

  if (!visible) return null;

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      statusBarTranslucent
      onRequestClose={onSkip}
    >
      <View style={styles.backdrop}>
        <Animated.View
          style={[
            styles.welcomeModal,
            { opacity: welcomeOpacity, transform: [{ scale }] },
          ]}
        >
          <View style={styles.badge}>
            <Ionicons name="sparkles" size={22} color={theme.colors.primary} />
          </View>

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
                <Ionicons name="book" size={22} color={theme.colors.primary} />
              </View>
              <Text style={styles.welcomeFeatureText}>
                {t("welcome.feature1", "Create custom wordlists")}
              </Text>
            </View>

            <View style={styles.welcomeFeature}>
              <View style={styles.welcomeFeatureIcon}>
                <Ionicons name="bulb" size={22} color={theme.colors.primary} />
              </View>
              <Text style={styles.welcomeFeatureText}>
                {t("welcome.feature2", "AI-powered definitions & images")}
              </Text>
            </View>

            <View style={styles.welcomeFeature}>
              <View style={styles.welcomeFeatureIcon}>
                <Ionicons
                  name="repeat"
                  size={22}
                  color={theme.colors.primary}
                />
              </View>
              <Text style={styles.welcomeFeatureText}>
                {t("welcome.feature3", "Smart spaced repetition")}
              </Text>
            </View>
          </View>

          <TouchableOpacity
            style={styles.welcomeButton}
            onPress={onGetStarted}
            activeOpacity={0.9}
            accessibilityRole="button"
            accessibilityLabel={t(
              "welcome.getStarted",
              "Create Your First Wordlist",
            )}
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
            accessibilityRole="button"
            accessibilityLabel={t("welcome.skip", "Skip for now")}
          >
            <Text style={styles.welcomeSkipText}>
              {t("welcome.skip", "Skip for now")}
            </Text>
          </TouchableOpacity>
        </Animated.View>
      </View>
    </Modal>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) =>
  StyleSheet.create({
    backdrop: {
      flex: 1,
      backgroundColor: "rgba(0, 0, 0, 0.6)",
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingTop: Platform.OS === "android" ? responsive.spacing.horizontal : 0,
    },
    welcomeModal: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.xl,
      padding: responsive.spacing.formPadding,
      maxWidth: responsive.getValueForSize(340, 380, 420, 460),
      width: "100%",
      ...theme.shadows.lg,
    },
    welcomeHeader: {
      alignItems: "center",
      marginBottom: responsive.spacing.formPadding,
      paddingTop: responsive.spacing.elementSpacing,
    },
    badge: {
      position: "absolute",
      top: -responsive.getValueForSize(14, 16, 18, 20),
      alignSelf: "center",
      backgroundColor: `${theme.colors.primary}20`,
      borderRadius: 999,
      padding: responsive.getValueForSize(8, 10, 12, 12),
    },
    welcomeTitle: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: responsive.spacing.elementSpacing,
      lineHeight: responsive.fontSizes.title * responsive.fontSizes.lineHeight,
    },
    welcomeSubtitle: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: responsive.fontSizes.body * responsive.fontSizes.lineHeight,
    },
    welcomeFeatures: {
      marginBottom: responsive.spacing.formPadding,
    },
    welcomeFeature: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: responsive.spacing.horizontal,
    },
    welcomeFeatureIcon: {
      width: responsive.getValueForSize(36, 40, 44, 48),
      height: responsive.getValueForSize(36, 40, 44, 48),
      borderRadius: responsive.getValueForSize(18, 20, 22, 24),
      backgroundColor: `${theme.colors.primary}15`,
      justifyContent: "center",
      alignItems: "center",
      marginRight: responsive.spacing.horizontal,
    },
    welcomeFeatureText: {
      flex: 1,
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.primary,
      fontWeight: "500",
      lineHeight: responsive.fontSizes.body * responsive.fontSizes.lineHeight,
    },
    welcomeButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: responsive.spacing.horizontal,
      paddingHorizontal: responsive.spacing.formPadding,
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      marginBottom: responsive.spacing.elementSpacing,
      minHeight: responsive.spacing.minTouchTarget,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
    },
    welcomeButtonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
      marginRight: responsive.spacing.elementSpacing,
    },
    welcomeSkip: {
      paddingVertical: responsive.spacing.elementSpacing,
      alignItems: "center",
      minHeight: responsive.spacing.minTouchTarget,
    },
    welcomeSkipText: {
      color: theme.colors.text.secondary,
      fontSize: responsive.fontSizes.body,
      fontWeight: "500",
    },
  });
