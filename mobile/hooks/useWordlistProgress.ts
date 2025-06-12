import { useQuery } from "@tanstack/react-query";
import { getProgressSummary, ProgressSummaryResponse } from "@/api/analytics";
import { useUserInfo } from "@/hooks/users";

export function useWordlistProgress() {
  const { userInfo, isPremium } = useUserInfo();

  const userId = userInfo?.id;

  // Similar cache strategy as useAnalytics for consistency
  const staleTime = isPremium ? 10 * 1000 : 15 * 60 * 1000; // 10s vs 15min
  const gcTime = isPremium ? 2 * 60 * 1000 : 60 * 60 * 1000; // 2min vs 1hr

  return useQuery<ProgressSummaryResponse>({
    queryKey: ["analytics", "progress-summary", userId],
    queryFn: getProgressSummary,
    staleTime,
    gcTime,
    refetchOnWindowFocus: isPremium, // Premium users get automatic refresh when returning to screen
    refetchOnMount: isPremium ? ("always" as const) : false, // Premium users always get fresh data on mount
    enabled: Boolean(userId),
  });
}
