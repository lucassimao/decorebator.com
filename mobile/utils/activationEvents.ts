export const ACTIVATION_EVENT_NAMES = {
  USER_SIGNED_UP: "user_signed_up",
  WORDLIST_CREATED: "wordlist_created",
  WORD_ADDED: "word_added",
  QUIZ_SESSION_STARTED: "quiz_session_started",
  QUIZ_SESSION_COMPLETED: "quiz_session_completed",
  QUIZ_ANSWERED: "quiz_answered",
  PRACTICE_CTA_OPENED: "practice_cta_opened",
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
};

type ActivationEventInput = Partial<ActivationEventProperties> &
  Record<string, unknown>;

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
}

export function captureActivationEvent(
  client: ActivationAnalyticsClient,
  event: ActivationEventName,
  input: ActivationEventInput = {},
  options: CaptureActivationEventOptions = {},
): void {
  const properties = createActivationEventProperties(input);
  const capturedEvent = { event, properties };

  if (options.dryRun) {
    options.onDryRun?.(capturedEvent);
    return;
  }

  void client.capture(event, properties);
}
