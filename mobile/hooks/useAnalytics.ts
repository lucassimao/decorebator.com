import {
  DashboardStats,
  getDashboardStats,
  getLearningProgress,
  getQuizPerformance,
  getWordMastery,
  LearningProgress,
  QuizTypePerformance,
  WordMasteryStats,
} from '@/api/analytics';
import { useQuery } from '@tanstack/react-query';

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

  isPending:boolean
}

export function useAnalytics(wordlistId: number): UseAnalyticsResult {
  // 1) Dashboard stats
  const {
    data: dashboardStats,
    isLoading: statsLoading,
    error: statsError,
  } = useQuery<DashboardStats, unknown>({
    queryKey: ['analytics', 'dashboard'],
    queryFn: getDashboardStats,
  });

  // 2) Word mastery for the given wordlist
  const {
    data: wordMastery,
    isLoading: masteryLoading,
    error: masteryError,
  } = useQuery<WordMasteryStats[], unknown>({
    queryKey: ['analytics', 'mastery', wordlistId],
    queryFn: () => getWordMastery(wordlistId),
    enabled: Boolean(wordlistId),
  });

  // 3) Learning progress for the last 7 days
  const {
    data: learningProgress,
    isLoading: progressLoading,
    error: progressError,
  } = useQuery<LearningProgress[], unknown>({
    queryKey: ['analytics', 'progress', wordlistId],
    queryFn: () => getLearningProgress(wordlistId, 7),
    enabled: Boolean(wordlistId),
  });

  // 4) Quiz performance
  const {
    data: quizPerformance,
    isLoading: quizPerfLoading,
    error: quizPerfError,
  } = useQuery<QuizTypePerformance[], unknown>({
    queryKey: ['analytics', 'quiz-performance'],
    queryFn: getQuizPerformance,
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

    isPending: statsLoading || masteryLoading ||progressLoading||quizPerfLoading
  };
}
