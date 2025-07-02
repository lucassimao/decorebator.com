import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { WordlistStats } from "@/api/analytics";
import { useResponsive } from "@/hooks/useResponsive";

interface StatsGridProps {
  stats?: WordlistStats;
}

export const StatsGrid: React.FC<StatsGridProps> = ({ stats }) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType } = useResponsive();
  const styles = createStyles(theme, isTablet, deviceType);

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

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  deviceType: string,
) =>
  StyleSheet.create({
    statsGrid: {
      flexDirection: "row",
      flexWrap: "wrap",
      paddingHorizontal: theme.spacing.sm,
      paddingVertical: theme.spacing.sm,
      justifyContent: "space-between",
      gap: theme.spacing.xs,
    },
    statCard: {
      backgroundColor: theme.colors.background.surface,
      width: isTablet && deviceType === "tablet-10" ? "23%" : "48%", // 4 columns on 10" tablets
      marginBottom: theme.spacing.sm,
      borderRadius: theme.borderRadius.lg,
      padding: theme.spacing.md,
      ...theme.shadows.md,
      alignItems: "center",
      minHeight: isTablet ? 120 : 100,
    },
    statCardHighlight: {
      backgroundColor:
        theme.mode === "light" ? "#FFDCC3" : theme.colors.background.elevated,
    },
    statIconContainer: {
      marginBottom: theme.spacing.xs,
    },
    statIcon: {
      fontSize: isTablet ? 32 : 24,
    },
    statValue: {
      fontSize: isTablet
        ? theme.typography.sizes.display
        : theme.typography.sizes.title,
      fontWeight: theme.typography.weights.bold,
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.xs,
    },
    statValueHighlight: {
      color: theme.colors.primary,
    },
    statValueSuccess: {
      color: theme.colors.success,
    },
    statLabel: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: theme.typography.lineHeights.caption,
    },
  });
