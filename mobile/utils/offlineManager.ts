import AsyncStorage from "@react-native-async-storage/async-storage";
import NetInfo, { NetInfoState } from "@react-native-community/netinfo";
import * as FileSystem from "expo-file-system/legacy";
import { Quiz, Word, Definition, Wordlist } from "@/api/wordlists";
import { Platform } from "react-native";

const CACHE_PREFIX = "decorebator_offline_";
const QUIZ_CACHE_KEY = `${CACHE_PREFIX}quiz_`;
const WORDS_CACHE_KEY = `${CACHE_PREFIX}words_`;
const WORDLISTS_CACHE_KEY = `${CACHE_PREFIX}wordlists`;
const DEFINITIONS_CACHE_KEY = `${CACHE_PREFIX}definitions_`;
const ASSET_CACHE_DIR = `${FileSystem.documentDirectory}decorebator_assets/`;
const CACHE_EXPIRY_HOURS = 72; // 3 days

// Network detection constants
const CONNECTIVITY_TEST_ENDPOINTS = [
  process.env.EXPO_PUBLIC_CONNECTIVITY_URL ||
    "https://clients3.google.com/generate_204", // Google connectivity check
];
const CONNECTIVITY_TIMEOUT = 5000; // 5 seconds
const CONNECTIVITY_CACHE_DURATION = 30000; // 30 seconds

interface CachedQuiz {
  quiz: Quiz;
  timestamp: number;
  wordlistId: number;
}

interface CachedAsset {
  localUri: string;
  timestamp: number;
}

interface CachedWords {
  words: Word[];
  timestamp: number;
  wordlistId: number;
}

interface CachedWordlists {
  wordlists: Wordlist[];
  timestamp: number;
}

interface CachedDefinitions {
  definitions: Definition[];
  timestamp: number;
  wordlistId: number;
  wordId: number;
}

interface ConnectivityTestResult {
  isConnected: boolean;
  responseTime: number;
  timestamp: number;
  endpoint: string;
}

interface NetworkState {
  isOnline: boolean;
  connectionQuality: "fast" | "slow" | "unknown";
  lastConnectivityTest: ConnectivityTestResult | null;
  consecutiveFailures: number;
}

interface CircuitBreakerState {
  isOpen: boolean;
  failureCount: number;
  lastFailureTime: number;
  nextAttemptTime: number;
}

interface RetryConfig {
  maxAttempts: number;
  baseDelayMs: number;
  maxDelayMs: number;
  backoffMultiplier: number;
}

class OfflineCacheIdentityChangedError extends Error {}

class OfflineManager {
  private isServer =
    Platform.OS === "web" &&
    (typeof window === "undefined" || typeof document === "undefined");
  private networkState: NetworkState = {
    isOnline: true,
    connectionQuality: "unknown",
    lastConnectivityTest: null,
    consecutiveFailures: 0,
  };
  private isPremium: boolean = false;
  private networkListeners: ((isOnline: boolean) => void)[] = [];
  private connectivityTestCache: Map<string, ConnectivityTestResult> =
    new Map();
  private netInfoUnsubscribe: (() => void) | null = null;
  private connectivityTestTimer: ReturnType<typeof setInterval> | null = null;
  private lastNetInfoState: NetInfoState | null = null;
  private circuitBreaker: CircuitBreakerState = {
    isOpen: false,
    failureCount: 0,
    lastFailureTime: 0,
    nextAttemptTime: 0,
  };
  private retryConfig: RetryConfig = {
    maxAttempts: 3,
    baseDelayMs: 1000,
    maxDelayMs: 30000,
    backoffMultiplier: 2,
  };
  private cacheGeneration = 0;
  private inFlightCacheWrites = new Set<Promise<void>>();

  private trackCacheWrite(
    operation: (generation: number) => Promise<void>,
    originatingGeneration = this.cacheGeneration,
  ): Promise<void> {
    const generation = originatingGeneration;
    const write = operation(generation).finally(() => {
      this.inFlightCacheWrites.delete(write);
    });
    this.inFlightCacheWrites.add(write);
    return write;
  }

  private assertCacheGeneration(generation: number) {
    if (generation !== this.cacheGeneration) {
      throw new OfflineCacheIdentityChangedError(
        "Offline cache identity changed during write",
      );
    }
  }

  public captureCacheGeneration(): number {
    return this.cacheGeneration;
  }

  private async guardCacheRead<T>(
    generation: number,
    operation: () => Promise<T>,
  ): Promise<T> {
    this.assertCacheGeneration(generation);
    const result = await operation();
    this.assertCacheGeneration(generation);
    return result;
  }

  private cacheReadFallback<T>(error: unknown): T | null {
    if (error instanceof OfflineCacheIdentityChangedError) throw error;
    return null;
  }

  private async writeCacheItem(generation: number, key: string, value: string) {
    this.assertCacheGeneration(generation);
    await AsyncStorage.setItem(key, value);
    this.assertCacheGeneration(generation);
  }

  constructor() {
    if (this.isServer) {
      return;
    }
    this.configureNetInfo();
    this.initializeNetworkStatus();
    this.initializeNetworkListener();
    this.ensureAssetDirectory();
    this.cleanupExpiredCache().catch((error) => {
      console.error("Error cleaning up offline cache:", error);
    });
  }

