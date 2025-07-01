import * as usersApi from "@/api/users";
import { useSnackbar } from "@/hooks/useSnackbar";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Link, router } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import { usePostHog } from "posthog-react-native";
import * as React from "react";
import { Controller, FieldError, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import i18n from "@/i18n";
import { getDetectedCountry } from "@/utils/countryDetection";
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
import { authLightTheme } from "@/theme/authTheme";
import type { Theme } from "@/contexts/ThemeContext";

const schema = z
  .object({
    firstName: z.string().min(2, "Required"),
    lastName: z.string().min(2, "Required"),
    email: z.string().email().min(2, "Required"),
    password: z.string().min(2, "Required"),
    agreeToTerms: z.boolean().refine((val) => val === true, {
      message: "You must agree to the terms",
    }),
  })
  .required();

const { width: screenWidth, height: screenHeight } = Dimensions.get("window");

type ErrorMessageProps = {
  error?: FieldError | null;
  style?: StyleProp<TextStyle>;
  errorStyle?: StyleProp<TextStyle>;
};

export default function SignUpScreen() {
  const [secureTextEntry, setSecureTextEntry] = React.useState(true);
  const [signUpError, setSignUpError] = React.useState<Error | null>(null);
  const [detectedCountry, setDetectedCountry] = React.useState<string>("");
  const snackbar = useSnackbar();
  const { t } = useTranslation();
  const posthog = usePostHog();
  // Always use light theme for auth screens
  const theme = authLightTheme;
  const styles = createStyles(theme);

  const ErrorMessage = ({ error, style, errorStyle }: ErrorMessageProps) => {
    if (!error) return null;

    return (
      <Text style={[styles.errorMessage, errorStyle, style]}>
        {error?.message || "Invalid"}
      </Text>
    );
  };

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
  }, [imageHeight, maxViewportHeight]);

  // Detect user's country on component mount
  React.useEffect(() => {
    try {
      const country = getDetectedCountry();
      setDetectedCountry(country);
      console.log("Detected country:", country);
    } catch (error) {
      console.warn("Failed to detect country:", error);
      setDetectedCountry("US"); // Fallback to US
    }
  }, []);

  const { mutate: signup } = useMutation<void, Error, usersApi.UserSignup>({
    mutationFn: (userData) => usersApi.signup(userData),
    onError: (error) => {
      setSignUpError(error);
    },
    onSuccess: (_, variables) => {
      posthog.capture("user_signed_up", {
        email: variables.email,
      });
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
      agreeToTerms: false,
    },
  });

  const onSubmit = (data: any) => {
    // Exclude agreeToTerms from API submission and add detected country
    const { agreeToTerms, ...submitData } = data;
    const signupData = {
      ...submitData,
      country: detectedCountry || "US", // Include detected country or fallback to US
    };
    signup(signupData);
  };
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
            <Text style={styles.title}>{t("auth.signup.title")}</Text>

            <Controller
              control={control}
              rules={{ required: true }}
              name="email"
              render={({ field: { onChange, onBlur, value } }) => (
                <View style={styles.inputGroup}>
                  <View style={styles.inputLabelRow}>
                    <MaterialIcons name="email" size={20} color="#636E72" />
                    <Text style={styles.inputLabel}>
                      {t("auth.signup.email")}
                    </Text>
                  </View>
                  <TextInput
                    style={[styles.input, errors.email && styles.inputError]}
                    placeholder={t("auth.signup.emailPlaceholder")}
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
                      <Text style={styles.inputLabel}>
                        {t("auth.signup.firstName")}
                      </Text>
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
                      placeholder={t("auth.signup.firstNamePlaceholder")}
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
                      <Text style={styles.inputLabel}>
                        {t("auth.signup.lastName")}
                      </Text>
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
                      placeholder={t("auth.signup.lastNamePlaceholder")}
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
                      <MaterialIcons
                        name="lock"
                        size={20}
                        color={theme.colors.text.secondary}
                      />
                      <Text style={styles.inputLabel}>
                        {t("auth.signup.password")}
                      </Text>
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
                        placeholderTextColor={theme.colors.text.placeholder}
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
                          color={theme.colors.text.secondary}
                        />
                      </TouchableOpacity>
                    </View>
                    <ErrorMessage error={errors.password} />
                  </View>
                )}
              />
            </View>

            {/* Terms and Privacy Agreement */}
            <Controller
              control={control}
              name="agreeToTerms"
              defaultValue={false}
              render={({ field: { onChange, value } }) => (
                <View style={styles.termsContainer}>
                  <TouchableOpacity
                    style={styles.checkboxContainer}
                    onPress={() => onChange(!value)}
                    activeOpacity={0.8}
                  >
                    <View
                      style={[styles.checkbox, value && styles.checkboxChecked]}
                    >
                      {value && (
                        <Ionicons
                          name="checkmark"
                          size={16}
                          color={theme.colors.text.inverse}
                        />
                      )}
                    </View>
                    <Text style={styles.termsText}>
                      {t("auth.signup.termsText")}{" "}
                      <Text
                        style={styles.termsLink}
                        onPress={async () => {
                          await WebBrowser.openBrowserAsync(
                            `https://decorebator.com/${i18n.language}/terms`,
                          );
                        }}
                      >
                        {t("auth.signup.termsOfService")}
                      </Text>{" "}
                      {t("common.and")}{" "}
                      <Text
                        style={styles.termsLink}
                        onPress={async () => {
                          await WebBrowser.openBrowserAsync(
                            `https://decorebator.com/${i18n.language}/privacy`,
                          );
                        }}
                      >
                        {t("auth.signup.privacyPolicy")}
                      </Text>
                    </Text>
                  </TouchableOpacity>
                  {errors.agreeToTerms && (
                    <Text style={styles.errorMessage}>
                      {t("auth.signup.mustAgreeToTerms")}
                    </Text>
                  )}
                </View>
              )}
            />

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

