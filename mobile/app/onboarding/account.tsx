import React, { useMemo } from "react";
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
} from "react-native";
import { useRouter } from "expo-router";
import { useTheme } from "@/contexts/ThemeContext";
import { useTranslation } from "react-i18next";
import { usePostHog } from "posthog-react-native";
import { OnboardingLayout } from "@/components/onboarding/OnboardingLayout";
import { Feather } from "@expo/vector-icons";

export default function OnboardingAccount() {
  const router = useRouter();
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const { t } = useTranslation();
  const posthog = usePostHog();

  const planSections = useMemo(
    () => [
      {
        key: "free",
        icon: "unlock",
        accent: "rgba(255, 123, 84, 0.16)",
        title: t("onboarding.account.free.title", "Free plan includes"),
        items: [
          t(
            "onboarding.account.free.features.starter",
            "Everything you need to get started.",
          ),
          t(
            "onboarding.account.free.features.wordlists",
            "Wordlists sync across every device.",
          ),
          t("onboarding.account.free.features.quizzes", "Adaptive quizzes."),
        ],
      },
      {
        key: "premium",
        icon: "star",
        accent: "rgba(156, 39, 176, 0.16)",
        title: t("onboarding.account.premium.title", "Premium unlocks"),
        items: [
          t(
            "onboarding.account.premium.features.quizzes",
            "All quiz modes, AI visuals, audio packs, and smarter review drills.",
          ),
          t(
            "onboarding.account.premium.features.offline",
            "Offline practice plus progress that syncs the moment you reconnect.",
          ),
          t(
            "onboarding.account.free.features.analytics",
            "7-day progress snapshots and stats.",
          ),
          t("onboarding.account.premium.features.analytics", "Cancel anytime."),
        ],
      },
    ],
    [t],
  );

  const finish = async (to: "/signup" | "/signin") => {
    posthog.capture("onboarding_completed");
    router.replace(to);
  };

  return (
    <OnboardingLayout
      step={3}
      totalSteps={3}
      showBack
      backLabel={t("common.back", "Back")}
      onBack={() => router.replace("/onboarding/features")}
      stepLabel={t("onboarding.stepIndicator", {
        step: 3,
        total: 3,
        defaultValue: "Step 3 of 3",
      })}
      contentStyle={styles.content}
      footer={
        <View style={styles.footerActions}>
          <TouchableOpacity
            onPress={() => finish("/signup")}
            style={styles.primaryButton}
            accessibilityRole="button"
          >
            <Text style={styles.primaryText}>
              {t("onboarding.account.cta", "Your learning starts now")}
            </Text>
            <Text style={styles.primarySubtext}>
              {t("onboarding.account.ctaTime", "Takes less than a minute")}
            </Text>
            <Text style={styles.primarySubtext}>
              {t(
                "onboarding.account.ctaNote",
                "Free to start - No credit card required",
              )}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            onPress={() => finish("/signin")}
            style={styles.secondaryLink}
            accessibilityRole="button"
          >
            <Text style={styles.secondaryText}>
              {t(
                "onboarding.account.haveAccount",
                "Already have an account? Sign in",
              )}
            </Text>
          </TouchableOpacity>
        </View>
      }
    >
      <ScrollView
        showsVerticalScrollIndicator={false}
        contentContainerStyle={styles.scrollContent}
        style={styles.scroll}
      >
        <View style={styles.containerCard}>
          <View style={styles.headerRow}>
            <Text style={styles.title}>
              {t("onboarding.account.title", "Create your account")}
            </Text>
          </View>
          <Text style={styles.subtitle}>
            {t("onboarding.account.subtitle", "Your words won’t forget you.")}
          </Text>

          <View style={styles.planSections}>
            {planSections.map((section) => (
              <View key={section.key} style={styles.planCard}>
                <View style={styles.planHeader}>
                  <View
                    style={[
                      styles.planIcon,
                      { backgroundColor: section.accent },
                    ]}
                  >
                    <Feather
                      name={section.icon as any}
                      size={18}
                      color={theme.colors.primary}
                    />
                  </View>
                  <Text style={styles.planTitle}>{section.title}</Text>
                </View>
                <View style={styles.planItems}>
                  {section.items.map((item, index) => (
                    <View
                      key={`${section.key}-${index}`}
                      style={styles.planItemRow}
                    >
                      <View
                        style={[
                          styles.planBullet,
                          { backgroundColor: theme.colors.primary },
                        ]}
                      />
                      <Text style={styles.planItemText}>{item}</Text>
                    </View>
                  ))}
                </View>
              </View>
            ))}
          </View>
        </View>
      </ScrollView>
    </OnboardingLayout>
  );
}

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    content: {
      flexGrow: 1,
    },
    scrollContent: {
      flexGrow: 1,
      gap: 20,
      paddingBottom: 160,
    },
    scroll: {
      flex: 1,
    },
    containerCard: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: 28,
      padding: 24,
      gap: 18,
      shadowColor: "#000",
      shadowOpacity: 0.08,
      shadowOffset: { width: 0, height: 12 },
      shadowRadius: 22,
      elevation: 4,
    },
    headerRow: {
      flexDirection: "row",
      alignItems: "center",
    },
    title: {
      fontSize: 26,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    subtitle: {
      fontSize: 15,
      lineHeight: 22,
      color: theme.colors.text.secondary,
    },
    planSections: {
      gap: 14,
    },
    planCard: {
      borderRadius: 18,
      padding: 16,
      borderWidth: 1,
      borderColor: "rgba(255, 123, 84, 0.08)",
      gap: 12,
    },
    planHeader: {
      flexDirection: "row",
      alignItems: "center",
      gap: 12,
    },
    planIcon: {
      width: 32,
      height: 32,
      borderRadius: 16,
      alignItems: "center",
      justifyContent: "center",
    },
    planTitle: {
      fontSize: 16,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    planItems: {
      gap: 10,
    },
    planItemRow: {
      flexDirection: "row",
      gap: 10,
    },
    planBullet: {
      width: 8,
      height: 8,
      borderRadius: 4,
      marginTop: 7,
    },
    planItemText: {
      flex: 1,
      color: theme.colors.text.primary,
      fontSize: 14,
      lineHeight: 20,
    },
    footerActions: {
      gap: 16,
    },
    primaryButton: {
      backgroundColor: theme.colors.primary,
      paddingVertical: 16,
      borderRadius: 18,
      alignItems: "center",
      shadowColor: "#FF7B54",
      shadowOpacity: 0.35,
      shadowOffset: { width: 0, height: 12 },
      shadowRadius: 18,
      elevation: 3,
    },
    primaryText: {
      color: theme.colors.text.inverse,
      fontWeight: "700",
      fontSize: 16,
    },
    secondaryLink: {
      alignItems: "center",
      paddingVertical: 4,
    },
    primarySubtext: {
      marginTop: 4,
      color: theme.colors.text.inverse,
      fontSize: 12,
      fontWeight: "600",
      opacity: 0.9,
    },
    secondaryText: {
      color: theme.colors.text.secondary,
      fontWeight: "600",
    },
  });
