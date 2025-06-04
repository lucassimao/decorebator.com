import { useQuery } from '@tanstack/react-query';
import { LinearGradient } from 'expo-linear-gradient';
import { useLocalSearchParams } from 'expo-router';
import React from 'react';
import { useTranslation } from 'react-i18next';
import {
    ActivityIndicator,
    Dimensions,
    ScrollView,
    StyleSheet,
    Text,
    View
} from 'react-native';
import { BarChart, LineChart, ProgressChart } from 'react-native-chart-kit';

const screenWidth = Dimensions.get('window').width;

interface AnalyticsDashboardProps {
  wordlistId: number;
}

interface DashboardStats {
  total_words: number;
  words_mastered: number;
  average_mastery: number;
  best_streak: number;
  current_streak: number;
  words_studied_today: number;
  quizzes_today: number;
  accuracy_today: number;
}

interface WordMasteryStats {
  word_id: number;
  word: string;
  mastery_level: number;
  accuracy: number;
  streak_count: number;
  highest_box: number;
}

interface LearningProgress {
  date: string;
  words_studied: number;
  accuracy_rate: number;
}

interface QuizTypePerformance {
  quiz_type: string;
  success_rate: number;
  total_attempts: number;
}

// Color palette from design guidelines
const colors = {
  primary: '#FF7B54',
  success: '#4CAF50',
  error: '#FF6B6B',
  gold: '#FFD700',
  background: '#FDF6E3',
  backgroundLight: '#FFF9F0',
  backgroundPeach: '#FFE8D6',
  backgroundOrange: '#FFDCC3',
  backgroundSage: '#F5F0E6',
  textDark: '#2D3436',
  textMedium: '#636E72',
  textLight: '#B2BEC3',
  white: '#FFFFFF',
  lightBackground: '#FAFAFA',
  borderGray: '#E0E0E0',
  divider: '#F0F0F0',
};

