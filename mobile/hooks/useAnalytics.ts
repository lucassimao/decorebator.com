import {
  OverviewStats,
  getWordlistOverviewStats,
  getLearningProgress,
  getQuizPerformance,
  getWordMastery,
  calculateWordlistProgressFromMastery,
  getCurrentBoxDistribution,
  getHistoricalBoxDistribution,
  getPracticeTime,
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
  overviewStats?: OverviewStats;
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
  const { isPremium } = useUserInfo();

  // Define cache times based on subscription tier
  // Premium users get fresher data for better UX after quiz sessions
  const staleTime = isPremium ? 10 * 1000 : 15 * 60 * 1000; // 10s vs 15min
  const gcTime = isPremium ? 2 * 60 * 1000 : 60 * 60 * 1000; // 2min vs 1hr
  
  // Common query options for analytics
  const commonQueryOptions = {
    staleTime,
    gcTime,
    refetchOnWindowFocus: isPremium, // Premium users get automatic refresh when returning to screen
    refetchOnMount: isPremium ? ("always" as const) : false, // Premium users always get fresh data on mount
  };

  // 1) Overview stats (wordlist-specific)
  const {
    data: overviewStats,
    isLoading: statsLoading,
    error: statsError,
  } = useQuery<OverviewStats, unknown>({
    queryKey: ["analytics", "overview", wordlistId, isPremium],
    queryFn: () => getWordlistOverviewStats(wordlistId),
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
    queryFn: () => getWordMastery(wordlistId),
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
    queryFn: () => getLearningProgress(wordlistId, 7),
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
    queryFn: () => getQuizPerformance(wordlistId),
    enabled: Boolean(wordlistId),
    ...commonQueryOptions,
  });

  // 5) Wordlist progress summary (calculated from word mastery data to avoid duplicate API calls)
  const wordlistProgress = wordMastery ? calculateWordlistProgressFromMastery(wordlistId, wordMastery) : undefined;
  const wordlistProgressLoading = masteryLoading;
  const wordlistProgressError = masteryError;

  // 6) Box distribution
  const {
    data: boxDistribution,
    isLoading: boxDistLoading,
    error: boxDistError,
  } = useQuery<BoxDistributionResponse, unknown>({
    queryKey: ["analytics", "boxDistribution", wordlistId, isPremium],
    queryFn: () => getCurrentBoxDistribution(wordlistId),
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
    queryFn: () => getHistoricalBoxDistribution(wordlistId, 7),
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
    queryFn: () => getPracticeTime(wordlistId, 7),
    enabled: Boolean(wordlistId),
    ...commonQueryOptions,
  });

  return {
    overviewStats,
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
