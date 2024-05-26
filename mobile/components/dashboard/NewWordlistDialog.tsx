import * as wordlistsApi from "@/api/wordlists";
import { useMutation } from "@tanstack/react-query";
import * as React from "react";
import { Controller, useForm } from "react-hook-form";
import { StyleSheet } from "react-native";
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
  // success is true if word was created
  onDismiss: (success?: boolean) => void;
};

type CreateWordlistDTO = wordlistsApi.CreateWordlistDTO;
const NewWordlistDialog = ({ onDismiss }: Props) => {
  const theme = useTheme();

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<wordlistsApi.CreateWordlistDTO>({
    defaultValues: {
      name: "",
      description: "",
    },
  });

  const [error, setError] = React.useState<any>(null);

  const { mutate: createWordlist, isPending } = useMutation<
    void,
    Error,
    CreateWordlistDTO
  >({
    mutationFn: ({ name, description }) =>
      wordlistsApi.addWordlist({ name, description }),
    onError: (error) => {
      setError(error);
    },
    onSuccess: () => {
      onDismiss(true);
    },
  });

  const onSubmit = (data: CreateWordlistDTO) => createWordlist(data);

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
        <Dialog.Title style={styles.title}>New Wordlist</Dialog.Title>
        <Dialog.Content>
          {isPending ? (
            <ActivityIndicator size={"large"} animating={true} theme={theme} />
          ) : (
            <>
              <Controller
                control={control}
                rules={{
                  required: true,
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    label="Name"
                    mode="outlined"
                    style={styles.inputs}
                    onBlur={onBlur}
                    onChangeText={onChange}
                    value={value}
                    error={!!errors.name}
                    autoFocus={true}
                  />
                )}
                name="name"
              />
              {errors.name && <HelperText type="error">Required</HelperText>}

              <Controller
                control={control}
                rules={{
                  required: false,
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    label="Description"
                    mode="outlined"
                    onBlur={onBlur}
                    onChangeText={onChange}
                    value={value}
                    error={!!errors.description}
                  />
                )}
                name="description"
              />
              <HelperText type="info">Optional</HelperText>
            </>
          )}
        </Dialog.Content>
        <Dialog.Actions>
          <Button onPress={() => onDismiss(false)}>Cancel</Button>
          <Button onPress={handleSubmit(onSubmit)}>Add</Button>
        </Dialog.Actions>
      </Dialog>
    </Portal>
  );
};

const styles = StyleSheet.create({
  title: {
    textAlign: "center",
  },
  inputs: {
    marginBottom: 10,
  },
});

export default NewWordlistDialog;
