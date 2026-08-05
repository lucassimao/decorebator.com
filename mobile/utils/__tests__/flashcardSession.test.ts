import {
  flashcardPositionKey,
  flashcardSavePositionKey,
  resetSavedFlashcardPosition,
  resolveFlashcardStartIndex,
  visitFlashcard,
} from "../flashcardSession";

describe("flashcard session state", () => {
  it("owns both persistence keys", () => {
    expect(flashcardPositionKey("42")).toBe("flashcard_position_42");
    expect(flashcardSavePositionKey("42")).toBe("flashcard_save_position_42");
  });
  it.each([
    ["true", "4", 12, 4],
    ["false", "4", 12, 0],
    [null, "4", 12, 0],
    ["true", null, 12, 0],
    ["true", "-1", 12, 0],
    ["true", "12", 12, 0],
    ["true", "1.5", 12, 0],
    ["true", "not-a-number", 12, 0],
  ])(
    "resolves save=%p index=%p total=%p to %p",
    (save, index, total, expected) => {
      expect(resolveFlashcardStartIndex(save, index, total)).toBe(expected);
    },
  );

  it("counts unique viewed indices without mutating the previous set", () => {
    const initial = new Set([8]);
    const revisited = visitFlashcard(initial, 8);
    const advanced = visitFlashcard(revisited, 9);

    expect(initial).toEqual(new Set([8]));
    expect(revisited).toEqual(new Set([8]));
    expect(advanced).toEqual(new Set([8, 9]));
  });

  it("resets persisted position only when save-position is enabled", async () => {
    const storage = { setItem: jest.fn(() => Promise.resolve()) };

    await resetSavedFlashcardPosition(storage, "42", false);
    expect(storage.setItem).not.toHaveBeenCalled();

    await resetSavedFlashcardPosition(storage, "42", true);
    expect(storage.setItem).toHaveBeenCalledWith("flashcard_position_42", "0");
  });
});
