import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { useTranslation } from "react-i18next";
import { WordMasteryStats } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";

interface TopWordsSectionProps {
  wordMastery?: WordMasteryStats[];
}

export const TopWordsSection: React.FC<TopWordsSectionProps> = ({
  wordMastery,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

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

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    chartCard: {
      backgroundColor: theme.colors.background.surface,
      marginHorizontal: 16,
      marginVertical: 8,
      borderRadius: 16,
      padding: 20,
      shadowColor: "#000",
      shadowOffset: {
        width: 0,
        height: 2,
      },
      shadowOpacity: 0.1,
      shadowRadius: 3.84,
      elevation: 5,
    },
    chartTitle: {
      fontSize: 20,
      fontWeight: "bold",
      color: theme.colors.text.primary,
      marginBottom: 4,
    },
    chartSubtitle: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginBottom: 16,
    },
    emptyChartContainer: {
      height: 200,
      justifyContent: "center",
      alignItems: "center",
      backgroundColor: theme.colors.background.default,
      borderRadius: 12,
    },
    emptyChartText: {
      fontSize: 16,
      color: theme.colors.text.placeholder,
    },
    wordsList: {
      marginTop: 8,
    },
    wordItem: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: 12,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.border,
    },
    wordItemLeft: {
      flexDirection: "row",
      alignItems: "center",
      flex: 1,
    },
    wordRankCircle: {
      width: 32,
      height: 32,
      borderRadius: 16,
      backgroundColor: theme.colors.background.subtle,
      justifyContent: "center",
      alignItems: "center",
      marginRight: 12,
    },
    wordRankGold: {
      backgroundColor: "#FFD700",
    },
    wordRank: {
      fontSize: 16,
      fontWeight: "bold",
      color: theme.colors.text.primary,
    },
    wordRankTextGold: {
      color: theme.mode === "dark" ? theme.colors.text.primary : "#FFFFFF",
    },
    wordName: {
      fontSize: 16,
      color: theme.colors.text.primary,
      flex: 1,
    },
    wordStats: {
      flexDirection: "row",
      alignItems: "center",
      gap: 8,
    },
    masteryBadge: {
      backgroundColor: theme.colors.success,
      paddingHorizontal: 10,
      paddingVertical: 4,
      borderRadius: 12,
    },
    wordMastery: {
      fontSize: 14,
      fontWeight: "bold",
      color: "#FFFFFF",
    },
    boxBadge: {
      backgroundColor: theme.colors.background.subtle,
      paddingHorizontal: 10,
      paddingVertical: 4,
      borderRadius: 12,
    },
    wordBox: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
  });
