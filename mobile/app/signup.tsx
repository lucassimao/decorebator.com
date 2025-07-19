import * as usersApi from "@/api/users";
import { useSnackbar } from "@/hooks/useSnackbar";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Link, router } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import { usePostHog } from "posthog-react-native";
import * as React from "react";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import i18n from "@/i18n";
import { getDetectedCountry } from "@/utils/countryDetection";
import { mapErrorToI18n } from "@/utils/errorMapping";
import AsyncStorage from "@react-native-async-storage/async-storage";
import {
  ImageBackground,
  Keyboard,
  KeyboardAvoidingView,
  Platform,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import z from "zod";
import { authLightTheme } from "@/theme/authTheme";
import type { Theme } from "@/contexts/ThemeContext";
import { useTheme } from "@/contexts/ThemeContext";
import type { ResponsiveValues } from "@/contexts/ThemeContext";

const schema = z
  .object({
    firstName: z.string().min(2, "Required"),
    lastName: z.string().min(2, "Required"),
    email: z.string().email().min(2, "Required"),
    password: z.string().min(5, "Required"),
  })
  .required();

export default function SignUpScreen() {
  const [secureTextEntry, setSecureTextEntry] = React.useState(true);
  const [signUpError, setSignUpError] = React.useState<Error | null>(null);
  const [detectedCountry, setDetectedCountry] = React.useState<string>("");
  const [keyboardVisible, setKeyboardVisible] = React.useState(false);
  const snackbar = useSnackbar();
  const { t } = useTranslation();
  const posthog = usePostHog();

  // Always use light theme for auth screens
  const theme = authLightTheme;

  // Get responsive values from the unified theme context
  const { responsive } = useTheme();

  // Memoize styles to prevent recreation on every render
  const styles = React.useMemo(
    () => createStyles(theme, responsive, keyboardVisible),
    [theme, responsive, keyboardVisible],
  );

  const scrollViewRef = React.useRef<ScrollView>(null);
  const emailInputRef = React.useRef<TextInput>(null);
  const firstNameInputRef = React.useRef<TextInput>(null);
  const lastNameInputRef = React.useRef<TextInput>(null);
  const passwordInputRef = React.useRef<TextInput>(null);

  React.useEffect(() => {
    if (signUpError) {
      // Only show snackbar for non-email errors (email errors are shown inline)
      snackbar.show(signUpError.message, "error", 2000);
    }
  }, [signUpError, snackbar]);

  // Keyboard listeners
  React.useEffect(() => {
    const keyboardWillShow = Keyboard.addListener(
      Platform.OS === "ios" ? "keyboardWillShow" : "keyboardDidShow",
      () => setKeyboardVisible(true),
    );

    const keyboardWillHide = Keyboard.addListener(
      Platform.OS === "ios" ? "keyboardWillHide" : "keyboardDidHide",
      () => setKeyboardVisible(false),
    );

    return () => {
      keyboardWillShow.remove();
      keyboardWillHide.remove();
    };
  }, []);

  // Detect user's country on component mount
  React.useEffect(() => {
    try {
      const country = getDetectedCountry();
      setDetectedCountry(country);
      if (__DEV__) {
        console.log("Detected country:", country);
      }
    } catch (error) {
      if (__DEV__) {
        console.warn("Failed to detect country:", error);
      }
      setDetectedCountry("US"); // Fallback to US
    }
  }, []);

  // Auto-focus email field on component mount for better UX
  // React.useEffect(() => {
  //   const timer = setTimeout(() => {
  //     emailInputRef.current?.focus();
  //   }, 300); // Small delay to ensure component is fully mounted

  //   return () => clearTimeout(timer);
  // }, []);

  const { mutate: signup } = useMutation<void, Error, usersApi.UserSignup>({
    mutationFn: (userData) => usersApi.signup(userData),
    onError: (error) => {
      const mappedError = mapErrorToI18n(error.message);

      if (mappedError.isFieldError && mappedError.field) {
        // Set error on specific field to display translated message below the input
        setError(mappedError.field as keyof typeof schema._type, {
          type: "manual",
          message: t(mappedError.i18nKey),
        });
      } else {
        // For other errors, use snackbar with translated message
        setSignUpError(new Error(t(mappedError.i18nKey)));
      }
    },
    onSuccess: async (_, variables) => {
      posthog.capture("user_signed_up", {
        email: variables.email,
      });

      // Set flag that user just signed up for welcome flow
      try {
        await AsyncStorage.setItem("justSignedUp", "true");
        if (__DEV__) {
          console.log("Set justSignedUp flag to true");
        }
      } catch (error) {
        console.warn("Failed to set signup flag:", error);
      }

      router.replace("/dashboard");
    },
  });

  const {
    control,
    handleSubmit,
    formState: { errors },
    setError,
  } = useForm({
    resolver: zodResolver(schema),
    mode: "onChange", // Validate on every change to enable real-time form state
    defaultValues: {
      firstName: "",
      lastName: "",
      email: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_EMAIL || "" : "",
      password: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_PASSWORD || "" : "",
    },
  });

  const onSubmit = React.useCallback(
    (data: z.infer<typeof schema>) => {
      // Add detected country to signup data
      const signupData = {
        ...data,
        country: detectedCountry || "US", // Include detected country or fallback to US
      };
      signup(signupData);
    },
    [detectedCountry, signup],
  );

  const toggleSecureTextEntry = React.useCallback(
    () => setSecureTextEntry((prev) => !prev),
    [],
  );

  // Function to handle scrolling to input with improved reliability
  const scrollToInput = React.useCallback(
    (inputRef: React.RefObject<TextInput | null>) => {
      if (!inputRef.current || !scrollViewRef.current) return;

      // Method 1: Try using KeyboardAvoidingView's automatic behavior
      // (We'll rely on the KeyboardAvoidingView component to handle basic scrolling)
      const tryBuiltInScroll = () => {
        // For now, return false to always use our manual implementation
        // In the future, we could implement platform-specific solutions here
        return false;
      };

      // Method 2: Improved manual implementation with better coordinate handling
      const tryManualScroll = () => {
        const scrollView = scrollViewRef.current;
        const input = inputRef.current;

        if (!scrollView || !input) return;

        // Use measureInWindow instead of measureLayout to avoid native component ref issues
        input.measureInWindow((x, y, width, height) => {
          // Only scroll if keyboard is visible and input might be covered
          if (!keyboardVisible) {
            return; // Don't scroll at all when keyboard isn't visible
          }

          // Conservative keyboard height estimation
          const keyboardHeight = 300;
          const inputBottom = y + height;
          const visibleScreenBottom = responsive.screenHeight - keyboardHeight;

          // Only scroll if the input is actually being covered by the keyboard
          if (inputBottom > visibleScreenBottom) {
            // Minimal scroll - just enough to show the input above keyboard
            const targetY = visibleScreenBottom - height - 20; // 20px padding above keyboard
            const scrollAmount = Math.max(0, y - targetY);

            scrollView.scrollTo({
              y: scrollAmount,
              animated: true,
            });
          }
        });
      };

      // Small delay to allow layout to settle after focus
      setTimeout(() => {
        // Try built-in method first, fallback to manual if it fails
        if (!tryBuiltInScroll()) {
          tryManualScroll();
        }
      }, 100);
    },
    [keyboardVisible, responsive.screenHeight],
  );

  return (
    <ImageBackground
      source={require("../assets/images/signup-bg3.png")}
      style={styles.backgroundImage}
      imageStyle={styles.backgroundImageStyle}
      resizeMode="cover"
    >
      <SafeAreaView style={styles.safeArea}>
        <KeyboardAvoidingView
          style={styles.container}
          behavior={responsive.keyboardBehavior}
          keyboardVerticalOffset={responsive.keyboardOffset}
        >
          <TouchableWithoutFeedback onPress={Keyboard.dismiss}>
            <ScrollView
              ref={scrollViewRef}
              contentContainerStyle={styles.scrollContent}
              showsVerticalScrollIndicator={false}
              keyboardShouldPersistTaps="handled"
            >
              {/* Add some space at the top for the background to show */}
              <View style={styles.topSpacer} />

              {/* Form Card */}
              <View style={styles.formCard}>
                {/* Logo or App Name */}
                <View style={styles.headerSection}>
                  <Text style={styles.title}>{t("auth.signup.title")}</Text>
                  <Text style={styles.subtitle}>
                    {t(
                      "auth.signup.subtitle",
                      "Start your language learning journey",
                    )}
                  </Text>
                </View>

                {/* Email Input */}
                <Controller
                  control={control}
                  rules={{ required: true }}
                  name="email"
                  render={({ field: { onChange, onBlur, value } }) => (
                    <View style={styles.inputGroup}>
                      <View style={styles.inputLabelRow}>
                        <MaterialIcons
                          name="email"
                          size={20}
                          color={theme.colors.text.secondary}
                        />
                        <Text style={styles.inputLabel}>
                          {t("auth.signup.email")}
                        </Text>
                      </View>
                      <TextInput
                        ref={emailInputRef}
                        style={[
                          styles.input,
                          errors.email && styles.inputError,
                        ]}
                        placeholder={t("auth.signup.emailPlaceholder")}
                        placeholderTextColor={theme.colors.text.placeholder}
                        value={value}
                        onChangeText={onChange}
                        // onFocus={() => scrollToInput(emailInputRef)}
                        onBlur={onBlur}
                        autoCapitalize="none"
                        keyboardType="email-address"
                        autoComplete="email"
                        textContentType="emailAddress"
                        returnKeyType="next"
                        onSubmitEditing={() =>
                          firstNameInputRef.current?.focus()
                        }
                      />
                      <ErrorMessage error={errors.email} />
                    </View>
                  )}
                />

                {/* Name Inputs Row */}
                <View style={styles.row}>
                  <Controller
                    control={control}
                    rules={{ required: true }}
                    name="firstName"
                    render={({ field: { onChange, onBlur, value } }) => (
                      <View style={[styles.inputGroup, styles.inputHalf]}>
                        <View style={styles.inputLabelRow}>
                          <MaterialIcons
                            name="person"
                            size={20}
                            color={theme.colors.text.secondary}
                          />
                          <Text style={styles.inputLabel}>
                            {t("auth.signup.firstName")}
                          </Text>
                        </View>
                        <TextInput
                          ref={firstNameInputRef}
                          style={[
                            styles.input,
                            errors.firstName && styles.inputError,
                          ]}
                          placeholder={t("auth.signup.firstNamePlaceholder")}
                          placeholderTextColor={theme.colors.text.placeholder}
                          value={value}
                          onChangeText={onChange}
                          onFocus={() => scrollToInput(firstNameInputRef)}
                          onBlur={onBlur}
                          autoCapitalize="words"
                          textContentType="givenName"
                          returnKeyType="next"
                          onSubmitEditing={() =>
                            lastNameInputRef.current?.focus()
                          }
                        />
                        <ErrorMessage error={errors.firstName} />
                      </View>
                    )}
                  />

                  <Controller
                    control={control}
                    rules={{ required: true }}
                    name="lastName"
                    render={({ field: { onChange, onBlur, value } }) => (
                      <View style={[styles.inputGroup, styles.inputHalf]}>
                        <View style={styles.inputLabelRow}>
                          <MaterialIcons
                            name="person"
                            size={20}
                            color={theme.colors.text.secondary}
                          />
                          <Text style={styles.inputLabel}>
                            {t("auth.signup.lastName")}
                          </Text>
                        </View>
                        <TextInput
                          ref={lastNameInputRef}
                          style={[
                            styles.input,
                            errors.lastName && styles.inputError,
                          ]}
                          placeholder={t("auth.signup.lastNamePlaceholder")}
                          placeholderTextColor={theme.colors.text.placeholder}
                          value={value}
                          onChangeText={onChange}
                          onFocus={() => scrollToInput(lastNameInputRef)}
                          onBlur={onBlur}
                          autoCapitalize="words"
                          textContentType="familyName"
                          returnKeyType="next"
                          onSubmitEditing={() =>
                            passwordInputRef.current?.focus()
                          }
                        />
                        <ErrorMessage error={errors.lastName} />
                      </View>
                    )}
                  />
                </View>

                {/* Password Input */}
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
                          ref={passwordInputRef}
                          style={[
                            styles.input,
                            styles.passwordInput,
                            errors.password && styles.inputError,
                          ]}
                          placeholder={t("auth.signup.passwordPlaceholder")}
                          placeholderTextColor={theme.colors.text.placeholder}
                          value={value}
                          onChangeText={onChange}
                          onFocus={() => scrollToInput(passwordInputRef)}
                          onBlur={onBlur}
                          secureTextEntry={secureTextEntry}
                          autoComplete="password-new"
                          textContentType="newPassword"
                          returnKeyType="done"
                          onSubmitEditing={() => Keyboard.dismiss()}
                        />
                        <TouchableOpacity
                          style={styles.passwordToggle}
                          onPress={toggleSecureTextEntry}
                          hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
                        >
                          <Ionicons
                            name={secureTextEntry ? "eye-off" : "eye"}
                            size={20}
                            color={theme.colors.text.secondary}
                          />
                        </TouchableOpacity>
                      </View>
                      <ErrorMessage error={errors.password} />
                    </View>
                  )}
                />

                {/* Implicit Terms Acceptance */}
                <View style={styles.termsContainer}>
                  <Text style={styles.implicitTermsText}>
                    {t(
                      "auth.signup.implicitTerms",
                      "By signing up, you agree to our",
                    )}{" "}
                    <Link
                      href={`https://decorebator.com/${i18n.language}/terms`}
                      style={styles.termsLink}
                      onPress={async (e) => {
                        e.preventDefault();
                        await WebBrowser.openBrowserAsync(
                          `https://decorebator.com/${i18n.language}/terms`,
                        );
                      }}
                      accessibilityRole="link"
                      accessibilityHint="Opens terms of service in web browser"
                    >
                      {t("auth.signup.termsOfService", "Terms of Service")}
                    </Link>{" "}
                    {t("common.and", "and")}{" "}
                    <Link
                      href={`https://decorebator.com/${i18n.language}/privacy`}
                      style={styles.termsLink}
                      onPress={async (e) => {
                        e.preventDefault();
                        await WebBrowser.openBrowserAsync(
                          `https://decorebator.com/${i18n.language}/privacy`,
                        );
                      }}
                      accessibilityRole="link"
                      accessibilityHint="Opens privacy policy in web browser"
                    >
                      {t("auth.signup.privacyPolicy", "Privacy Policy")}
                    </Link>
                    .
                  </Text>
                </View>
              </View>

              {/* Bottom spacer for fixed footer */}
              <View style={styles.bottomSpacer} />
            </ScrollView>
          </TouchableWithoutFeedback>
        </KeyboardAvoidingView>

        {/* Fixed Footer */}
        <View style={styles.fixedFooter}>
          {/* Submit Button */}
          <TouchableOpacity
            style={styles.buttonInFooter}
            onPress={handleSubmit(onSubmit)}
            activeOpacity={0.8}
          >
            <Text style={styles.buttonText}>
              {t("auth.signup.signUpButton")}
            </Text>
          </TouchableOpacity>

          {/* Footer Link */}
          <Text style={styles.footer}>
            {t("auth.signup.alreadyHaveAccount")}{" "}
            <Link replace style={styles.link} href={"/signin"}>
              <Text style={styles.link}>{t("auth.signup.signIn")}</Text>
            </Link>
          </Text>
        </View>
      </SafeAreaView>
    </ImageBackground>
  );
}

