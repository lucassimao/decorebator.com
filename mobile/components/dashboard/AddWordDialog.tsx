import * as wordlistsApi from "@/api/wordlists";
import { useMutation } from "@tanstack/react-query";
import * as React from "react";
import { Controller, useForm } from "react-hook-form";
import { Keyboard, StyleSheet } from "react-native";
import {
  ActivityIndicator,
  Button,
  Dialog,
  HelperText,
  Portal,
  Snackbar,
  TextInput,
  useTheme,
} from "react-native-paper";

type Props = {
  onWordAdded: (dismiss: boolean) => void;
  onDismiss: () => void;
  wordlistId: number;
};

const AddWordDialog = ({ onWordAdded, wordlistId, onDismiss }: Props) => {
  const theme = useTheme();

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm({
    defaultValues: {
      name: "",
    },
  });

  const [error, setError] = React.useState<any>(null);
  const [closeAfterSubmit, setCloseAfterSubmit] = React.useState(true);

  const { mutate: addWord, isPending } = useMutation<
    void,
    Error,
    { name: string }
  >({
    mutationFn: ({ name }) => wordlistsApi.addWord({ wordlistId, name }),
    onError: (error) => {
      setError(error);
      Keyboard.dismiss();
    },
    onSuccess: () => {
      if (closeAfterSubmit) {
        onWordAdded(true);
      } else {
        onWordAdded(false);
        reset();
      }
    },
  });

  const onSubmitAndClose = (data: any) => {
    setCloseAfterSubmit(true);
    addWord({ name: data.name });
  };
  const onSubmitAndAddMore = (data: any) => {
    setCloseAfterSubmit(false);
    addWord({ name: data.name });
  };

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
        onDismiss={onDismiss}
        dismissable
        dismissableBackButton
      >
        <Dialog.Title style={styles.title}>Add Word</Dialog.Title>
        <Dialog.Content>
          {isPending ? (
            <ActivityIndicator size={"large"} animating={true} theme={theme} />
          ) : (
            <Controller
              control={control}
              rules={{
                required: true,
              }}
              render={({ field: { onChange, onBlur, value } }) => (
                <TextInput
                  label="Word"
                  mode="outlined"
                  // style={styles.inputs}
                  onBlur={onBlur}
                  onChangeText={onChange}
                  value={value}
                  error={!!errors.name}
                  autoFocus={error == null}
                />
              )}
              name="name"
            />
          )}
          {errors.name && <HelperText type="error">Required</HelperText>}
        </Dialog.Content>
        <Dialog.Actions>
          <Button onPress={onDismiss}>Cancel</Button>
          <Button onPress={handleSubmit(onSubmitAndClose)}>Save & Close</Button>
          <Button onPress={handleSubmit(onSubmitAndAddMore)}>
            Save & Add More
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
});

export default AddWordDialog;
