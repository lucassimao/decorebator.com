interface CompletePasswordChangeOptions {
  clearCredentials: () => Promise<void>;
  presentSuccess: (onConfirm: () => void) => void;
  resetForm: () => void;
  close: () => void;
  redirectToSignIn: () => void;
  clearInMemoryState: () => void;
}

export async function completePasswordChange({
  clearCredentials,
  presentSuccess,
  resetForm,
  close,
  redirectToSignIn,
  clearInMemoryState,
}: CompletePasswordChangeOptions): Promise<void> {
  await clearCredentials();
  clearInMemoryState();
  presentSuccess(() => {
    resetForm();
    close();
    redirectToSignIn();
  });
}
