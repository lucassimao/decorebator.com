import { callAPI } from "./api";

export interface DashboardStats {
  totalWords: number;
  wordsMastered: number;
  averageMastery: number;
  bestStreak: number;
  currentStreak: number;
  wordsStudiedToday: number;
  quizzesToday: number;
  accuracyToday: number;
}

export interface WordMasteryStats {
  wordId: number;
  word: string;
  masteryLevel: number;
  accuracy: number;
  streakCount: number;
  highestBox: number;
}

export interface LearningProgress {
  date: string;
  wordsStudied: number;
  accuracyRate: number;
}

export interface QuizTypePerformance {
  quizType: string;
  successRate: number;
  totalAttempts: number;
}

const BASE_URL = process.env.EXPO_PUBLIC_API_URL;

// 1) Get dashboard statistics
export async function getDashboardStats(): Promise<DashboardStats> {
  const endpoint = `${BASE_URL}/analytics/dashboard`;
  const body = await callAPI<{ stats: DashboardStats }>("GET", endpoint);
  return body.stats;
}

// 2) Get word mastery for a specific wordlist
export async function getWordMastery(
  wordlistId: number,
): Promise<WordMasteryStats[]> {
  const endpoint = `${BASE_URL}/analytics/wordlists/${wordlistId}/mastery`;
  const payload = await callAPI<{ stats: WordMasteryStats[] }>("GET", endpoint);
  return payload.stats;
}

// 3) Get learning progress for the last N days
export async function getLearningProgress(
  wordlistId: number,
  days: number,
): Promise<LearningProgress[]> {
  const endpoint = `${BASE_URL}/analytics/wordlists/${wordlistId}/progress?days=${days}`;
  const payload = await callAPI<{ progress: LearningProgress[] }>(
    "GET",
    endpoint,
  );
  return payload.progress;
}

// 4) Get quiz performance across all quiz types
export async function getQuizPerformance(): Promise<QuizTypePerformance[]> {
  const endpoint = `${BASE_URL}/analytics/quiz-performance`;
  const payload = await callAPI<{ quiz_performance: QuizTypePerformance[] }>(
    "GET",
    endpoint,
  );
  return payload.quiz_performance;
}

// 5) Get wordlist progress summary with mastery-based calculations
export interface WordlistProgressSummary {
  wordlistId: number;
  totalWords: number;
  wordsMastered: number;
  averageMastery: number;
  progressPercentage: number;
}

export async function getWordlistProgressSummary(
  wordlistId: number,
): Promise<WordlistProgressSummary> {
  const wordMasteryStats = await getWordMastery(wordlistId);
  
  const totalWords = wordMasteryStats.length;
  const wordsMastered = wordMasteryStats.filter(
    (word) => word.masteryLevel >= 0.8 // Consider 80%+ as mastered
  ).length;
  
  const averageMastery = totalWords > 0 
    ? wordMasteryStats.reduce((sum, word) => sum + word.masteryLevel, 0) / totalWords
    : 0;
  
  const progressPercentage = totalWords > 0 
    ? Math.round((wordsMastered / totalWords) * 100)
    : 0;

  return {
    wordlistId: wordlistId,
    totalWords: totalWords,
    wordsMastered: wordsMastered,
    averageMastery: averageMastery,
    progressPercentage: progressPercentage,
  };
}
