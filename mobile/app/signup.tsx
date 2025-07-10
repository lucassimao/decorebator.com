import * as usersApi from "@/api/users";
import { useSnackbar } from "@/hooks/useSnackbar";
import { useResponsive } from "@/hooks/useResponsive";
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

  // Get responsive values using the optimized hook
  const responsive = useResponsive();

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
      email: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_EMAIL || "" : "",
      password: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_PASSWORD || "" : "",
      agreeToTerms: false,
    },
  });

  const onSubmit = React.useCallback(
    (data: z.infer<typeof schema>) => {
      // Exclude agreeToTerms from API submission and add detected country
      const { agreeToTerms, ...submitData } = data;
      const signupData = {
        ...submitData,
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

  // Function to handle scrolling to input with robust timing
  const scrollToInput = React.useCallback(
    (inputRef: React.RefObject<TextInput | null>) => {
      if (!inputRef.current || !scrollViewRef.current) return;

      // Use requestAnimationFrame for better timing coordination with layout
      const scrollToElement = () => {
        inputRef.current?.measureInWindow((x, y, width, height) => {
          const offset = Math.max(0, y - 150); // 150px padding from top
          scrollViewRef.current?.scrollTo({
            y: offset,
            animated: true,
          });
        });
      };

      // First try immediate scroll, then fallback with short delay if needed
      requestAnimationFrame(() => {
        scrollToElement();
        // Fallback for cases where keyboard animation hasn't completed
        setTimeout(scrollToElement, 50);
      });
    },
    [],
  );

  return (
    <ImageBackground
      source={require("../assets/images/signup-bg.png")}
      style={styles.backgroundImage}
      imageStyle={{ opacity: responsive.imageConfig.opacity }}
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
                        onFocus={() => scrollToInput(emailInputRef)}
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
                      </View>
                    )}
                  />
                </View>

                {/* Error messages row for names */}
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
                        accessibilityRole="checkbox"
                        accessibilityState={{ checked: value }}
                        accessibilityLabel={t("auth.signup.agreeToTerms")}
                        accessibilityHint={t("auth.signup.termsCheckboxHint")}
                      >
                        <View
                          style={[
                            styles.checkbox,
                            value && styles.checkboxChecked,
                          ]}
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
                            {t("auth.signup.termsOfService")}
                          </Link>{" "}
                          {t("common.and")}{" "}
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
                            {t("auth.signup.privacyPolicy")}
                          </Link>
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

                {/* Submit Button */}
                <TouchableOpacity
                  style={styles.button}
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

              {/* Bottom spacer for keyboard */}
              <View style={styles.bottomSpacer} />
            </ScrollView>
          </TouchableWithoutFeedback>
        </KeyboardAvoidingView>
      </SafeAreaView>
    </ImageBackground>
  );
}

const createStyles = (
  theme: Theme,
  responsive: ReturnType<typeof useResponsive>,
  keyboardVisible: boolean,
) =>
  StyleSheet.create({
    backgroundImage: {
      flex: 1,
      width: responsive.screenWidth,
      height: responsive.screenHeight,
    },
    safeArea: {
      flex: 1,
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
        : responsive.screenHeight * 0.1,
    },
    bottomSpacer: {
      height: responsive.spacing.vertical * 2,
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
      marginTop: -6,
      marginBottom: 6,
    },
    errorColumn: {
      flex: 1,
      paddingHorizontal: 4,
    },
    errorMessage: {
      fontSize: responsive.fontSizes.caption,
      color: theme.colors.error,
      marginTop: 4,
    },
    termsContainer: {
      marginTop: responsive.spacing.elementSpacing,
      marginBottom: responsive.spacing.elementSpacing / 2,
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
      // Ensure minimum touch target
      minWidth: 24,
      minHeight: 24,
    },
    checkboxChecked: {
      backgroundColor: theme.colors.primary,
      borderColor: theme.colors.primary,
    },
    termsText: {
      flex: 1,
      fontSize: responsive.fontSizes.label,
      color: theme.colors.text.secondary,
      lineHeight: 20,
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
    buttonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.fontSizes.button,
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
    },
  });
