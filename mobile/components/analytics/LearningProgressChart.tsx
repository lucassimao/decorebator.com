import React from "react";
import { View, Text, StyleSheet, Dimensions } from "react-native";
import { LineChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { colors, chartConfig } from "./theme";
import { LearningProgress } from "@/api/analytics";

const screenWidth = Dimensions.get("window").width;

interface LearningProgressChartProps {
  learningProgress?: LearningProgress[];
}

export const LearningProgressChart: React.FC<LearningProgressChartProps> = ({
  learningProgress,
}) => {
  const { t } = useTranslation();

  const lineChartData = {
    labels:
      learningProgress?.slice(-7).map((p) => {
        const date = new Date(p.date);
        return `${date.getMonth() + 1}/${date.getDate()}`;
      }) || [],
    datasets: [
      {
        data: learningProgress?.slice(-7).map((p) => p.wordsStudied) || [],
        color: (opacity = 1) => `rgba(255, 123, 84, ${opacity})`, // Primary orange
        strokeWidth: 3,
      },
    ],
  };

  return (
    <View style={styles.chartCard}>
      <Text style={styles.chartTitle}>
        {t("analytics.charts.progress.title")}
      </Text>
      <Text style={styles.chartSubtitle}>
        {t("analytics.charts.progress.subtitle")}
      </Text>
      {lineChartData.labels.length > 0 ? (
        <LineChart
          data={lineChartData}
          width={screenWidth - 40}
          height={220}
          chartConfig={{
            ...chartConfig,
            propsForDots: {
              r: "6",
              strokeWidth: "2",
              stroke: colors.primary,
              fill: colors.white,
            },
          }}
          bezier
          style={styles.chart}
        />
      ) : (
        <View style={styles.emptyChartContainer}>
          <Text style={styles.emptyChartText}>
            {t("analytics.empty.noProgress")}
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
});