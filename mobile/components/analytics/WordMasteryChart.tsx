import React from "react";
import { View, Text, StyleSheet, Dimensions } from "react-native";
import { ProgressChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { getChartColors, getChartConfig } from "./theme";
import { WordMasteryStats } from "@/api/analytics";

const screenWidth = Dimensions.get("window").width;

interface WordMasteryChartProps {
  wordMastery?: WordMasteryStats[];
}

export const WordMasteryChart: React.FC<WordMasteryChartProps> = ({
  wordMastery,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const chartColors = getChartColors(theme);
  const chartConfig = getChartConfig(theme);

  const progressChartData = {
    labels: wordMastery?.slice(0, 6).map((w) => w.word.substring(0, 8)) || [],
    data: wordMastery?.slice(0, 6).map((w) => w.masteryLevel) || [],
  };

  return (
    <View style={styles.chartCard}>
      <Text style={styles.chartTitle}>
        {t("analytics.charts.wordMastery.title")}
      </Text>
      <Text style={styles.chartSubtitle}>
        {t("analytics.charts.wordMastery.subtitle")}
      </Text>
      {progressChartData.data.length > 0 ? (
        <>
          <ProgressChart
            data={progressChartData}
            width={screenWidth - 40}
            height={200}
            strokeWidth={16}
            radius={28}
            chartConfig={{
              ...chartConfig,
              color: (opacity = 1, index = 0) => {
                const color =
                  chartColors.colors[index % chartColors.colors.length];
                return `rgba(${parseInt(color.slice(1, 3), 16)}, ${parseInt(color.slice(3, 5), 16)}, ${parseInt(color.slice(5, 7), 16)}, ${opacity})`;
              },
            }}
            hideLegend={true}
            style={styles.chart}
          />
          {/* Custom Legend */}
          <View style={styles.legendContainer}>
            {progressChartData.labels.map((label, index) => (
              <View key={index} style={styles.legendItem}>
                <View
                  style={[
                    styles.legendDot,
                    {
                      backgroundColor:
                        chartColors.colors[index % chartColors.colors.length],
                    },
                  ]}
                />
                <Text style={styles.legendText}>{label}</Text>
                <Text style={styles.legendValue}>
                  {Math.round(progressChartData.data[index] * 100)}%
                </Text>
              </View>
            ))}
          </View>
        </>
      ) : (
        <View style={styles.emptyChartContainer}>
          <Text style={styles.emptyChartText}>
            {t("analytics.empty.noData")}
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
      shadowColor: theme.colors.text.primary,
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
      color: theme.colors.text.disabled,
    },
    legendContainer: {
      flexDirection: "row",
      flexWrap: "wrap",
      marginTop: 16,
      paddingHorizontal: 8,
    },
    legendItem: {
      flexDirection: "row",
      alignItems: "center",
      marginRight: 16,
      marginBottom: 8,
    },
    legendDot: {
      width: 12,
      height: 12,
      borderRadius: 6,
      marginRight: 6,
    },
    legendText: {
      fontSize: 12,
      color: theme.colors.text.secondary,
      marginRight: 4,
    },
    legendValue: {
      fontSize: 12,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
  });