const AnalyticsDashboard: React.FC<AnalyticsDashboardProps> = () => {
  const { t } = useTranslation();
    const { wordlistId } = useLocalSearchParams();
  
  // Fetch dashboard stats
  const { data: dashboardStats, isLoading: statsLoading } = useQuery<DashboardStats>({
    queryKey: ['analytics', 'dashboard'],
    queryFn: async () => {
      const response = await fetch('/api/analytics/dashboard');
      return response.json();
    },
  });

  // Fetch word mastery
  const { data: wordMastery, isLoading: masteryLoading } = useQuery<{ stats: WordMasteryStats[] }>({
    queryKey: ['analytics', 'mastery', wordlistId],
    queryFn: async () => {
      const response = await fetch(`/api/analytics/wordlists/${wordlistId}/mastery`);
      return response.json();
    },
  });

  // Fetch learning progress (last 7 days)
  const { data: learningProgress, isLoading: progressLoading } = useQuery<{ progress: LearningProgress[] }>({
    queryKey: ['analytics', 'progress', wordlistId],
    queryFn: async () => {
      const response = await fetch(`/api/analytics/wordlists/${wordlistId}/progress?days=7`);
      return response.json();
    },
  });

  // Fetch quiz performance
  const { data: quizPerformance, isLoading: quizPerfLoading } = useQuery<{ quiz_performance: QuizTypePerformance[] }>({
    queryKey: ['analytics', 'quiz-performance'],
    queryFn: async () => {
      const response = await fetch('/api/analytics/quiz-performance');
      return response.json();
    },
  });

  if (statsLoading || masteryLoading || progressLoading || quizPerfLoading) {
    return (
      <LinearGradient
        colors={[colors.backgroundLight, colors.backgroundPeach, colors.backgroundSage]}
        style={styles.loadingContainer}
      >
        <ActivityIndicator size="large" color={colors.primary} />
      </LinearGradient>
    );
  }

  // Prepare data for charts
  const progressChartData = {
    labels: wordMastery?.stats.slice(0, 6).map(w => w.word.substring(0, 8)) || [],
    data: wordMastery?.stats.slice(0, 6).map(w => w.mastery_level) || [],
  };

  const lineChartData = {
    labels: learningProgress?.progress.slice(-7).map(p => {
      const date = new Date(p.date);
      return `${date.getMonth() + 1}/${date.getDate()}`;
    }) || [],
    datasets: [
      {
        data: learningProgress?.progress.slice(-7).map(p => p.words_studied) || [],
        color: (opacity = 1) => `rgba(255, 123, 84, ${opacity})`, // Primary orange
        strokeWidth: 3,
      },
    ],
  };

  // Translate quiz type names
  const translateQuizType = (quizType: string): string => {
    const quizTypeKey = `analytics.quizTypes.${quizType.replace(/([A-Z])/g, '_$1').toLowerCase().trim()}`;
    return t(quizTypeKey);
  };

  const quizTypeData = {
    labels: quizPerformance?.quiz_performance.map(q => 
      translateQuizType(q.quiz_type).substring(0, 10)
    ) || [],
    datasets: [
      {
        data: quizPerformance?.quiz_performance.map(q => q.success_rate) || [],
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
      strokeDasharray: '',
      stroke: colors.divider,
    },
  };

  return (
    <LinearGradient
      colors={[colors.backgroundLight, colors.backgroundPeach, colors.backgroundSage]}
      style={styles.container}
    >
      <ScrollView showsVerticalScrollIndicator={false}>
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.headerTitle}>{t('analytics.header.title')}</Text>
          <Text style={styles.headerSubtitle}>{t('analytics.header.subtitle')}</Text>
        </View>

        {/* Dashboard Stats Cards */}
        <View style={styles.statsGrid}>
          <View style={styles.statCard}>
            <View style={styles.statIconContainer}>
              <Text style={styles.statIcon}>📚</Text>
            </View>
            <Text style={styles.statValue}>{dashboardStats?.words_studied_today || 0}</Text>
            <Text style={styles.statLabel}>{t('analytics.stats.wordsToday')}</Text>
          </View>

          <View style={[styles.statCard, styles.statCardHighlight]}>
            <View style={styles.statIconContainer}>
              <Text style={styles.statIcon}>🔥</Text>
            </View>
            <Text style={[styles.statValue, styles.statValueHighlight]}>
              {dashboardStats?.current_streak || 0}
            </Text>
            <Text style={styles.statLabel}>{t('analytics.stats.dayStreak')}</Text>
          </View>

          <View style={styles.statCard}>
            <View style={styles.statIconContainer}>
              <Text style={styles.statIcon}>🏆</Text>
            </View>
            <Text style={[styles.statValue, styles.statValueSuccess]}>
              {dashboardStats?.words_mastered || 0}
            </Text>
            <Text style={styles.statLabel}>{t('analytics.stats.mastered')}</Text>
          </View>

          <View style={styles.statCard}>
            <View style={styles.statIconContainer}>
              <Text style={styles.statIcon}>🎯</Text>
            </View>
            <Text style={styles.statValue}>
              {Math.round(dashboardStats?.accuracy_today || 0)}%
            </Text>
            <Text style={styles.statLabel}>{t('analytics.stats.accuracy')}</Text>
          </View>
        </View>

        {/* Word Mastery Progress */}
        <View style={styles.chartCard}>
          <Text style={styles.chartTitle}>{t('analytics.charts.wordMastery.title')}</Text>
          <Text style={styles.chartSubtitle}>{t('analytics.charts.wordMastery.subtitle')}</Text>
          {progressChartData.data.length > 0 ? (
            <ProgressChart
              data={progressChartData}
              width={screenWidth - 32}
              height={220}
              strokeWidth={16}
              radius={28}
              chartConfig={{
                ...chartConfig,
                color: (opacity = 1, index=0) => {
                  const gradientColors = [
                    colors.primary,
                    colors.success,
                    colors.gold,
                    '#9C27B0',
                    '#2196F3',
                    '#FF6B3D',
                  ];
                  return `rgba(${parseInt(gradientColors[index % gradientColors.length].slice(1, 3), 16)}, ${parseInt(gradientColors[index % gradientColors.length].slice(3, 5), 16)}, ${parseInt(gradientColors[index % gradientColors.length].slice(5, 7), 16)}, ${opacity})`;
                },
              }}
              hideLegend={false}
              style={styles.chart}
            />
          ) : (
            <View style={styles.emptyChartContainer}>
              <Text style={styles.emptyChartText}>{t('analytics.empty.noData')}</Text>
            </View>
          )}
        </View>

        {/* Learning Progress Line Chart */}
        <View style={styles.chartCard}>
          <Text style={styles.chartTitle}>{t('analytics.charts.progress.title')}</Text>
          <Text style={styles.chartSubtitle}>{t('analytics.charts.progress.subtitle')}</Text>
          {lineChartData.labels.length > 0 ? (
            <LineChart
              data={lineChartData}
              width={screenWidth - 32}
              height={220}
              chartConfig={{
                ...chartConfig,
                propsForDots: {
                  r: '6',
                  strokeWidth: '2',
                  stroke: colors.primary,
                  fill: colors.white,
                },
              }}
              bezier
              style={styles.chart}
            />
          ) : (
            <View style={styles.emptyChartContainer}>
              <Text style={styles.emptyChartText}>{t('analytics.empty.noProgress')}</Text>
            </View>
          )}
        </View>

        {/* Quiz Type Performance */}
        <View style={styles.chartCard}>
          <Text style={styles.chartTitle}>{t('analytics.charts.quizPerformance.title')}</Text>
          <Text style={styles.chartSubtitle}>{t('analytics.charts.quizPerformance.subtitle')}</Text>
          {quizTypeData.labels.length > 0 ? (
            <BarChart
              data={quizTypeData}
              width={screenWidth - 32}
              height={220}
              yAxisLabel=""
              yAxisSuffix="%"
              chartConfig={{
                ...chartConfig,
                color: (opacity = 1) => `rgba(76, 175, 80, ${opacity})`,
                barPercentage: 0.7,
              }}
              style={styles.chart}
              showValuesOnTopOfBars={true}
            />
          ) : (
            <View style={styles.emptyChartContainer}>
              <Text style={styles.emptyChartText}>{t('analytics.empty.noQuizzes')}</Text>
            </View>
          )}
        </View>

        {/* Top Words by Mastery */}
        <View style={styles.chartCard}>
          <Text style={styles.chartTitle}>{t('analytics.charts.topWords.title')}</Text>
          <Text style={styles.chartSubtitle}>{t('analytics.charts.topWords.subtitle')}</Text>
          {wordMastery?.stats && wordMastery.stats.length > 0 ? (
            <View style={styles.wordsList}>
              {wordMastery.stats.slice(0, 5).map((word, index) => (
                <View key={word.word_id} style={styles.wordItem}>
                  <View style={styles.wordItemLeft}>
                    <View style={[styles.wordRankCircle, index === 0 && styles.wordRankGold]}>
                      <Text style={[styles.wordRank, index === 0 && styles.wordRankTextGold]}>
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
                        {t('analytics.box', { number: word.highest_box })}
                      </Text>
                    </View>
                  </View>
                </View>
              ))}
            </View>
          ) : (
            <View style={styles.emptyChartContainer}>
              <Text style={styles.emptyChartText}>{t('analytics.empty.noMastered')}</Text>
            </View>
          )}
        </View>

        {/* Bottom spacing */}
        <View style={{ height: 20 }} />
      </ScrollView>
    </LinearGradient>
  );
};

