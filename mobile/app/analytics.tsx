import { useNavigation } from "@react-navigation/native";
import { useAnalytics } from "@/hooks/useAnalytics";
import { useI18n } from "@/hooks/useI18n";
import { Ionicons } from "@expo/vector-icons";
import { LinearGradient } from "expo-linear-gradient";
import { useLocalSearchParams } from "expo-router";
import {
  ActivityIndicator,
  Dimensions,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  SafeAreaView,
} from "react-native";
import { BarChart, LineChart, ProgressChart } from "react-native-chart-kit";

const screenWidth = Dimensions.get("window").width;

interface AnalyticsDashboardProps {}

// Color palette from design guidelines
const colors = {
  primary: "#FF7B54",
  success: "#4CAF50",
  error: "#FF6B6B",
  gold: "#FFD700",
  background: "#FDF6E3",
  backgroundLight: "#FFF9F0",
  backgroundPeach: "#FFE8D6",
  backgroundOrange: "#FFDCC3",
  backgroundSage: "#F5F0E6",
  textDark: "#2D3436",
  textMedium: "#636E72",
  textLight: "#B2BEC3",
  white: "#FFFFFF",
  lightBackground: "#FAFAFA",
  borderGray: "#E0E0E0",
  divider: "#F0F0F0",
};