const createStyles = (theme: Theme) =>
  StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
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
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.xl,
      padding: theme.spacing.lg,
      ...theme.shadows.md,
    },
    title: {
      fontSize: 26,
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: theme.spacing.lg,
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
      color: theme.colors.text.primary,
    },
    input: {
      backgroundColor: theme.colors.ui.inputBackground,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      borderRadius: theme.borderRadius.md,
      paddingHorizontal: theme.spacing.md,
      paddingVertical: 14,
      fontSize: 16,
      color: theme.colors.text.primary,
    },
    inputError: {
      borderColor: theme.colors.error,
      backgroundColor: theme.colors.state.incorrectBackground,
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
      color: theme.colors.error,
      marginTop: 4,
    },
    button: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md,
      alignItems: "center",
      marginTop: theme.spacing.sm,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
    },
    buttonText: {
      color: theme.colors.text.inverse,
      fontSize: 16,
      fontWeight: "600",
    },
    footer: {
      textAlign: "center",
      marginTop: theme.spacing.md,
      color: theme.colors.text.secondary,
      fontSize: 14,
    },
    link: {
      color: theme.colors.primary,
      fontWeight: "600",
    },

    contentStyleError: {
      color: theme.colors.error,
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
    termsContainer: {
      marginTop: 16,
      marginBottom: 8,
    },
    checkboxContainer: {
      flexDirection: "row",
      alignItems: "flex-start",
    },
    checkbox: {
      width: 20,
      height: 20,
      borderWidth: 2,
      borderColor: theme.colors.ui.border,
      borderRadius: 4,
      marginRight: theme.spacing.sm,
      alignItems: "center",
      justifyContent: "center",
      marginTop: 2,
    },
    checkboxChecked: {
      backgroundColor: theme.colors.primary,
      borderColor: theme.colors.primary,
    },
    termsText: {
      flex: 1,
      fontSize: 14,
      color: theme.colors.text.secondary,
      lineHeight: 20,
    },
    termsLink: {
      color: theme.colors.primary,
      fontWeight: "600",
      textDecorationLine: "underline",
    },
  });
