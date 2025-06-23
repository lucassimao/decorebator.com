import React from "react";
import { View, Text, StyleSheet, Dimensions } from "react-native";
import { LineChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { getChartColors, getChartConfig } from "./theme";
import { LearningProgress } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";

const screenWidth = Dimensions.get("window").width;

interface LearningProgressChartProps {
  learningProgress?: LearningProgress[];
}

export const LearningProgressChart: React.FC<LearningProgressChartProps> = ({
  learningProgress,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const chartConfig = getChartConfig(theme);
  const chartColors = getChartColors(theme);
  const styles = createStyles(theme);

  const lineChartData = {
    labels:
      learningProgress?.slice(-7).map((p) => {
        // Parse date as local time to avoid timezone conversion issues
        // Backend sends ISO timestamps like "2025-06-11T00:00:00Z", extract date part
        const datePart = p.date.split("T")[0];
        const [year, month, day] = datePart.split("-").map(Number);
        const date = new Date(year, month - 1, day);
        return `${date.getMonth() + 1}/${date.getDate()}`;
      }) || [],
    datasets: [
      {
        data: learningProgress?.slice(-7).map((p) => p.wordsStudied) || [],
        color: (opacity = 1) => chartColors.primary(opacity),
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
              stroke: theme.colors.primary,
              fill: theme.colors.background.default,
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
    chart: {
      marginVertical: 8,
      borderRadius: 16,
      marginLeft: -20,
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
  });
