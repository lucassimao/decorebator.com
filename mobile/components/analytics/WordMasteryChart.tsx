import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { ProgressChart } from "react-native-chart-kit";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { getChartColors, getChartConfig } from "./theme";
import { WordMasteryStats } from "@/api/analytics";
import { useResponsive } from "@/hooks/useResponsive";

interface WordMasteryChartProps {
  wordMastery?: WordMasteryStats[];
}

export const WordMasteryChart: React.FC<WordMasteryChartProps> = ({
  wordMastery,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType, contentWidth } = useResponsive();
  const styles = createStyles(theme, isTablet, deviceType);
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
            width={contentWidth - 40}
            height={isTablet ? 250 : 200}
            strokeWidth={isTablet ? 20 : 16}
            radius={isTablet ? 35 : 28}
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
      height: isTablet ? 250 : 200,
      justifyContent: "center",
      alignItems: "center",
      backgroundColor: theme.colors.background.default,
      borderRadius: theme.borderRadius.md,
    },
    emptyChartText: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.disabled,
    },
    legendContainer: {
      flexDirection: "row",
      flexWrap: "wrap",
      marginTop: theme.spacing.md,
      paddingHorizontal: theme.spacing.xs,
      gap: theme.spacing.xs,
    },
    legendItem: {
      flexDirection: "row",
      alignItems: "center",
      marginRight: theme.spacing.md,
      marginBottom: theme.spacing.xs,
      minHeight: theme.touchTargets.minimum,
    },
    legendDot: {
      width: isTablet ? 16 : 12,
      height: isTablet ? 16 : 12,
      borderRadius: isTablet ? 8 : 6,
      marginRight: theme.spacing.xs,
    },
    legendText: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
      marginRight: theme.spacing.xs,
    },
    legendValue: {
      fontSize: theme.typography.sizes.caption,
      fontWeight: theme.typography.weights.semibold,
      color: theme.colors.text.primary,
    },
  });
