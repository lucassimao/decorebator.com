import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { BarChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { getChartColors, getChartConfig } from "./theme";
import { QuizTypePerformance } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive } from "@/hooks/useResponsive";

interface QuizPerformanceChartProps {
  quizPerformance?: QuizTypePerformance[];
}

export const QuizPerformanceChart: React.FC<QuizPerformanceChartProps> = ({
  quizPerformance,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType, contentWidth } = useResponsive();
  const chartConfig = getChartConfig(theme);
  const chartColors = getChartColors(theme);
  const styles = createStyles(theme, isTablet, deviceType);

  // Translate quiz type names
  const translateQuizType = (quizType: string): string => {
    const quizTypeKey = `analytics.quizTypes.${quizType.slice(0).toLowerCase().trim()}`;
    return t(quizTypeKey);
  };

  const quizTypeLabels =
    quizPerformance?.map((q) => translateQuizType(q.quizType)) || [];
  const quizTypeData = {
    labels: quizTypeLabels.map((label) => label.substring(0, 10)),
    datasets: [
      {
        data: quizPerformance?.map((q) => q.successRate) || [],
      },
    ],
  };

  return (
    <View style={styles.chartCard}>
      <Text style={styles.chartTitle}>
        {t("analytics.charts.quizPerformance.title")}
      </Text>
      <Text style={styles.chartSubtitle}>
        {t("analytics.charts.quizPerformance.subtitle")}
      </Text>
      {quizTypeData.labels.length > 0 ? (
        <>
          <BarChart
            data={quizTypeData}
            width={contentWidth}
            height={isTablet ? 280 : 200}
            yAxisLabel=""
            yAxisSuffix="%"
            chartConfig={{
              ...chartConfig,
              color: (opacity = 1) => chartColors.success(opacity),
              barPercentage: 0.7,
              strokeWidth: isTablet ? 3 : 2,
            }}
            style={styles.chart}
            showValuesOnTopOfBars={true}
            showBarTops={false}
            withHorizontalLabels={true}
            withVerticalLabels={false}
          />
          {/* Custom Legend */}
          <View style={styles.barLegendContainer}>
            {quizPerformance?.map((quiz, index) => (
              <View key={index} style={styles.barLegendItem}>
                <View style={styles.barLegendLeft}>
                  <View
                    style={[
                      styles.barLegendDot,
                      { backgroundColor: theme.colors.success },
                    ]}
                  />
                  <Text style={styles.barLegendText} numberOfLines={1}>
                    {translateQuizType(quiz.quizType)}
                  </Text>
                </View>
                <Text style={styles.barLegendValue}>
                  {Math.round(quiz.successRate)}%
                </Text>
              </View>
            ))}
          </View>
        </>
      ) : (
        <View style={styles.emptyChartContainer}>
          <Text style={styles.emptyChartText}>
            {t("analytics.empty.noQuizzes")}
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
    barLegendContainer: {
      marginTop: theme.spacing.md,
    },
    barLegendItem: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: theme.spacing.sm,
      paddingHorizontal: theme.spacing.sm,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.border,
    },
    barLegendLeft: {
      flexDirection: "row",
      alignItems: "center",
      flex: 1,
    },
    barLegendDot: {
      width: isTablet ? 12 : 8,
      height: isTablet ? 12 : 8,
      borderRadius: isTablet ? 6 : 4,
      marginRight: theme.spacing.sm,
    },
    barLegendText: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.primary,
      flex: 1,
    },
    barLegendValue: {
      fontSize: theme.typography.sizes.body,
      fontWeight: "600",
      color: theme.colors.success,
      marginLeft: theme.spacing.sm,
    },
  });
