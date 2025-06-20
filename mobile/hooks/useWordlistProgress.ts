import { useQuery } from "@tanstack/react-query";
import { getProgressSummary, ProgressSummaryResponse } from "@/api/analytics";
import { useUserInfo } from "@/hooks/users";

export function useWordlistProgress() {
  const { userInfo, cacheConfig } = useUserInfo();

  const userId = userInfo?.id;

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
