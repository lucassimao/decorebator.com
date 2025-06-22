import { useNavigation } from "@react-navigation/native";
import { useAnalytics } from "@/hooks/useAnalytics";
import { useLocalSearchParams } from "expo-router";
import { useTheme } from "@/contexts/ThemeContext";
import { createCommonStyles } from "@/styles/common";
import {
  ActivityIndicator,
  ScrollView,
  StyleSheet,
  View,
  SafeAreaView,
} from "react-native";
import {
  AnalyticsHeader,
  StatsGrid,
  WordMasteryChart,
  LearningProgressChart,
  QuizPerformanceChart,
  TopWordsSection,
  BoxDistributionChart,
  HistoricalBoxDistributionChart,
  PracticeTimeChart,
} from "@/components/analytics";

const AnalyticsDashboard = () => {
  const { wordlistId: rawWordlistId } = useLocalSearchParams();
  const wordlistId = Number(rawWordlistId);
  const {
    stats,
    wordMastery,
    learningProgress,
    practiceTime,
    quizPerformance,
    boxDistribution,
    historicalBoxDistribution,
    isPending,
  } = useAnalytics(wordlistId);
  const navigation = useNavigation();
  const { theme } = useTheme();
  const commonStyles = createCommonStyles(theme);
  const styles = createStyles(theme);

  if (isPending) {
    return (
      <SafeAreaView style={[commonStyles.safeArea, styles.loadingContainer]}>
        <ActivityIndicator size="large" color={theme.colors.primary} />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={[commonStyles.safeArea, styles.container]}>
      <AnalyticsHeader onBackPress={() => navigation.goBack()} />

      <ScrollView 
        showsVerticalScrollIndicator={false}
        style={styles.scrollView}
        contentContainerStyle={styles.scrollContent}
      >
        <StatsGrid stats={stats} />

          <WordMasteryChart wordMastery={wordMastery} />

          <LearningProgressChart learningProgress={learningProgress} />

          <PracticeTimeChart practiceTime={practiceTime} />

          <QuizPerformanceChart quizPerformance={quizPerformance} />

          <TopWordsSection wordMastery={wordMastery} />

          <BoxDistributionChart boxDistribution={boxDistribution} />

          <HistoricalBoxDistributionChart
            historicalBoxDistribution={historicalBoxDistribution}
          />

          {/* Bottom spacing */}
          <View style={{ height: 20 }} />
      </ScrollView>
    </SafeAreaView>
  );
};

export default AnalyticsDashboard;

const createStyles = (theme: ReturnType<typeof useTheme>['theme']) => StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: theme.colors.background.default,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: theme.colors.background.default,
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    paddingBottom: 20,
  },
});
