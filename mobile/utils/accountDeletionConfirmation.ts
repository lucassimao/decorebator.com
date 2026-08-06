type ConfirmationButton = {
  text: string;
  style?: "default" | "cancel" | "destructive";
  onPress?: () => void;
};

type AlertFunction = (
  title: string,
  message?: string,
  buttons?: ConfirmationButton[],
) => void;

type ShowAccountDeletionConfirmationOptions = {
  alert: AlertFunction;
  onConfirm: () => void;
  t: (key: string) => string;
};

export function showAccountDeletionConfirmation({
  alert,
  onConfirm,
  t,
}: ShowAccountDeletionConfirmationOptions) {
  alert(
    t("settings.account.deleteAccount"),
    t("settings.account.deleteWarning"),
    [
      { text: t("common.cancel"), style: "cancel" },
      {
        text: t("profile.continueToDeletion"),
        style: "destructive",
        onPress: () => {
          alert(t("profile.confirmDeletion"), t("profile.finalDeleteWarning"), [
            { text: t("common.cancel"), style: "cancel" },
            {
              text: t("profile.deleteNow"),
              style: "destructive",
              onPress: onConfirm,
            },
          ]);
        },
      },
    ],
  );
}
