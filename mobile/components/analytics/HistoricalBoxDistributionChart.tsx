import React from "react";
import { View, Text, StyleSheet, Dimensions } from "react-native";
import { BarChart } from "react-native-chart-kit";
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

  const boxColorGradient = [
    "#FF6B6B", // Box 1 - Red (hardest)
    "#FF8E53", // Box 2 - Orange Red
    "#FFB74D", // Box 3 - Orange
    "#FFD54F", // Box 4 - Yellow
    "#AED581", // Box 5 - Light Green
    "#81C784", // Box 6 - Green
    "#4CAF50", // Box 7 - Full Green (mastered)
  ];

  const prepareChartData = () => {
    if (!historicalBoxDistribution?.distribution || historicalBoxDistribution.distribution.length === 0) {
      return null;
    }

    const data = historicalBoxDistribution.distribution
      .slice()
      .reverse(); // Show oldest to newest

    // Take every 3rd day to avoid crowding (for 30 days, show ~10 points)
    const sampledData = data.filter((_, index) => index % 3 === 0).slice(0, 10);

    // Prepare labels (dates)
    const labels = sampledData.map((item) => {
      const date = new Date(item.date);
      return `${date.getMonth() + 1}/${date.getDate()}`;
    });

    // Show progression: Box 7 (mastered) vs Box 1 (new/difficult)
    // This gives a good sense of learning progress over time
    const masteredWordsData = sampledData.map((item) => item.boxes.box7);
    const newWordsData = sampledData.map((item) => item.boxes.box1);

    // Create a progress ratio: mastered vs new words
    const progressData = sampledData.map((item) => {
      const total = item.boxes.box1 + item.boxes.box2 + item.boxes.box3 + 
                   item.boxes.box4 + item.boxes.box5 + item.boxes.box6 + item.boxes.box7;
      return total > 0 ? Math.round((item.boxes.box7 / total) * 100) : 0;
    });

    return {
      labels,
      datasets: [
        {
          data: progressData,
        },
      ],
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
          <BarChart
            data={chartData}
            width={screenWidth - 40}
            height={220}
            yAxisLabel=""
            yAxisSuffix="%"
            chartConfig={{
              ...chartConfig,
              color: (opacity = 1) => boxColorGradient[6], // Green for mastery progress
              barPercentage: 0.7,
            }}
            style={styles.chart}
            showValuesOnTopOfBars={true}
            withHorizontalLabels={true}
            fromZero={true}
          />
          {/* Progress Legend */}
          <View style={styles.progressLegendContainer}>
            <View style={styles.progressLegendItem}>
              <View style={styles.progressLegendLeft}>
                <View
                  style={[
                    styles.boxLegendDot,
                    { backgroundColor: boxColorGradient[6] }, // Green
                  ]}
                />
                <Text style={styles.boxLegendText}>
                  {t("analytics.charts.historicalBoxDistribution.masteryPercentage")}
                </Text>
              </View>
            </View>
          </View>
          {/* Explanation */}
          <View style={styles.explanationContainer}>
            <MaterialIcons
              name="info-outline"
              size={20}
              color={colors.textMedium}
            />
            <Text style={styles.explanationText}>
              {t("analytics.charts.historicalBoxDistribution.explanation")}
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
  boxLegendContainer: {
    marginTop: 16,
  },
  progressLegendContainer: {
    marginTop: 16,
  },
  boxLegendItem: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingVertical: 6,
    paddingHorizontal: 8,
  },
  progressLegendItem: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    paddingVertical: 8,
    paddingHorizontal: 8,
  },
  boxLegendLeft: {
    flexDirection: "row",
    alignItems: "center",
    flex: 1,
  },
  progressLegendLeft: {
    flexDirection: "row",
    alignItems: "center",
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