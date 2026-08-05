export type NotificationDestination =
  | { kind: "dashboard" }
  | { kind: "quiz"; wordlistId: string };

export function getNotificationDestination(
  data: Record<string, unknown> | undefined,
): NotificationDestination | null {
  if (data?.type === "daily_practice_reminder") {
    return { kind: "dashboard" };
  }

  if (data?.type !== "due_items_reminder") {
    return null;
  }

  const wordlistId =
    typeof data.wordlistId === "number"
      ? data.wordlistId
      : typeof data.wordlistId === "string" && /^\d+$/.test(data.wordlistId)
        ? Number(data.wordlistId)
        : Number.NaN;

  if (!Number.isSafeInteger(wordlistId) || wordlistId <= 0) {
    return null;
  }

  return { kind: "quiz", wordlistId: String(wordlistId) };
}

export function isAuthenticationPath(pathname: string): boolean {
  return (
    pathname === "/signin" ||
    pathname === "/signup" ||
    pathname === "/forgotPassword" ||
    pathname === "/onboarding" ||
    pathname.startsWith("/onboarding/")
  );
}
