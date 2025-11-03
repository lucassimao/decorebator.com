import React, { useEffect, useRef } from "react";
import {
  Animated,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { usePaymentProvider } from "@/hooks/useRevenueCat";

export type PremiumUpsellContext = "analytics" | "public_quiz" | "chat";

type Props = {
  visible: boolean;
  onClose: () => void;
  context?: PremiumUpsellContext;
  onUpgradePress?: () => void;
};

export default function PremiumUpsellModal({
  visible,
  onClose,
  context = "analytics",
  onUpgradePress,
}: Props) {
  const { theme, responsive } = useTheme();
  const { t } = useTranslation();
  const providerInfo = usePaymentProvider().data;

  const opacity = useRef(new Animated.Value(0)).current;
  const scale = useRef(new Animated.Value(0.96)).current;

  useEffect(() => {
    if (visible) {
      Animated.parallel([
        Animated.timing(opacity, {
          toValue: 1,
          duration: 160,
          useNativeDriver: true,
        }),
        Animated.spring(scale, {
          toValue: 1,
          friction: 10,
          tension: 70,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      Animated.parallel([
        Animated.timing(opacity, {
          toValue: 0,
          duration: 140,
          useNativeDriver: true,
        }),
        Animated.timing(scale, {
          toValue: 0.96,
          duration: 140,
          useNativeDriver: true,
        }),
      ]).start();
    }
  }, [visible, opacity, scale]);

  const featureLines =
    context === "analytics"
      ? [
          t("dashboard.stats.premium.title"),
          t("dashboard.stats.feature.historicalTrends", "Historical trends"),
          t("dashboard.stats.feature.accuracyInsights", "Accuracy insights"),
          t("dashboard.stats.feature.boxDistribution", "Box distribution"),
        ]
      : context === "chat"
        ? [
            t(
              "wordlistItem.feature.voiceConversations",
              "Voice conversations with AI",
            ),
            t(
              "wordlistItem.feature.vocabularyPractice",
              "Practice your vocabulary naturally",
            ),
            t(
              "wordlistItem.feature.pronunciationHelp",
              "Get pronunciation guidance",
            ),
            t(
              "wordlistItem.feature.unlimitedChatTime",
              "Unlimited conversation time",
            ),
          ]
        : [
            t("wordlistItem.feature.publishQuizzes", "Publish public quizzes"),
            t("wordlistItem.feature.shareLinks", "Share links & leaderboards"),
            t(
              "wordlistItem.feature.betterBranding",
              "Better visuals & branding",
            ),
          ];

  const handleUpgrade = () => {
    if (onUpgradePress) {
      onUpgradePress();
      onClose();
      return;
    }
    // Default: let caller decide navigation; modal only frames the upsell.
    // For hints, providerInfo?.provider === "revenuecat" → RevenueCatPaywall; otherwise Settings.
    void providerInfo; // referenced for lint - caller handles routing by provider
    onClose();
  };

  const styles = createStyles(theme, responsive);

  return (
    <Modal
      visible={visible}
      transparent
      animationType="none"
      onRequestClose={onClose}
    >
      <TouchableWithoutFeedback onPress={onClose}>
        <Animated.View style={[styles.overlay, { opacity }]}>
          <TouchableWithoutFeedback>
            <Animated.View style={[styles.card, { transform: [{ scale }] }]}>
              <View style={styles.header}>
                <View style={styles.iconBadge}>
                  <MaterialIcons
                    name="workspace-premium"
                    size={24}
                    color={theme.colors.text.inverse}
                  />
                </View>
                <Text style={styles.title}>
                  {context === "analytics"
                    ? t("dashboard.stats.premium.title")
                    : context === "chat"
                      ? t(
                          "wordlistItem.chatUpsell.title",
                          "Unlock AI Conversations",
                        )
                      : t("wordlistItem.premiumUpsell.title", "Unlock Premium")}
                </Text>
                <Text style={styles.subtitle}>
                  {context === "analytics"
                    ? t("dashboard.stats.premium.subtitle")
                    : context === "chat"
                      ? t(
                          "wordlistItem.chatUpsell.subtitle",
                          "Practice vocabulary through natural voice conversations",
                        )
                      : t(
                          "wordlistItem.premiumUpsell.subtitle",
                          "Get access to advanced features",
                        )}
                </Text>
              </View>

              <View style={styles.features}>
                {featureLines.map((line, idx) => (
                  <View key={idx} style={styles.featureRow}>
                    <MaterialIcons
                      name="check-circle"
                      size={18}
                      color={theme.colors.success}
                    />
                    <Text style={styles.featureText}>{line}</Text>
                  </View>
                ))}
              </View>

              <View style={styles.actions}>
                <TouchableOpacity
                  style={styles.secondary}
                  onPress={onClose}
                  accessibilityLabel={t("upgradePrompt.maybeLater")}
                >
                  <Text style={styles.secondaryText}>
                    {t("upgradePrompt.maybeLater")}
                  </Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={styles.primary}
                  onPress={handleUpgrade}
                  accessibilityLabel={t("settings.subscription.upgradeButton")}
                >
                  <Text style={styles.primaryText}>
                    {t("settings.subscription.upgradeButton")}
                  </Text>
                </TouchableOpacity>
              </View>

              <Text style={styles.finePrint}>
                {t(
                  "wordlistItem.premiumUpsell.finePrint",
                  "Cancel anytime. Existing data remains safe.",
                )}
              </Text>
            </Animated.View>
          </TouchableWithoutFeedback>
        </Animated.View>
      </TouchableWithoutFeedback>
    </Modal>
  );
}

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) =>
  StyleSheet.create({
    overlay: {
      flex: 1,
      backgroundColor:
        theme.mode === "light" ? "rgba(0,0,0,0.5)" : "rgba(0,0,0,0.7)",
      justifyContent: "center",
      alignItems: "center",
      padding: responsive.spacing.horizontal,
    },
    card: {
      width: "100%",
      maxWidth: responsive.getValueForSize(320, 360, 380, 400),
      borderRadius: theme.borderRadius.xl,
      backgroundColor: theme.colors.background.surface,
      padding: responsive.spacing.formPadding,
      ...theme.shadows.lg,
    },
    header: { alignItems: "center" },
    iconBadge: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.premium,
      alignItems: "center",
      justifyContent: "center",
      marginBottom: 8,
    },
    title: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
    },
    subtitle: {
      marginTop: 4,
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    features: { marginTop: responsive.spacing.elementSpacing, gap: 8 },
    featureRow: { flexDirection: "row", alignItems: "center", gap: 8 },
    featureText: {
      color: theme.colors.text.primary,
      fontSize: responsive.getScaledFont("body"),
    },
    actions: {
      marginTop: responsive.spacing.elementSpacing * 1.25,
      flexDirection: "row",
      gap: 10,
    },
    secondary: {
      flex: 1,
      borderRadius: 12,
      paddingVertical: 12,
      backgroundColor: theme.colors.background.default,
      alignItems: "center",
    },
    secondaryText: { color: theme.colors.text.secondary, fontWeight: "600" },
    primary: {
      flex: 1,
      borderRadius: 12,
      paddingVertical: 12,
      backgroundColor: theme.colors.primary,
      alignItems: "center",
    },
    primaryText: { color: theme.colors.text.inverse, fontWeight: "700" },
    finePrint: {
      marginTop: 10,
      textAlign: "center",
      color: theme.colors.text.secondary,
      fontSize: responsive.getScaledFont("label"),
    },
  });
