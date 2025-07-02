import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { useTranslation } from "react-i18next";
import { WordMasteryStats } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive } from "@/hooks/useResponsive";

interface TopWordsSectionProps {
  wordMastery?: WordMasteryStats[];
}

export const TopWordsSection: React.FC<TopWordsSectionProps> = ({
  wordMastery,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType } = useResponsive();
  const styles = createStyles(theme, isTablet, deviceType);

  return (
    <View style={styles.chartCard}>
      <Text style={styles.chartTitle}>
        {t("analytics.charts.topWords.title")}
      </Text>
      <Text style={styles.chartSubtitle}>
        {t("analytics.charts.topWords.subtitle")}
      </Text>
      {wordMastery && wordMastery.length > 0 ? (
        <View style={styles.wordsList}>
          {wordMastery.slice(0, 5).map((word, index) => (
            <View key={word.wordId} style={styles.wordItem}>
              <View style={styles.wordItemLeft}>
                <View
                  style={[
                    styles.wordRankCircle,
                    index === 0 && styles.wordRankGold,
                  ]}
                >
                  <Text
                    style={[
                      styles.wordRank,
                      index === 0 && styles.wordRankTextGold,
                    ]}
                  >
                    {index + 1}
                  </Text>
                </View>
                <Text style={styles.wordName}>{word.word}</Text>
              </View>
              <View style={styles.wordStats}>
                <View style={styles.masteryBadge}>
                  <Text style={styles.wordMastery}>
                    {Math.round(word.masteryLevel * 100)}%
                  </Text>
                </View>
                <View style={styles.boxBadge}>
                  <Text style={styles.wordBox}>
                    {t("analytics.box", { number: word.highestBox })}
                  </Text>
                </View>
              </View>
            </View>
          ))}
        </View>
      ) : (
        <View style={styles.emptyChartContainer}>
          <Text style={styles.emptyChartText}>
            {t("analytics.empty.noMastered")}
          </Text>
        </View>
      )}
    </View>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  deviceType: string,
) =>
  StyleSheet.create({
    chartCard: {
      backgroundColor: theme.colors.background.surface,
      marginVertical: theme.spacing.sm,
      borderRadius: theme.borderRadius.lg,
      padding: theme.spacing.lg,
      ...theme.shadows.sm,
    },
    chartTitle: {
      fontSize: theme.typography.sizes.title,
      fontWeight: "bold",
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.xs,
    },
    chartSubtitle: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.secondary,
      marginBottom: theme.spacing.md,
    },
    emptyChartContainer: {
      height: isTablet ? 280 : 200,
      justifyContent: "center",
      alignItems: "center",
      backgroundColor: theme.colors.background.default,
      borderRadius: theme.borderRadius.md,
    },
    emptyChartText: {
      fontSize: theme.typography.sizes.bodyLarge,
      color: theme.colors.text.placeholder,
    },
    wordsList: {
      marginTop: theme.spacing.sm,
    },
    wordItem: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: theme.spacing.md,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.border,
    },
    wordItemLeft: {
      flexDirection: "row",
      alignItems: "center",
      flex: 1,
    },
    wordRankCircle: {
      width: isTablet ? 40 : 32,
      height: isTablet ? 40 : 32,
      borderRadius: isTablet ? 20 : 16,
      backgroundColor: theme.colors.background.subtle,
      justifyContent: "center",
      alignItems: "center",
      marginRight: theme.spacing.md,
    },
    wordRankGold: {
      backgroundColor: "#FFD700",
    },
    wordRank: {
      fontSize: theme.typography.sizes.bodyLarge,
      fontWeight: "bold",
      color: theme.colors.text.primary,
    },
    wordRankTextGold: {
      color: theme.mode === "dark" ? theme.colors.text.primary : "#FFFFFF",
    },
    wordName: {
      fontSize: theme.typography.sizes.bodyLarge,
      color: theme.colors.text.primary,
      flex: 1,
    },
    wordStats: {
      flexDirection: "row",
      alignItems: "center",
      gap: theme.spacing.sm,
    },
    masteryBadge: {
      backgroundColor: theme.colors.success,
      paddingHorizontal: theme.spacing.sm + 2,
      paddingVertical: theme.spacing.xs,
      borderRadius: theme.borderRadius.md,
    },
    wordMastery: {
      fontSize: theme.typography.sizes.body,
      fontWeight: "bold",
      color: "#FFFFFF",
    },
    boxBadge: {
      backgroundColor: theme.colors.background.subtle,
      paddingHorizontal: theme.spacing.sm + 2,
      paddingVertical: theme.spacing.xs,
      borderRadius: theme.borderRadius.md,
    },
    wordBox: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.secondary,
    },
  });
