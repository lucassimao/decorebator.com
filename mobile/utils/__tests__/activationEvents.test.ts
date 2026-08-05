import {
  ACTIVATION_EVENT_NAMES,
  captureActivationEvent,
  createActivationSessionId,
  createActivationEventProperties,
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
});
