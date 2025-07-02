import { useNavigation } from "@react-navigation/native";
import { useAnalytics } from "@/hooks/useAnalytics";
import { useLocalSearchParams } from "expo-router";
import { useTheme } from "@/contexts/ThemeContext";
import { createCommonStyles } from "@/styles/common";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";
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
  const { isTablet, contentWidth } = useResponsive();
  const spacing = useResponsiveSpacing();
  const commonStyles = createCommonStyles(theme);
  const styles = createStyles(theme, isTablet, contentWidth, spacing);

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
        contentContainerStyle={[
          styles.scrollContent,
          isTablet && styles.tabletScrollContent,
        ]}
      >
        <View
          style={[
            styles.contentContainer,
            isTablet && styles.tabletContentContainer,
          ]}
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
          <View style={{ height: spacing.lg }} />
        </View>
      </ScrollView>
    </SafeAreaView>
  );
};

export default AnalyticsDashboard;

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  contentWidth: number,
  spacing: ReturnType<typeof useResponsiveSpacing>,
) =>
  StyleSheet.create({
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
      paddingBottom: spacing.lg,
    },
    tabletScrollContent: {
      alignItems: "center",
      paddingHorizontal: spacing.lg,
    },
    contentContainer: {
      width: "100%",
    },
    tabletContentContainer: {
      width: contentWidth,
      maxWidth: 800,
      alignSelf: "center",
    },
  });
