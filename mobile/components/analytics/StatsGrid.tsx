import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { useTranslation } from "react-i18next";
import { colors } from "./theme";
import { OverviewStats } from "@/api/analytics";

interface StatsGridProps {
  overviewStats?: OverviewStats;
}

export const StatsGrid: React.FC<StatsGridProps> = ({ overviewStats }) => {
  const { t } = useTranslation();

  return (
    <View style={styles.statsGrid}>
      <View style={styles.statCard}>
        <View style={styles.statIconContainer}>
          <Text style={styles.statIcon}>📚</Text>
        </View>
        <Text style={styles.statValue}>
          {overviewStats?.wordsStudiedToday || 0}
        </Text>
        <Text style={styles.statLabel}>{t("analytics.stats.wordsToday")}</Text>
      </View>

      <View style={[styles.statCard, styles.statCardHighlight]}>
        <View style={styles.statIconContainer}>
          <Text style={styles.statIcon}>🔥</Text>
        </View>
        <Text style={[styles.statValue, styles.statValueHighlight]}>
          {overviewStats?.currentStreak || 0}
        </Text>
        <Text style={styles.statLabel}>{t("analytics.stats.dayStreak")}</Text>
      </View>

      <View style={styles.statCard}>
        <View style={styles.statIconContainer}>
          <Text style={styles.statIcon}>🏆</Text>
        </View>
        <Text style={[styles.statValue, styles.statValueSuccess]}>
          {overviewStats?.wordsMastered || 0}
        </Text>
        <Text style={styles.statLabel}>{t("analytics.stats.mastered")}</Text>
      </View>

      <View style={styles.statCard}>
        <View style={styles.statIconContainer}>
          <Text style={styles.statIcon}>🎯</Text>
        </View>
        <Text style={styles.statValue}>
          {Math.round(overviewStats?.accuracyToday || 0)}%
        </Text>
        <Text style={styles.statLabel}>{t("analytics.stats.accuracy")}</Text>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  statsGrid: {
    flexDirection: "row",
    flexWrap: "wrap",
    paddingHorizontal: 12,
    paddingVertical: 10,
    justifyContent: "space-between",
  },
  statCard: {
    backgroundColor: colors.white,
    width: "48%",
    marginBottom: 12,
    borderRadius: 16,
    padding: 16,
    shadowColor: "#000",
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
    backgroundColor: colors.backgroundOrange,
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
    color: colors.textDark,
    marginBottom: 4,
  },
  statValueHighlight: {
    color: colors.primary,
  },
  statValueSuccess: {
    color: colors.success,
  },
  statLabel: {
    fontSize: 14,
    color: colors.textMedium,
    textAlign: "center",
  },
});
