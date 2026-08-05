export const ACTIVATION_EVENT_NAMES = {
  USER_SIGNED_UP: "user_signed_up",
  WORDLIST_CREATED: "wordlist_created",
  WORD_ADDED: "word_added",
  QUIZ_SESSION_STARTED: "quiz_session_started",
  QUIZ_SESSION_COMPLETED: "quiz_session_completed",
  QUIZ_ANSWERED: "quiz_answered",
  PRACTICE_CTA_OPENED: "practice_cta_opened",
  USER_SIGNED_IN: "user_signed_in",
  ONBOARDING_STARTED: "onboarding_started",
  ONBOARDING_SKIPPED: "onboarding_skipped",
  ONBOARDING_FEATURE_VIEWED: "onboarding_feature_viewed",
  ONBOARDING_COMPLETED: "onboarding_completed",
  PAYWALL_IMPRESSION: "paywall_impression",
  PAYWALL_PLAN_SELECTED: "paywall_plan_selected",
  PURCHASE_PENDING: "purchase_pending",
  PURCHASE_SUCCEEDED: "purchase_succeeded",
  PURCHASE_FAILED: "purchase_failed",
  NOTIFICATION_OPENED: "notification_opened",
  RESTORE_COMPLETED: "restore_completed",
} as const;

export type ActivationEventName =
  (typeof ACTIVATION_EVENT_NAMES)[keyof typeof ACTIVATION_EVENT_NAMES];

type ActivationPropertyValue = boolean | number | string;

export type ActivationEventProperties = {
  eventVersion: 1;
  source?: string;
  entryPoint?: string;
  appVersion?: string;
  platform?: "ios" | "android" | "web";
  userId?: number;
  wordlistId?: number;
  language?: string;
  sessionId?: string;
  quizMode?: string;
  quizType?: string;
  correct?: boolean;
  answeredCount?: number;
  correctCount?: number;
  wordCount?: number;
  durationMs?: number;
  responseTimeMs?: number;
  outcome?: "success" | "failure" | "cancelled";
  errorCode?: string;
  step?: string;
  slide?: string;
  destination?: "signup";
  store?: "apple" | "google";
  productId?: string;
  billingPeriod?: "monthly" | "annual";
  notificationType?:
    | "daily_practice_reminder"
    | "due_items_reminder"
    | "unknown";
  restoreStatus?:
    | "restored"
    | "partial"
    | "no_purchases"
    | "pending"
    | "failed";
};

type ActivationEventInput = Partial<ActivationEventProperties> &
  Record<string, unknown>;

const REQUIRED_PROPERTIES: Record<
  ActivationEventName,
  readonly (keyof ActivationEventProperties)[]
> = {
  user_signed_up: ["source"],
  wordlist_created: ["wordlistId", "language", "source"],
  word_added: ["wordlistId", "wordCount", "source"],
  quiz_session_started: ["sessionId", "wordlistId", "source"],
  quiz_session_completed: [
    "sessionId",
    "wordlistId",
    "answeredCount",
    "correctCount",
    "outcome",
  ],
  quiz_answered: ["sessionId", "wordlistId", "correct", "responseTimeMs"],
  practice_cta_opened: ["source"],
  user_signed_in: ["source"],
  onboarding_started: ["source"],
  onboarding_skipped: ["step"],
  onboarding_feature_viewed: ["slide"],
  onboarding_completed: ["destination"],
  paywall_impression: ["source", "store"],
  paywall_plan_selected: ["source", "store", "billingPeriod"],
  purchase_pending: ["store", "productId", "billingPeriod"],
  purchase_succeeded: ["store", "productId", "billingPeriod"],
  purchase_failed: ["store", "productId", "errorCode"],
  notification_opened: ["source", "notificationType"],
  restore_completed: ["store", "restoreStatus"],
};

const ALLOWED_PROPERTIES = new Set<keyof ActivationEventProperties>([
  "eventVersion",
  "source",
  "entryPoint",
  "appVersion",
  "platform",
  "userId",
  "wordlistId",
  "language",
  "sessionId",
  "quizMode",
  "quizType",
  "correct",
  "answeredCount",
  "correctCount",
  "wordCount",
  "durationMs",
  "responseTimeMs",
  "outcome",
  "errorCode",
  "step",
  "slide",
  "destination",
  "store",
  "productId",
  "billingPeriod",
  "notificationType",
  "restoreStatus",
]);

