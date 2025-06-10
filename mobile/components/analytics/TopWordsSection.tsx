import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { useTranslation } from "react-i18next";
import { colors } from "./theme";
import { WordMasteryStats } from "@/api/analytics";

interface TopWordsSectionProps {
  wordMastery?: WordMasteryStats[];
}

export const TopWordsSection: React.FC<TopWordsSectionProps> = ({ wordMastery }) => {
  const { t } = useTranslation();

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

const styles = StyleSheet.create({
  chartCard: {
    backgroundColor: colors.white,
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
    color: colors.textDark,
    marginBottom: 4,
  },
  chartSubtitle: {
    fontSize: 14,
    color: colors.textMedium,
    marginBottom: 16,
  },
  emptyChartContainer: {
    height: 200,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: colors.lightBackground,
    borderRadius: 12,
  },
  emptyChartText: {
    fontSize: 16,
    color: colors.textLight,
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
    borderBottomColor: colors.divider,
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
    backgroundColor: colors.lightBackground,
    justifyContent: "center",
    alignItems: "center",
    marginRight: 12,
  },
  wordRankGold: {
    backgroundColor: colors.gold,
  },
  wordRank: {
    fontSize: 16,
    fontWeight: "bold",
    color: colors.textDark,
  },
  wordRankTextGold: {
    color: colors.white,
  },
  wordName: {
    fontSize: 16,
    color: colors.textDark,
    flex: 1,
  },
  wordStats: {
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
  },
  masteryBadge: {
    backgroundColor: colors.success,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  wordMastery: {
    fontSize: 14,
    fontWeight: "bold",
    color: colors.white,
  },
  boxBadge: {
    backgroundColor: colors.lightBackground,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  wordBox: {
    fontSize: 14,
    color: colors.textMedium,
  },
});