import AsyncStorage from '@react-native-async-storage/async-storage';
import NetInfo from '@react-native-community/netinfo';
import * as FileSystem from 'expo-file-system';
import { Quiz, Word } from '@/api/wordlists';

const CACHE_PREFIX = 'decorebator_offline_';
const QUIZ_CACHE_KEY = `${CACHE_PREFIX}quiz_`;
const WORDS_CACHE_KEY = `${CACHE_PREFIX}words_`;
const ASSET_CACHE_DIR = `${FileSystem.documentDirectory}decorebator_assets/`;
const CACHE_EXPIRY_HOURS = 72; // 3 days

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

class OfflineManager {
  private isOnline: boolean = true;
  private isPremium: boolean = false;
  private networkListeners: ((isOnline: boolean) => void)[] = [];

  constructor() {
    this.initializeNetworkListener();
    this.ensureAssetDirectory();
  }

  private async ensureAssetDirectory() {
    const dirInfo = await FileSystem.getInfoAsync(ASSET_CACHE_DIR);
    if (!dirInfo.exists) {
      await FileSystem.makeDirectoryAsync(ASSET_CACHE_DIR, { intermediates: true });
    }
  }

  private initializeNetworkListener() {
    NetInfo.addEventListener((state) => {
      const wasOnline = this.isOnline;
      this.isOnline = state.isConnected ?? false;
      
      if (wasOnline !== this.isOnline) {
        this.networkListeners.forEach(listener => listener(this.isOnline));
      }
    });
  }

  public subscribeToNetworkChanges(listener: (isOnline: boolean) => void) {
    this.networkListeners.push(listener);
    return () => {
      this.networkListeners = this.networkListeners.filter(l => l !== listener);
    };
  }

  public setUserPremiumStatus(isPremium: boolean) {
    this.isPremium = isPremium;
  }

  public isOfflineAvailable(): boolean {
    return this.isPremium && !this.isOnline;
  }

  public getNetworkStatus(): boolean {
    return this.isOnline;
  }

  // Quiz caching methods
  public async cacheQuiz(wordlistId: number, quiz: Quiz): Promise<void> {
    if (!this.isPremium) return;

    const cacheKey = `${QUIZ_CACHE_KEY}${wordlistId}-${quiz.id}`;
    const cachedData: CachedQuiz = {
      quiz,
      timestamp: Date.now(),
      wordlistId,
    };

    try {
      // Cache the quiz data
      await AsyncStorage.setItem(cacheKey, JSON.stringify(cachedData));

      // Cache associated assets
      await this.cacheQuizAssets(quiz);
    } catch (error) {
      console.error('Error caching quiz:', error);
    }
  }

  private async cacheQuizAssets(quiz: Quiz): Promise<void> {
    const assetPromises: Promise<void>[] = [];

    // Cache image for WORD_FROM_IMAGE quiz type (value contains image URL)
    if (quiz.type === 'WORD_FROM_IMAGE' && quiz.value) {
      assetPromises.push(this.cacheAsset(quiz.value, 'image'));
    }

    // Cache audio URL if present
    if (quiz.audioURL) {
      assetPromises.push(this.cacheAsset(quiz.audioURL, 'audio'));
    }

    await Promise.allSettled(assetPromises);
  }

  private async cacheAsset(url: string, type: 'image' | 'audio'): Promise<void> {
    try {
      const filename = this.getFilenameFromUrl(url);
      const localUri = `${ASSET_CACHE_DIR}${filename}`;
      
      // Check if already cached
      const fileInfo = await FileSystem.getInfoAsync(localUri);
      if (fileInfo.exists) {
        // Update timestamp
        const metaKey = `${CACHE_PREFIX}asset_meta_${filename}`;
        await AsyncStorage.setItem(metaKey, JSON.stringify({
          localUri,
          timestamp: Date.now(),
        }));
        return;
      }

      // Download and cache
      const downloadResult = await FileSystem.downloadAsync(url, localUri);
      
      if (downloadResult.status === 200) {
        // Store metadata
        const metaKey = `${CACHE_PREFIX}asset_meta_${filename}`;
        await AsyncStorage.setItem(metaKey, JSON.stringify({
          localUri,
          timestamp: Date.now(),
        }));
      }
    } catch (error) {
      console.error(`Error caching ${type} asset:`, error);
    }
  }