export function createActivationEventProperties(
  input: ActivationEventInput = {},
): ActivationEventProperties {
  const properties: Record<string, ActivationPropertyValue> = {
    eventVersion: 1,
  };

  for (const key of ALLOWED_PROPERTIES) {
    if (key === "eventVersion" || input[key] === undefined) {
      continue;
    }

    const value = input[key];
    if (
      typeof value === "string" ||
      typeof value === "number" ||
      typeof value === "boolean"
    ) {
      properties[key] = value;
    }
  }

  return properties as ActivationEventProperties;
}

export function validateActivationEvent(
  event: ActivationEventName,
  properties: ActivationEventProperties,
): string[] {
  return REQUIRED_PROPERTIES[event].filter(
    (property) => properties[property] === undefined,
  );
}

export interface ActivationAnalyticsClient {
  capture: (
    event: ActivationEventName,
    properties: ActivationEventProperties,
  ) => void;
}

export interface DryRunActivationEvent {
  event: ActivationEventName;
  properties: ActivationEventProperties;
}

export interface CaptureActivationEventOptions {
  dryRun?: boolean;
  onDryRun?: (event: DryRunActivationEvent) => void;
  dedupeKey?: string;
  /** Caller-owned process-scoped store; this is not durable delivery tracking. */
  dedupeStore?: Set<string>;
  onInvalid?: (
    event: ActivationEventName,
    properties: ActivationEventProperties,
    missingProperties: string[],
  ) => void;
}

export function createActivationSessionId(): string {
  return `quiz-session-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}

export type NotificationAnalyticsType = NonNullable<
  ActivationEventProperties["notificationType"]
>;

export function normalizeNotificationType(
  value: unknown,
): NotificationAnalyticsType {
  if (value === "daily_practice_reminder" || value === "due_items_reminder") {
    return value;
  }
  return "unknown";
}

export interface AnalyticsIdentityClient {
  identify: (distinctId: string) => void;
  reset: () => void;
}

export function identifyAnalyticsUser(
  client: AnalyticsIdentityClient,
  userId: number | string,
): boolean {
  const normalized = String(userId);
  if (!/^\d+$/.test(normalized) || Number(normalized) <= 0) {
    return false;
  }
  try {
    client.identify(normalized);
    return true;
  } catch {
    return false;
  }
}

export function resetAnalyticsIdentity(client: AnalyticsIdentityClient): void {
  try {
    client.reset();
  } catch {
    // Analytics cleanup must never block local authentication cleanup.
  }
}

const analyticsIdentityResetListeners = new Set<() => void>();

export function subscribeAnalyticsIdentityReset(
  listener: () => void,
): () => void {
  analyticsIdentityResetListeners.add(listener);
  return () => analyticsIdentityResetListeners.delete(listener);
}

export function requestAnalyticsIdentityReset(): void {
  for (const listener of analyticsIdentityResetListeners) {
    try {
      listener();
    } catch {
      // Authentication cleanup cannot depend on analytics availability.
    }
  }
}

export function captureActivationEvent(
  client: ActivationAnalyticsClient,
  event: ActivationEventName,
  input: ActivationEventInput = {},
  options: CaptureActivationEventOptions = {},
): boolean {
  const properties = createActivationEventProperties(input);
  const missingProperties = validateActivationEvent(event, properties);
  if (missingProperties.length > 0) {
    options.onInvalid?.(event, properties, missingProperties);
    return false;
  }

  if (options.dedupeKey && options.dedupeStore?.has(options.dedupeKey)) {
    return false;
  }

  const capturedEvent = { event, properties };

  if (options.dryRun) {
    options.onDryRun?.(capturedEvent);
    return true;
  }

  try {
    void client.capture(event, properties);
    if (options.dedupeKey && options.dedupeStore) {
      options.dedupeStore.add(options.dedupeKey);
    }
    return true;
  } catch {
    return false;
  }
}
