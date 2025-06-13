import React from "react";
import { View, Text, StyleSheet, Dimensions, ScrollView } from "react-native";
import { StackedBarChart } from "react-native-chart-kit";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { colors, chartConfig } from "./theme";
import { HistoricalBoxDistributionResponse } from "@/api/analytics";

const screenWidth = Dimensions.get("window").width;

interface HistoricalBoxDistributionChartProps {
  historicalBoxDistribution?: HistoricalBoxDistributionResponse;
}

export const HistoricalBoxDistributionChart: React.FC<
  HistoricalBoxDistributionChartProps
> = ({ historicalBoxDistribution }) => {
  const { t } = useTranslation();

  const boxColors = {
    box1: "#FF6B6B", // Red - New/Difficult
    box2: "#FF8E53", // Orange Red
    box3: "#FFB74D", // Orange
    box4: "#FFD54F", // Yellow
    box5: "#AED581", // Light Green
    box6: "#81C784", // Green
    box7: "#4CAF50", // Full Green - Mastered
  };

  const boxNames = [
    {
      key: "box1",
      name: t("analytics.charts.boxDistribution.box1"),
      color: boxColors.box1,
    },
    {
      key: "box2",
      name: t("analytics.charts.boxDistribution.box2"),
      color: boxColors.box2,
    },
    {
      key: "box3",
      name: t("analytics.charts.boxDistribution.box3"),
      color: boxColors.box3,
    },
    {
      key: "box4",
      name: t("analytics.charts.boxDistribution.box4"),
      color: boxColors.box4,
    },
    {
      key: "box5",
      name: t("analytics.charts.boxDistribution.box5"),
      color: boxColors.box5,
    },
    {
      key: "box6",
      name: t("analytics.charts.boxDistribution.box6"),
      color: boxColors.box6,
    },
    {
      key: "box7",
      name: t("analytics.charts.boxDistribution.box7"),
      color: boxColors.box7,
    },
  ];

  const prepareChartData = () => {
    if (
      !historicalBoxDistribution?.distribution ||
      historicalBoxDistribution.distribution.length === 0
    ) {
      return null;
    }

    // For 7 days, show all data points (no sampling needed)
    const data = historicalBoxDistribution.distribution.slice().reverse(); // Show oldest to newest

    // Prepare labels (dates) - use shorter format for 7 days
    const labels = data.map((item) => {
      // Parse date as local time to avoid timezone conversion issues
      // Backend sends ISO timestamps like "2025-06-11T00:00:00Z", extract date part
      const datePart = item.date.split("T")[0];
      const [year, month, day] = datePart.split("-").map(Number);
      const date = new Date(year, month - 1, day); // month is 0-indexed
      const today = new Date();

      // Calculate difference using just the date parts to avoid timezone issues
      const dateOnly = new Date(
        date.getFullYear(),
        date.getMonth(),
        date.getDate(),
      );
      const todayOnly = new Date(
        today.getFullYear(),
        today.getMonth(),
        today.getDate(),
      );
      const diffTime = todayOnly.getTime() - dateOnly.getTime();
      const diffDays = Math.round(diffTime / (1000 * 60 * 60 * 24));

      if (diffDays === 0) {
        return "Today";
      } else if (diffDays === 1) {
        return "Yesterday";
      }
      return `${date.getMonth() + 1}/${date.getDate()}`;
    });

    // Prepare data for each box - use actual counts, not percentages for better readability
    const stackedData = data.map((item) => {
      return [
        item.boxes.box1 || 0,
        item.boxes.box2 || 0,
        item.boxes.box3 || 0,
        item.boxes.box4 || 0,
        item.boxes.box5 || 0,
        item.boxes.box6 || 0,
        item.boxes.box7 || 0,
      ];
    });

    return {
      labels,
      legend: ["Box 1", "Box 2", "Box 3", "Box 4", "Box 5", "Box 6", "Box 7"],
      data: stackedData,
      barColors: Object.values(boxColors),
    };
  };

  const chartData = prepareChartData();

  return (
    <View style={styles.chartCard}>
      <Text style={styles.chartTitle}>
        {t("analytics.charts.historicalBoxDistribution.title")}
      </Text>
      <Text style={styles.chartSubtitle}>
        {t("analytics.charts.historicalBoxDistribution.subtitle")}
      </Text>
      {chartData ? (
        <>
          <StackedBarChart
            data={chartData}
            width={screenWidth - 72}
            height={260}
            chartConfig={{
              ...chartConfig,
              color: (opacity = 1) => `rgba(134, 65, 244, ${opacity})`,
              barPercentage: 0.7,
              decimalPlaces: 0,
              propsForLabels: {
                fontSize: 12,
              },
              formatYLabel: (value: string) =>
                `${Math.round(parseFloat(value))}`,
            }}
            style={styles.chart}
            withHorizontalLabels={true}
            hideLegend={true}
            yAxisLabel=""
          />
          {/* Box Legend - Vertical Layout */}
          <View style={styles.legendContainer}>
            <Text style={styles.legendTitle}>
              {t(
                "analytics.charts.historicalBoxDistribution.currentDistribution",
              )}
            </Text>
            <View style={styles.boxLegendWrapper}>
              {boxNames.map((box, index) => (
                <View key={box.key} style={styles.boxLegendItem}>
                  <View style={styles.boxLegendLeft}>
                    <View
                      style={[
                        styles.boxLegendDot,
                        { backgroundColor: box.color },
                      ]}
                    />
                    <Text style={styles.boxLegendText}>{box.name}</Text>
                  </View>
                  {chartData.data.length > 0 && (
                    <Text style={styles.boxLegendPercentage}>
                      {chartData.data[chartData.data.length - 1][index]}
                    </Text>
                  )}
                </View>
              ))}
            </View>
          </View>
          {/* Divider */}
          <View style={styles.divider} />
          {/* Progress Summary */}
          <View style={styles.progressSummaryContainer}>
            {chartData.data.length > 0 && (
              <View style={styles.progressSummaryRow}>
                <Text style={styles.progressSummaryLabel}>
                  {t(
                    "analytics.charts.historicalBoxDistribution.currentMastery",
                  )}
                </Text>
                <Text style={styles.progressSummaryValue}>
                  {chartData.data[chartData.data.length - 1][6]}
                </Text>
              </View>
            )}
          </View>
          {/* Explanation */}
          <View style={styles.explanationContainer}>
            <MaterialIcons
              name="info-outline"
              size={20}
              color={colors.textMedium}
            />
            <Text style={styles.explanationText}>
              {t(
                "analytics.charts.historicalBoxDistribution.stackedExplanation",
              )}
            </Text>
          </View>
        </>
      ) : (
        <View style={styles.emptyChartContainer}>
          <Text style={styles.emptyChartText}>
            {t("analytics.empty.noHistoricalData")}
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
    marginTop: 16,
    marginBottom: 8,
    paddingHorizontal: 8,
  },
  legendTitle: {
    fontSize: 16,
    fontWeight: "600",
    color: colors.textDark,
    marginBottom: 12,
  },
  boxLegendWrapper: {
    flexDirection: "column",
  },
  boxLegendItem: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingVertical: 6,
  },
  boxLegendLeft: {
    flexDirection: "row",
    alignItems: "center",
    flex: 1,
  },
  boxLegendDot: {
    width: 12,
    height: 12,
    borderRadius: 6,
    marginRight: 8,
  },
  boxLegendText: {
    fontSize: 13,
    color: colors.textDark,
    flex: 1,
  },
  boxLegendPercentage: {
    fontSize: 14,
    fontWeight: "600",
    color: colors.textDark,
    minWidth: 45,
    textAlign: "right",
  },
  divider: {
    height: 1,
    backgroundColor: colors.divider,
    marginHorizontal: 8,
    marginVertical: 12,
  },
  progressSummaryContainer: {
    paddingHorizontal: 8,
  },
  progressSummaryRow: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingVertical: 8,
  },
  progressSummaryLabel: {
    fontSize: 14,
    color: colors.textMedium,
  },
  progressSummaryValue: {
    fontSize: 18,
    fontWeight: "bold",
    color: "#4CAF50", // Box 7 green color
  },
  explanationContainer: {
    flexDirection: "row",
    alignItems: "flex-start",
    marginTop: 16,
    paddingHorizontal: 8,
    paddingVertical: 12,
    backgroundColor: colors.backgroundLight,
    borderRadius: 8,
  },
  explanationText: {
    fontSize: 13,
    color: colors.textMedium,
    lineHeight: 18,
    flex: 1,
    marginLeft: 8,
  },
});
