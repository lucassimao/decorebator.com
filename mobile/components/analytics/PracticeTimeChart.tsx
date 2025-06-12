import React from "react";
import { View, Text, StyleSheet, Dimensions } from "react-native";
import { BarChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { colors, chartConfig } from "./theme";
import { PracticeTimeResponse } from "@/api/analytics";

const screenWidth = Dimensions.get("window").width;

interface PracticeTimeChartProps {
  practiceTime?: PracticeTimeResponse;
}

export const PracticeTimeChart: React.FC<PracticeTimeChartProps> = ({
  practiceTime,
}) => {
  const { t } = useTranslation();

  const formatDateLabel = (dateString: string): string => {
    try {
      // Extract just the date part from ISO timestamp (e.g., "2025-06-11T00:00:00Z" -> "2025-06-11")
      const datePart = dateString.split('T')[0];
      const [year, month, day] = datePart.split('-').map(Number);
      
      // Validate parts exist and are numbers
      if (!year || !month || !day || isNaN(year) || isNaN(month) || isNaN(day)) {
        console.error("Invalid date parts:", { dateString, datePart, year, month, day });
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
      return remainingMinutes > 0 ? `${hours}h ${remainingMinutes}m` : `${hours}h`;
    }
  };

  // Sort by date ascending and take last 7 days
  const sortedData = practiceTime?.practiceTime
    ?.slice()
    .sort((a, b) => {
      // Parse dates as local time for proper sorting
      // Extract date part from ISO timestamp
      const datePartA = a.date.split('T')[0];
      const datePartB = b.date.split('T')[0];
      const [yearA, monthA, dayA] = datePartA.split('-').map(Number);
      const [yearB, monthB, dayB] = datePartB.split('-').map(Number);
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

  const totalPracticeTime = sortedData.reduce((sum, p) => sum + p.practiceTimeMinutes, 0);
  const avgPracticeTime = sortedData.length > 0 ? totalPracticeTime / sortedData.length : 0;

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
          <Text style={styles.statValue}>{formatMinutes(totalPracticeTime)}</Text>
          <Text style={styles.statLabel}>{t("analytics.charts.practiceTime.totalTime")}</Text>
        </View>
        <View style={styles.statItem}>
          <Text style={styles.statValue}>{formatMinutes(avgPracticeTime)}</Text>
          <Text style={styles.statLabel}>{t("analytics.charts.practiceTime.avgDaily")}</Text>
        </View>
      </View>

      {chartData.labels.length > 0 ? (
        <BarChart
          data={chartData}
          width={screenWidth - 40}
          height={200}
          yAxisLabel=""
          yAxisSuffix=""
          chartConfig={{
            ...chartConfig,
            color: (opacity = 1) => `rgba(74, 144, 226, ${opacity})`, // Blue color
            barPercentage: 0.7,
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
  statsContainer: {
    flexDirection: "row",
    justifyContent: "space-around",
    marginBottom: 16,
    paddingVertical: 12,
    backgroundColor: colors.lightBackground,
    borderRadius: 12,
  },
  statItem: {
    alignItems: "center",
  },
  statValue: {
    fontSize: 18,
    fontWeight: "bold",
    color: colors.primary,
    marginBottom: 4,
  },
  statLabel: {
    fontSize: 12,
    color: colors.textMedium,
    textAlign: "center",
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