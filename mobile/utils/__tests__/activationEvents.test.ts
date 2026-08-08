import {
  ACTIVATION_EVENT_NAMES,
  captureActivationEvent,
  createActivationSessionId,
  createActivationEventProperties,
  identifyAnalyticsUser,
  normalizeNotificationType,
  requestAnalyticsIdentityReset,
  resetAnalyticsIdentity,
  subscribeAnalyticsIdentityReset,
  validateActivationEvent,
  type ActivationEventName,
} from "@/utils/activationEvents";

describe("activation analytics contract", () => {
  it("exposes one canonical name for every activation event", () => {
    expect(Object.values(ACTIVATION_EVENT_NAMES)).toEqual([
      "user_signed_up",
      "wordlist_created",
      "word_added",
      "quiz_session_started",
      "quiz_session_completed",
      "quiz_answered",
      "practice_cta_opened",
      "user_signed_in",
      "onboarding_started",
      "onboarding_skipped",
      "onboarding_feature_viewed",
      "onboarding_completed",
      "paywall_impression",
      "paywall_plan_selected",
      "purchase_pending",
      "purchase_succeeded",
      "purchase_failed",
      "notification_opened",
      "restore_completed",
    ]);
  });

  it("adds the contract version and keeps only privacy-safe properties", () => {
    const properties = createActivationEventProperties({
      source: "signup_screen",
      appVersion: "1.1.1",
      userId: 42,
      wordlistId: 7,
      email: "secret@example.com",
      wordlistName: "Private words",
      word: "secret",
    });

    expect(properties).toEqual({
      eventVersion: 1,
      source: "signup_screen",
      appVersion: "1.1.1",
      userId: 42,
      wordlistId: 7,
    });
  });

  it("supports a dry-run sink without calling PostHog", () => {
    const posthog = { capture: jest.fn() };
    const dryRunEvents: {
      event: ActivationEventName;
      properties: Record<string, unknown>;
    }[] = [];

    captureActivationEvent(
      posthog,
      ACTIVATION_EVENT_NAMES.QUIZ_ANSWERED,
      {
        sessionId: "session-1",
        wordlistId: 7,
        correct: true,
        responseTimeMs: 100,
      },
      { dryRun: true, onDryRun: (event) => dryRunEvents.push(event) },
    );

    expect(posthog.capture).not.toHaveBeenCalled();
    expect(dryRunEvents).toEqual([
      {
        event: "quiz_answered",
        properties: {
          eventVersion: 1,
          sessionId: "session-1",
          wordlistId: 7,
          correct: true,
          responseTimeMs: 100,
        },
      },
    ]);
  });

  it("reports missing event-specific properties before capture", () => {
    expect(
      validateActivationEvent(ACTIVATION_EVENT_NAMES.WORD_ADDED, {
        eventVersion: 1,
        wordlistId: 7,
      }),
    ).toEqual(["wordCount", "source"]);
  });

  it("suppresses a repeated event when the caller supplies a dedupe key", () => {
    const posthog = { capture: jest.fn() };
    const dedupeStore = new Set<string>();
    const event = ACTIVATION_EVENT_NAMES.WORDLIST_CREATED;
    const properties = { wordlistId: 7, language: "es", source: "test" };

    expect(
      captureActivationEvent(posthog, event, properties, {
        dedupeKey: "wordlist-created-7",
        dedupeStore,
      }),
    ).toBe(true);
    expect(
      captureActivationEvent(posthog, event, properties, {
        dedupeKey: "wordlist-created-7",
        dedupeStore,
      }),
    ).toBe(false);
    expect(posthog.capture).toHaveBeenCalledTimes(1);
  });

  it("does not consume a dedupe key during dry-run", () => {
    const posthog = { capture: jest.fn() };
    const dedupeStore = new Set<string>();

    expect(
      captureActivationEvent(
        posthog,
        ACTIVATION_EVENT_NAMES.WORDLIST_CREATED,
        { wordlistId: 7, language: "es", source: "test" },
        { dryRun: true, dedupeKey: "wordlist-created-7", dedupeStore },
      ),
    ).toBe(true);
    expect(dedupeStore).toEqual(new Set());
    expect(posthog.capture).not.toHaveBeenCalled();
  });

  it("creates a non-empty session identifier for one learning session", () => {
    expect(createActivationSessionId()).toMatch(/^quiz-session-/);
  });

  it("requires privacy-safe fields for commerce and notification outcomes", () => {
    const properties = createActivationEventProperties({
      store: "apple",
      productId: "premium.monthly",
      billingPeriod: "monthly",
      errorCode: "provider_unavailable",
      receipt: "sensitive-receipt",
      purchaseToken: "sensitive-token",
    });

    expect(properties).toEqual({
      eventVersion: 1,
      store: "apple",
      productId: "premium.monthly",
      billingPeriod: "monthly",
      errorCode: "provider_unavailable",
    });
    expect(
      validateActivationEvent(
        ACTIVATION_EVENT_NAMES.PURCHASE_FAILED,
        properties,
      ),
    ).toEqual([]);
  });

  it("normalizes notification types to a bounded catalog", () => {
    expect(normalizeNotificationType("daily_practice_reminder")).toBe(
      "daily_practice_reminder",
    );
    expect(normalizeNotificationType("due_items_reminder")).toBe(
      "due_items_reminder",
    );
    expect(normalizeNotificationType("private-user-content")).toBe("unknown");
    expect(normalizeNotificationType(undefined)).toBe("unknown");
  });

  it("identifies by opaque server id without user properties and resets on sign out", () => {
    const posthog = { identify: jest.fn(), reset: jest.fn() };

    expect(identifyAnalyticsUser(posthog, 42)).toBe(true);
    expect(identifyAnalyticsUser(posthog, "invalid")).toBe(false);
    resetAnalyticsIdentity(posthog);

    expect(posthog.identify).toHaveBeenCalledWith("42");
    expect(posthog.identify).toHaveBeenCalledTimes(1);
    expect(posthog.reset).toHaveBeenCalledTimes(1);
  });

  it("delivers mounted resets and latches a reset while the bridge is absent", () => {
    const listener = jest.fn();
    const unsubscribe = subscribeAnalyticsIdentityReset(listener);

    requestAnalyticsIdentityReset();
    unsubscribe();
    requestAnalyticsIdentityReset();
    const replacementListener = jest.fn();
    const unsubscribeReplacement =
      subscribeAnalyticsIdentityReset(replacementListener);
    unsubscribeReplacement();

    expect(listener).toHaveBeenCalledTimes(1);
    expect(replacementListener).toHaveBeenCalledTimes(1);
  });
});
