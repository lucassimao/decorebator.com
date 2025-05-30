import React from "react";
import { StyleSheet } from "react-native";
import { Button, Dialog, Portal, Text, useTheme } from "react-native-paper";
import { router } from "expo-router";

type Props = {
  visible: boolean;
  onDismiss: () => void;
};

const UpgradePromptDialog = ({ visible, onDismiss }: Props) => {
  const theme = useTheme();

  const handleUpgrade = () => {
    onDismiss();
    router.push({ pathname: "/subscription" });
  };

  return (
    <Portal>
      <Dialog visible={visible} onDismiss={onDismiss}>
        <Dialog.Icon icon="crown" size={48} />
        <Dialog.Title style={styles.title}>Enjoying Decorebator?</Dialog.Title>
        <Dialog.Content>
          <Text variant="bodyMedium" style={styles.content}>
            You've reached the free plan limit of 1 wordlist.
          </Text>
          <Text variant="bodyMedium" style={styles.content}>
            Upgrade to Premium to create unlimited wordlists and unlock all
            learning features!
          </Text>
        </Dialog.Content>
        <Dialog.Actions>
          <Button onPress={onDismiss}>Maybe Later</Button>
          <Button mode="contained" onPress={handleUpgrade}>
            View Plans
          </Button>
        </Dialog.Actions>
      </Dialog>
    </Portal>
  );
};

const styles = StyleSheet.create({
  title: {
    textAlign: "center",
  },
  content: {
    textAlign: "center",
    marginBottom: 10,
  },
});

export default UpgradePromptDialog;
