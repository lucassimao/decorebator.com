import {
  claimQuizSubmission,
  prepareQuizTypeSelection,
  resetQuizSubmission,
} from "@/utils/quizSession";

describe("quiz session submission lock", () => {
  it("allows exactly one answer or skip for a displayed quiz", () => {
    const lock = { current: false };

    expect(claimQuizSubmission(lock)).toBe(true);
    expect(claimQuizSubmission(lock)).toBe(false);
    expect(lock.current).toBe(true);
  });

  it("unlocks only when the next quiz is displayed", () => {
    const lock = { current: true };

    resetQuizSubmission(lock);

    expect(claimQuizSubmission(lock)).toBe(true);
  });

  it("unlocks when an applied quiz-type selection replaces the question", () => {
    const lock = { current: true };

    expect(prepareQuizTypeSelection(lock, ["GUESS_MEANING"])).toBe(true);
    expect(claimQuizSubmission(lock)).toBe(true);
  });

  it("does not unlock for an invalid empty quiz-type selection", () => {
    const lock = { current: true };

    expect(prepareQuizTypeSelection(lock, [])).toBe(false);
    expect(lock.current).toBe(true);
  });
});
