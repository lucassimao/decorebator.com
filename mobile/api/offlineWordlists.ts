import offlineManager from "@/utils/offlineManager";
import * as wordlistsApi from "./wordlists";

export async function getUserWordlists(): Promise<wordlistsApi.Wordlist[]> {
  const isOnline = offlineManager.getNetworkStatus();

  if (isOnline) {
    const cacheGeneration = offlineManager.captureCacheGeneration();
    try {
      const wordlists = await wordlistsApi.getUserWordlists();
      offlineManager
        .cacheWordlists(wordlists, cacheGeneration)
        .catch(console.error);
      return wordlists;
    } catch (error) {
      const cached = await offlineManager.getCachedWordlists(cacheGeneration);
      if (cached) {
        return cached;
      }
      throw error;
    }
  }

  const cached = await offlineManager.getCachedWordlists(
    offlineManager.captureCacheGeneration(),
  );
  if (!cached) {
    throw new Error("No cached wordlists available for offline use");
  }

  return cached;
}

export async function newQuiz(
  wordlistId: number,
  quizTypes?: wordlistsApi.QuizType[],
): Promise<wordlistsApi.Quiz> {
  const isOnline = offlineManager.getNetworkStatus();

  if (isOnline) {
    const cacheGeneration = offlineManager.captureCacheGeneration();
    // Online mode: fetch from API and cache
    try {
      const quiz = await wordlistsApi.newQuiz(wordlistId, quizTypes);

      // Cache for offline use (async, don't wait)
      offlineManager
        .cacheQuiz(wordlistId, quiz, cacheGeneration)
        .catch(console.error);

      return quiz;
    } catch (error) {
      // If online request fails, try offline
      const cachedQuiz = await offlineManager.getCachedQuiz(
        wordlistId,
        cacheGeneration,
      );
      if (cachedQuiz) {
        return cachedQuiz;
      }
      throw error;
    }
  } else {
    // Offline mode: get from cache
    const cachedQuiz = await offlineManager.getCachedQuiz(
      wordlistId,
      offlineManager.captureCacheGeneration(),
    );

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

// Words API with offline support
export async function getWords(
  wordlistId: number,
  onlyWithDefinitions = false,
): Promise<wordlistsApi.Word[]> {
  const isOnline = offlineManager.getNetworkStatus();

  if (isOnline) {
    const cacheGeneration = offlineManager.captureCacheGeneration();
    // Online mode: fetch from API and cache
    try {
      const words = await wordlistsApi.getWords(
        wordlistId,
        onlyWithDefinitions,
      );

      // Cache for offline use (async, don't wait)
      offlineManager
        .cacheWords(wordlistId, words, cacheGeneration)
        .catch(console.error);

      return words;
    } catch (error) {
      // If online request fails, try offline
      const cachedWords = await offlineManager.getCachedWords(
        wordlistId,
        cacheGeneration,
      );
      if (cachedWords) {
        return cachedWords;
      }
      throw error;
    }
  } else {
    // Offline mode: get from cache
    const cachedWords = await offlineManager.getCachedWords(
      wordlistId,
      offlineManager.captureCacheGeneration(),
    );

    if (!cachedWords) {
      throw new Error("No cached words available for offline use");
    }

    return cachedWords;
  }
}

// Definitions API with offline support
export async function getWordDefinitions(
  wordlistId: number,
  wordId: number,
): Promise<wordlistsApi.Definition[]> {
  const isOnline = offlineManager.getNetworkStatus();

  if (isOnline) {
    const cacheGeneration = offlineManager.captureCacheGeneration();
    // Online mode: fetch from API and cache
    try {
      const definitions = await wordlistsApi.getWordDefinitions(
        wordlistId,
        wordId,
      );

      // Cache for offline use (async, don't wait)
      offlineManager
        .cacheDefinitions(wordlistId, wordId, definitions, cacheGeneration)
        .catch(console.error);

      return definitions;
    } catch (error) {
      // If online request fails, try offline
      const cachedDefinitions = await offlineManager.getCachedDefinitions(
        wordlistId,
        wordId,
        cacheGeneration,
      );
      if (cachedDefinitions) {
        return cachedDefinitions;
      }
      throw error;
    }
  } else {
    // Offline mode: get from cache
    const cachedDefinitions = await offlineManager.getCachedDefinitions(
      wordlistId,
      wordId,
      offlineManager.captureCacheGeneration(),
    );

    if (!cachedDefinitions) {
      throw new Error("No cached definitions available for offline use");
    }

    return cachedDefinitions;
  }
}

// Preload a wordlist for offline use
export async function preloadWordlistForOffline(
  wordlistId: number,
): Promise<void> {
  const cacheGeneration = offlineManager.captureCacheGeneration();
  const words = await wordlistsApi.getWords(wordlistId, true);

  await offlineManager.preloadWordlistForOffline(
    wordlistId,
    words,
    (wordId) => wordlistsApi.getWordDefinitions(wordlistId, wordId),
    cacheGeneration,
  );
}

// Check if wordlist is available offline
export async function isWordlistAvailableOffline(
  wordlistId: number,
): Promise<boolean> {
  return offlineManager.isWordlistCachedForOffline(wordlistId);
}

// Get cache statistics for a wordlist
export async function getWordlistCacheStats(wordlistId: number) {
  return offlineManager.getWordlistCacheStats(wordlistId);
}
