import {
  boundFlashcardIndex,
  flashcardSettleOffset,
  resolveFlashcardLoadState,
} from "../flashcardPresentation";

const readyWords = [{ id: 1 }];

describe("flashcard presentation state", () => {
  it("keeps loading and retry-in-flight ahead of terminal states", () => {
    expect(
      resolveFlashcardLoadState({
        isLoading: true,
        isLoadingPosition: false,
        isRetrying: false,
        isFetching: false,
        error: new Error("network"),
        words: undefined,
        isOnline: true,
        isOfflineAvailable: false,
      }),
    ).toBe("loading");

    expect(
      resolveFlashcardLoadState({
        isLoading: false,
        isLoadingPosition: false,
        isRetrying: true,
        isFetching: true,
        error: new Error("network"),
        words: undefined,
        isOnline: true,
        isOfflineAvailable: false,
      }),
    ).toBe("loading");
  });

  it.each([
    ["no definitions are ready", "processing"],
    ["no words found while processing", "processing"],
    ["request failed", "error"],
  ] as const)("maps %s to %s", (message, expected) => {
    expect(
      resolveFlashcardLoadState({
        isLoading: false,
        isLoadingPosition: false,
        isRetrying: false,
        isFetching: false,
        error: new Error(message),
        words: undefined,
        isOnline: true,
        isOfflineAvailable: false,
      }),
    ).toBe(expected);
  });

  it("separates unavailable offline, empty, and ready states", () => {
    const base = {
      isLoading: false,
      isLoadingPosition: false,
      isRetrying: false,
      isFetching: false,
      isOnline: false,
      isOfflineAvailable: false,
    };

    expect(
      resolveFlashcardLoadState({
        ...base,
        error: new Error("network"),
        words: undefined,
      }),
    ).toBe("offline-unavailable");
    expect(resolveFlashcardLoadState({ ...base, error: null, words: [] })).toBe(
      "empty",
    );
    expect(
      resolveFlashcardLoadState({
        ...base,
        error: null,
        words: readyWords,
      }),
    ).toBe("ready");
  });

  it("settles new content from the navigation direction", () => {
    expect(flashcardSettleOffset("next")).toBe(12);
    expect(flashcardSettleOffset("prev")).toBe(-12);
  });

  it("synchronously bounds an index when a refetch shortens the wordlist", () => {
    expect(boundFlashcardIndex(7, 3)).toBe(2);
    expect(boundFlashcardIndex(-1, 3)).toBe(0);
    expect(boundFlashcardIndex(4, 0)).toBe(0);
  });
});
