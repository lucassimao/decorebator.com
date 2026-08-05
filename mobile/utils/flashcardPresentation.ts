export type FlashcardLoadState =
  | "loading"
  | "processing"
  | "offline-unavailable"
  | "error"
  | "empty"
  | "ready";

interface ResolveFlashcardLoadStateInput {
  isLoading: boolean;
  isLoadingPosition: boolean;
  isRetrying: boolean;
  isFetching: boolean;
  error: Error | null | undefined;
  words: readonly unknown[] | undefined;
  isOnline: boolean;
  isOfflineAvailable: boolean;
}

export function isFlashcardProcessingError(
  error: Error | null | undefined,
): boolean {
  const message = error?.message.toLowerCase() ?? "";
  return (
    message.includes("no words found") || message.includes("no definitions")
  );
}

export function resolveFlashcardLoadState({
  isLoading,
  isLoadingPosition,
  isRetrying,
  isFetching,
  error,
  words,
  isOnline,
  isOfflineAvailable,
}: ResolveFlashcardLoadStateInput): FlashcardLoadState {
  if (isLoading || isLoadingPosition || (isRetrying && isFetching)) {
    return "loading";
  }
  if (isFlashcardProcessingError(error)) return "processing";
  if (error && !isOnline && !isOfflineAvailable) {
    return "offline-unavailable";
  }
  if (error) return "error";
  if (!words || words.length === 0) return "empty";
  return "ready";
}

export function flashcardSettleOffset(direction: "next" | "prev"): number {
  return direction === "next" ? 12 : -12;
}

export function boundFlashcardIndex(index: number, wordCount: number): number {
  if (wordCount <= 0) return 0;
  return Math.min(Math.max(index, 0), wordCount - 1);
}