const AnalyticsDashboard: React.FC<AnalyticsDashboardProps> = () => {
  const { t } = useI18n();
  const { wordlistId: rawWordlistId } = useLocalSearchParams();
  const wordlistId = Number(rawWordlistId);
  const {
    dashboardStats,
    wordMastery,
    learningProgress,
    quizPerformance,
    isPending,
  } = useAnalytics(wordlistId);
  const navigation = useNavigation();

  if (isPending) {
    return (
      <LinearGradient
        colors={[
          colors.backgroundLight,
          colors.backgroundPeach,
          colors.backgroundSage,
        ]}
        style={styles.loadingContainer}
      >
        <ActivityIndicator size="large" color={colors.primary} />
      </LinearGradient>
    );
  }

  // Color palette for charts
  const chartColors = [
    colors.primary,
    colors.success,
    colors.gold,
    "#9C27B0",
    "#2196F3",
    "#FF6B3D",
  ];

  // Prepare data for charts
  const progressChartData = {
    labels: wordMastery?.slice(0, 6).map((w) => w.word.substring(0, 8)) || [],
    data: wordMastery?.slice(0, 6).map((w) => w.mastery_level) || [],
  };

  const lineChartData = {
    labels:
      learningProgress?.slice(-7).map((p) => {
        const date = new Date(p.date);
        return `${date.getMonth() + 1}/${date.getDate()}`;
      }) || [],
    datasets: [
      {
        data: learningProgress?.slice(-7).map((p) => p.words_studied) || [],
        color: (opacity = 1) => `rgba(255, 123, 84, ${opacity})`, // Primary orange
        strokeWidth: 3,
      },
    ],
  };

  // Translate quiz type names
  const translateQuizType = (quizType: string): string => {
    const quizTypeKey = `analytics.quizTypes.${quizType.slice(0).toLowerCase().trim()}`;
    return t(quizTypeKey);
  };

  const quizTypeLabels =
    quizPerformance?.map((q) => translateQuizType(q.quiz_type)) || [];
  const quizTypeData = {
    labels: quizTypeLabels.map((label) => label.substring(0, 10)),
    datasets: [
      {
        data: quizPerformance?.map((q) => q.success_rate) || [],
      },
    ],
  };

  const chartConfig = {
    backgroundColor: colors.white,
    backgroundGradientFrom: colors.white,
    backgroundGradientTo: colors.white,
    decimalPlaces: 0,
    color: (opacity = 1) => `rgba(255, 123, 84, ${opacity})`,
    labelColor: (opacity = 1) => `rgba(45, 52, 54, ${opacity})`,
    style: {
      borderRadius: 16,
    },
    propsForBackgroundLines: {
      strokeDasharray: "",
      stroke: colors.divider,
    },
  };

  return (
    <LinearGradient
      colors={[
        colors.backgroundLight,
        colors.backgroundPeach,
        colors.backgroundSage,
      ]}
      style={styles.container}
    >
      <SafeAreaView style={styles.safeArea}>
        {/* Header */}
        <View style={styles.header}>
          <TouchableOpacity
            style={styles.backButton}
            onPress={() => navigation.goBack()}
          >
            <Ionicons name="arrow-back" size={24} color={colors.textDark} />
          </TouchableOpacity>
          <View style={styles.headerTextContainer}>
            <Text style={styles.headerTitle}>
              {t("analytics.header.title")}
            </Text>
            <Text style={styles.headerSubtitle}>
              {t("analytics.header.subtitle")}
            </Text>
            <Text style={styles.headerNote}>
              {t("analytics.header.updateNote")}
            </Text>
          </View>
          <View style={styles.headerSpacer} />
        </View>

        <ScrollView showsVerticalScrollIndicator={false}>
          {/* Dashboard Stats Cards */}
          <View style={styles.statsGrid}>
            <View style={styles.statCard}>
              <View style={styles.statIconContainer}>
                <Text style={styles.statIcon}>📚</Text>
              </View>
              <Text style={styles.statValue}>
                {dashboardStats?.words_studied_today || 0}
              </Text>
              <Text style={styles.statLabel}>
                {t("analytics.stats.wordsToday")}
              </Text>
            </View>

            <View style={[styles.statCard, styles.statCardHighlight]}>
              <View style={styles.statIconContainer}>
                <Text style={styles.statIcon}>🔥</Text>
              </View>
              <Text style={[styles.statValue, styles.statValueHighlight]}>
                {dashboardStats?.current_streak || 0}
              </Text>
              <Text style={styles.statLabel}>
                {t("analytics.stats.dayStreak")}
              </Text>
            </View>

            <View style={styles.statCard}>
              <View style={styles.statIconContainer}>
                <Text style={styles.statIcon}>🏆</Text>
              </View>
              <Text style={[styles.statValue, styles.statValueSuccess]}>
                {dashboardStats?.words_mastered || 0}
              </Text>
              <Text style={styles.statLabel}>
                {t("analytics.stats.mastered")}
              </Text>
            </View>

            <View style={styles.statCard}>
              <View style={styles.statIconContainer}>
                <Text style={styles.statIcon}>🎯</Text>
              </View>
              <Text style={styles.statValue}>
                {Math.round(dashboardStats?.accuracy_today || 0)}%
              </Text>
              <Text style={styles.statLabel}>
                {t("analytics.stats.accuracy")}
              </Text>
            </View>
          </View>

          {/* Word Mastery Progress */}
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
                      const color = chartColors[index % chartColors.length];
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
                              chartColors[index % chartColors.length],
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

          {/* Learning Progress Line Chart */}
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
                width={screenWidth - 40}
                height={220}
                chartConfig={{
                  ...chartConfig,
                  propsForDots: {
                    r: "6",
                    strokeWidth: "2",
                    stroke: colors.primary,
                    fill: colors.white,
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

          {/* Quiz Type Performance */}
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
                    color: (opacity = 1) => `rgba(76, 175, 80, ${opacity})`,
                    barPercentage: 0.7,
                  }}
                  style={styles.chart}
                  showValuesOnTopOfBars={true}
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
                            { backgroundColor: colors.success },
                          ]}
                        />
                        <Text style={styles.barLegendText} numberOfLines={1}>
                          {translateQuizType(quiz.quiz_type)}
                        </Text>
                      </View>
                      <Text style={styles.barLegendValue}>
                        {Math.round(quiz.success_rate)}%
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

          {/* Top Words by Mastery */}
          <View style={styles.chartCard}>
            <Text style={styles.chartTitle}>
              {t("analytics.charts.topWords.title")}
            </Text>
            <Text style={styles.chartSubtitle}>
              {t("analytics.charts.topWords.subtitle")}
            </Text>
            {wordMastery && wordMastery.length > 0 ? (
              <View style={styles.wordsList}>
                {wordMastery.slice(0, 5).map((word, index) => (
                  <View key={word.word_id} style={styles.wordItem}>
                    <View style={styles.wordItemLeft}>
                      <View
                        style={[
                          styles.wordRankCircle,
                          index === 0 && styles.wordRankGold,
                        ]}
                      >
                        <Text
                          style={[
                            styles.wordRank,
                            index === 0 && styles.wordRankTextGold,
                          ]}
                        >
                          {index + 1}
                        </Text>
                      </View>
                      <Text style={styles.wordName}>{word.word}</Text>
                    </View>
                    <View style={styles.wordStats}>
                      <View style={styles.masteryBadge}>
                        <Text style={styles.wordMastery}>
                          {Math.round(word.mastery_level * 100)}%
                        </Text>
                      </View>
                      <View style={styles.boxBadge}>
                        <Text style={styles.wordBox}>
                          {t("analytics.box", { number: word.highest_box })}
                        </Text>
                      </View>
                    </View>
                  </View>
                ))}
              </View>
            ) : (
              <View style={styles.emptyChartContainer}>
                <Text style={styles.emptyChartText}>
                  {t("analytics.empty.noMastered")}
                </Text>
              </View>
            )}
          </View>

          {/* Bottom spacing */}
          <View style={{ height: 20 }} />
        </ScrollView>
      </SafeAreaView>
    </LinearGradient>
  );
};