export default AnalyticsDashboard;


const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  header: {
    paddingHorizontal: 16,
    paddingTop: 20,
    paddingBottom: 10,
  },
  headerTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: colors.textDark,
    marginBottom: 4,
  },
  headerSubtitle: {
    fontSize: 16,
    color: colors.textMedium,
  },
  statsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    paddingHorizontal: 12,
    paddingVertical: 10,
    justifyContent: 'space-between',
  },
  statCard: {
    backgroundColor: colors.white,
    width: '48%',
    marginBottom: 12,
    borderRadius: 16,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 3.84,
    elevation: 5,
    alignItems: 'center',
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
    fontWeight: 'bold',
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
    textAlign: 'center',
  },
  chartCard: {
    backgroundColor: colors.white,
    marginHorizontal: 16,
    marginVertical: 8,
    borderRadius: 16,
    padding: 20,
    shadowColor: '#000',
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
    fontWeight: 'bold',
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
    marginLeft: -16,
  },
  emptyChartContainer: {
    height: 200,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: colors.lightBackground,
    borderRadius: 12,
  },
  emptyChartText: {
    fontSize: 16,
    color: colors.textLight,
  },
  wordsList: {
    marginTop: 8,
  },
  wordItem: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: colors.divider,
  },
  wordItemLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  wordRankCircle: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: colors.lightBackground,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  wordRankGold: {
    backgroundColor: colors.gold,
  },
  wordRank: {
    fontSize: 16,
    fontWeight: 'bold',
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
    flexDirection: 'row',
    alignItems: 'center',
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
    fontWeight: 'bold',
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