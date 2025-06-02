import * as usersApi from "@/api/users";
import { useSnackbar } from "@/hooks/useSnackbar";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Link, router } from "expo-router";
import * as React from "react";
import { Controller, FieldError, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  Animated,
  Dimensions,
  Image,
  Keyboard,
  Platform,
  ScrollView,
  StyleProp,
  StyleSheet,
  Text,
  TextInput,
  TextStyle,
  TouchableOpacity,
  View,
} from "react-native";
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
  if (!error) return null;

  return (
    <Text style={[styles.errorMessage, style]}>
      {error?.message || "Invalid"}
    </Text>
  );
};

export default function SignUpScreen() {
  const [secureTextEntry, setSecureTextEntry] = React.useState(true);
  const [signUpError, setSignUpError] = React.useState<Error | null>(null);
  const snackbar = useSnackbar();
  const { t } = useTranslation();

  const scrollViewRef = React.useRef<ScrollView>(null);

  // Assuming image native resolution is 1024x1536
  const imageAspectRatio = 1024 / 1536;

  // Calculate full width height respecting aspect ratio
  const scaledImageHeight = screenWidth / imageAspectRatio;

  // Viewport constraint (max visible area)
  const maxViewportHeight = screenHeight * 0.7;

  const [, setKeyboardVisible] = React.useState(false);
  const imageHeight = React.useRef(
    new Animated.Value(maxViewportHeight),
  ).current;

  React.useEffect(() => {
    if (signUpError) {
      snackbar.show(signUpError.message, "error", 2000);
    }
  }, [signUpError, snackbar]);

  React.useEffect(() => {
    const keyboardWillShow = Keyboard.addListener(
      Platform.OS === "ios" ? "keyboardWillShow" : "keyboardDidShow",
      () => {
        setKeyboardVisible(true);
        Animated.timing(imageHeight, {
          toValue: maxViewportHeight * 0.5, // Shrink image when keyboard shows
          duration: 250,
          useNativeDriver: false,
        }).start();
      },
    );

    const keyboardWillHide = Keyboard.addListener(
      Platform.OS === "ios" ? "keyboardWillHide" : "keyboardDidHide",
      () => {
        setKeyboardVisible(false);
        Animated.timing(imageHeight, {
          toValue: maxViewportHeight,
          duration: 250,
          useNativeDriver: false,
        }).start();
      },
    );

    return () => {
      keyboardWillShow.remove();
      keyboardWillHide.remove();
    };
  }, []);

  const { mutate: signup } = useMutation<void, Error, usersApi.UserSignup>({
    mutationFn: (userData) => usersApi.signup(userData),
    onError: (error) => {
      setSignUpError(error);
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
      email: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_EMAIL : "",
      password: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_PASSWORD : "",
    },
  });

  const onSubmit = (data: any) => signup(data);
  const toggleSecureTextEntry = () => setSecureTextEntry(!secureTextEntry);

  return (
    <View style={styles.container}>
      {/* Viewport with max 70% of screen height - use Animated.View for the image viewport */}
      <Animated.View style={[styles.imageViewport, { height: imageHeight }]}>
        <Image
          source={require("../assets/images/signup-bg.png")}
          style={[
            styles.illustration,
            { width: screenWidth, height: scaledImageHeight },
          ]}
          resizeMode="contain"
        />
      </Animated.View>
      <View style={{ flex: 1, marginTop: -(screenHeight * 0.29) }}>
        <ScrollView contentContainerStyle={styles.content}>
          <View style={styles.formCard}>
            <Text style={styles.title}>Sign Up</Text>

            <Controller
              control={control}
              rules={{ required: true }}
              name="email"
              render={({ field: { onChange, onBlur, value } }) => (
                <View style={styles.inputGroup}>
                  <View style={styles.inputLabelRow}>
                    <MaterialIcons name="email" size={20} color="#636E72" />
                    <Text style={styles.inputLabel}>Email</Text>
                  </View>
                  <TextInput
                    style={[styles.input, errors.email && styles.inputError]}
                    placeholder="Enter your email"
                    placeholderTextColor="#B2BEC3"
                    value={value}
                    onChangeText={onChange}
                    onFocus={(event) => {
                      // Scroll to the input after a short delay
                      setTimeout(() => {
                        scrollViewRef.current?.scrollToEnd({ animated: true });
                      }, 300);
                    }}
                    onBlur={onBlur}
                    autoCapitalize="none"
                    keyboardType="email-address"
                    autoComplete="email"
                  />
                  <ErrorMessage error={errors.email} />
                </View>
              )}
            />

            <View style={styles.row}>
              <Controller
                control={control}
                rules={{ required: true }}
                name="firstName"
                render={({ field: { onChange, onBlur, value } }) => (
                  <View style={[styles.inputGroup, styles.inputHalf]}>
                    <View style={styles.inputLabelRow}>
                      <MaterialIcons name="person" size={20} color="#636E72" />
                      <Text style={styles.inputLabel}>First Name</Text>
                    </View>
                    <TextInput
                      style={[
                        styles.input,
                        errors.firstName && styles.inputError,
                      ]}
                      onFocus={(event) => {
                        // Scroll to the input after a short delay
                        setTimeout(() => {
                          scrollViewRef.current?.scrollToEnd({
                            animated: true,
                          });
                        }, 300);
                      }}
                      placeholder="Enter first name"
                      placeholderTextColor="#B2BEC3"
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                      autoCapitalize="words"
                    />
                  </View>
                )}
              />

              <Controller
                control={control}
                rules={{ required: true }}
                name="lastName"
                render={({ field: { onChange, onBlur, value } }) => (
                  <View
                    style={[
                      styles.inputGroup,
                      styles.inputHalf,
                      { marginLeft: 8 },
                    ]}
                  >
                    <View style={styles.inputLabelRow}>
                      <MaterialIcons name="person" size={20} color="#636E72" />
                      <Text style={styles.inputLabel}>Last Name</Text>
                    </View>
                    <TextInput
                      onFocus={(event) => {
                        // Scroll to the input after a short delay
                        setTimeout(() => {
                          scrollViewRef.current?.scrollToEnd({
                            animated: true,
                          });
                        }, 300);
                      }}
                      style={[
                        styles.input,
                        errors.lastName && styles.inputError,
                      ]}
                      placeholder="Enter last name"
                      placeholderTextColor="#B2BEC3"
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                      autoCapitalize="words"
                    />
                  </View>
                )}
              />
            </View>

            {/* Error messages row */}
            {(errors.firstName || errors.lastName) && (
              <View style={styles.errorRow}>
                <View style={styles.errorColumn}>
                  <ErrorMessage error={errors.firstName} />
                </View>
                <View style={styles.errorColumn}>
                  <ErrorMessage error={errors.lastName} />
                </View>
              </View>
            )}

            <View style={{ flex: 1 }}>
              <Controller
                control={control}
                rules={{ required: true }}
                name="password"
                render={({ field: { onChange, onBlur, value } }) => (
                  <View style={styles.inputGroup}>
                    <View style={styles.inputLabelRow}>
                      <MaterialIcons name="lock" size={20} color="#636E72" />
                      <Text style={styles.inputLabel}>Password</Text>
                    </View>
                    <View style={styles.passwordContainer}>
                      <TextInput
                        onFocus={(event) => {
                          // Scroll to the input after a short delay
                          setTimeout(() => {
                            scrollViewRef.current?.scrollToEnd({
                              animated: true,
                            });
                          }, 300);
                        }}
                        style={[
                          styles.input,
                          styles.passwordInput,
                          errors.password && styles.inputError,
                        ]}
                        placeholder={t("auth.signup.passwordPlaceholder")}
                        placeholderTextColor="#B2BEC3"
                        value={value}
                        onChangeText={onChange}
                        onBlur={onBlur}
                        secureTextEntry={secureTextEntry}
                        autoComplete="password"
                      />
                      <TouchableOpacity
                        style={styles.passwordToggle}
                        onPress={toggleSecureTextEntry}
                      >
                        <Ionicons
                          name={secureTextEntry ? "eye" : "eye-off"}
                          size={20}
                          color="#636E72"
                        />
                      </TouchableOpacity>
                    </View>
                    <ErrorMessage error={errors.password} />
                  </View>
                )}
              />
            </View>

            <TouchableOpacity
              style={styles.button}
              onPress={handleSubmit(onSubmit)}
              activeOpacity={0.8}
            >
              <Text style={styles.buttonText}>
                {t("auth.signup.signUpButton")}
              </Text>
            </TouchableOpacity>

            <Text style={styles.footer}>
              {t("auth.signup.alreadyHaveAccount")}{" "}
              <Link replace style={styles.link} href={"/signin"}>
                <Text style={styles.link}>{t("auth.signup.signIn")}</Text>
              </Link>
            </Text>
          </View>
        </ScrollView>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#FDF6E3", // Updated to match your color scheme
  },
  imageViewport: {
    width: "100%",
    overflow: "hidden",
  },
  illustration: {
    width: "100%",
  },
  content: {
    // paddingHorizontal: 24,
    // paddingBottom: 24,

    paddingHorizontal: 24,
    paddingBottom: 100, // Increase this to ensure space at bottom
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
    color: "#2D3436",
    textAlign: "center",
    marginBottom: 24,
  },
  inputGroup: {
    marginBottom: 10,
  },
  inputLabelRow: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 8,
    gap: 6,
  },
  inputLabel: {
    fontSize: 14,
    fontWeight: "500",
    color: "#2D3436",
  },
  input: {
    backgroundColor: "#FAFAFA",
    borderWidth: 1,
    borderColor: "#E0E0E0",
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 16,
    color: "#2D3436",
  },
  inputError: {
    borderColor: "#FF6B6B",
    backgroundColor: "#FFF5F5",
  },
  passwordContainer: {
    position: "relative",
  },
  passwordInput: {
    paddingRight: 48,
  },
  passwordToggle: {
    position: "absolute",
    right: 16,
    top: "50%",
    transform: [{ translateY: -10 }],
  },
  row: {
    flexDirection: "row",
    gap: 8,
  },
  inputHalf: {
    flex: 1,
  },
  errorRow: {
    flexDirection: "row",
    marginTop: -6,
    marginBottom: 6,
  },
  errorColumn: {
    flex: 1,
    paddingHorizontal: 4,
  },
  errorMessage: {
    fontSize: 12,
    color: "#FF6B6B",
    marginTop: 4,
  },
  button: {
    backgroundColor: "#FF7B54", // Updated to match your color scheme
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: "center",
    marginTop: 8,
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 5,
  },
  buttonText: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  footer: {
    textAlign: "center",
    marginTop: 16,
    color: "#636E72",
    fontSize: 14,
  },
  link: {
    color: "#FF7B54", // Updated to match your color scheme
    fontWeight: "600",
  },

  contentStyleError: {
    color: "#c60000",
  },
  gridRow: {
    flexDirection: "row",
  },
  gridColumn: {
    flex: 1,
  },
  placeholder: {
    minHeight: 16, // Height matching the ErrorMessage height
  },
});
