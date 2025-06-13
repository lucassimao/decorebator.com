# State Management Patterns & Architecture

## Overview

The Decorebator mobile app implements a sophisticated multi-layered state management architecture that optimizes for performance, user experience, and maintainability. This document details the state management patterns, caching strategies, and data flow architectures used throughout the application.

## Multi-Layered State Architecture

### 1. Server State (React Query)

#### Global Configuration

```typescript
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        if (error?.message?.includes('timeout')) return false;
        return isOnline ? failureCount < 2 : false;
      },
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      staleTime: 5 * 60 * 1000, // 5 minutes default
      gcTime: 30 * 60 * 1000,   // 30 minutes garbage collection
    },
  },
});
```

#### Tier-Based Caching Strategy

```typescript
export function useAnalytics(wordlistId: number, isPremium: boolean) {
  // Premium users get fresher data and faster updates
  const staleTime = isPremium ? 10 * 1000 : 15 * 60 * 1000;
  const gcTime = isPremium ? 2 * 60 * 1000 : 60 * 60 * 1000;
  
  return useQuery({
    queryKey: ["analytics", wordlistId, isPremium],
    queryFn: () => fetchAnalytics(wordlistId),
    staleTime,
    gcTime,
    refetchOnWindowFocus: isPremium, // Real-time updates for premium
  });
}
```

#### Cache Invalidation Strategies

```typescript
export function useInvalidateAnalytics() {
  const queryClient = useQueryClient();
  
  const invalidateAllAnalytics = useCallback((wordlistId: number) => {
    // Premium users get immediate cache invalidation
    queryClient.invalidateQueries({
      predicate: (query) => {
        const [resource, id, isPremium] = query.queryKey;
        return resource === "analytics" && 
               id === wordlistId && 
               isPremium === true;
      },
    });
  }, [queryClient]);
  
  return { invalidateAllAnalytics };
}
```

### 2. Local State Management

#### Context Providers for Global State

```typescript
// Snackbar Provider - Global notification system
export function SnackbarProvider({ children }: { children: React.ReactNode }) {
  const [snackbar, setSnackbar] = useState<SnackbarState | null>(null);
  
  const showSnackbar = useCallback((message: string, type: SnackbarType = 'info') => {
    setSnackbar({ message, type, id: Date.now() });
  }, []);
  
  return (
    <SnackbarContext.Provider value={{ showSnackbar }}>
      {children}
      {snackbar && (
        <SnackBar
          message={snackbar.message}
          type={snackbar.type}
          onDismiss={() => setSnackbar(null)}
        />
      )}
    </SnackbarContext.Provider>
  );
}

// Upgrade Prompt Provider - Premium feature gating
export function UpgradePromptDialogProvider({ children }: PropsWithChildren) {
  const [isVisible, setIsVisible] = useState(false);
  const [context, setContext] = useState<UpgradeContext>('general');
  
  const showUpgradePrompt = useCallback((context: UpgradeContext = 'general') => {
    setContext(context);
    setIsVisible(true);
  }, []);
  
  return (
    <UpgradePromptDialogContext.Provider value={{ showUpgradePrompt }}>
      {children}
      <UpgradePromptDialog 
        visible={isVisible}
        context={context}
        onClose={() => setIsVisible(false)}
      />
    </UpgradePromptDialogContext.Provider>
  );
}
```

#### Custom Hooks for Encapsulated Logic

```typescript
// Error Reporting Hook - Centralized error handling
export function useErrorReporting({
  wordId,
  definitionId,
  isOnline,
  context,
  onSuccess,
}: UseErrorReportingProps) {
  const [showReportModal, setShowReportModal] = useState(false);
  const [isReporting, setIsReporting] = useState(false);
  
  const handleReportError = useCallback(async (errorType: ErrorType, description?: string) => {
    if (!isOnline) {
      showSnackbar(t("errors.offlineErrorReporting"), "error");
      return;
    }
    
    setIsReporting(true);
    try {
      await reportError({
        wordId,
        definitionId,
        errorType,
        description,
        context,
      });
      
      showSnackbar(t("errors.reportSuccess"), "success");
      setShowReportModal(false);
      onSuccess?.();
    } catch (error) {
      handleErrorReportingError(error);
    } finally {
      setIsReporting(false);
    }
  }, [wordId, definitionId, isOnline, context, onSuccess]);
  
  return {
    showReportModal,
    isReporting,
    handleReportError,
    openReportModal: () => setShowReportModal(true),
    closeReportModal: () => setShowReportModal(false),
  };
}

// Offline Hook - Network state management
export function useOffline() {
  const [isOnline, setIsOnline] = useState(true);
  const [isOfflineAvailable, setIsOfflineAvailable] = useState(false);
  const { data: userInfo } = useUserInfo();
  
  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener(state => {
      setIsOnline(state.isConnected ?? false);
    });
    
    return unsubscribe;
  }, []);
  
  useEffect(() => {
    // Only premium users have offline access
    setIsOfflineAvailable(userInfo?.subscriptionPlan !== 'free');
  }, [userInfo?.subscriptionPlan]);
  
  return { isOnline, isOfflineAvailable };
}
```

