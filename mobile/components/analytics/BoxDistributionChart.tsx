import { BoxDistributionResponse } from "@/api/analytics";
import { useTheme } from "@/contexts/ThemeContext";
import { MaterialIcons } from "@expo/vector-icons";
import React from "react";
import { useTranslation } from "react-i18next";
import { StyleSheet, Text, View } from "react-native";
import { BarChart } from "react-native-chart-kit";
import { getChartConfig } from "./theme";
import { useResponsive } from "@/hooks/useResponsive";

interface BoxDistributionChartProps {
  boxDistribution?: BoxDistributionResponse;
}

export const BoxDistributionChart: React.FC<BoxDistributionChartProps> = ({
  boxDistribution,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType, contentWidth } = useResponsive();
  const styles = createStyles(theme, isTablet, deviceType);
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
            width={contentWidth}
            height={isTablet ? 300 : 220}
            yAxisLabel=""
            yAxisSuffix=""
            chartConfig={{
              ...chartConfig,
              barPercentage: 0.8,
              strokeWidth: isTablet ? 3 : 2,
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
              size={isTablet ? 24 : 20}
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
      height: isTablet ? 300 : 200,
      justifyContent: "center",
      alignItems: "center",
      backgroundColor: theme.colors.background.secondary,
      borderRadius: theme.borderRadius.md,
    },
    emptyChartText: {
      fontSize: theme.typography.sizes.bodyLarge,
      color: theme.colors.text.tertiary,
    },
    boxLegendContainer: {
      marginTop: theme.spacing.md,
    },
    boxLegendItem: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: theme.spacing.sm,
      paddingHorizontal: theme.spacing.sm,
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
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.primary,
      flex: 1,
    },
    boxLegendCount: {
      fontSize: theme.typography.sizes.body,
      fontWeight: "600",
      color: theme.colors.text.secondary,
      marginLeft: theme.spacing.sm,
    },
    explanationContainer: {
      flexDirection: "row",
      alignItems: "flex-start",
      marginTop: theme.spacing.md,
      paddingHorizontal: theme.spacing.sm,
      paddingVertical: theme.spacing.md,
      backgroundColor: theme.colors.background.secondary,
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
