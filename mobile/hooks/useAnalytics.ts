import {
  DashboardStats,
  getDashboardStats,
  getLearningProgress,
  getQuizPerformance,
  getWordMastery,
  getWordlistProgressSummary,
  getCurrentBoxDistribution,
  getHistoricalBoxDistribution,
  LearningProgress,
  QuizTypePerformance,
  WordMasteryStats,
  WordlistProgressSummary,
  BoxDistributionResponse,
  HistoricalBoxDistributionResponse,
} from "@/api/analytics";
import { useQuery } from "@tanstack/react-query";

type UseAnalyticsResult = {
  dashboardStats?: DashboardStats;
  statsLoading: boolean;
  statsError?: unknown;

  wordMastery?: WordMasteryStats[];
  masteryLoading: boolean;
  masteryError?: unknown;

  learningProgress?: LearningProgress[];
  progressLoading: boolean;
  progressError?: unknown;

  quizPerformance?: QuizTypePerformance[];
  quizPerfLoading: boolean;
  quizPerfError?: unknown;

  wordlistProgress?: WordlistProgressSummary;
  wordlistProgressLoading: boolean;
  wordlistProgressError?: unknown;
  progressPercentage: number;

  boxDistribution?: BoxDistributionResponse;
  boxDistLoading: boolean;
  boxDistError?: unknown;

  historicalBoxDistribution?: HistoricalBoxDistributionResponse;
  historicalBoxDistLoading: boolean;
  historicalBoxDistError?: unknown;

  isPending: boolean;
};

export function useAnalytics(wordlistId: number): UseAnalyticsResult {
  // 1) Dashboard stats
  const {
    data: dashboardStats,
    isLoading: statsLoading,
    error: statsError,
  } = useQuery<DashboardStats, unknown>({
    queryKey: ["analytics", "dashboard"],
    queryFn: getDashboardStats,
  });

  // 2) Word mastery for the given wordlist
  const {
    data: wordMastery,
    isLoading: masteryLoading,
    error: masteryError,
  } = useQuery<WordMasteryStats[], unknown>({
    queryKey: ["analytics", "mastery", wordlistId],
    queryFn: () => getWordMastery(wordlistId),
    enabled: Boolean(wordlistId),
  });

  // 3) Learning progress for the last 7 days
  const {
    data: learningProgress,
    isLoading: progressLoading,
    error: progressError,
  } = useQuery<LearningProgress[], unknown>({
    queryKey: ["analytics", "progress", wordlistId],
    queryFn: () => getLearningProgress(wordlistId, 7),
    enabled: Boolean(wordlistId),
  });

  // 4) Quiz performance
  const {
    data: quizPerformance,
    isLoading: quizPerfLoading,
    error: quizPerfError,
  } = useQuery<QuizTypePerformance[], unknown>({
    queryKey: ["analytics", "quiz-performance", wordlistId],
    queryFn: () => getQuizPerformance(wordlistId),
    enabled: Boolean(wordlistId),
  });

  // 5) Wordlist progress summary
  const {
    data: wordlistProgress,
    isLoading: wordlistProgressLoading,
    error: wordlistProgressError,
  } = useQuery<WordlistProgressSummary, unknown>({
    queryKey: ["analytics", "wordlistProgress", wordlistId],
    queryFn: () => getWordlistProgressSummary(wordlistId),
    enabled: Boolean(wordlistId),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // 6) Box distribution
  const {
    data: boxDistribution,
    isLoading: boxDistLoading,
    error: boxDistError,
  } = useQuery<BoxDistributionResponse, unknown>({
    queryKey: ["analytics", "boxDistribution", wordlistId],
    queryFn: () => getCurrentBoxDistribution(wordlistId),
    enabled: Boolean(wordlistId),
    staleTime: Infinity, // Cache until quiz session invalidates it
  });

  // 7) Historical box distribution
  const {
    data: historicalBoxDistribution,
    isLoading: historicalBoxDistLoading,
    error: historicalBoxDistError,
  } = useQuery<HistoricalBoxDistributionResponse, unknown>({
    queryKey: ["analytics", "historicalBoxDistribution", wordlistId],
    queryFn: () => getHistoricalBoxDistribution(wordlistId, 30),
    enabled: Boolean(wordlistId),
    staleTime: 5 * 60 * 1000, // 5 minutes cache for historical data
  });

  return {
    dashboardStats,
    statsLoading,
    statsError,

    wordMastery,
    masteryLoading,
    masteryError,

    learningProgress,
    progressLoading,
    progressError,

    quizPerformance,
    quizPerfLoading,
    quizPerfError,

    wordlistProgress,
    wordlistProgressLoading,
    wordlistProgressError,
    progressPercentage: wordlistProgress?.progressPercentage ?? 0,

    boxDistribution,
    boxDistLoading,
    boxDistError,

    historicalBoxDistribution,
    historicalBoxDistLoading,
    historicalBoxDistError,

    isPending:
      statsLoading ||
      masteryLoading ||
      progressLoading ||
      quizPerfLoading ||
      wordlistProgressLoading ||
      boxDistLoading ||
      historicalBoxDistLoading,
  };
}
