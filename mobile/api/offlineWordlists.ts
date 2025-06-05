import offlineManager from "@/utils/offlineManager";
import * as wordlistsApi from "./wordlists";

export async function newQuiz(wordlistId: number): Promise<wordlistsApi.Quiz> {
  const isOnline = offlineManager.getNetworkStatus();

  if (isOnline) {
    // Online mode: fetch from API and cache
    try {
      const quiz = await wordlistsApi.newQuiz(wordlistId);

      // Cache for offline use (async, don't wait)
      offlineManager.cacheQuiz(wordlistId, quiz).catch(console.error);

      return quiz;
    } catch (error) {
      // If online request fails, try offline
      const cachedQuiz = await offlineManager.getCachedQuiz(wordlistId);
      if (cachedQuiz) {
        return cachedQuiz;
      }
      throw error;
    }
  } else {
    // Offline mode: get from cache
    const cachedQuiz = await offlineManager.getCachedQuiz(wordlistId);

    if (!cachedQuiz) {
      throw new Error("No cached quiz available for offline use");
    }

    return cachedQuiz;
  }
}

export async function answerQuiz(
  input: wordlistsApi.AnswerQuizInput,
): Promise<void> {
  const isOnline = offlineManager.getNetworkStatus();

  if (!isOnline) {
    // In offline mode, don't track answers
    console.log("Offline mode: quiz answer not tracked");
    return;
  }

  return wordlistsApi.answerQuiz(input);
}
