import { showAccountDeletionConfirmation } from "../accountDeletionConfirmation";
import en from "../../i18n/locales/en.json";

function translate(key: string): string {
  let value: unknown = en;

  for (const part of key.split(".")) {
    if (typeof value !== "object" || value === null || !(part in value)) {
      return key;
    }
    value = (value as Record<string, unknown>)[part];
  }

  return typeof value === "string" ? value : key;
}

describe("showAccountDeletionConfirmation", () => {
  it("warns about store cancellation and requires a truthful final confirmation", () => {
    const alert = jest.fn();
    const onConfirm = jest.fn();
    showAccountDeletionConfirmation({
      alert,
      onConfirm,
      t: translate,
    });

    expect(alert).toHaveBeenCalledTimes(1);
    expect(alert.mock.calls[0][1]).toContain("Apple or Google subscription");
    expect(alert.mock.calls[0][1]).toContain("store first");

    const firstButtons = alert.mock.calls[0][2];
    firstButtons[1].onPress();

    expect(alert).toHaveBeenCalledTimes(2);
    expect(alert.mock.calls[1][1]).toBe(
      "Delete your account now? This cannot be undone.",
    );
    expect(alert.mock.calls[1][1]).not.toContain("type");
    expect(alert.mock.calls[1][1]).not.toContain('DELETE"');

    const finalButtons = alert.mock.calls[1][2];
    expect(finalButtons[1].text).toBe(en.profile.deleteNow);
    finalButtons[1].onPress();
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });
});
