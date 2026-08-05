import {
  getNotificationDestination,
  isAuthenticationPath,
} from "../notificationRouting";

describe("notification routing", () => {
  it("maps the bounded reminder catalog to owned routes", () => {
    expect(
      getNotificationDestination({ type: "daily_practice_reminder" }),
    ).toEqual({ kind: "dashboard" });
    expect(
      getNotificationDestination({
        type: "due_items_reminder",
        wordlistId: 42,
      }),
    ).toEqual({ kind: "quiz", wordlistId: "42" });
    expect(
      getNotificationDestination({
        type: "due_items_reminder",
        wordlistId: "7",
      }),
    ).toEqual({ kind: "quiz", wordlistId: "7" });
  });

  it.each([undefined, {}, { type: "unknown" }])(
    "ignores unknown payload %p",
    (payload) => expect(getNotificationDestination(payload)).toBeNull(),
  );

  it.each([0, -1, 1.5, Number.MAX_SAFE_INTEGER + 1, "", "1.5", "1/quiz"])(
    "rejects unsafe wordlist id %p",
    (wordlistId) =>
      expect(
        getNotificationDestination({
          type: "due_items_reminder",
          wordlistId,
        }),
      ).toBeNull(),
  );

  it("recognizes routes where protected notification navigation must stop", () => {
    expect(isAuthenticationPath("/signin")).toBe(true);
    expect(isAuthenticationPath("/onboarding/features")).toBe(true);
    expect(isAuthenticationPath("/dashboard")).toBe(false);
    expect(isAuthenticationPath("/quiz")).toBe(false);
  });
});
