interface FlashcardStorage {
  setItem(key: string, value: string): Promise<void>;
}

export const flashcardPositionKey = (wordlistId: string) =>
  `flashcard_position_${wordlistId}`;

export const flashcardSavePositionKey = (wordlistId: string) =>
  `flashcard_save_position_${wordlistId}`;

export function resolveFlashcardStartIndex(
  savePositionValue: string | null,
  savedIndexValue: string | null,
  totalWords: number,
): number {
  if (savePositionValue !== "true" || savedIndexValue === null) return 0;

  const savedIndex = Number(savedIndexValue);
  return Number.isSafeInteger(savedIndex) &&
    savedIndex >= 0 &&
    savedIndex < totalWords
    ? savedIndex
    : 0;
}

export function visitFlashcard(
  visitedIndices: Set<number>,
  index: number,
): Set<number> {
  const next = new Set(visitedIndices);
  next.add(index);
  return next;
}

export async function resetSavedFlashcardPosition(
  storage: FlashcardStorage,
  wordlistId: string,
  savePosition: boolean,
): Promise<void> {
  if (!savePosition) return;
  await storage.setItem(flashcardPositionKey(wordlistId), "0");
}
