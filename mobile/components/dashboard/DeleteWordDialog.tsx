import * as wordlistsApi from "@/api/wordlists";
import { useMutation } from "@tanstack/react-query";
import * as React from "react";
import { StyleSheet } from "react-native";
import {
  ActivityIndicator,
  Button,
  Dialog,
  Portal,
  Snackbar,
  Text,
  useTheme,
} from "react-native-paper";

type Props = {
  onDismiss: (success: boolean) => void;
  word: wordlistsApi.Word;
};

const DeleteWordDialog = ({ onDismiss, word }: Props) => {
  const theme = useTheme();

  const [error, setError] = React.useState<any>(null);

  const { mutate: deleteWord, isPending } = useMutation<void, Error>({
    mutationFn: () => wordlistsApi.deleteWord(word),
    onError: (error) => {
      setError(error);
    },
    onSuccess: () => {
      onDismiss(true);
    },
  });

  const cancel = () => onDismiss(false);

  return (
    <Portal>
      <Snackbar
        visible={!!error}
        onDismiss={() => setError(null)}
        action={{
          label: "Hide",
        }}
      >
        {error?.message}
      </Snackbar>

      <Dialog
        visible={true}
        onDismiss={cancel}
        dismissable
        dismissableBackButton
      >
        <Dialog.Title style={styles.title}>Delete Word</Dialog.Title>
        <Dialog.Content>
          {isPending ? (
            <ActivityIndicator size={"large"} animating={true} theme={theme} />
          ) : (
            <Text>Are your sure? {word.name} will be removed altogether</Text>
          )}
        </Dialog.Content>
        <Dialog.Actions>
          <Button onPress={cancel}>Cancel</Button>
          <Button onPress={() => deleteWord()}>Delete</Button>
        </Dialog.Actions>
      </Dialog>
    </Portal>
  );
};

const styles = StyleSheet.create({
  title: {
    textAlign: "center",
  },
});

export default DeleteWordDialog;
