import React, { useMemo, useState } from "react";
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
import AsyncStorage from "@react-native-async-storage/async-storage";
import { LANGUAGES } from "@/components/dashboard/CreateWordlistModal";
import { usePostHog } from "posthog-react-native";
import { OnboardingLayout } from "@/components/onboarding/OnboardingLayout";
import { Feather } from "@expo/vector-icons";

export default function OnboardingPersonalize() {
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const { t } = useTranslation();
  const router = useRouter();
  const posthog = usePostHog();

  const [language, setLanguage] = useState(LANGUAGES[0].code);
  const [goal, setGoal] = useState(10);

  const languageOptions = useMemo(() => LANGUAGES, []);
  const goalOptions = useMemo(() => [5, 10, 15, 20], []);

  const onContinue = async () => {
    await AsyncStorage.multiSet([
      ["@onboarding_language", language],
      ["@onboarding_daily_goal", String(goal)],
    ]);
    posthog.capture("onboarding_personalized", { language, goal });
    router.replace("/onboarding/account");
  };

  return (
    <OnboardingLayout
      step={3}
      totalSteps={4}
      showBack
      backLabel={t("common.back", "Back")}
      onBack={() => router.replace("/onboarding/features")}
      stepLabel={t("onboarding.stepIndicator", {
        step: 3,
        total: 4,
        defaultValue: "Step 3 of 4",
      })}
      contentStyle={styles.content}
      footer={
        <TouchableOpacity
          onPress={onContinue}
          style={styles.primaryButton}
          accessibilityRole="button"
        >
          <Text style={styles.primaryText}>
            {t("common.continue", "Continue")}
          </Text>
        </TouchableOpacity>
      }
    >
      <ScrollView
        showsVerticalScrollIndicator={false}
        contentContainerStyle={styles.scrollContent}
      >
        <View style={styles.header}>
          <Text style={styles.title}>
            {t("onboarding.personalize.title", "Let’s personalize this")}
          </Text>
          <Text style={styles.subtitle}>
            {t(
              "onboarding.personalize.subtitle",
              "Pick a language and daily focus so we can shape lessons around your schedule.",
            )}
          </Text>
        </View>
        <Text style={styles.sectionLabel}>
          {t("onboarding.personalize.language", "Learning language")}
        </Text>
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.languageRow}
          snapToAlignment="start"
          decelerationRate="fast"
        >
          {languageOptions.map((item) => {
            const active = language === item.code;
            return (
              <TouchableOpacity
                key={item.code}
                style={[
                  styles.languageCard,
                  {
                    borderColor: active ? theme.colors.primary : "transparent",
                    backgroundColor: theme.colors.background.surface,
                  },
                  active && styles.languageCardActive,
                ]}
                onPress={() => setLanguage(item.code)}
                accessibilityRole="button"
              >
                <Text style={styles.languageFlag}>{item.flag}</Text>
                <Text
                  style={[
                    styles.languageName,
                    active && styles.languageNameActive,
                  ]}
                >
                  {item.name}
                </Text>
              </TouchableOpacity>
            );
          })}
        </ScrollView>

        <Text style={styles.sectionLabel}>
          {t("onboarding.personalize.goal", "Daily focus")}
        </Text>
        <View style={styles.goalsRow}>
          {goalOptions.map((g) => {
            const active = goal === g;
            return (
              <TouchableOpacity
                key={g}
                style={[
                  styles.goalChip,
                  active && [
                    styles.goalChipActive,
                    { borderColor: theme.colors.primary },
                  ],
                ]}
                onPress={() => setGoal(g)}
                accessibilityRole="button"
              >
                <Feather
                  name="clock"
                  size={16}
                  color={
                    active ? theme.colors.primary : theme.colors.text.secondary
                  }
                />
                <Text
                  style={[styles.goalText, active && styles.goalTextActive]}
                >
                  {t(
                    "onboarding.personalize.goalValue",
                    "{{minutes}} min/day",
                    { minutes: g },
                  )}
                </Text>
              </TouchableOpacity>
            );
          })}
        </View>

        <View
          style={[
            styles.previewCard,
            { backgroundColor: theme.colors.background.surface },
          ]}
        >
          <View style={styles.previewHeader}>
            <Feather name="star" size={18} color={theme.colors.primary} />
            <Text style={styles.previewTitle}>
              {t("onboarding.personalize.preview.title", "Your study preview")}
            </Text>
          </View>
          <View style={styles.previewRow}>
            <Feather
              name="bookmark"
              size={16}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.previewText}>
              {t(
                "onboarding.personalize.preview.language",
                "Daily cards in {{language}}",
                {
                  language:
                    languageOptions.find((l) => l.code === language)?.name ??
                    "",
                },
              )}
            </Text>
          </View>
          <View style={styles.previewRow}>
            <Feather
              name="activity"
              size={16}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.previewText}>
              {t(
                "onboarding.personalize.preview.goal",
                "Goal set to {{minutes}} minutes",
                { minutes: goal },
              )}
            </Text>
          </View>
          <View style={styles.previewFooter}>
            <Feather name="gift" size={16} color={theme.colors.premium} />
            <Text style={styles.previewFooterText}>
              {t(
                "onboarding.personalize.preview.hint",
                "You can tweak this anytime in Settings.",
              )}
            </Text>
          </View>
        </View>
      </ScrollView>
    </OnboardingLayout>
  );
}

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    content: {
      paddingBottom: 16,
    },
    scrollContent: {
      paddingBottom: 32,
      gap: 22,
    },
    header: {
      gap: 12,
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
    sectionLabel: {
      fontSize: 14,
      fontWeight: "700",
      textTransform: "uppercase",
      letterSpacing: 1,
      color: "rgba(45, 52, 54, 0.65)",
    },
    languageRow: {
      gap: 12,
      paddingRight: 24,
      paddingVertical: 8,
    },
    languageCard: {
      width: 110,
      paddingVertical: 18,
      borderRadius: 20,
      borderWidth: 1,
      borderColor: "transparent",
      alignItems: "center",
      gap: 12,
      shadowColor: "#000",
      shadowOpacity: 0.05,
      shadowOffset: { width: 0, height: 6 },
      shadowRadius: 12,
      elevation: 2,
    },
    languageCardActive: {
      shadowOpacity: 0.12,
      shadowRadius: 16,
    },
    languageFlag: {
      fontSize: 32,
    },
    languageName: {
      fontSize: 15,
      fontWeight: "600",
      color: theme.colors.text.secondary,
    },
    languageNameActive: {
      color: theme.colors.primary,
    },
    goalsRow: {
      flexDirection: "row",
      flexWrap: "wrap",
      gap: 12,
    },
    goalChip: {
      flexDirection: "row",
      alignItems: "center",
      gap: 8,
      paddingVertical: 10,
      paddingHorizontal: 16,
      borderRadius: 18,
      backgroundColor: theme.colors.background.surface,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
    },
    goalChipActive: {
      backgroundColor: "rgba(255,123,84,0.12)",
    },
    goalText: {
      fontWeight: "600",
      color: theme.colors.text.secondary,
    },
    goalTextActive: {
      color: theme.colors.primary,
    },
    previewCard: {
      borderRadius: 24,
      padding: 20,
      gap: 14,
      shadowColor: "#000",
      shadowOpacity: 0.06,
      shadowOffset: { width: 0, height: 10 },
      shadowRadius: 16,
      elevation: 3,
    },
    previewHeader: {
      flexDirection: "row",
      alignItems: "center",
      gap: 10,
    },
    previewTitle: {
      fontSize: 16,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    previewRow: {
      flexDirection: "row",
      alignItems: "center",
      gap: 10,
    },
    previewText: {
      flex: 1,
      color: theme.colors.text.secondary,
      fontSize: 15,
    },
    previewFooter: {
      flexDirection: "row",
      alignItems: "center",
      gap: 10,
      paddingTop: 6,
      borderTopWidth: StyleSheet.hairlineWidth,
      borderTopColor: theme.colors.ui.divider,
    },
    previewFooterText: {
      flex: 1,
      color: theme.colors.text.secondary,
      fontSize: 14,
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
