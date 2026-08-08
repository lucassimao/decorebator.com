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
    const clearInMemoryState = jest.fn();
    let confirm: (() => void) | undefined;

    await completePasswordChange({
      clearCredentials,
      presentSuccess: (onConfirm) => {
        confirm = onConfirm;
      },
      resetForm,
      close,
      redirectToSignIn,
      clearInMemoryState,
    });

    expect(deleteAccess).toHaveBeenCalledTimes(1);
    expect(deleteRefresh).toHaveBeenCalledTimes(1);
    expect(confirm).toBeDefined();
    expect(clearInMemoryState).toHaveBeenCalledTimes(1);
    confirm?.();
    expect(resetForm).toHaveBeenCalledTimes(1);
    expect(close).toHaveBeenCalledTimes(1);
    expect(redirectToSignIn).toHaveBeenCalledTimes(1);
  });
});
