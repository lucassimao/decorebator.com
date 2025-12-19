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
            "onboarding.account.free.features.wordlists",
            "Create and practice your own wordlists with synced progress on every device.",
          ),
          t(
            "onboarding.account.free.features.quizzes",
            "Adaptive quizzes with streak tracking and daily goal reminders.",
          ),
          t(
            "onboarding.account.free.features.analytics",
            "Core analytics to see wins from the past 7 days.",
          ),
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
            "All eight quiz modes, AI visuals, audio packs, and smarter review drills.",
          ),
          t(
            "onboarding.account.premium.features.offline",
            "Offline practice plus progress that syncs the moment you reconnect.",
          ),
          t(
            "onboarding.account.premium.features.analytics",
            "Deep-dive analytics and weekly insights to keep momentum high.",
          ),
        ],
      },
    ],
    [t],
  );

  const upgradeSteps = useMemo(
    () => [
      t(
        "onboarding.account.upgrade.steps.settings",
        "Open Settings from the dashboard menu once you're inside the app.",
      ),
      t(
        "onboarding.account.upgrade.steps.option",
        'Tap "Upgrade to Premium" under Account to review plan details.',
      ),
      t(
        "onboarding.account.upgrade.steps.confirm",
        "Choose your plan, confirm the upgrade, or start a free trial anytime.",
      ),
    ],
    [t],
  );

  const finish = async (to: "/signup" | "/signin") => {
    await AsyncStorage.setItem("@onboarding_completed", "1");
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
              {t("signup.title", "Sign Up")}
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
          <View
            style={[
              styles.iconHero,
              { backgroundColor: "rgba(255, 123, 84, 0.18)" },
            ]}
          >
            <Feather name="smile" size={26} color={theme.colors.primary} />
          </View>
          <Text style={styles.title}>
            {t("onboarding.account.title", "Create your account")}
          </Text>
          <Text style={styles.subtitle}>
            {t(
              "onboarding.account.subtitle",
              "Lock in your progress today, see what's included for free, and know exactly how to upgrade.",
            )}
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
                  {section.items.map((item) => (
                    <View key={item} style={styles.planItemRow}>
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

          <View style={styles.upgradeCard}>
            <View style={styles.planHeader}>
              <View
                style={[
                  styles.planIcon,
                  { backgroundColor: "rgba(76, 175, 80, 0.16)" },
                ]}
              >
                <Feather
                  name="arrow-up-circle"
                  size={18}
                  color={theme.colors.primary}
                />
              </View>
              <Text style={styles.planTitle}>
                {t("onboarding.account.upgrade.title", "How to upgrade later")}
              </Text>
            </View>
            <View style={styles.upgradeSteps}>
              {upgradeSteps.map((step, index) => (
                <View key={step} style={styles.upgradeRow}>
                  <View
                    style={[
                      styles.stepBadge,
                      { borderColor: theme.colors.primary },
                    ]}
                  >
                    <Text
                      style={[
                        styles.stepBadgeText,
                        { color: theme.colors.primary },
                      ]}
                    >
                      {index + 1}
                    </Text>
                  </View>
                  <Text style={styles.upgradeText}>{step}</Text>
                </View>
              ))}
            </View>
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
    iconHero: {
      width: 56,
      height: 56,
      borderRadius: 28,
      alignItems: "center",
      justifyContent: "center",
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
    upgradeCard: {
      marginTop: 12,
      borderRadius: 18,
      padding: 18,
      borderWidth: 1,
      borderColor: "rgba(76, 175, 80, 0.14)",
      gap: 14,
    },
    upgradeSteps: {
      gap: 14,
    },
    upgradeRow: {
      flexDirection: "row",
      gap: 12,
      alignItems: "flex-start",
    },
    stepBadge: {
      width: 28,
      height: 28,
      borderRadius: 14,
      alignItems: "center",
      justifyContent: "center",
      borderWidth: 1,
      marginTop: 2,
    },
    stepBadgeText: {
      fontWeight: "700",
    },
    upgradeText: {
      flex: 1,
      fontSize: 15,
      lineHeight: 22,
      color: theme.colors.text.primary,
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
    secondaryText: {
      color: theme.colors.text.secondary,
      fontWeight: "600",
    },
  });
