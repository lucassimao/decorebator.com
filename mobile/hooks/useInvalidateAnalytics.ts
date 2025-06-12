import { useQueryClient } from "@tanstack/react-query";

/**
 * Hook to invalidate analytics cache after quiz sessions
 * All analytics are now real-time and should be refreshed after quiz completion
 */
export function useInvalidateAnalytics() {
  const queryClient = useQueryClient();

  const invalidateBoxDistribution = (wordlistId: number) => {
    queryClient.invalidateQueries({
      queryKey: ["analytics", "boxDistribution", wordlistId],
    });
  };

  const invalidateAllAnalytics = (wordlistId: number) => {
    // Invalidate all analytics for the specific wordlist
    // Use predicate function to match partial keys (React Query v5 pattern)
    queryClient.invalidateQueries({
      predicate: (query) => {
        const key = query.queryKey;
        return (
          Array.isArray(key) &&
          key.length >= 3 &&
          key[0] === "analytics" &&
          key[2] === wordlistId
        );
      },
    });
    
    // Also invalidate global progress summary since it aggregates all wordlists
    queryClient.invalidateQueries({
      queryKey: ["analytics", "progress-summary"],
    });
  };

  return {
    invalidateBoxDistribution, // Keep for backward compatibility
    invalidateAllAnalytics,
  };
}