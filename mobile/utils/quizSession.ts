export interface QuizSubmissionLock {
  current: boolean;
}

export function claimQuizSubmission(lock: QuizSubmissionLock): boolean {
  if (lock.current) return false;
  lock.current = true;
  return true;
}

export function resetQuizSubmission(lock: QuizSubmissionLock): void {
  lock.current = false;
}

export function prepareQuizTypeSelection(
  lock: QuizSubmissionLock,
  selectedTypes: readonly unknown[],
): boolean {
  if (selectedTypes.length === 0) return false;
  resetQuizSubmission(lock);
  return true;
}