  private configureNetInfo() {
    // Configure NetInfo with optimized settings for better reliability
    NetInfo.configure({
      reachabilityUrl: "https://clients3.google.com/generate_204",
      reachabilityTest: async (response) => response.status === 204,
      reachabilityLongTimeout: 60 * 1000, // 60s
      reachabilityShortTimeout: 5 * 1000, // 5s
      reachabilityRequestTimeout: 15 * 1000, // 15s
      reachabilityShouldRun: () => true,
      shouldFetchWiFiSSID: false, // Avoid memory leaks
      useNativeReachability: true,
    });
  }

  private async ensureAssetDirectory() {
    if (Platform.OS === "web") return;

    const dirInfo = await FileSystem.getInfoAsync(ASSET_CACHE_DIR);
    if (!dirInfo.exists) {
      await FileSystem.makeDirectoryAsync(ASSET_CACHE_DIR, {
        intermediates: true,
      });
    }
  }

  private async initializeNetworkStatus() {
    try {
      const state = await NetInfo.fetch();
      this.lastNetInfoState = state;
      const basicOnlineStatus = this.determineBasicOnlineStatus(state);

      if (basicOnlineStatus && state.isInternetReachable == null) {
        // If internet reachability is unknown, verify with connectivity test
        const connectivityResult = await this.performConnectivityTest();
        this.updateNetworkState(
          connectivityResult.isConnected,
          connectivityResult.responseTime,
        );
      } else {
        // If we have a clear signal from NetInfo, trust it
        this.updateNetworkState(basicOnlineStatus, -1);
      }
    } catch (error) {
      console.warn("Failed to fetch initial network state:", error);
      // Keep default value of true to avoid blocking user
      this.updateNetworkState(true, -1);
    }
  }

  private determineBasicOnlineStatus(state: NetInfoState): boolean {
    // Check if we have a clear connection
    if (state.isConnected === true && state.isInternetReachable === true) {
      return true;
    }

    // If isConnected is false, we're definitely offline
    if (state.isConnected === false) {
      return false;
    }

    // If isInternetReachable is false, we're offline
    if (state.isInternetReachable === false) {
      return false;
    }

    // If we have basic connectivity but internet reachability is unknown, test it
    if (state.isConnected === true && state.isInternetReachable == null) {
      return true; // Will be verified by connectivity test
    }

    // If both are null/undefined, assume online but verify with connectivity test
    return true;
  }

  private async performConnectivityTest(): Promise<ConnectivityTestResult> {
    // Check circuit breaker
    if (this.isCircuitBreakerOpen()) {
      return {
        isConnected: false,
        responseTime: -1,
        timestamp: Date.now(),
        endpoint: "circuit_breaker_open",
      };
    }

    const testStartTime = Date.now();

    // Check cache first
    const cachedResult = this.getCachedConnectivityResult();
    if (
      cachedResult &&
      Date.now() - cachedResult.timestamp < CONNECTIVITY_CACHE_DURATION
    ) {
      return cachedResult;
    }

    // Try each endpoint with enhanced retry logic
    for (let attempt = 0; attempt < this.retryConfig.maxAttempts; attempt++) {
      for (const endpoint of CONNECTIVITY_TEST_ENDPOINTS) {
        try {
          const result = await this.executeWithRetry(
            () => this.testEndpointConnectivity(endpoint),
            attempt,
          );

          if (result.isConnected) {
            // Reset circuit breaker on success
            this.resetCircuitBreaker();
            // Cache successful result
            this.connectivityTestCache.set("last_successful", result);
            return result;
          }
        } catch (error) {
          console.warn(`Connectivity test failed for ${endpoint}:`, error);
          this.recordCircuitBreakerFailure();
        }
      }

      // Enhanced exponential backoff between attempts
      if (attempt < this.retryConfig.maxAttempts - 1) {
        const delay = this.calculateBackoffDelay(attempt);
        await this.delay(delay);
      }
    }

    // All tests failed - record failure and open circuit breaker if needed
    this.recordCircuitBreakerFailure();
    const result: ConnectivityTestResult = {
      isConnected: false,
      responseTime: Date.now() - testStartTime,
      timestamp: Date.now(),
      endpoint: "all_failed",
    };

    this.connectivityTestCache.set("last_failed", result);
    return result;
  }

