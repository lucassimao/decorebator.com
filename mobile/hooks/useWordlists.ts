import * as wordlistsApi from "@/api/wordlists";
import { QueryClient, useQuery } from "@tanstack/react-query";

/**
 * Hook to fetch and manage user's wordlists
 */
export const useWordlists = () => {
  return useQuery<wordlistsApi.Wordlist[], Error>({
    queryFn: () => wordlistsApi.getUserWordlists(),
    queryKey: ["wordlists"],
    staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
    gcTime: 10 * 60 * 1000, // Keep in cache for 10 minutes (formerly cacheTime)
  });
};

/**
 * Prefetch wordlists data into React Query cache
 * This is non-blocking and will populate the cache for when the dashboard loads
 */
export const prefetchWordlists = async (queryClient: QueryClient) => {
  await queryClient.prefetchQuery({
    queryKey: ["wordlists"],
    queryFn: () => wordlistsApi.getUserWordlists(),
    staleTime: 5 * 60 * 1000, // Same stale time as the hook
  });
};
