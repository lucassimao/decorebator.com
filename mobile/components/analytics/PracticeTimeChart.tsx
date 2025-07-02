import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { BarChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { getChartColors, getChartConfig } from "./theme";
import { PracticeTimeResponse } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive } from "@/hooks/useResponsive";

interface PracticeTimeChartProps {
  practiceTime?: PracticeTimeResponse;
}

export const PracticeTimeChart: React.FC<PracticeTimeChartProps> = ({
  practiceTime,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType, contentWidth } = useResponsive();
  const chartConfig = getChartConfig(theme);
  const chartColors = getChartColors(theme);
  const styles = createStyles(theme, isTablet, deviceType);

  const formatDateLabel = (dateString: string): string => {
    try {
      // Extract just the date part from ISO timestamp (e.g., "2025-06-11T00:00:00Z" -> "2025-06-11")
      const datePart = dateString.split("T")[0];
      const [year, month, day] = datePart.split("-").map(Number);

      // Validate parts exist and are numbers
      if (
        !year ||
        !month ||
        !day ||
        isNaN(year) ||
        isNaN(month) ||
        isNaN(day)
      ) {
        console.error("Invalid date parts:", {
          dateString,
          datePart,
          year,
          month,
          day,
        });
        // Fallback to UTC parsing
        const fallbackDate = new Date(dateString);
        return `${fallbackDate.getMonth() + 1}/${fallbackDate.getDate()}`;
      }

      const date = new Date(year, month - 1, day);

      // Check if date creation was successful
      if (isNaN(date.getTime())) {
        console.error("Failed to create date:", dateString);
        // Fallback to UTC parsing
        const fallbackDate = new Date(dateString);
        return `${fallbackDate.getMonth() + 1}/${fallbackDate.getDate()}`;
      }

      return `${date.getMonth() + 1}/${date.getDate()}`;
    } catch (error) {
      console.error("Error parsing date:", dateString, error);
      // Ultimate fallback - just show the original string
      return dateString;
    }
  };

  const formatMinutes = (minutes: number): string => {
    if (minutes < 1) {
      return "< 1m";
    } else if (minutes < 60) {
      return `${Math.round(minutes)}m`;
    } else {
      const hours = Math.floor(minutes / 60);
      const remainingMinutes = Math.round(minutes % 60);
      return remainingMinutes > 0
        ? `${hours}h ${remainingMinutes}m`
        : `${hours}h`;
    }
  };

  // Sort by date ascending and take last 7 days
  const sortedData =
    practiceTime?.practiceTime
      ?.slice()
      .sort((a, b) => {
        // Parse dates as local time for proper sorting
        // Extract date part from ISO timestamp
        const datePartA = a.date.split("T")[0];
        const datePartB = b.date.split("T")[0];
        const [yearA, monthA, dayA] = datePartA.split("-").map(Number);
        const [yearB, monthB, dayB] = datePartB.split("-").map(Number);
        const dateA = new Date(yearA, monthA - 1, dayA);
        const dateB = new Date(yearB, monthB - 1, dayB);

        // Check if dates are valid
        if (isNaN(dateA.getTime()) || isNaN(dateB.getTime())) {
          console.error("Invalid date during sorting:", a.date, b.date);
          return 0; // Keep original order if dates are invalid
        }

        return dateA.getTime() - dateB.getTime();
      })
      .slice(-7) || [];

  const chartData = {
    labels: sortedData.map((p) => formatDateLabel(p.date)),
    datasets: [
      {
        data: sortedData.map((p) => p.practiceTimeMinutes),
        // Custom format for chart values
        withDots: false,
      },
    ],
  };

  const totalPracticeTime = sortedData.reduce(
    (sum, p) => sum + p.practiceTimeMinutes,
    0,
  );
  const avgPracticeTime =
    sortedData.length > 0 ? totalPracticeTime / sortedData.length : 0;

  return (
    <View style={styles.chartCard}>
      <Text style={styles.chartTitle}>
        {t("analytics.charts.practiceTime.title")}
      </Text>
      <Text style={styles.chartSubtitle}>
        {t("analytics.charts.practiceTime.subtitle")}
      </Text>

      {/* Summary Stats */}
      <View style={styles.statsContainer}>
        <View style={styles.statItem}>
          <Text style={styles.statValue}>
            {formatMinutes(totalPracticeTime)}
          </Text>
          <Text style={styles.statLabel}>
            {t("analytics.charts.practiceTime.totalTime")}
          </Text>
        </View>
        <View style={styles.statItem}>
          <Text style={styles.statValue}>{formatMinutes(avgPracticeTime)}</Text>
          <Text style={styles.statLabel}>
            {t("analytics.charts.practiceTime.avgDaily")}
          </Text>
        </View>
      </View>

      {chartData.labels.length > 0 ? (
        <BarChart
          data={chartData}
          width={contentWidth}
          height={isTablet ? 280 : 200}
          yAxisLabel=""
          yAxisSuffix=""
          chartConfig={{
            ...chartConfig,
            color: (opacity = 1) => chartColors.secondary(opacity),
            barPercentage: 0.7,
            strokeWidth: isTablet ? 3 : 2,
            decimalPlaces: 1,
            formatTopBarValue: (value: number) => formatMinutes(value),
            formatYLabel: (value: string) => formatMinutes(parseFloat(value)),
          }}
          style={styles.chart}
          showValuesOnTopOfBars={true}
          showBarTops={false}
          withHorizontalLabels={true}
          withVerticalLabels={true}
        />
      ) : (
        <View style={styles.emptyChartContainer}>
          <Text style={styles.emptyChartText}>
            {t("analytics.empty.noPracticeTime")}
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
    statsContainer: {
      flexDirection: "row",
      justifyContent: "space-around",
      marginBottom: theme.spacing.md,
      paddingVertical: theme.spacing.md,
      backgroundColor: theme.colors.background.subtle,
      borderRadius: theme.borderRadius.md,
    },
    statItem: {
      alignItems: "center",
    },
    statValue: {
      fontSize: theme.typography.sizes.bodyLarge,
      fontWeight: "bold",
      color: theme.colors.primary,
      marginBottom: theme.spacing.xs,
    },
    statLabel: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    chart: {
      marginVertical: theme.spacing.sm,
      borderRadius: theme.borderRadius.lg,
      marginLeft: -theme.spacing.lg,
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
  });
