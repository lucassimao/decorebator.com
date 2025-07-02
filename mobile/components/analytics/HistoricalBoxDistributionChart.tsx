import { HistoricalBoxDistributionResponse } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";
import { MaterialIcons } from "@expo/vector-icons";
import React from "react";
import { useTranslation } from "react-i18next";
import { StyleSheet, Text, View } from "react-native";
import { StackedBarChart } from "react-native-chart-kit";
import { getChartConfig } from "./theme";
import { useResponsive } from "@/hooks/useResponsive";

interface HistoricalBoxDistributionChartProps {
  historicalBoxDistribution?: HistoricalBoxDistributionResponse;
}

export const HistoricalBoxDistributionChart: React.FC<
  HistoricalBoxDistributionChartProps
> = ({ historicalBoxDistribution }) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType, contentWidth } = useResponsive();
  const chartConfig = getChartConfig(theme);
  const styles = createStyles(theme, isTablet, deviceType);

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
            width={contentWidth - theme.spacing.lg}
            height={isTablet ? 360 : 260}
            chartConfig={{
              ...chartConfig,
              color: (opacity = 1) => `rgba(134, 65, 244, ${opacity})`,
              barPercentage: 0.7,
              strokeWidth: isTablet ? 3 : 2,
              decimalPlaces: 0,
              propsForLabels: {
                fontSize: isTablet ? 14 : 12,
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
              size={isTablet ? 24 : 20}
              color={theme.colors.text.secondary}
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
    chart: {
      marginVertical: theme.spacing.sm,
      borderRadius: theme.borderRadius.lg,
    },
    emptyChartContainer: {
      height: isTablet ? 360 : 200,
      justifyContent: "center",
      alignItems: "center",
      backgroundColor: theme.colors.background.default,
      borderRadius: theme.borderRadius.md,
    },
    emptyChartText: {
      fontSize: theme.typography.sizes.bodyLarge,
      color: theme.colors.text.placeholder,
    },
    legendContainer: {
      marginTop: theme.spacing.md,
      marginBottom: theme.spacing.sm,
      paddingHorizontal: theme.spacing.sm,
    },
    legendTitle: {
      fontSize: theme.typography.sizes.bodyLarge,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.md,
    },
    boxLegendWrapper: {
      flexDirection: "column",
    },
    boxLegendItem: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: theme.spacing.xs,
    },
    boxLegendLeft: {
      flexDirection: "row",
      alignItems: "center",
      flex: 1,
    },
    boxLegendDot: {
      width: isTablet ? 16 : 12,
      height: isTablet ? 16 : 12,
      borderRadius: isTablet ? 8 : 6,
      marginRight: theme.spacing.sm,
    },
    boxLegendText: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.primary,
      flex: 1,
    },
    boxLegendPercentage: {
      fontSize: theme.typography.sizes.body,
      fontWeight: "600",
      color: theme.colors.text.primary,
      minWidth: isTablet ? 60 : 45,
      textAlign: "right",
    },
    divider: {
      height: 1,
      backgroundColor: theme.colors.ui.border,
      marginHorizontal: theme.spacing.sm,
      marginVertical: theme.spacing.md,
    },
    progressSummaryContainer: {
      paddingHorizontal: theme.spacing.sm,
    },
    progressSummaryRow: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      paddingVertical: theme.spacing.sm,
    },
    progressSummaryLabel: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.secondary,
    },
    progressSummaryValue: {
      fontSize: theme.typography.sizes.bodyLarge,
      fontWeight: "bold",
      color: "#4CAF50", // Box 7 green color
    },
    explanationContainer: {
      flexDirection: "row",
      alignItems: "flex-start",
      marginTop: theme.spacing.md,
      paddingHorizontal: theme.spacing.sm,
      paddingVertical: theme.spacing.md,
      backgroundColor: theme.colors.background.subtle,
      borderRadius: theme.borderRadius.sm,
    },
    explanationText: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
      lineHeight: theme.typography.lineHeights.caption,
      flex: 1,
      marginLeft: theme.spacing.sm,
    },
  });