export default AnalyticsDashboard;

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  safeArea: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: "rgba(255, 255, 255, 0.3)",
    borderBottomWidth: 1,
    borderBottomColor: colors.divider,
  },
  backButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: colors.white,
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 3.84,
    elevation: 3,
  },
  headerTextContainer: {
    flex: 1,
    marginLeft: 12,
  },
  headerTitle: {
    fontSize: 24,
    fontWeight: "bold",
    color: colors.textDark,
  },
  headerSubtitle: {
    fontSize: 14,
    color: colors.textMedium,
    marginTop: 2,
  },
  headerNote: {
    fontSize: 12,
    color: colors.textLight,
    marginTop: 2,
    fontStyle: "italic",
  },
  headerSpacer: {
    width: 40,
  },
  statsGrid: {
    flexDirection: "row",
    flexWrap: "wrap",
    paddingHorizontal: 12,
    paddingVertical: 10,
    justifyContent: "space-between",
  },
  statCard: {
    backgroundColor: colors.white,
    width: "48%",
    marginBottom: 12,
    borderRadius: 16,
    padding: 16,
    shadowColor: "#000",
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 3.84,
    elevation: 5,
    alignItems: "center",
  },
  statCardHighlight: {
    backgroundColor: colors.backgroundOrange,
  },
  statIconContainer: {
    marginBottom: 8,
  },
  statIcon: {
    fontSize: 24,
  },
  statValue: {
    fontSize: 32,
    fontWeight: "bold",
    color: colors.textDark,
    marginBottom: 4,
  },
  statValueHighlight: {
    color: colors.primary,
  },
  statValueSuccess: {
    color: colors.success,
  },
  statLabel: {
    fontSize: 14,
    color: colors.textMedium,
    textAlign: "center",
  },
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
  // Custom Legend Styles
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
    color: colors.textMedium,
    marginRight: 4,
  },
  legendValue: {
    fontSize: 12,
    fontWeight: "600",
    color: colors.textDark,
  },
  // Bar Chart Legend Styles
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
    borderBottomColor: colors.divider,
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
    color: colors.textDark,
    flex: 1,
  },
  barLegendValue: {
    fontSize: 14,
    fontWeight: "600",
    color: colors.success,
    marginLeft: 8,
  },
  // Word List Styles
  wordsList: {
    marginTop: 8,
  },
  wordItem: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: colors.divider,
  },
  wordItemLeft: {
    flexDirection: "row",
    alignItems: "center",
    flex: 1,
  },
  wordRankCircle: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: colors.lightBackground,
    justifyContent: "center",
    alignItems: "center",
    marginRight: 12,
  },
  wordRankGold: {
    backgroundColor: colors.gold,
  },
  wordRank: {
    fontSize: 16,
    fontWeight: "bold",
    color: colors.textDark,
  },
  wordRankTextGold: {
    color: colors.white,
  },
  wordName: {
    fontSize: 16,
    color: colors.textDark,
    flex: 1,
  },
  wordStats: {
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
  },
  masteryBadge: {
    backgroundColor: colors.success,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  wordMastery: {
    fontSize: 14,
    fontWeight: "bold",
    color: colors.white,
  },
  boxBadge: {
    backgroundColor: colors.lightBackground,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  wordBox: {
    fontSize: 14,
    color: colors.textMedium,
  },
});
