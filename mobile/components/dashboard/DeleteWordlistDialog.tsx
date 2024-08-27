import * as wordlistsApi from "@/api/wordlists";
import { useMutation, useQuery } from "@tanstack/react-query";
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
  wordlist: wordlistsApi.Wordlist;
};

const DeleteWordlistDialog = ({ onDismiss, wordlist }: Props) => {
  const theme = useTheme();

  const [error, setError] = React.useState<any>(null);

  const { mutate: deleteWordlist, isPending } = useMutation<void, Error>({
    mutationFn: () => wordlistsApi.deleteWordlist(wordlist.id),
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
        <Dialog.Title style={styles.title}>Delete wordlist</Dialog.Title>
        <Dialog.Content>
          {isPending ? (
            <ActivityIndicator size={"large"} animating={true} theme={theme} />
          ) : (
            <Text>
              Are your sure? {wordlist.name} will be removed altogether,
              including the words
            </Text>
          )}
        </Dialog.Content>
        <Dialog.Actions>
          <Button onPress={cancel}>Cancel</Button>
          <Button onPress={() => deleteWordlist()}>Delete</Button>
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

export default DeleteWordlistDialog;
