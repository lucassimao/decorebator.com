import {
  WordlistStats,
  getWordlistStats,
  getWordlistLearningProgress,
  getWordlistQuizTypePerformance,
  getWordlistWordMastery,
  calculateWordlistProgressFromMastery,
  getWordlistCurrentBoxDistribution,
  getWordlistBoxDistributionHistory,
  getWordlistPracticeTime,
  LearningProgress,
  QuizTypePerformance,
  WordMasteryStats,
  WordlistProgressSummary,
  BoxDistributionResponse,
  HistoricalBoxDistributionResponse,
  PracticeTimeResponse,
} from "@/api/analytics";
import { useQuery } from "@tanstack/react-query";
import { useUserInfo } from "@/hooks/users";

type UseAnalyticsResult = {
  stats?: WordlistStats;
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

  practiceTime?: PracticeTimeResponse;
  practiceTimeLoading: boolean;
  practiceTimeError?: unknown;

  isPending: boolean;
};

export function useAnalytics(wordlistId: number): UseAnalyticsResult {
  const { isPremium, cacheConfig } = useUserInfo();

  // Common query options for analytics using centralized cache configuration
  const commonQueryOptions = {
    staleTime: cacheConfig.dataFreshnessDuration,
    gcTime: cacheConfig.memoryRetentionTime,
    refetchOnWindowFocus: cacheConfig.autoRefreshOnFocus,
    refetchOnMount: cacheConfig.alwaysFetchOnMount,
  };

  // 1) Stats (wordlist-specific)
  const {
    data: stats,
    isLoading: statsLoading,
    error: statsError,
  } = useQuery<WordlistStats, unknown>({
    queryKey: ["analytics", "stats", wordlistId, isPremium],
    queryFn: () => getWordlistStats(wordlistId),
    enabled: Boolean(wordlistId),
    ...commonQueryOptions,
  });

  // 2) Word mastery for the given wordlist
  const {
    data: wordMastery,
    isLoading: masteryLoading,
    error: masteryError,
  } = useQuery<WordMasteryStats[], unknown>({
    queryKey: ["analytics", "mastery", wordlistId, isPremium],
    queryFn: () => getWordlistWordMastery(wordlistId),
    enabled: Boolean(wordlistId),
    ...commonQueryOptions,
  });

  // 3) Learning progress for the last 7 days
  const {
    data: learningProgress,
    isLoading: progressLoading,
    error: progressError,
  } = useQuery<LearningProgress[], unknown>({
    queryKey: ["analytics", "progress", wordlistId, isPremium],
    queryFn: () => getWordlistLearningProgress(wordlistId, 7),
    enabled: Boolean(wordlistId),
    ...commonQueryOptions,
  });

  // 4) Quiz performance
  const {
    data: quizPerformance,
    isLoading: quizPerfLoading,
    error: quizPerfError,
  } = useQuery<QuizTypePerformance[], unknown>({
    queryKey: ["analytics", "quiz-performance", wordlistId, isPremium],
    queryFn: () => getWordlistQuizTypePerformance(wordlistId),
    enabled: Boolean(wordlistId),
    ...commonQueryOptions,
  });

  // 5) Wordlist progress summary (calculated from word mastery data to avoid duplicate API calls)
  const wordlistProgress = wordMastery
    ? calculateWordlistProgressFromMastery(wordlistId, wordMastery)
    : undefined;
  const wordlistProgressLoading = masteryLoading;
  const wordlistProgressError = masteryError;

  // 6) Box distribution
  const {
    data: boxDistribution,
    isLoading: boxDistLoading,
    error: boxDistError,
  } = useQuery<BoxDistributionResponse, unknown>({
    queryKey: ["analytics", "boxDistribution", wordlistId, isPremium],
    queryFn: () => getWordlistCurrentBoxDistribution(wordlistId),
    enabled: Boolean(wordlistId),
    ...commonQueryOptions,
  });

  // 7) Historical box distribution
  const {
    data: historicalBoxDistribution,
    isLoading: historicalBoxDistLoading,
    error: historicalBoxDistError,
  } = useQuery<HistoricalBoxDistributionResponse, unknown>({
    queryKey: ["analytics", "historicalBoxDistribution", wordlistId, isPremium],
    queryFn: () => getWordlistBoxDistributionHistory(wordlistId, 7),
    enabled: Boolean(wordlistId),
    ...commonQueryOptions,
  });

  // 8) Practice time for the last 7 days
  const {
    data: practiceTime,
    isLoading: practiceTimeLoading,
    error: practiceTimeError,
  } = useQuery<PracticeTimeResponse, unknown>({
    queryKey: ["analytics", "practiceTime", wordlistId, isPremium],
    queryFn: () => getWordlistPracticeTime(wordlistId, 7),
    enabled: Boolean(wordlistId),
    ...commonQueryOptions,
  });

  return {
    stats,
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

    practiceTime,
    practiceTimeLoading,
    practiceTimeError,

    isPending:
      statsLoading ||
      masteryLoading ||
      progressLoading ||
      quizPerfLoading ||
      wordlistProgressLoading ||
      boxDistLoading ||
      historicalBoxDistLoading ||
      practiceTimeLoading,
  };
}
