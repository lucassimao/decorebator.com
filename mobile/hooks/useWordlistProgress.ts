import { useQuery } from "@tanstack/react-query";
import { getProgressSummary, ProgressSummaryResponse } from "@/api/analytics";
import { useUserSession } from "@/hooks/useUserSession";

export function useWordlistProgress() {
  const { userId, cacheConfig } = useUserSession();

  return useQuery<ProgressSummaryResponse>({
    queryKey: ["analytics", "progress-summary", userId],
    queryFn: getProgressSummary,
    staleTime: cacheConfig.dataFreshnessDuration,
    gcTime: cacheConfig.memoryRetentionTime,
    refetchOnWindowFocus: cacheConfig.autoRefreshOnFocus,
    refetchOnMount: cacheConfig.alwaysFetchOnMount,
    enabled: Boolean(userId),
  });
}