  private async testEndpointConnectivity(
    endpoint: string,
  ): Promise<ConnectivityTestResult> {
    const startTime = Date.now();

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        reject(new Error(`Connectivity test timeout for ${endpoint}`));
      }, CONNECTIVITY_TIMEOUT);

      fetch(endpoint, {
        method: "GET",
        cache: "no-cache",
        headers: {
          "Cache-Control": "no-cache, no-store, must-revalidate",
          Pragma: "no-cache",
        },
      })
        .then((response) => {
          clearTimeout(timeoutId);
          const responseTime = Date.now() - startTime;

          resolve({
            isConnected: response.ok || response.status === 204,
            responseTime,
            timestamp: Date.now(),
            endpoint,
          });
        })
        .catch((error) => {
          clearTimeout(timeoutId);
          reject(error);
        });
    });
  }

  private getCachedConnectivityResult(): ConnectivityTestResult | null {
    const cached = this.connectivityTestCache.get("last_successful");
    return cached || null;
  }

  private updateNetworkState(isOnline: boolean, responseTime: number) {
    const wasOnline = this.networkState.isOnline;

    this.networkState.isOnline = isOnline;
    this.networkState.connectionQuality =
      this.determineConnectionQuality(responseTime);

    if (isOnline) {
      this.networkState.consecutiveFailures = 0;
    } else {
      this.networkState.consecutiveFailures++;
    }

    // Notify listeners if status changed
    if (wasOnline !== isOnline) {
      console.log(
        `Network status changed: ${wasOnline ? "online" : "offline"} → ${isOnline ? "online" : "offline"} (quality: ${this.networkState.connectionQuality}, response: ${responseTime}ms)`,
      );
      this.networkListeners.forEach((listener) => listener(isOnline));
    }
  }

  private determineConnectionQuality(
    responseTime: number,
  ): "fast" | "slow" | "unknown" {
    if (responseTime < 0) return "unknown";
    if (responseTime < 1000) return "fast"; // Under 1 second
    if (responseTime < 3000) return "slow"; // 1-3 seconds
    return "slow"; // Over 3 seconds
  }

  private delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  // Promise.allSettled polyfill for React Native compatibility
  private async promiseAllSettledPolyfill<T>(
    promises: Promise<T>[],
  ): Promise<{ status: "fulfilled" | "rejected"; value?: T; reason?: any }[]> {
    const results = await Promise.all(
      promises.map(async (promise) => {
        try {
          const value = await promise;
          return { status: "fulfilled" as const, value };
        } catch (reason) {
          return { status: "rejected" as const, reason };
        }
      }),
    );
    return results;
  }

  private async executeWithRetry<T>(
    operation: () => Promise<T>,
    attemptIndex: number,
  ): Promise<T> {
    try {
      return await operation();
    } catch (error) {
      if (attemptIndex < this.retryConfig.maxAttempts - 1) {
        const delay = this.calculateBackoffDelay(attemptIndex);
        await this.delay(delay);
        return this.executeWithRetry(operation, attemptIndex + 1);
      }
      throw error;
    }
  }

  private calculateBackoffDelay(attemptIndex: number): number {
    const exponentialDelay =
      this.retryConfig.baseDelayMs *
      Math.pow(this.retryConfig.backoffMultiplier, attemptIndex);

    // Add jitter to prevent thundering herd
    const jitter = Math.random() * 0.1 * exponentialDelay;

    return Math.min(exponentialDelay + jitter, this.retryConfig.maxDelayMs);
  }

  private isCircuitBreakerOpen(): boolean {
    if (!this.circuitBreaker.isOpen) {
      return false;
    }

    // Check if we can attempt a half-open state
    const now = Date.now();
    if (now >= this.circuitBreaker.nextAttemptTime) {
      this.circuitBreaker.isOpen = false;
      console.log("Circuit breaker transitioning to half-open state");
      return false;
    }

    return true;
  }

  private recordCircuitBreakerFailure(): void {
    this.circuitBreaker.failureCount++;
    this.circuitBreaker.lastFailureTime = Date.now();

    // Open circuit breaker after threshold failures
    const failureThreshold = 5;
    if (this.circuitBreaker.failureCount >= failureThreshold) {
      this.circuitBreaker.isOpen = true;
      // Exponential backoff for next attempt (starts at 1 minute, caps at 10 minutes)
      const backoffMinutes = Math.min(
        Math.pow(
          2,
          Math.floor(this.circuitBreaker.failureCount / failureThreshold),
        ),
        10,
      );
      this.circuitBreaker.nextAttemptTime =
        Date.now() + backoffMinutes * 60 * 1000;

      console.log(
        `Circuit breaker opened. Next attempt in ${backoffMinutes} minutes.`,
      );
    }
  }

  private resetCircuitBreaker(): void {
    if (this.circuitBreaker.failureCount > 0 || this.circuitBreaker.isOpen) {
      console.log("Circuit breaker reset - connectivity restored");
    }

    this.circuitBreaker.isOpen = false;
    this.circuitBreaker.failureCount = 0;
    this.circuitBreaker.lastFailureTime = 0;
    this.circuitBreaker.nextAttemptTime = 0;
  }

  private initializeNetworkListener() {
    this.netInfoUnsubscribe = NetInfo.addEventListener(async (state) => {
      this.lastNetInfoState = state;
      const basicOnlineStatus = this.determineBasicOnlineStatus(state);

      if (basicOnlineStatus && state.isInternetReachable == null) {
        // If basic check indicates online but reachability is unknown, verify
        try {
          const connectivityResult = await this.performConnectivityTest();
          this.updateNetworkState(
            connectivityResult.isConnected,
            connectivityResult.responseTime,
          );
        } catch (error) {
          console.warn("Connectivity test failed in network listener:", error);
          // Fall back to basic status if connectivity test fails
          this.updateNetworkState(basicOnlineStatus, -1);
        }
      } else {
        // If we have a clear signal from NetInfo, trust it immediately
        this.updateNetworkState(basicOnlineStatus, -1);
      }
    });

    // Start periodic connectivity verification for enhanced reliability
    this.startPeriodicConnectivityCheck();
  }

  private startPeriodicConnectivityCheck() {
    // Clear existing timer
    if (this.connectivityTestTimer) {
      clearInterval(this.connectivityTestTimer);
    }

    // Run connectivity test every 30 seconds when online
    this.connectivityTestTimer = setInterval(async () => {
      if (
        this.networkState.isOnline &&
        this.lastNetInfoState?.isInternetReachable == null
      ) {
        try {
          const result = await this.performConnectivityTest();
          if (!result.isConnected && this.networkState.isOnline) {
            // Connection lost, update state
            this.updateNetworkState(false, result.responseTime);
          }
        } catch (error) {
          console.warn("Periodic connectivity check failed:", error);
        }
      }
    }, 30000); // 30 seconds
  }

  public subscribeToNetworkChanges(listener: (isOnline: boolean) => void) {
    this.networkListeners.push(listener);
    return () => {
      this.networkListeners = this.networkListeners.filter(
        (l) => l !== listener,
      );
    };
  }

  public setUserPremiumStatus(isPremium: boolean) {
    this.isPremium = isPremium;
  }

  public isOfflineAvailable(): boolean {
    return this.isPremium && !this.networkState.isOnline;
  }

  public getNetworkStatus(): boolean {
    return this.networkState.isOnline;
  }

  public getNetworkState(): NetworkState {
    return { ...this.networkState };
  }

  public getConnectionQuality(): "fast" | "slow" | "unknown" {
    return this.networkState.connectionQuality;
  }

  public async forceConnectivityTest(): Promise<boolean> {
    try {
      const result = await this.performConnectivityTest();
      this.updateNetworkState(result.isConnected, result.responseTime);
      return result.isConnected;
    } catch (error) {
      console.warn("Force connectivity test failed:", error);
      return false;
    }
  }

  // Quiz caching methods with atomic operations
  public async cacheQuiz(
    wordlistId: number,
    quiz: Quiz,
    originatingGeneration = this.cacheGeneration,
  ): Promise<void> {
    if (!this.isPremium) return;

    await this.trackCacheWrite(
      (generation) =>
        this.withErrorRecovery(async () => {
          const cacheKey = `${QUIZ_CACHE_KEY}${wordlistId}-${quiz.id}`;
          const cachedData: CachedQuiz = {
            quiz,
            timestamp: Date.now(),
            wordlistId,
          };

          // Atomic operation with rollback capability
          await this.atomicCacheOperation(async () => {
            // Cache the quiz data
            await this.writeCacheItem(
              generation,
              cacheKey,
              JSON.stringify(cachedData),
            );

            // Cache associated assets
            await this.cacheQuizAssets(quiz, generation);
          }, [
            // Rollback: remove quiz data if asset caching fails
            async () => AsyncStorage.removeItem(cacheKey),
          ]);
        }, "Quiz caching"),
      originatingGeneration,
    );
  }

  private async cacheQuizAssets(quiz: Quiz, generation: number): Promise<void> {
    const assetPromises: Promise<void>[] = [];

    // Cache image for WORD_FROM_IMAGE quiz type (value contains image URL)
    if (quiz.type === "WORD_FROM_IMAGE" && quiz.value) {
      assetPromises.push(this.cacheAsset(quiz.value, "image", generation));
    }

    // Cache audio URL if present (for audio-based quiz types)
    if (quiz.audioURL) {
      assetPromises.push(this.cacheAsset(quiz.audioURL, "audio", generation));
    }

    await this.promiseAllSettledPolyfill(assetPromises);
  }

  private async cacheAsset(
    url: string,
    type: "image" | "audio",
    generation: number,
  ): Promise<void> {
    try {
      const filename = this.getFilenameFromUrl(url);
      const localUri = `${ASSET_CACHE_DIR}${filename}`;

      // Check if already cached
      const fileInfo = await FileSystem.getInfoAsync(localUri);
      if (fileInfo.exists) {
        // Update timestamp
        const metaKey = `${CACHE_PREFIX}asset_meta_${filename}`;
        await this.writeCacheItem(
          generation,
          metaKey,
          JSON.stringify({
            localUri,
            timestamp: Date.now(),
          }),
        );
        return;
      }

      // Download and cache
      const downloadResult = await FileSystem.downloadAsync(url, localUri);
      this.assertCacheGeneration(generation);

      if (downloadResult.status === 200) {
        // Store metadata
        const metaKey = `${CACHE_PREFIX}asset_meta_${filename}`;
        await this.writeCacheItem(
          generation,
          metaKey,
          JSON.stringify({
            localUri,
            timestamp: Date.now(),
          }),
        );
      }
    } catch (error) {
      console.error(`Error caching ${type} asset:`, error);
    }
  }

  private getFilenameFromUrl(url: string): string {
    const urlParts = url.split("/");
    const filename = urlParts[urlParts.length - 1].split("?")[0];
    return filename || `cached_${Date.now()}`;
  }

  public async getCachedQuiz(
    wordlistId: number,
    originatingGeneration = this.cacheGeneration,
  ): Promise<Quiz | null> {
    this.assertCacheGeneration(originatingGeneration);
    if (!this.isPremium || this.networkState.isOnline) return null;

    try {
      return await this.guardCacheRead(originatingGeneration, async () => {
        // Get all keys from AsyncStorage
        const allKeys = await AsyncStorage.getAllKeys();

        // Find keys that match the pattern for this wordlist
        const wordlistQuizKeys = allKeys.filter((key) =>
          key.startsWith(`${QUIZ_CACHE_KEY}${wordlistId}-`),
        );

        if (wordlistQuizKeys.length === 0) return null;

        // Randomize the order of keys to avoid always returning the same quiz
        const shuffledKeys = this.shuffleArray([...wordlistQuizKeys]);

        // Try to find a valid cached quiz from the randomized list
        for (const cacheKey of shuffledKeys) {
          const cachedDataStr = await AsyncStorage.getItem(cacheKey);

          if (!cachedDataStr) continue;

          const cachedData: CachedQuiz = JSON.parse(cachedDataStr);

          // Check if cache is expired
          if (this.isCacheExpired(cachedData.timestamp)) {
            await AsyncStorage.removeItem(cacheKey);
            continue;
          }

          // Validate all required assets are available
          const validatedQuiz = await this.validateQuizAssets(cachedData.quiz);

          if (validatedQuiz) {
            return validatedQuiz;
          }
        }

        return null;
      });
    } catch (error) {
      if (error instanceof OfflineCacheIdentityChangedError) throw error;
      console.error("Error retrieving cached quiz:", error);
      return null;
    }
  }

  // Fisher-Yates shuffle algorithm
  private shuffleArray<T>(array: T[]): T[] {
    const shuffled = [...array];
    for (let i = shuffled.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
    }
    return shuffled;
  }

  // Get all cached quizzes for a wordlist (for future use)
  public async getAllCachedQuizzes(wordlistId: number): Promise<Quiz[]> {
    if (!this.isPremium) return [];

    const quizzes: Quiz[] = [];

    try {
      const allKeys = await AsyncStorage.getAllKeys();
      const wordlistQuizKeys = allKeys.filter((key) =>
        key.startsWith(`${QUIZ_CACHE_KEY}${wordlistId}-`),
      );

      for (const cacheKey of wordlistQuizKeys) {
        const cachedDataStr = await AsyncStorage.getItem(cacheKey);
        if (!cachedDataStr) continue;

        const cachedData: CachedQuiz = JSON.parse(cachedDataStr);

        if (!this.isCacheExpired(cachedData.timestamp)) {
          const validatedQuiz = await this.validateQuizAssets(cachedData.quiz);
          if (validatedQuiz) {
            quizzes.push(validatedQuiz);
          }
        } else {
          // Clean up expired quiz
          await AsyncStorage.removeItem(cacheKey);
        }
      }
    } catch (error) {
      console.error("Error retrieving all cached quizzes:", error);
    }

    return quizzes;
  }

  // Clean up expired cache data (can be called periodically)
  public async cleanupExpiredCache(): Promise<void> {
    try {
      const allKeys = await AsyncStorage.getAllKeys();
      const cacheKeys = allKeys.filter((key) => key.startsWith(CACHE_PREFIX));

      for (const cacheKey of cacheKeys) {
        const cachedDataStr = await AsyncStorage.getItem(cacheKey);
        if (!cachedDataStr) continue;

        try {
          const cachedData = JSON.parse(cachedDataStr);

          if (
            cachedData.timestamp &&
            this.isCacheExpired(cachedData.timestamp)
          ) {
            await AsyncStorage.removeItem(cacheKey);
          }
        } catch {
          // If we can't parse the data, remove it
          await AsyncStorage.removeItem(cacheKey);
        }
      }
    } catch (error) {
      console.error("Error cleaning up expired cache:", error);
    }
  }

  // Legacy method for backwards compatibility
  public async cleanupExpiredQuizzes(): Promise<void> {
    return this.cleanupExpiredCache();
  }

  private async validateQuizAssets(quiz: Quiz): Promise<Quiz | null> {
    const validatedQuiz = { ...quiz };
    let isValid = true;

    // Check image availability for WORD_FROM_IMAGE quiz
    if (quiz.type === "WORD_FROM_IMAGE" && quiz.value) {
      const localUri = await this.getLocalAssetUri(quiz.value);
      if (localUri) {
        validatedQuiz.value = localUri;
      } else {
        isValid = false;
      }
    }

    // Check audio availability for audio-based quiz types
    if (quiz.audioURL) {
      const localUri = await this.getLocalAssetUri(quiz.audioURL);
      if (localUri) {
        validatedQuiz.audioURL = localUri;
      } else {
        // Audio is required for audio-based quiz types
        if (
          quiz.type === "WORD_FROM_AUDIO" ||
          quiz.type === "MEANING_FROM_AUDIO" ||
          quiz.type === "WORD_FROM_EXAMPLE_AUDIO"
        ) {
          isValid = false;
        }
      }
    }

    // Return null if assets are missing
    if (!isValid) {
      return null;
    }

    return validatedQuiz;
  }

  private async getLocalAssetUri(url: string): Promise<string | null> {
    try {
      const filename = this.getFilenameFromUrl(url);
      const localUri = `${ASSET_CACHE_DIR}${filename}`;

      // Check if file exists
      const fileInfo = await FileSystem.getInfoAsync(localUri);
      if (!fileInfo.exists) {
        return null;
      }

      // Check metadata for expiry
      const metaKey = `${CACHE_PREFIX}asset_meta_${filename}`;
      const metaStr = await AsyncStorage.getItem(metaKey);

      if (metaStr) {
        const meta: CachedAsset = JSON.parse(metaStr);
        if (this.isCacheExpired(meta.timestamp)) {
          // Clean up expired asset
          await FileSystem.deleteAsync(localUri, { idempotent: true });
          await AsyncStorage.removeItem(metaKey);
          return null;
        }
      }

      return localUri;
    } catch (error) {
      console.error("Error getting local asset:", error);
      return null;
    }
  }

  private isCacheExpired(timestamp: number): boolean {
    const expiryTime = CACHE_EXPIRY_HOURS * 60 * 60 * 1000;
    return Date.now() - timestamp > expiryTime;
  }

  // Words caching methods with atomic operations
  public async cacheWords(
    wordlistId: number,
    words: Word[],
    originatingGeneration = this.cacheGeneration,
  ): Promise<void> {
    if (!this.isPremium) return;

    await this.trackCacheWrite(
      (generation) =>
        this.withErrorRecovery(async () => {
          const cacheKey = `${WORDS_CACHE_KEY}${wordlistId}`;
          const cachedData: CachedWords = {
            words,
            timestamp: Date.now(),
            wordlistId,
          };

          await this.atomicCacheOperation(async () => {
            await this.writeCacheItem(
              generation,
              cacheKey,
              JSON.stringify(cachedData),
            );
          });
        }, "Words caching"),
      originatingGeneration,
    );
  }

  // Wordlists caching for offline dashboard access
  public async cacheWordlists(
    wordlists: Wordlist[],
    originatingGeneration = this.cacheGeneration,
  ): Promise<void> {
    await this.trackCacheWrite(
      (generation) =>
        this.withErrorRecovery(async () => {
          const cachedData: CachedWordlists = {
            wordlists,
            timestamp: Date.now(),
          };

          await this.atomicCacheOperation(async () => {
            await this.writeCacheItem(
              generation,
              WORDLISTS_CACHE_KEY,
              JSON.stringify(cachedData),
            );
          });
        }, "Wordlists caching"),
      originatingGeneration,
    );
  }

  public async getCachedWordlists(
    originatingGeneration = this.cacheGeneration,
  ): Promise<Wordlist[] | null> {
    return this.guardCacheRead(originatingGeneration, () =>
      this.withErrorRecovery(async () => {
        const cachedDataStr = await AsyncStorage.getItem(WORDLISTS_CACHE_KEY);
        if (!cachedDataStr) return null;

        const cachedData: CachedWordlists = JSON.parse(cachedDataStr);

        if (this.isCacheExpired(cachedData.timestamp)) {
          await AsyncStorage.removeItem(WORDLISTS_CACHE_KEY);
          return null;
        }

        return cachedData.wordlists;
      }, "Cached wordlists retrieval"),
    ).catch((error) => this.cacheReadFallback<Wordlist[]>(error));
  }

  public async getCachedWords(
    wordlistId: number,
    originatingGeneration = this.cacheGeneration,
  ): Promise<Word[] | null> {
    this.assertCacheGeneration(originatingGeneration);
    if (!this.isPremium) return null;

    return this.guardCacheRead(originatingGeneration, () =>
      this.withErrorRecovery(async () => {
        const cacheKey = `${WORDS_CACHE_KEY}${wordlistId}`;
        const cachedDataStr = await AsyncStorage.getItem(cacheKey);

        if (!cachedDataStr) return null;

        const cachedData: CachedWords = JSON.parse(cachedDataStr);

        // Check if cache is expired
        if (this.isCacheExpired(cachedData.timestamp)) {
          await AsyncStorage.removeItem(cacheKey);
          return null;
        }

        return cachedData.words;
      }, "Cached words retrieval"),
    ).catch((error) => this.cacheReadFallback<Word[]>(error));
  }

  // Definitions caching methods with atomic operations
  public async cacheDefinitions(
    wordlistId: number,
    wordId: number,
    definitions: Definition[],
    originatingGeneration = this.cacheGeneration,
  ): Promise<void> {
    if (!this.isPremium) return;

    await this.trackCacheWrite(
      (generation) =>
        this.withErrorRecovery(async () => {
          const cacheKey = `${DEFINITIONS_CACHE_KEY}${wordlistId}-${wordId}`;
          const cachedData: CachedDefinitions = {
            definitions,
            timestamp: Date.now(),
            wordlistId,
            wordId,
          };

          await this.atomicCacheOperation(async () => {
            await this.writeCacheItem(
              generation,
              cacheKey,
              JSON.stringify(cachedData),
            );

            // Cache any images referenced in definitions
            await this.cacheDefinitionAssets(definitions, generation);
          }, [
            // Rollback: remove definitions if asset caching fails
            async () => AsyncStorage.removeItem(cacheKey),
          ]);
        }, "Definitions caching"),
      originatingGeneration,
    );
  }

  private async cacheDefinitionAssets(
    definitions: Definition[],
    generation: number,
  ): Promise<void> {
    const assetPromises: Promise<void>[] = [];

    if (!Array.isArray(definitions) || definitions.length === 0) {
      return;
    }

    for (const definition of definitions) {
      // Cache sounds if present
      if (definition.sounds && definition.sounds.length > 0) {
        for (const sound of definition.sounds) {
          if (sound.link) {
            assetPromises.push(
              this.cacheAsset(sound.link, "audio", generation),
            );
          }
        }
      }
    }

    await this.promiseAllSettledPolyfill(assetPromises);
  }

  public async getCachedDefinitions(
    wordlistId: number,
    wordId: number,
    originatingGeneration = this.cacheGeneration,
  ): Promise<Definition[] | null> {
    this.assertCacheGeneration(originatingGeneration);
    if (!this.isPremium) return null;

    return this.guardCacheRead(originatingGeneration, () =>
      this.withErrorRecovery(async () => {
        const cacheKey = `${DEFINITIONS_CACHE_KEY}${wordlistId}-${wordId}`;
        const cachedDataStr = await AsyncStorage.getItem(cacheKey);

        if (!cachedDataStr) return null;

        const cachedData: CachedDefinitions = JSON.parse(cachedDataStr);

        // Check if cache is expired
        if (this.isCacheExpired(cachedData.timestamp)) {
          await AsyncStorage.removeItem(cacheKey);
          return null;
        }

        // Validate and update asset URLs to local paths if offline
        const validatedDefinitions = await this.validateDefinitionAssets(
          cachedData.definitions,
        );

        return validatedDefinitions;
      }, "Cached definitions retrieval"),
    ).catch((error) => this.cacheReadFallback<Definition[]>(error));
  }

  private async validateDefinitionAssets(
    definitions: Definition[],
  ): Promise<Definition[]> {
    const validatedDefinitions: Definition[] = [];

    for (const definition of definitions) {
      const validatedDefinition = { ...definition };

      // Update sound URLs to local paths if available
      if (definition.sounds && definition.sounds.length > 0) {
        const validatedSounds = await Promise.all(
          definition.sounds.map(async (sound) => {
            if (sound.link) {
              const localUri = await this.getLocalAssetUri(sound.link);
              return {
                ...sound,
                link: localUri || sound.link, // Keep original if local not available
              };
            }
            return sound;
          }),
        );
        validatedDefinition.sounds = validatedSounds;
      }

      validatedDefinitions.push(validatedDefinition);
    }

    return validatedDefinitions;
  }

  // Get all cached definitions for a wordlist (useful for pre-loading)
  public async getAllCachedDefinitions(
    wordlistId: number,
  ): Promise<{ wordId: number; definitions: Definition[] }[]> {
    if (!this.isPremium) return [];

    const result: { wordId: number; definitions: Definition[] }[] = [];

    try {
      const allKeys = await AsyncStorage.getAllKeys();
      const definitionKeys = allKeys.filter((key) =>
        key.startsWith(`${DEFINITIONS_CACHE_KEY}${wordlistId}-`),
      );

      for (const cacheKey of definitionKeys) {
        const cachedDataStr = await AsyncStorage.getItem(cacheKey);
        if (!cachedDataStr) continue;

        const cachedData: CachedDefinitions = JSON.parse(cachedDataStr);

        if (!this.isCacheExpired(cachedData.timestamp)) {
          const validatedDefinitions = await this.validateDefinitionAssets(
            cachedData.definitions,
          );
          result.push({
            wordId: cachedData.wordId,
            definitions: validatedDefinitions,
          });
        } else {
          // Clean up expired definitions
          await AsyncStorage.removeItem(cacheKey);
        }
      }
    } catch (error) {
      console.error("Error retrieving all cached definitions:", error);
    }

    return result;
  }

  // Bulk caching methods for flash cards
  public async preloadWordlistForOffline(
    wordlistId: number,
    words: Word[],
    getDefinitions: (wordId: number) => Promise<Definition[]>,
    originatingGeneration = this.cacheGeneration,
  ): Promise<void> {
    if (!this.isPremium) return;

    try {
      // Cache the words list
      await this.cacheWords(wordlistId, words, originatingGeneration);

      // Cache definitions for each word (with some throttling to avoid overwhelming the API)
      for (let i = 0; i < words.length; i++) {
        const word = words[i];

        try {
          const definitions = await getDefinitions(word.id);
          await this.cacheDefinitions(
            wordlistId,
            word.id,
            definitions,
            originatingGeneration,
          );

          // Small delay to avoid overwhelming the server
          if (i < words.length - 1) {
            await new Promise((resolve) => setTimeout(resolve, 100));
          }
        } catch (error) {
          if (error instanceof OfflineCacheIdentityChangedError) throw error;
          console.error(
            `Error caching definitions for word ${word.name}:`,
            error,
          );
          // Continue with other words even if one fails
        }
      }
    } catch (error) {
      if (error instanceof OfflineCacheIdentityChangedError) throw error;
      console.error("Error preloading wordlist for offline:", error);
    }
  }

  // Check if a wordlist is fully cached for offline use
  public async isWordlistCachedForOffline(
    wordlistId: number,
  ): Promise<boolean> {
    if (!this.isPremium) return false;

    try {
      // Check if words are cached
      const cachedWords = await this.getCachedWords(wordlistId);
      if (!cachedWords || cachedWords.length === 0) return false;

      // Check if definitions are cached for all words
      const cachedDefinitions = await this.getAllCachedDefinitions(wordlistId);
      const cachedWordIds = new Set(
        cachedDefinitions.map((item) => item.wordId),
      );

      // All words should have cached definitions
      const allWordsCached = cachedWords.every((word) =>
        cachedWordIds.has(word.id),
      );

      return allWordsCached;
    } catch (error) {
      console.error("Error checking wordlist cache status:", error);
      return false;
    }
  }

  // Get cache statistics for a wordlist
  public async getWordlistCacheStats(wordlistId: number): Promise<{
    totalWords: number;
    cachedWords: number;
    cachedDefinitions: number;
    cachePercentage: number;
  }> {
    if (!this.isPremium) {
      return {
        totalWords: 0,
        cachedWords: 0,
        cachedDefinitions: 0,
        cachePercentage: 0,
      };
    }

    try {
      const cachedWords = await this.getCachedWords(wordlistId);
      const cachedDefinitions = await this.getAllCachedDefinitions(wordlistId);

      const totalWords = cachedWords?.length || 0;
      const cachedDefinitionsCount = cachedDefinitions.length;
      const cachePercentage =
        totalWords > 0 ? (cachedDefinitionsCount / totalWords) * 100 : 0;

      return {
        totalWords,
        cachedWords: totalWords,
        cachedDefinitions: cachedDefinitionsCount,
        cachePercentage: Math.round(cachePercentage),
      };
    } catch (error) {
      console.error("Error getting cache stats:", error);
      return {
        totalWords: 0,
        cachedWords: 0,
        cachedDefinitions: 0,
        cachePercentage: 0,
      };
    }
  }

  public async clearCache(): Promise<void> {
    this.cacheGeneration += 1;
    const pendingWrites = Array.from(this.inFlightCacheWrites);
    if (pendingWrites.length > 0) {
      await Promise.allSettled(pendingWrites);
    }
    let keys: readonly string[] = [];
    const failures: unknown[] = [];
    try {
      keys = await AsyncStorage.getAllKeys();
    } catch (error) {
      console.error("Error listing offline cache keys:", error);
      failures.push(error);
    }
    try {
      const cacheKeys = keys.filter((key) => key.startsWith(CACHE_PREFIX));
      await AsyncStorage.multiRemove(cacheKeys);
    } catch (error) {
      console.error("Error clearing offline cache storage:", error);
      failures.push(error);
    }
    try {
      await FileSystem.deleteAsync(ASSET_CACHE_DIR, { idempotent: true });
      await this.ensureAssetDirectory();
    } catch (error) {
      console.error("Error clearing offline cache assets:", error);
      failures.push(error);
    }
    if (failures.length > 0) {
      throw new AggregateError(failures, "Offline cache cleanup incomplete");
    }
  }

  public async getCacheSize(): Promise<number> {
    try {
      const dirInfo = await FileSystem.getInfoAsync(ASSET_CACHE_DIR);
      if (!dirInfo.exists) return 0;

      // This is a simplified version - you might want to recursively calculate
      // the actual size of all files in the directory
      return 0; // Placeholder
    } catch (error) {
      console.error("Error getting cache size:", error);
      return 0;
    }
  }

  public destroy(): void {
    // Clean up network listener
    if (this.netInfoUnsubscribe) {
      this.netInfoUnsubscribe();
      this.netInfoUnsubscribe = null;
    }

    // Clean up periodic connectivity timer
    if (this.connectivityTestTimer) {
      clearInterval(this.connectivityTestTimer);
      this.connectivityTestTimer = null;
    }

    // Clear caches
    this.connectivityTestCache.clear();
    this.networkListeners = [];

    console.log("OfflineManager destroyed and cleaned up");
  }

  // Method to get connectivity statistics for debugging
  public getConnectivityStats(): {
    consecutiveFailures: number;
    lastTest: ConnectivityTestResult | null;
    cacheSize: number;
    circuitBreakerState: CircuitBreakerState;
  } {
    return {
      consecutiveFailures: this.networkState.consecutiveFailures,
      lastTest: this.networkState.lastConnectivityTest,
      cacheSize: this.connectivityTestCache.size,
      circuitBreakerState: { ...this.circuitBreaker },
    };
  }

  // Atomic cache operations to prevent corruption
  private async atomicCacheOperation<T>(
    operation: () => Promise<T>,
    rollbackOperations: (() => Promise<void>)[] = [],
  ): Promise<T> {
    try {
      return await operation();
    } catch (error) {
      // Execute rollback operations in reverse order
      for (let i = rollbackOperations.length - 1; i >= 0; i--) {
        try {
          await rollbackOperations[i]();
        } catch (rollbackError) {
          console.error("Rollback operation failed:", rollbackError);
        }
      }
      throw error;
    }
  }

  // Enhanced error recovery with automatic retry
  private async withErrorRecovery<T>(
    operation: () => Promise<T>,
    context: string,
  ): Promise<T> {
    try {
      return await operation();
    } catch (error) {
      if (error instanceof OfflineCacheIdentityChangedError) throw error;
      console.warn(`${context} failed, attempting recovery:`, error);

      // Attempt to clean up corrupted cache entries
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      if (errorMessage.includes("JSON") || errorMessage.includes("parse")) {
        console.log("Detected cache corruption, cleaning up...");
        await this.cleanupCorruptedCache();
      }

      // Retry once after cleanup
      try {
        return await operation();
      } catch (retryError) {
        console.error(`${context} failed after recovery attempt:`, retryError);
        throw retryError;
      }
    }
  }

  private async cleanupCorruptedCache(): Promise<void> {
    try {
      const keys = await AsyncStorage.getAllKeys();
      const cacheKeys = keys.filter((key) => key.startsWith(CACHE_PREFIX));

      for (const key of cacheKeys) {
        try {
          const data = await AsyncStorage.getItem(key);
          if (data) {
            JSON.parse(data); // Test if it's valid JSON
          }
        } catch {
          console.log(`Removing corrupted cache entry: ${key}`);
          await AsyncStorage.removeItem(key);
        }
      }
    } catch (error) {
      console.error("Error during cache cleanup:", error);
    }
  }
}

export default new OfflineManager();