const createStyles = (
  theme: Theme,
  responsive: ResponsiveValues,
  keyboardVisible: boolean,
) =>
  StyleSheet.create({
    backgroundImage: {
      flex: 1,
      width: "100%",
      height: "100%",
      backgroundColor: "#FFF9F0", // Fallback warm background color
    },
    backgroundImageStyle: {
      // Force full coverage regardless of aspect ratio
      width: "100%",
      height: "100%",
    },
    safeArea: {
      flex: 1,
      backgroundColor: "transparent", // Let image show through
    },
    container: {
      flex: 1,
    },
    scrollContent: {
      flexGrow: 1,
      paddingHorizontal: responsive.spacing.horizontal,
    },
    topSpacer: {
      height: keyboardVisible
        ? responsive.spacing.vertical
        : responsive.screenHeight * 0.12, // Balanced spacing
    },
    bottomSpacer: {
      height: responsive.spacing.vertical * 4, // Increased to account for fixed footer
    },
    formCard: {
      backgroundColor: "rgba(255, 255, 255, 0.95)",
      borderRadius: theme.borderRadius.xl,
      padding: responsive.spacing.formPadding,
      ...theme.shadows.lg,
      // Add subtle border for definition
      borderWidth: 1,
      borderColor: "rgba(255, 255, 255, 0.3)",
    },
    headerSection: {
      alignItems: "center",
      marginBottom: responsive.spacing.elementSpacing * 1.5,
    },
    title: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: responsive.spacing.elementSpacing / 2,
    },
    subtitle: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    inputGroup: {
      marginBottom: responsive.spacing.elementSpacing,
    },
    inputLabelRow: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: 8,
      gap: 6,
    },
    inputLabel: {
      fontSize: responsive.fontSizes.label,
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
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.primary,
      // Ensure minimum height for touch targets
      minHeight: 48,
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
      gap: responsive.spacing.elementSpacing / 2,
    },
    inputHalf: {
      flex: 1,
    },
    errorRow: {
      flexDirection: "row",
      marginTop: 0,
      marginBottom: 0,
    },
    errorColumn: {
      flex: 1,
      paddingHorizontal: 0,
    },
    termsContainer: {
      marginTop: responsive.spacing.elementSpacing,
      marginBottom: responsive.spacing.elementSpacing / 2,
    },
    implicitTermsText: {
      fontSize: responsive.fontSizes.label,
      color: theme.colors.text.secondary,
      lineHeight: 20,
      textAlign: "center",
    },
    termsLink: {
      color: theme.colors.primary,
      fontWeight: "600",
      textDecorationLine: "underline",
    },
    button: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md,
      alignItems: "center",
      marginTop: responsive.spacing.elementSpacing,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
      minHeight: responsive.spacing.buttonHeight,
      justifyContent: "center",
    },
    buttonInFooter: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md,
      alignItems: "center",
      marginTop: 0, // No top margin in footer
      marginBottom: responsive.spacing.elementSpacing / 2,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
      minHeight: responsive.spacing.buttonHeight,
      justifyContent: "center",
    },
    buttonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
    },
    footer: {
      textAlign: "center",
      marginTop: responsive.spacing.elementSpacing,
      color: theme.colors.text.secondary,
      fontSize: responsive.fontSizes.label,
    },
    link: {
      color: theme.colors.primary,
      fontWeight: "600",
      fontSize: responsive.isSmallPhone
        ? responsive.fontSizes.body
        : responsive.fontSizes.label,
    },
    fixedFooter: {
      position: "absolute",
      bottom: 0,
      left: 0,
      right: 0,
      backgroundColor: "rgba(255, 255, 255, 0.98)",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.vertical,
      paddingBottom: responsive.spacing.vertical + (keyboardVisible ? 0 : 20), // Extra padding for safe area
      borderTopWidth: 1,
      borderTopColor: "rgba(0, 0, 0, 0.05)",
      ...theme.shadows.lg,
    },
  });
