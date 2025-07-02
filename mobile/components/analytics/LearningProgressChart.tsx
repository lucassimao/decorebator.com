import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { LineChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { getChartColors, getChartConfig } from "./theme";
import { LearningProgress } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive } from "@/hooks/useResponsive";

interface LearningProgressChartProps {
  learningProgress?: LearningProgress[];
}

export const LearningProgressChart: React.FC<LearningProgressChartProps> = ({
  learningProgress,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType, contentWidth } = useResponsive();
  const chartConfig = getChartConfig(theme);
  const chartColors = getChartColors(theme);
  const styles = createStyles(theme, isTablet, deviceType);

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
          width={contentWidth - 40}
          height={isTablet ? 280 : 220}
          chartConfig={{
            ...chartConfig,
            propsForDots: {
              r: isTablet ? "8" : "6",
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

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  deviceType: string,
) =>
  StyleSheet.create({
    chartCard: {
      backgroundColor: theme.colors.background.surface,
      marginHorizontal: 0, // Remove for responsive container
      marginVertical: theme.spacing.xs,
      borderRadius: theme.borderRadius.lg,
      padding: theme.spacing.md,
      ...theme.shadows.md,
    },
    chartTitle: {
      fontSize: theme.typography.sizes.title,
      fontWeight: theme.typography.weights.bold,
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.xs,
    },
    chartSubtitle: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
      marginBottom: theme.spacing.md,
    },
    chart: {
      marginVertical: theme.spacing.xs,
      borderRadius: theme.borderRadius.lg,
      marginLeft: -theme.spacing.md,
    },
    emptyChartContainer: {
      height: isTablet ? 280 : 220,
      justifyContent: "center",
      alignItems: "center",
      backgroundColor: theme.colors.background.default,
      borderRadius: theme.borderRadius.md,
    },
    emptyChartText: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.disabled,
    },
  });
