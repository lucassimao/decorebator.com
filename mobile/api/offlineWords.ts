import offlineManager from "@/utils/offlineManager";
import * as wordlistsApi from "./wordlists";

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
