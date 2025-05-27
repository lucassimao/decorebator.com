import * as usersApi from "@/api/users";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Link, router } from "expo-router";
import * as React from "react";
import SnackBar from '@/components/SnackBar'
import { Controller, FieldError, useForm } from "react-hook-form";
import {
  Dimensions,
  Image,
  ScrollView,
  StyleProp,
  StyleSheet,
  Text,
  TextStyle,
  View
} from "react-native";
import { Button, TextInput, useTheme } from "react-native-paper";
import z from "zod";

const schema = z
  .object({
    firstName: z.string().min(2, "Required"),
    lastName: z.string().min(2, "Required"),
    email: z.string().email().min(2, "Required"),
    password: z.string().min(2, "Required"),
  })
  .required();

const { width: screenWidth, height: screenHeight } = Dimensions.get("window");

type ErrorMessageProps = {
  error?: FieldError | null;
  style?: StyleProp<TextStyle>;
};
const ErrorMessage = ({ error, style }: ErrorMessageProps) => {
  const theme = useTheme();

  if (!error) return null;

  return (
    <Text style={[styles.errorMessage, { color: theme.colors.error }, style]}>
      {error?.message || "Invalid"}
    </Text>
  );
};

export default function SignUpScreen() {
  const theme = useTheme();
  const [secureTextEntry, setSecureTextEntry] = React.useState(true);
  const [signUpError, setSignUpError] = React.useState<string | null>(null);

  const { mutate: signup } = useMutation<void, Error, usersApi.UserSignup>({
    mutationFn: (userData) => usersApi.signup(userData),
    onError: (error) => {
      setSignUpError(error.message);
    },
    onSuccess: () => {
      router.replace("/dashboard/welcome");
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

  // Assuming image native resolution is 1024x1536
  const imageAspectRatio = 1024 / 1536;

  // Calculate full width height respecting aspect ratio
  const scaledImageHeight = screenWidth / imageAspectRatio;

  // Viewport constraint (max visible area)
  const maxViewportHeight = screenHeight * 0.7;

  return (
    <View style={styles.container}>
      {/* Viewport with max 70% of screen height */}
      <View style={[styles.imageViewport, { height: maxViewportHeight }]}>
        <Image
          source={require("../assets/images/signup-bg.png")}
          style={[
            styles.illustration,
            { width: screenWidth, height: scaledImageHeight },
          ]}
          resizeMode="contain"
        />
      </View>

      <View style={{ flex: 1, marginTop: -(screenHeight * 0.29) }}>
        <ScrollView contentContainerStyle={styles.content}>
          <View style={styles.formCard}>
            <Text style={styles.title}>Sign Up</Text>

            <Controller
              control={control}
              rules={{
                required: true,
              }}
              name="email"
              render={({ field: { onChange, onBlur, value } }) => (
                <View>
                  <TextInput
                    label="Email"
                    mode="outlined"
                    style={styles.input}
                    error={!!errors.email}
                    textContentType="emailAddress"
                    left={<TextInput.Icon icon="email" />}
                    onBlur={onBlur}
                    onChangeText={onChange}
                    value={value}
                    outlineStyle={errors.email ? styles.inputError : undefined}
                    contentStyle={
                      errors.email ? styles.contentStyleError : undefined
                    }
                  />
                  <ErrorMessage error={errors.email} />
                </View>
              )}
            />
            <View style={styles.row}>
              <Controller
                control={control}
                rules={{
                  required: true,
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    label="First Name"
                    mode="outlined"
                    left={<TextInput.Icon icon="account" />}
                    textContentType="name"
                    error={!!errors.firstName}
                    style={[styles.input, styles.inputHalf]}
                    outlineStyle={
                      errors.firstName ? styles.inputError : undefined
                    }
                    contentStyle={
                      errors.firstName ? styles.contentStyleError : undefined
                    }
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
                    label="Last Name"
                    mode="outlined"
                    textContentType="middleName"
                    style={[styles.input, styles.inputHalf, { marginLeft: 8 }]}
                    left={<TextInput.Icon icon="account" />}
                    outlineStyle={
                      errors.lastName ? styles.inputError : undefined
                    }
                    contentStyle={
                      errors.lastName ? styles.contentStyleError : undefined
                    }
                    onBlur={onBlur}
                    error={!!errors.lastName}
                    onChangeText={onChange}
                    value={value}
                  />
                )}
                name="lastName"
              />
            </View>

           {(errors.firstName || errors.lastName) && <View style={styles.gridRow}>
              <View style={styles.gridColumn}>
                {errors.firstName ? (
                  <ErrorMessage error={errors.firstName} />
                ) : (
                  <View style={styles.placeholder} />
                )}
              </View>

              <View style={styles.gridColumn}>
                {errors.lastName ? (
                  <ErrorMessage error={errors.lastName} />
                ) : (
                  <View style={styles.placeholder} />
                )}
              </View>
            </View>}

            <View style={{ flex: 1 }}>
              <Controller
                control={control}
                rules={{
                  required: true,
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    label="Password"
                    secureTextEntry={secureTextEntry}
                    mode="outlined"
                    textContentType="password"
                    style={[styles.input]}
                    left={<TextInput.Icon icon="lock" />}
                    right={
                      <TextInput.Icon
                        icon="eye"
                        onPress={toggleSecureTextEntry}
                      />
                    }
                    outlineStyle={
                      errors.password ? styles.inputError : undefined
                    }
                    contentStyle={
                      errors.password ? styles.contentStyleError : undefined
                    }
                    onBlur={onBlur}
                    error={!!errors.password}
                    onChangeText={onChange}
                    value={value}
                  />
                )}
                name="password"
              />
              <ErrorMessage error={errors.password} />
            </View>

            <Button
              mode="contained"
              style={styles.button}
              onPress={handleSubmit(onSubmit)}
            >
              Create Account
            </Button>

            <Text style={styles.footer}>
              Already have an account?{" "}
              <Link replace style={styles.link} href={"/signin"}>
                Sign in
              </Link>
            </Text>
          </View>
        </ScrollView>
      </View>
      {signUpError && <SnackBar message={signUpError} type="error" onDismiss={()=> setSignUpError(null)} duration={2000}  />}
    </View>
  );
}
const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff2d6",
  },
  imageViewport: {
    width: "100%",
    overflow: "hidden",
  },
  illustration: {
    width: "100%",
  },
  content: {
    paddingHorizontal: 24,
    paddingBottom: 24,
  },
  formCard: {
    backgroundColor: "white",
    borderRadius: 24,
    padding: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 6,
  },
  title: {
    fontSize: 26,
    fontWeight: "700",
    color: "#1D3B29",
    textAlign: "center",
    marginBottom: 24,
  },
  input: {
    backgroundColor: "#fffaeb",
    marginBottom: 16,
    color: "red",
  },
  row: {
    flexDirection: "row",
  },
  inputHalf: {
    flex: 1,
  },
  button: {
    backgroundColor: "#2A5D4D",
    borderRadius: 12,
    paddingVertical: 8,
    marginTop: 8,
  },
  buttonLabel: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  footer: {
    textAlign: "center",
    marginTop: 16,
    color: "#5C5C5C",
    fontSize: 14,
  },
  link: {
    color: "#2A5D4D",
    fontWeight: "600",
  },
  inputError: {
    borderColor: "#f7d7bf",
    backgroundColor: "#fff4e9",
  },
  contentStyleError: {
    color: "#c60000",
  },
  errorMessage: {
    fontSize: 12,
    marginLeft: 4,
    position: "relative",
    top: -4,
  },
  gridRow: {
    flexDirection: 'row',
  },
  gridColumn: {
    flex: 1,
  },
  placeholder: {
    minHeight: 16, // Height matching the ErrorMessage height
  },
});