  private getFilenameFromUrl(url: string): string {
    const urlParts = url.split('/');
    const filename = urlParts[urlParts.length - 1].split('?')[0];
    return filename || `cached_${Date.now()}`;
  }

  public async getCachedQuiz(wordlistId: number): Promise<Quiz | null> {
    if (!this.isPremium || this.isOnline) return null;

    try {
      // Get all keys from AsyncStorage
      const allKeys = await AsyncStorage.getAllKeys();
      
      // Find keys that match the pattern for this wordlist
      const wordlistQuizKeys = allKeys.filter(key => 
        key.startsWith(`${QUIZ_CACHE_KEY}${wordlistId}-`)
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
    } catch (error) {
      console.error('Error retrieving cached quiz:', error);
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
      const wordlistQuizKeys = allKeys.filter(key => 
        key.startsWith(`${QUIZ_CACHE_KEY}${wordlistId}-`)
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
      console.error('Error retrieving all cached quizzes:', error);
    }
    
    return quizzes;
  }
  
  // Clean up expired quizzes (can be called periodically)
  public async cleanupExpiredQuizzes(): Promise<void> {
    try {
      const allKeys = await AsyncStorage.getAllKeys();
      const quizKeys = allKeys.filter(key => key.startsWith(QUIZ_CACHE_KEY));
      
      for (const cacheKey of quizKeys) {
        const cachedDataStr = await AsyncStorage.getItem(cacheKey);
        if (!cachedDataStr) continue;
        
        const cachedData: CachedQuiz = JSON.parse(cachedDataStr);
        
        if (this.isCacheExpired(cachedData.timestamp)) {
          await AsyncStorage.removeItem(cacheKey);
        }
      }
    } catch (error) {
      console.error('Error cleaning up expired quizzes:', error);
    }
  }

  private async validateQuizAssets(quiz: Quiz): Promise<Quiz | null> {
    const validatedQuiz = { ...quiz };
    let isValid = true;

    // Check image availability for WORD_FROM_IMAGE quiz
    if (quiz.type === 'WORD_FROM_IMAGE' && quiz.value) {
      const localUri = await this.getLocalAssetUri(quiz.value);
      if (localUri) {
        validatedQuiz.value = localUri;
      } else {
        isValid = false;
      }
    }

    // Check audio availability
    if (quiz.audioURL) {
      const localUri = await this.getLocalAssetUri(quiz.audioURL);
      if (localUri) {
        validatedQuiz.audioURL = localUri;
      } else {
        isValid = false;
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
      console.error('Error getting local asset:', error);
      return null;
    }
  }

  private isCacheExpired(timestamp: number): boolean {
    const expiryTime = CACHE_EXPIRY_HOURS * 60 * 60 * 1000;
    return Date.now() - timestamp > expiryTime;
  }
  
  // Words caching methods
  public async cacheWords(wordlistId: number, words: Word[]): Promise<void> {
    if (!this.isPremium) return;
    
    const cacheKey = `${WORDS_CACHE_KEY}${wordlistId}`;
    const cachedData: CachedWords = {
      words,
      timestamp: Date.now(),
      wordlistId,
    };
    
    try {
      await AsyncStorage.setItem(cacheKey, JSON.stringify(cachedData));
    } catch (error) {
      console.error('Error caching words:', error);
    }
  }
  
  public async getCachedWords(wordlistId: number): Promise<Word[] | null> {
    if (!this.isPremium) return null;
    
    try {
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
    } catch (error) {
      console.error('Error retrieving cached words:', error);
      return null;
    }
  }

  public async clearCache(): Promise<void> {
    try {
      // Clear quiz cache
      const keys = await AsyncStorage.getAllKeys();
      const cacheKeys = keys.filter(key => key.startsWith(CACHE_PREFIX));
      await AsyncStorage.multiRemove(cacheKeys);

      // Clear asset files
      await FileSystem.deleteAsync(ASSET_CACHE_DIR, { idempotent: true });
      await this.ensureAssetDirectory();
    } catch (error) {
      console.error('Error clearing cache:', error);
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
      console.error('Error getting cache size:', error);
      return 0;
    }
  }
}

export default new OfflineManager();