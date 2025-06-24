import { BoxDistributionResponse } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";
import { MaterialIcons } from "@expo/vector-icons";
import React from "react";
import { useTranslation } from "react-i18next";
import { Dimensions, StyleSheet, Text, View } from "react-native";
import { BarChart } from "react-native-chart-kit";
import { getChartConfig } from "./theme";

const screenWidth = Dimensions.get("window").width;

interface BoxDistributionChartProps {
  boxDistribution?: BoxDistributionResponse;
}

export const BoxDistributionChart: React.FC<BoxDistributionChartProps> = ({
  boxDistribution,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const chartConfig = getChartConfig(theme);

  const boxColorGradient = [
    theme.colors.error, // Box 1 - Red (hardest)
    "#FF8E53", // Box 2 - Orange Red
    "#FFB74D", // Box 3 - Orange
    theme.colors.semantic.warning, // Box 4 - Yellow
    "#AED581", // Box 5 - Light Green
    "#81C784", // Box 6 - Green
    theme.colors.success, // Box 7 - Full Green (mastered)
  ];

  return (
    <View style={styles.chartCard}>
      <Text style={styles.chartTitle}>
        {t("analytics.charts.boxDistribution.title")}
      </Text>
      <Text style={styles.chartSubtitle}>
        {t("analytics.charts.boxDistribution.subtitle")}
      </Text>
      {boxDistribution && boxDistribution.totalWords > 0 ? (
        <>
          <BarChart
            data={{
              labels: ["1", "2", "3", "4", "5", "6", "7"],
              datasets: [
                {
                  data: [
                    boxDistribution.distribution.box1,
                    boxDistribution.distribution.box2,
                    boxDistribution.distribution.box3,
                    boxDistribution.distribution.box4,
                    boxDistribution.distribution.box5,
                    boxDistribution.distribution.box6,
                    boxDistribution.distribution.box7,
                  ],
                  colors: boxColorGradient.map(
                    (hex) =>
                      (opacity = 1) =>
                        hex,
                  ),
                },
              ],
            }}
            width={screenWidth - 40}
            height={220}
            yAxisLabel=""
            yAxisSuffix=""
            chartConfig={{
              ...chartConfig,
              barPercentage: 0.8,
              // color: () => boxColorGradient[6]
            }}
            style={styles.chart}
            showValuesOnTopOfBars={true}
            showBarTops={false}
            withHorizontalLabels={true}
            fromZero={true}
            withCustomBarColorFromData={true}
            flatColor={true}
          />
          {/* Box Labels Legend */}
          <View style={styles.boxLegendContainer}>
            {[1, 2, 3, 4, 5, 6, 7].map((boxNum) => {
              const boxKey =
                `box${boxNum}` as keyof typeof boxDistribution.distribution;
              const count = boxDistribution.distribution[boxKey];
              const labelKey = `analytics.charts.boxDistribution.box${boxNum}`;
              return (
                <View key={boxNum} style={styles.boxLegendItem}>
                  <View style={styles.boxLegendLeft}>
                    <View
                      style={[
                        styles.boxLegendDot,
                        {
                          backgroundColor: boxColorGradient[boxNum - 1],
                        },
                      ]}
                    />
                    <Text style={styles.boxLegendText}>
                      Box {boxNum}: {t(labelKey)}
                    </Text>
                  </View>
                  <Text style={styles.boxLegendCount}>
                    {count} {count === 1 ? "word" : "words"}
                  </Text>
                </View>
              );
            })}
          </View>
          {/* Explanation */}
          <View style={styles.explanationContainer}>
            <MaterialIcons
              name="info-outline"
              size={20}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.explanationText}>
              {t("analytics.charts.boxDistribution.explanation")}
            </Text>
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
      backgroundColor: theme.colors.background.secondary,
      borderRadius: 12,
    },
    emptyChartText: {
      fontSize: 16,
      color: theme.colors.text.tertiary,
    },
    boxLegendContainer: {
      marginTop: 16,
    },
    boxLegendItem: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: 8,
      paddingHorizontal: 8,
    },
    boxLegendLeft: {
      flexDirection: "row",
      alignItems: "center",
      flex: 1,
    },
    boxLegendDot: {
      width: 12,
      height: 12,
      borderRadius: 6,
      marginRight: 8,
    },
    boxLegendText: {
      fontSize: 14,
      color: theme.colors.text.primary,
      flex: 1,
    },
    boxLegendCount: {
      fontSize: 14,
      fontWeight: "600",
      color: theme.colors.text.secondary,
      marginLeft: 8,
    },
    explanationContainer: {
      flexDirection: "row",
      alignItems: "flex-start",
      marginTop: 16,
      paddingHorizontal: 8,
      paddingVertical: 12,
      backgroundColor: theme.colors.background.secondary,
      borderRadius: 8,
    },
    explanationText: {
      fontSize: 13,
      color: theme.colors.text.secondary,
      lineHeight: 18,
      flex: 1,
      marginLeft: 8,
    },
  });
