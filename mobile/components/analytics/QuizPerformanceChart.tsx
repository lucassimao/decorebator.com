import React from "react";
import { View, Text, StyleSheet, Dimensions } from "react-native";
import { BarChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { getChartColors, getChartConfig } from "./theme";
import { QuizTypePerformance } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";

const screenWidth = Dimensions.get("window").width;

interface QuizPerformanceChartProps {
  quizPerformance?: QuizTypePerformance[];
}

export const QuizPerformanceChart: React.FC<QuizPerformanceChartProps> = ({
  quizPerformance,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const chartConfig = getChartConfig(theme);
  const chartColors = getChartColors(theme);
  const styles = createStyles(theme);

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
            width={screenWidth - 40}
            height={200}
            yAxisLabel=""
            yAxisSuffix="%"
            chartConfig={{
              ...chartConfig,
              color: (opacity = 1) => chartColors.success(opacity),
              barPercentage: 0.7,
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
    barLegendContainer: {
      marginTop: 16,
    },
    barLegendItem: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: 8,
      paddingHorizontal: 8,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.border,
    },
    barLegendLeft: {
      flexDirection: "row",
      alignItems: "center",
      flex: 1,
    },
    barLegendDot: {
      width: 8,
      height: 8,
      borderRadius: 4,
      marginRight: 8,
    },
    barLegendText: {
      fontSize: 14,
      color: theme.colors.text.primary,
      flex: 1,
    },
    barLegendValue: {
      fontSize: 14,
      fontWeight: "600",
      color: theme.colors.success,
      marginLeft: 8,
    },
  });
