import { completePasswordChange } from "../completePasswordChange";

describe("completePasswordChange", () => {
  it("presents confirmation and redirects after credential cleanup handles a storage rejection", async () => {
    const deleteAccess = jest.fn(() =>
      Promise.reject(new Error("secure store unavailable")),
    );
    const deleteRefresh = jest.fn(() => Promise.resolve());
    const clearCredentials = jest.fn(async () => {
      await Promise.allSettled([deleteAccess(), deleteRefresh()]);
    });
    const resetForm = jest.fn();
    const close = jest.fn();
    const redirectToSignIn = jest.fn();
    let confirm: (() => void) | undefined;

    await completePasswordChange({
      clearCredentials,
      presentSuccess: (onConfirm) => {
        confirm = onConfirm;
      },
      resetForm,
      close,
      redirectToSignIn,
    });

    expect(deleteAccess).toHaveBeenCalledTimes(1);
    expect(deleteRefresh).toHaveBeenCalledTimes(1);
    expect(confirm).toBeDefined();
    confirm?.();
    expect(resetForm).toHaveBeenCalledTimes(1);
    expect(close).toHaveBeenCalledTimes(1);
    expect(redirectToSignIn).toHaveBeenCalledTimes(1);
  });
});
