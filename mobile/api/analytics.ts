import { callAPI } from "./api";

export interface DashboardStats {
  total_words: number;
  words_mastered: number;
  average_mastery: number;
  best_streak: number;
  current_streak: number;
  words_studied_today: number;
  quizzes_today: number;
  accuracy_today: number;
}

export interface WordMasteryStats {
  word_id: number;
  word: string;
  mastery_level: number;
  accuracy: number;
  streak_count: number;
  highest_box: number;
}

export interface LearningProgress {
  date: string;
  words_studied: number;
  accuracy_rate: number;
}

export interface QuizTypePerformance {
  quiz_type: string;
  success_rate: number;
  total_attempts: number;
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
