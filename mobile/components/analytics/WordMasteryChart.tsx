import React from "react";
import { View, Text, StyleSheet, Dimensions } from "react-native";
import { ProgressChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { colors, chartColors, chartConfig } from "./theme";
import { WordMasteryStats } from "@/api/analytics";

const screenWidth = Dimensions.get("window").width;

interface WordMasteryChartProps {
  wordMastery?: WordMasteryStats[];
}

export const WordMasteryChart: React.FC<WordMasteryChartProps> = ({
  wordMastery,
}) => {
  const { t } = useTranslation();

  const progressChartData = {
    labels: wordMastery?.slice(0, 6).map((w) => w.word.substring(0, 8)) || [],
    data: wordMastery?.slice(0, 6).map((w) => w.masteryLevel) || [],
  };

  return (
    <View style={styles.chartCard}>
      <Text style={styles.chartTitle}>
        {t("analytics.charts.wordMastery.title")}
      </Text>
      <Text style={styles.chartSubtitle}>
        {t("analytics.charts.wordMastery.subtitle")}
      </Text>
      {progressChartData.data.length > 0 ? (
        <>
          <ProgressChart
            data={progressChartData}
            width={screenWidth - 40}
            height={200}
            strokeWidth={16}
            radius={28}
            chartConfig={{
              ...chartConfig,
              color: (opacity = 1, index = 0) => {
                const color = chartColors[index % chartColors.length];
                return `rgba(${parseInt(color.slice(1, 3), 16)}, ${parseInt(color.slice(3, 5), 16)}, ${parseInt(color.slice(5, 7), 16)}, ${opacity})`;
              },
            }}
            hideLegend={true}
            style={styles.chart}
          />
          {/* Custom Legend */}
          <View style={styles.legendContainer}>
            {progressChartData.labels.map((label, index) => (
              <View key={index} style={styles.legendItem}>
                <View
                  style={[
                    styles.legendDot,
                    {
                      backgroundColor: chartColors[index % chartColors.length],
                    },
                  ]}
                />
                <Text style={styles.legendText}>{label}</Text>
                <Text style={styles.legendValue}>
                  {Math.round(progressChartData.data[index] * 100)}%
                </Text>
              </View>
            ))}
          </View>
        </>
      ) : (
        <View style={styles.emptyChartContainer}>
          <Text style={styles.emptyChartText}>
            {t("analytics.empty.noData")}
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
  chart: {
    marginVertical: 8,
    borderRadius: 16,
    marginLeft: -20,
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
  legendContainer: {
    flexDirection: "row",
    flexWrap: "wrap",
    marginTop: 16,
    paddingHorizontal: 8,
  },
  legendItem: {
    flexDirection: "row",
    alignItems: "center",
    marginRight: 16,
    marginBottom: 8,
  },
  legendDot: {
    width: 12,
    height: 12,
    borderRadius: 6,
    marginRight: 6,
  },
  legendText: {
    fontSize: 12,
    color: colors.textMedium,
    marginRight: 4,
  },
  legendValue: {
    fontSize: 12,
    fontWeight: "600",
    color: colors.textDark,
  },
});
