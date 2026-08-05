import {
  ACTIVATION_EVENT_NAMES,
  captureActivationEvent,
  createActivationEventProperties,
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
      { sessionId: "session-1", correct: true },
      { dryRun: true, onDryRun: (event) => dryRunEvents.push(event) },
    );

    expect(posthog.capture).not.toHaveBeenCalled();
    expect(dryRunEvents).toEqual([
      {
        event: "quiz_answered",
        properties: {
          eventVersion: 1,
          sessionId: "session-1",
          correct: true,
        },
      },
    ]);
  });
});
