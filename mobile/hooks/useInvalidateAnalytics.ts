import { useQueryClient } from "@tanstack/react-query";

/**
 * Hook to invalidate box distribution cache after quiz sessions
 * Note: Other analytics metrics are computed hourly via materialized views
 * so they don't need cache invalidation
 */
export function useInvalidateAnalytics() {
  const queryClient = useQueryClient();

  const invalidateBoxDistribution = (wordlistId: number) => {
    queryClient.invalidateQueries({
      queryKey: ["analytics", "boxDistribution", wordlistId],
    });
  };

  return {
    invalidateBoxDistribution,
  };
}