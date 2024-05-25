import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Link, router } from "expo-router";
import * as React from "react";
import { Controller, useForm } from "react-hook-form";
import { StyleSheet, Text, View } from "react-native";
import {
  Button,
  HelperText,
  Snackbar,
  TextInput,
  useTheme,
} from "react-native-paper";
import z from "zod";
import * as usersApi from "@/api/users";

const schema = z
  .object({
    firstName: z.string().min(2),
    lastName: z.string().min(2),
    email: z.string().email(),
    password: z.string().min(2),
  })
  .required();

export default function SignupScreen() {
  const theme = useTheme();
  const [secureTextEntry, setSecureTextEntry] = React.useState(true);
  const [signUpError, setSignUpError] = React.useState<string | null>(null);

  const { mutate: signup } = useMutation<void, Error, usersApi.UserSignup>({
    mutationFn: (userData) => usersApi.signup(userData),
    onError: (error) => {
      setSignUpError(error.message);
    },
    onSuccess: () => {
      router.replace("/dashboard");
    },
  });

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(schema),
    defaultValues: {
      firstName: "",
      lastName: "",
      email: "",
      password: "",
    },
  });

  const hasErrors =
    errors.email || errors.firstName || errors.lastName || errors.password;

  const onSubmit = (data: any) => signup(data);
  const toggleSecureTextEntry = () => setSecureTextEntry(!secureTextEntry);

  return (
    <View style={{ ...styles.container }}>
      <Text style={styles.title}>Sign up new account</Text>

      <Controller
        control={control}
        rules={{
          required: true,
        }}
        render={({ field: { onChange, onBlur, value } }) => (
          <TextInput
            label="First name"
            mode="outlined"
            textContentType="name"
            error={!!errors.firstName}
            style={styles.inputs}
            left={<TextInput.Icon icon="account" />}
            onBlur={onBlur}
            onChangeText={onChange}
            value={value}
          />
        )}
        name="firstName"
      />

      <Controller
        control={control}
        rules={{
          required: true,
        }}
        render={({ field: { onChange, onBlur, value } }) => (
          <TextInput
            label="Last name"
            mode="outlined"
            textContentType="middleName"
            style={styles.inputs}
            left={<TextInput.Icon icon="account" />}
            onBlur={onBlur}
            error={!!errors.lastName}
            onChangeText={onChange}
            value={value}
          />
        )}
        name="lastName"
      />

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

      <Controller
        control={control}
        rules={{
          required: true,
        }}
        render={({ field: { onChange, onBlur, value } }) => (
          <TextInput
            label="Password"
            mode="outlined"
            secureTextEntry={secureTextEntry}
            textContentType="password"
            placeholder="Type something"
            left={<TextInput.Icon icon="lock" />}
            right={
              <TextInput.Icon icon="eye" onPress={toggleSecureTextEntry} />
            }
            style={styles.inputs}
            onBlur={onBlur}
            onChangeText={onChange}
            error={!!errors.password}
            value={value}
          />
        )}
        name="password"
      />
      {hasErrors && (
        <HelperText type="error">
          Oups, some fields need some love :-)
        </HelperText>
      )}

      <Button
        style={styles.signUpButton}
        labelStyle={{ fontSize: 20 }}
        rippleColor={theme.colors.inversePrimary}
        onPress={handleSubmit(onSubmit)}
        mode="contained"
      >
        Sign Up
      </Button>
      <Text style={{ ...styles.signInHelper }}>
        Already have an account?{" "}
        <Link replace style={{ color: theme.colors.primary }} href={"/signin"}>
          Sign in
        </Link>
      </Text>

      <Snackbar
        visible={!!signUpError}
        onDismiss={() => setSignUpError(null)}
        action={{
          label: "Hide",
        }}
      >
        {signUpError}
      </Snackbar>
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
  signInHelper: {
    fontSize: 20,
  },
});
