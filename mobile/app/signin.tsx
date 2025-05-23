import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Link, router } from "expo-router";
import * as React from "react";
import { Controller, useForm } from "react-hook-form";
import { Keyboard, StyleSheet, Text, View } from "react-native";
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
    email: z.string().email(),
    password: z.string().min(1),
  })
  .required();

export default function SignInScreen() {
  const theme = useTheme();
  const [secureTextEntry, setSecureTextEntry] = React.useState(true);
  const [signInError, setSignInError] = React.useState<string | null>(null);

  const { mutate: signup } = useMutation<void, Error, usersApi.UserSignin>({
    mutationFn: (userData) => usersApi.signin(userData),
    onError: (error) => {
      setSignInError(error.message);
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
      email: "",
      password: "",
    },
  });

  const hasErrors = errors.email || errors.password;

  const onSubmit = (data: any) => {
    Keyboard.dismiss()
    signup(data)
  };
  const toggleSecureTextEntry = () => setSecureTextEntry(!secureTextEntry);

  return (
    <View style={{ ...styles.container }}>
      <Text style={styles.title}>Sign in</Text>

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
        <HelperText type="error">Oups, some fields are missing</HelperText>
      )}

      <Button
        style={styles.signUpButton}
        labelStyle={{ fontSize: 20 }}
        rippleColor={theme.colors.inversePrimary}
        onPress={handleSubmit(onSubmit)}
        mode="contained"
      >
        Sign In
      </Button>

      <View style={styles.bottom}>
        <Link replace style={{ color: theme.colors.primary,fontSize:18 }} href={"/resetPassword"}>
          Forgot your password?
        </Link>
        <Link replace style={{ color: theme.colors.primary,fontSize:18 }} href={"/signup"}>
          Sign up
        </Link>
      </View>

      <Snackbar
        visible={!!signInError}
        onDismiss={() => setSignInError(null)}
        action={{
          label: "Hide",
        }}
      >
        {signInError}
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
  bottom: {
    fontSize: 20,
    width: "80%",
    flexDirection: "row",
    justifyContent: "space-between",
  },
});
