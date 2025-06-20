import { useNavigation } from "@react-navigation/native";
import { useAnalytics } from "@/hooks/useAnalytics";
import { LinearGradient } from "expo-linear-gradient";
import { useLocalSearchParams } from "expo-router";
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
  colors,
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
        <AnalyticsHeader onBackPress={() => navigation.goBack()} />

        <ScrollView showsVerticalScrollIndicator={false}>
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
});
