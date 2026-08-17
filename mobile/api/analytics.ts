import { callAPI } from "./api";
import { getApiBaseUrl } from "./baseUrl";
import { getAllPages, PaginationError } from "./pagination";

export interface WordlistStats {
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

export interface BoxDistribution {
  box1: number;
  box2: number;
  box3: number;
  box4: number;
  box5: number;
  box6: number;
  box7: number;
}

export interface BoxDistributionResponse {
  wordlistId: number;
  distribution: BoxDistribution;
  totalWords: number;
}

export interface HistoricalBoxDistribution {
  date: string;
  boxes: BoxDistribution;
}

export interface HistoricalBoxDistributionResponse {
  wordlistId: number;
  days: number;
  distribution: HistoricalBoxDistribution[];
}

export interface PracticeTimeStats {
  date: string;
  practiceTimeMs: number;
  practiceTimeMinutes: number;
  quizCount: number;
}

export interface PracticeTimeResponse {
  wordlistId: number;
  days: number;
  practiceTime: PracticeTimeStats[];
}

const BASE_URL = getApiBaseUrl();

// 1) Get statistics for a specific wordlist
export async function getWordlistStats(
  wordlistId: number,
): Promise<WordlistStats> {
  const endpoint = `${BASE_URL}/analytics/wordlists/${wordlistId}/stats`;
  const body = await callAPI<{ stats: WordlistStats }>("GET", endpoint);

  // Ensure valid stats structure with safe defaults
  const stats = body.stats || {};
  return {
    totalWords: stats.totalWords || 0,
    wordsMastered: stats.wordsMastered || 0,
    averageMastery: stats.averageMastery || 0,
    bestStreak: stats.bestStreak || 0,
    currentStreak: stats.currentStreak || 0,
    wordsStudiedToday: stats.wordsStudiedToday || 0,
    quizzesToday: stats.quizzesToday || 0,
    accuracyToday: stats.accuracyToday || 0,
  };
}

// 2) Get word mastery for a specific wordlist
export async function getWordlistWordMastery(
  wordlistId: number,
): Promise<WordMasteryStats[]> {
  const endpoint = `${BASE_URL}/analytics/wordlists/${wordlistId}/word-mastery`;
  return getAllPages<{ stats: WordMasteryStats[] }, WordMasteryStats>({
    endpoint,
    getItems: (payload) => requireArray(payload?.stats, "word mastery stats"),
    getItemKey: (stat) => requireItemKey(stat.wordId, "word mastery stat"),
  });
}

// 3) Get learning progress for the last N days
export async function getWordlistLearningProgress(
  wordlistId: number,
  days: number,
): Promise<LearningProgress[]> {
  const endpoint = `${BASE_URL}/analytics/wordlists/${wordlistId}/learning-progress?days=${days}`;
  const payload = await callAPI<{ progress: LearningProgress[] }>(
    "GET",
    endpoint,
  );
  return payload.progress || [];
}

// 4) Get quiz performance across all quiz types for a specific wordlist
export async function getWordlistQuizTypePerformance(
  wordlistId: number,
): Promise<QuizTypePerformance[]> {
  const endpoint = `${BASE_URL}/analytics/wordlists/${wordlistId}/quiz-type-performance`;
  const payload = await callAPI<{ quizPerformance: QuizTypePerformance[] }>(
    "GET",
    endpoint,
  );
  return payload.quizPerformance || [];
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
  const wordMasteryStats = await getWordlistWordMastery(wordlistId);
  return calculateWordlistProgressFromMastery(wordlistId, wordMasteryStats);
}

// Helper function to calculate progress summary from pre-fetched mastery data
export function calculateWordlistProgressFromMastery(
  wordlistId: number,
  wordMasteryStats: WordMasteryStats[],
): WordlistProgressSummary {
  // Safety check: ensure wordMasteryStats is valid array
  const safeStats = Array.isArray(wordMasteryStats) ? wordMasteryStats : [];

  const totalWords = safeStats.length;
  const wordsMastered = safeStats.filter(
    (word) =>
      word && typeof word.masteryLevel === "number" && word.masteryLevel >= 0.8, // Consider 80%+ as mastered
  ).length;

  const averageMastery =
    totalWords > 0
      ? safeStats.reduce((sum, word) => sum + (word?.masteryLevel || 0), 0) /
        totalWords
      : 0;

  const progressPercentage =
    totalWords > 0 ? Math.round(averageMastery * 100) : 0;

  return {
    wordlistId: wordlistId,
    totalWords: totalWords,
    wordsMastered: wordsMastered,
    averageMastery: averageMastery,
    progressPercentage: progressPercentage,
  };
}

// 6) Get current box distribution for a wordlist
export async function getWordlistCurrentBoxDistribution(
  wordlistId: number,
): Promise<BoxDistributionResponse> {
  const endpoint = `${BASE_URL}/analytics/wordlists/${wordlistId}/current-box-distribution`;
  const response = await callAPI<BoxDistributionResponse>("GET", endpoint);

  // Ensure valid response structure
  return {
    wordlistId: response.wordlistId || wordlistId,
    distribution: response.distribution || {
      box1: 0,
      box2: 0,
      box3: 0,
      box4: 0,
      box5: 0,
      box6: 0,
      box7: 0,
    },
    totalWords: response.totalWords || 0,
  };
}

// 7) Get historical box distribution for a wordlist
export async function getWordlistBoxDistributionHistory(
  wordlistId: number,
  days: number = 30,
): Promise<HistoricalBoxDistributionResponse> {
  const endpoint = `${BASE_URL}/analytics/wordlists/${wordlistId}/box-distribution-history?days=${days}`;
  const response = await callAPI<HistoricalBoxDistributionResponse>(
    "GET",
    endpoint,
  );

  // Ensure valid response structure
  return {
    wordlistId: response.wordlistId || wordlistId,
    days: response.days || days,
    distribution: Array.isArray(response.distribution)
      ? response.distribution
      : [],
  };
}

// 8) Get practice time for a wordlist
export async function getWordlistPracticeTime(
  wordlistId: number,
  days: number = 7,
): Promise<PracticeTimeResponse> {
  const endpoint = `${BASE_URL}/analytics/wordlists/${wordlistId}/practice-time?days=${days}`;
  const response = await callAPI<PracticeTimeResponse>("GET", endpoint);

  // Ensure valid response structure
  return {
    wordlistId: response.wordlistId || wordlistId,
    days: response.days || days,
    practiceTime: Array.isArray(response.practiceTime)
      ? response.practiceTime
      : [],
  };
}

// 9) Get progress summary for all wordlists - NEW BATCH ENDPOINT
export interface WordlistProgress {
  wordlistId: number;
  wordlistName: string;
  languageCode: string;
  totalWords: number;
  wordsMastered: number;
  progressPercent: number;
  dueCount: number;
  currentStreak: number;
  lastActivityDate?: string;
}

export interface ProgressSummaryResponse {
  wordlists: WordlistProgress[];
}

export async function getProgressSummary(): Promise<ProgressSummaryResponse> {
  const endpoint = `${BASE_URL}/analytics/progress-summary`;
  return {
    wordlists: await getAllPages<ProgressSummaryResponse, WordlistProgress>({
      endpoint,
      getItems: (response) =>
        requireArray(response?.wordlists, "progress summary wordlists"),
      getItemKey: (wordlist) =>
        requireItemKey(wordlist.wordlistId, "progress summary wordlist"),
    }),
  };
}

function requireArray<T>(value: unknown, responseField: string): T[] {
  if (!Array.isArray(value)) {
    throw new PaginationError(
      `Invalid paginated ${responseField} response: expected an array`,
    );
  }
  return value as T[];
}

function requireItemKey(value: unknown, itemName: string): string {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value <= 0) {
    throw new PaginationError(
      `Invalid paginated ${itemName} response: expected a positive ID`,
    );
  }
  return String(value);
}
