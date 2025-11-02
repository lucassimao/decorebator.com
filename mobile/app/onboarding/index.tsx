import React, { useMemo } from "react";
import { View, Text, StyleSheet, TouchableOpacity } from "react-native";
import { useRouter } from "expo-router";
import { useTheme } from "@/contexts/ThemeContext";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { useTranslation } from "react-i18next";
import { usePostHog } from "posthog-react-native";
import { OnboardingLayout } from "@/components/onboarding/OnboardingLayout";
import { Feather } from "@expo/vector-icons";

const OnboardingWelcome = () => {
  const router = useRouter();
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const { t } = useTranslation();
  const posthog = usePostHog();

  const highlights = useMemo(
    () => [
      {
        icon: "zap",
        accent: "rgba(255, 123, 84, 0.16)",
        label: t(
          "onboarding.welcome.highlight.generator",
          "Create wordlists and our AI generates 8 tests, images, audio, and challenges",
        ),
      },
      {
        icon: "mic",
        accent: "rgba(0, 188, 212, 0.16)",
        label: t(
          "onboarding.welcome.highlight.chat",
          "AI voice coach drills the words you save out loud",
        ),
      },
      {
        icon: "layers",
        accent: "rgba(156, 39, 176, 0.14)",
        label: t(
          "onboarding.welcome.highlight.modes",
          "Eight quiz modes that keep practice fresh",
        ),
      },
      {
        icon: "bar-chart-2",
        accent: "rgba(76, 175, 80, 0.14)",
        label: t(
          "onboarding.welcome.highlight.progress",
          "Real-time analytics to celebrate your streaks",
        ),
      },
    ],
    [t],
  );

  const onContinue = () => {
    posthog.capture("onboarding_start");
    router.replace("/onboarding/features");
  };

  const onSkip = async () => {
    posthog.capture("onboarding_skipped", { at_step: "welcome" });
    await AsyncStorage.setItem("@onboarding_completed", "1");
    router.replace("/signup");
  };

  const skipLabel = t("common.skip", "Skip");
  const continueLabel = t("common.continue", "Continue");

  return (
    <OnboardingLayout
      step={1}
      totalSteps={4}
      showSkip
      skipLabel={skipLabel}
      onSkip={onSkip}
      stepLabel={t("onboarding.stepIndicator", {
        step: 1,
        total: 4,
        defaultValue: "Step 1 of 4",
      })}
      contentStyle={styles.content}
      footer={
        <TouchableOpacity
          onPress={onContinue}
          style={styles.primaryButton}
          accessibilityRole="button"
        >
          <Text style={styles.primaryText}>{continueLabel}</Text>
        </TouchableOpacity>
      }
    >
      <View style={styles.hero}>
        <Text style={styles.pillText}>
          {t("onboarding.welcome.greeting", "Welcome to")}
        </Text>
        <Text style={styles.title}>{t("welcome.title", "Decorebator")}</Text>
        <Text style={styles.subtitle}>
          {t(
            "welcome.subtitle",
            "Create your wordlist and Decorebator's AI will spin it into adaptive quizzes, vivid imagery, and daily practice.",
          )}
        </Text>
      </View>
      <View style={styles.highlights}>
        {highlights.map((item) => (
          <View
            key={item.label}
            style={[
              styles.highlightCard,
              {
                backgroundColor: theme.colors.background.surface,
                borderColor: "rgba(255, 123, 84, 0.08)",
              },
            ]}
          >
            <View style={[styles.iconChip, { backgroundColor: item.accent }]}>
              <Feather
                name={item.icon as any}
                size={18}
                color={theme.colors.primary}
              />
            </View>
            <Text
              style={[
                styles.highlightText,
                { color: theme.colors.text.primary },
              ]}
            >
              {item.label}
            </Text>
          </View>
        ))}
      </View>
    </OnboardingLayout>
  );
};

export default OnboardingWelcome;

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    content: {
      justifyContent: "center",
      paddingBottom: 16,
    },
    hero: {
      gap: 12,
      marginBottom: 28,
    },
    pillText: {
      fontSize: 14,
      fontWeight: "700",
      letterSpacing: 1,
      textTransform: "uppercase",
      color: "rgba(45, 52, 54, 0.6)",
    },
    title: {
      fontSize: 36,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    subtitle: {
      fontSize: 17,
      lineHeight: 24,
      color: theme.colors.text.secondary,
    },
    highlights: {
      gap: 16,
    },
    highlightCard: {
      borderRadius: 18,
      padding: 18,
      flexDirection: "row",
      alignItems: "center",
      gap: 12,
      borderWidth: 1,
      shadowColor: "#000",
      shadowOpacity: 0.05,
      shadowOffset: { width: 0, height: 6 },
      shadowRadius: 12,
      elevation: 2,
    },
    iconChip: {
      width: 40,
      height: 40,
      borderRadius: 20,
      alignItems: "center",
      justifyContent: "center",
    },
    highlightText: {
      flex: 1,
      fontSize: 15,
      fontWeight: "600",
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
  });
