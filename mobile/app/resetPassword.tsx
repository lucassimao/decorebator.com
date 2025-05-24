import * as usersApi from "@/api/users";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { router } from "expo-router";
import * as React from "react";
import { Controller, useForm } from "react-hook-form";
import { Keyboard, StyleSheet, Text, View } from "react-native";
import { Button, HelperText, TextInput, useTheme } from "react-native-paper";
import z from "zod";
import Snackbar from "@/components/SnackBar";

const schema = z
  .object({
    email: z.string().email(),
  })
  .required();

export default function ResetPasswordScreen() {
  const theme = useTheme();

  const {
    mutate: requestResetEmailPassword,
    status,
    reset,
    error,
    isPending
  } = useMutation<void, Error, string>({
    mutationFn: usersApi.requestResetEmailPassword,
  });

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(schema),
    defaultValues: {
      email: "",
    },
  });

  const onSubmit = (data: any) => {
    Keyboard.dismiss();
    requestResetEmailPassword(data.email);
  };
  const onDismiss = () => {
    reset();
    if (status == "success") {
      router.replace("/signin");
    }
  };

  const snackbarType = status == "success" ? "success" : "error";
  let snackbarMessage;

  if (status == "success") {
    snackbarMessage = "A recovery email was sent to you";
  } else {
    snackbarMessage = error?.message || "Oups ... could you try again please ?";
  }

  return (
    <View style={{ ...styles.container }}>
      <Text style={styles.title}>Reset your password</Text>

      <Controller
        control={control}
        rules={{
          required: true,
        }}
        render={({ field: { onChange, onBlur, value } }) => (
          <TextInput
            label="Email"
            mode="outlined"
            keyboardType="email-address"
            textContentType="emailAddress"
            style={styles.inputs}
            left={<TextInput.Icon icon="email" />}
            onBlur={onBlur}
            onChangeText={onChange}
            value={value}
            error={!!errors.email}
          />
        )}
        name="email"
      />

      {errors.email && (
        <HelperText type="error"> {errors.email.message}</HelperText>
      )}

      <Button
      loading={isPending}
        style={styles.signUpButton}
        labelStyle={{ fontSize: 18 }}
        rippleColor={theme.colors.inversePrimary}
        onPress={handleSubmit(onSubmit)}
        mode="contained"
      >
        Send password reset email
      </Button>

      {["success", "error"].includes(status) && (
        <Snackbar
          type={snackbarType}
          onDismiss={onDismiss}
          duration={2000}
          message={snackbarMessage}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
  },
  title: {
    fontSize: 30,
    fontWeight: "bold",
    marginBottom: 60,
  },
  inputs: {
    width: "80%",
    marginBottom: 10,
  },
  signUpButton: {
    width: "80%",
    marginTop: 10,
    marginBottom: 20,
    padding: 10,
  },
  bottom: {
    fontSize: 20,
    width: "80%",
    flexDirection: "row",
    justifyContent: "space-between",
  },
});
