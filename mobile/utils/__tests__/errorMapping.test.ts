import { mapErrorToI18n } from "../errorMapping";

describe("password policy error mapping", () => {
  it.each([
    ["password must contain at least 8 characters", "errors.shortPassword"],
    ["password must contain at most 72 UTF-8 bytes", "errors.longPassword"],
  ])("maps %s to %s", (message, i18nKey) => {
    expect(mapErrorToI18n(message)).toEqual({
      i18nKey,
      isFieldError: true,
      field: "password",
      originalMessage: message,
    });
  });
});
