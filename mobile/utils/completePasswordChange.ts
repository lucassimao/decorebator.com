interface CompletePasswordChangeOptions {
  clearCredentials: () => Promise<void>;
  presentSuccess: (onConfirm: () => void) => void;
  resetForm: () => void;
  close: () => void;
  redirectToSignIn: () => void;
}

export async function completePasswordChange({
  clearCredentials,
  presentSuccess,
  resetForm,
  close,
  redirectToSignIn,
}: CompletePasswordChangeOptions): Promise<void> {
  await clearCredentials();
  presentSuccess(() => {
    resetForm();
    close();
    redirectToSignIn();
  });
}