### 3. Form State Management

#### React Hook Form with Zod Validation

```typescript
// Type-safe form handling with validation
const schema = z.object({
  email: z.string().email("Invalid email format"),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

type FormData = z.infer<typeof schema>;

export function LoginForm() {
  const { control, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: "",
      password: "",
    },
  });
  
  const onSubmit = async (data: FormData) => {
    try {
      await loginUser(data);
    } catch (error) {
      // Handle login error
    }
  };
  
  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Controller
        name="email"
        control={control}
        render={({ field }) => (
          <TextInput
            {...field}
            placeholder="Email"
            error={!!errors.email}
            helperText={errors.email?.message}
          />
        )}
      />
    </form>
  );
}
```

### 4. Persistent State Management

#### Secure Storage Strategy

```typescript
// JWT Token Management
class TokenManager {
  private static readonly TOKEN_KEY = 'jwt_token';
  private static readonly REFRESH_KEY = 'refresh_token';
  
  static async storeTokens(token: string, refreshToken: string): Promise<void> {
    await Promise.all([
      SecureStore.setItemAsync(this.TOKEN_KEY, token),
      SecureStore.setItemAsync(this.REFRESH_KEY, refreshToken),
    ]);
  }
  
  static async getToken(): Promise<string | null> {
    return await SecureStore.getItemAsync(this.TOKEN_KEY);
  }
  
  static async clearTokens(): Promise<void> {
    await Promise.all([
      SecureStore.deleteItemAsync(this.TOKEN_KEY),
      SecureStore.deleteItemAsync(this.REFRESH_KEY),
    ]);
  }
}

// Preferences Management
class PreferencesManager {
  private static readonly LANGUAGE_KEY = 'preferred_language';
  private static readonly FLASHCARD_POSITION_KEY = 'flashcard_save_position';
  
  static async setLanguage(language: string): Promise<void> {
    await AsyncStorage.setItem(this.LANGUAGE_KEY, language);
  }
  
  static async getLanguage(): Promise<string | null> {
    return await AsyncStorage.getItem(this.LANGUAGE_KEY);
  }
}
```

## Advanced State Patterns

### 1. Optimistic Updates

```typescript
// Quiz answer with optimistic update
const answerMutation = useMutation({
  mutationFn: ({ success }: { success: boolean }) =>
    offlineQuizApi.answerQuiz({
      definitionID: quiz.definitionId,
      isCorrect: success,
      responseTimeMs: Date.now() - quizDisplayedAtRef.current,
    }),
  onMutate: async ({ success }) => {
    // Optimistically update UI
    setQuizCount(prev => prev + 1);
    if (success) {
      setCorrectCount(prev => prev + 1);
    }
  },
  onSuccess: () => {
    // Invalidate analytics for real-time updates
    invalidateAllAnalytics(wordlistId);
  },
  onError: (error, variables, context) => {
    // Rollback optimistic update
    setQuizCount(prev => prev - 1);
    if (variables.success) {
      setCorrectCount(prev => prev - 1);
    }
  },
});
```

### 2. Background Sync Pattern

```typescript
// Offline-first data synchronization
export class OfflineManager {
  private static async syncWhenOnline(): Promise<void> {
    const pendingData = await this.getPendingSync();
    
    if (pendingData.length > 0) {
      try {
        await Promise.all(
          pendingData.map(data => this.syncDataToServer(data))
        );
        await this.clearPendingSync();
      } catch (error) {
        console.error('Sync failed:', error);
      }
    }
  }
  
  static async handleNetworkStateChange(isOnline: boolean): Promise<void> {
    if (isOnline) {
      await this.syncWhenOnline();
    }
  }
}
```

### 3. Cache-First Strategy

```typescript
// Offline-aware API calls
export async function getWords(
  wordlistId: number, 
  onlyWithDefinitions: boolean = false
): Promise<Word[]> {
  const { isOnline, isOfflineAvailable } = useOffline();
  
  if (!isOnline && isOfflineAvailable) {
    // Try offline cache first
    const cachedWords = await OfflineManager.getCachedWords(wordlistId);
    if (cachedWords) {
      return onlyWithDefinitions 
        ? cachedWords.filter(word => word.hasDefinitions)
        : cachedWords;
    }
  }
  
  if (!isOnline && !isOfflineAvailable) {
    throw new Error('Offline access requires premium subscription');
  }
  
  // Fetch from API and cache
  const words = await callAPI<Word[]>('GET', `/wordlists/${wordlistId}/words`);
  
  if (isOfflineAvailable) {
    await OfflineManager.cacheWords(wordlistId, words);
  }
  
  return words;
}
```

## State Synchronization Patterns

### 1. Real-Time Updates for Premium Users

