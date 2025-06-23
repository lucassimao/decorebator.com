import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { WordlistStats } from "@/api/analytics";

interface StatsGridProps {
  stats?: WordlistStats;
}

export const StatsGrid: React.FC<StatsGridProps> = ({ stats }) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

  return (
    <View style={styles.statsGrid}>
      <View style={styles.statCard}>
        <View style={styles.statIconContainer}>
          <Text style={styles.statIcon}>📚</Text>
        </View>
        <Text style={styles.statValue}>{stats?.wordsStudiedToday || 0}</Text>
        <Text style={styles.statLabel}>{t("analytics.stats.wordsToday")}</Text>
      </View>

      <View style={[styles.statCard, styles.statCardHighlight]}>
        <View style={styles.statIconContainer}>
          <Text style={styles.statIcon}>🔥</Text>
        </View>
        <Text style={[styles.statValue, styles.statValueHighlight]}>
          {stats?.currentStreak || 0}
        </Text>
        <Text style={styles.statLabel}>{t("analytics.stats.dayStreak")}</Text>
      </View>

      <View style={styles.statCard}>
        <View style={styles.statIconContainer}>
          <Text style={styles.statIcon}>🏆</Text>
        </View>
        <Text style={[styles.statValue, styles.statValueSuccess]}>
          {stats?.wordsMastered || 0}
        </Text>
        <Text style={styles.statLabel}>{t("analytics.stats.mastered")}</Text>
      </View>

      <View style={styles.statCard}>
        <View style={styles.statIconContainer}>
          <Text style={styles.statIcon}>🎯</Text>
        </View>
        <Text style={styles.statValue}>
          {Math.round(stats?.accuracyToday || 0)}%
        </Text>
        <Text style={styles.statLabel}>{t("analytics.stats.accuracy")}</Text>
      </View>
    </View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    statsGrid: {
      flexDirection: "row",
      flexWrap: "wrap",
      paddingHorizontal: 12,
      paddingVertical: 10,
      justifyContent: "space-between",
    },
    statCard: {
      backgroundColor: theme.colors.background.surface,
      width: "48%",
      marginBottom: 12,
      borderRadius: 16,
      padding: 16,
      shadowColor: theme.colors.text.primary,
      shadowOffset: {
        width: 0,
        height: 2,
      },
      shadowOpacity: 0.1,
      shadowRadius: 3.84,
      elevation: 5,
      alignItems: "center",
    },
    statCardHighlight: {
      backgroundColor:
        theme.mode === "light" ? "#FFDCC3" : theme.colors.background.elevated,
    },
    statIconContainer: {
      marginBottom: 8,
    },
    statIcon: {
      fontSize: 24,
    },
    statValue: {
      fontSize: 32,
      fontWeight: "bold",
      color: theme.colors.text.primary,
      marginBottom: 4,
    },
    statValueHighlight: {
      color: theme.colors.primary,
    },
    statValueSuccess: {
      color: theme.colors.success,
    },
    statLabel: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
  });