```typescript
export function useRealtimeAnalytics(wordlistId: number) {
  const { isPremium } = useUserInfo();
  const queryClient = useQueryClient();
  
  useEffect(() => {
    if (!isPremium) return;
    
    // Premium users get real-time updates every 10 seconds
    const interval = setInterval(() => {
      queryClient.invalidateQueries({
        queryKey: ["analytics", wordlistId],
      });
    }, 10000);
    
    return () => clearInterval(interval);
  }, [isPremium, wordlistId, queryClient]);
}
```

### 2. Cross-Screen State Coordination

```typescript
// Shared wordlist progress hook to avoid duplicate API calls
export function useWordlistProgress() {
  const { data: userInfo } = useUserInfo();
  const isPremium = userInfo?.subscriptionPlan !== 'free';
  
  return useQuery({
    queryKey: ["wordlist-progress", isPremium],
    queryFn: () => api.getProgressSummary(),
    staleTime: isPremium ? 10 * 1000 : 15 * 60 * 1000,
    // Shared across Dashboard and Analytics screens
    refetchOnMount: false,
    refetchOnWindowFocus: isPremium,
  });
}
```

### 3. Authentication State Flow

```typescript
// Automatic session management
export function useAuthFlow() {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
  const queryClient = useQueryClient();
  
  useEffect(() => {
    const validateSession = async () => {
      try {
        const token = await TokenManager.getToken();
        if (!token) {
          setIsAuthenticated(false);
          return;
        }
        
        // Validate token with server
        await api.validateToken(token);
        setIsAuthenticated(true);
      } catch (error) {
        // Token invalid, clear storage
        await TokenManager.clearTokens();
        queryClient.clear();
        setIsAuthenticated(false);
      }
    };
    
    validateSession();
  }, []);
  
  // Refresh session on app focus
  useFocusEffect(
    useCallback(() => {
      if (isAuthenticated) {
        validateSession();
      }
    }, [isAuthenticated])
  );
  
  return { isAuthenticated };
}
```

## Performance Optimization Patterns

### 1. Query Deduplication

```typescript
// Automatic request deduplication with React Query
export function useAnalyticsData(wordlistId: number) {
  // Multiple components can call this hook
  // React Query automatically deduplicates requests
  return useQueries({
    queries: [
      {
        queryKey: ["stats", wordlistId],
        queryFn: () => api.getStats(wordlistId),
      },
      {
        queryKey: ["progress", wordlistId], 
        queryFn: () => api.getProgress(wordlistId),
      },
      {
        queryKey: ["mastery", wordlistId],
        queryFn: () => api.getMastery(wordlistId),
      },
    ],
  });
}
```

### 2. Selective Re-renders

```typescript
// Memoized components to prevent unnecessary re-renders
export const WordlistItem = React.memo(({ wordlist, onPress }: Props) => {
  const handlePress = useCallback(() => {
    onPress(wordlist.id);
  }, [wordlist.id, onPress]);
  
  return (
    <TouchableOpacity onPress={handlePress}>
      <Text>{wordlist.name}</Text>
    </TouchableOpacity>
  );
});
```

### 3. Background State Updates

```typescript
// Update state without blocking UI
export function useBackgroundDataRefresh() {
  const queryClient = useQueryClient();
  
  const refreshInBackground = useCallback(async () => {
    // Prefetch data in background
    await queryClient.prefetchQuery({
      queryKey: ["analytics"],
      queryFn: api.getAnalytics,
      staleTime: 0, // Force refresh
    });
  }, [queryClient]);
  
  // Refresh when app comes to foreground
  useFocusEffect(
    useCallback(() => {
      refreshInBackground();
    }, [refreshInBackground])
  );
}
```

## Error State Management

### 1. Global Error Boundary

```typescript
export function GlobalErrorBoundary({ children }: { children: React.ReactNode }) {
  return (
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={(error, errorInfo) => {
        // Log to error reporting service
        console.error('Global error:', error, errorInfo);
      }}
    >
      {children}
    </ErrorBoundary>
  );
}

function ErrorFallback({ error, resetErrorBoundary }: ErrorFallbackProps) {
  return (
    <View style={styles.errorContainer}>
      <Text>Something went wrong:</Text>
      <Text>{error.message}</Text>
      <Button onPress={resetErrorBoundary} title="Try again" />
    </View>
  );
}
```

### 2. Network Error Recovery

```typescript
export function useNetworkErrorRecovery() {
  const { isOnline } = useOffline();
  const queryClient = useQueryClient();
  
  useEffect(() => {
    if (isOnline) {
      // Retry failed queries when network comes back
      queryClient.getQueryCache().getAll().forEach(query => {
        if (query.state.error) {
          query.fetch();
        }
      });
    }
  }, [isOnline, queryClient]);
}
```

This comprehensive state management architecture ensures optimal performance, offline capability, and excellent user experience while maintaining code clarity and maintainability across the entire application.