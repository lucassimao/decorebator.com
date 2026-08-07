import React from "react";
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  Animated,
  StyleSheet,
} from "react-native";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";
import { Controller, UseFormReturn } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import type { Theme } from "@/contexts/ThemeContext";
import type { ResponsiveValues } from "@/contexts/ThemeContext";
import * as WebBrowser from "expo-web-browser";
import { Link } from "expo-router";
import i18n from "@/i18n";

interface SignupFormData {
  email: string;
  fullName: string;
  password: string;
}

interface ProgressiveSignupFormProps {
  form: UseFormReturn<SignupFormData>;
  onSubmit: (data: SignupFormData) => void;
  theme: Theme;
  responsive: ResponsiveValues;
}

export const ProgressiveSignupForm: React.FC<ProgressiveSignupFormProps> = ({
  form,
  onSubmit,
  theme,
  responsive,
}) => {
  const { t } = useTranslation();
  const [currentStep, setCurrentStep] = React.useState<
    "email" | "name" | "password"
  >("email");
  const [secureTextEntry, setSecureTextEntry] = React.useState(true);

  // Animation values
  const fadeAnim = React.useRef(new Animated.Value(1)).current;

  const {
    control,
    handleSubmit,
    formState: { errors },
    watch,
    trigger,
    setError,
  } = form;

  // Watch field values
  const email = watch("email");
  const fullName = watch("fullName");
  const password = watch("password");

  // Refs for inputs
  const emailInputRef = React.useRef<TextInput>(null);
  const fullNameInputRef = React.useRef<TextInput>(null);
  const passwordInputRef = React.useRef<TextInput>(null);

  const styles = React.useMemo(
    () => createStyles(theme, responsive),
    [theme, responsive],
  );

  // Auto-focus first field
  React.useEffect(() => {
    setTimeout(() => {
      emailInputRef.current?.focus();
    }, 300);
  }, []);

  const animateTransition = (callback: () => void) => {
    Animated.timing(fadeAnim, {
      toValue: 0,
      duration: 150,
      useNativeDriver: true,
    }).start(() => {
      callback();
      Animated.timing(fadeAnim, {
        toValue: 1,
        duration: 150,
        useNativeDriver: true,
      }).start();
    });
  };

  const handleEmailNext = async () => {
    const isValid = await trigger("email");
    if (isValid) {
      animateTransition(() => {
        setCurrentStep("name");
        setTimeout(() => fullNameInputRef.current?.focus(), 100);
      });
    }
  };

  const handleNameNext = async () => {
    const isValid = await trigger("fullName");
    if (isValid) {
      animateTransition(() => {
        setCurrentStep("password");
        setTimeout(() => passwordInputRef.current?.focus(), 100);
      });
    } else if (fullName && !fullName.includes(" ")) {
      setError("fullName", {
        type: "manual",
        message: t("auth.signup.fullNameError", "Please enter your full name"),
      });
    }
  };

  const handleBack = () => {
    animateTransition(() => {
      if (currentStep === "password") {
        setCurrentStep("name");
        setTimeout(() => fullNameInputRef.current?.focus(), 100);
      } else if (currentStep === "name") {
        setCurrentStep("email");
        setTimeout(() => emailInputRef.current?.focus(), 100);
      }
    });
  };

  const isFormValid =
    email &&
    fullName &&
    password &&
    !errors.email &&
    !errors.fullName &&
    !errors.password;

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.headerSection}>
        {/* Back Button and Title Row */}
        <View style={styles.titleRow}>
          {currentStep !== "email" && (
            <TouchableOpacity
              testID="back-button"
              style={styles.backButton}
              onPress={handleBack}
              activeOpacity={0.7}
              accessibilityRole="button"
              accessibilityLabel={t("common.back")}
              accessibilityHint={t(
                "auth.signup.backHint",
                "Go back to previous step",
              )}
              hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
            >
              <Ionicons
                name="arrow-back"
                size={24}
                color={theme.colors.text.primary}
              />
            </TouchableOpacity>
          )}
          <Text
            style={[
              styles.title,
              currentStep !== "email" && styles.titleWithBackButton,
            ]}
          >
            {t("auth.signup.title")}
          </Text>
          {/* Spacer to balance the layout when back button is visible */}
          {currentStep !== "email" && <View style={styles.spacer} />}
        </View>
        <Text style={styles.subtitle}>
          {t("auth.signup.subtitle", "Start your language learning journey")}
        </Text>

        {/* Step Indicator */}
        <View style={styles.stepIndicator}>
          <View style={[styles.stepDot, styles.stepDotActive]} />
          <View
            style={[
              styles.stepDot,
              (currentStep === "name" || currentStep === "password") &&
                styles.stepDotActive,
            ]}
          />
          <View
            style={[
              styles.stepDot,
              currentStep === "password" && styles.stepDotActive,
            ]}
          />
        </View>
      </View>

      <Animated.View style={{ opacity: fadeAnim }}>
        {/* Email Step */}
        {currentStep === "email" && (
          <Controller
            control={control}
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
                  testID="email-input"
                  style={[styles.input, errors.email && styles.inputError]}
                  placeholder={t("auth.signup.emailPlaceholder")}
                  placeholderTextColor={theme.colors.text.placeholder}
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  autoCapitalize="none"
                  keyboardType="email-address"
                  autoComplete="email"
                  textContentType="emailAddress"
                  returnKeyType="next"
                  onSubmitEditing={handleEmailNext}
                />
                <ErrorMessage error={errors.email} />
              </View>
            )}
          />
        )}

        {/* Name Step */}
        {currentStep === "name" && (
          <Controller
            control={control}
            name="fullName"
            render={({ field: { onChange, onBlur, value } }) => (
              <View style={styles.inputGroup}>
                <View style={styles.inputLabelRow}>
                  <MaterialIcons
                    name="person"
                    size={20}
                    color={theme.colors.text.secondary}
                  />
                  <Text style={styles.inputLabel}>
                    {t("auth.signup.fullName")}
                  </Text>
                </View>
                <TextInput
                  ref={fullNameInputRef}
                  testID="fullname-input"
                  style={[styles.input, errors.fullName && styles.inputError]}
                  placeholder={t("auth.signup.fullNamePlaceholder")}
                  placeholderTextColor={theme.colors.text.placeholder}
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  autoCapitalize="words"
                  textContentType="name"
                  returnKeyType="next"
                  onSubmitEditing={handleNameNext}
                />
                <ErrorMessage error={errors.fullName} />
              </View>
            )}
          />
        )}

        {/* Password Step */}
        {currentStep === "password" && (
          <Controller
            control={control}
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
                    testID="password-input"
                    style={[
                      styles.input,
                      styles.passwordInput,
                      errors.password && styles.inputError,
                    ]}
                    placeholder={t("auth.signup.passwordPlaceholder")}
                    maxLength={72}
                    placeholderTextColor={theme.colors.text.placeholder}
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    secureTextEntry={secureTextEntry}
                    autoComplete="password-new"
                    textContentType="newPassword"
                    returnKeyType="done"
                    onSubmitEditing={handleSubmit(onSubmit)}
                  />
                  <TouchableOpacity
                    testID="password-toggle"
                    style={styles.passwordToggle}
                    onPress={() => setSecureTextEntry(!secureTextEntry)}
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
        )}
      </Animated.View>

      {/* Next/Submit Button */}
      {currentStep === "email" && (
        <TouchableOpacity
          testID="email-next-button"
          style={[styles.button, !email && styles.buttonDisabled]}
          onPress={handleEmailNext}
          activeOpacity={0.8}
          disabled={!email}
        >
          <Text
            style={[styles.buttonText, !email && styles.buttonTextDisabled]}
          >
            {t("common.next")}
          </Text>
        </TouchableOpacity>
      )}

      {currentStep === "name" && (
        <TouchableOpacity
          testID="name-next-button"
          style={[
            styles.button,
            (!fullName || fullName.length < 2) && styles.buttonDisabled,
          ]}
          onPress={handleNameNext}
          activeOpacity={0.8}
          disabled={!fullName || fullName.length < 2}
        >
          <Text
            style={[
              styles.buttonText,
              (!fullName || fullName.length < 2) && styles.buttonTextDisabled,
            ]}
          >
            {t("common.next")}
          </Text>
        </TouchableOpacity>
      )}

      {currentStep === "password" && (
        <TouchableOpacity
          testID="signup-submit-button"
          style={[styles.button, !isFormValid && styles.buttonDisabled]}
          onPress={handleSubmit(onSubmit)}
          activeOpacity={0.8}
          disabled={!isFormValid}
        >
          <Text
            style={[
              styles.buttonText,
              !isFormValid && styles.buttonTextDisabled,
            ]}
          >
            {t("auth.signup.signUpButton")}
          </Text>
        </TouchableOpacity>
      )}

      {/* Terms - Only show on password step */}
      {currentStep === "password" && (
        <View style={styles.termsContainer}>
          <Text style={styles.implicitTermsText}>
            {t("auth.signup.termsText", "By signing up, you agree to our")}{" "}
            <Link
              href={`https://decorebator.com/${i18n.language}/terms`}
              style={styles.termsLink}
              onPress={async (e) => {
                e.preventDefault();
                await WebBrowser.openBrowserAsync(
                  `https://decorebator.com/${i18n.language}/terms`,
                );
              }}
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
            >
              {t("auth.signup.privacyPolicy", "Privacy Policy")}
            </Link>
            .
          </Text>
        </View>
      )}

      {/* Footer */}
      <Text style={styles.footer}>
        {t("auth.signup.alreadyHaveAccount")}{" "}
        <Link replace style={styles.link} href={"/signin"}>
          <Text style={styles.link}>{t("auth.signup.signIn")}</Text>
        </Link>
      </Text>
    </View>
  );
};

const createStyles = (theme: Theme, responsive: ResponsiveValues) =>
  StyleSheet.create({
    container: {
      flex: 1,
    },
    headerSection: {
      alignItems: "center",
      marginBottom: responsive.spacing.elementSpacing * 1.5,
    },
    titleRow: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      width: "100%",
      marginBottom: responsive.spacing.elementSpacing / 2,
    },
    backButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
      position: "absolute",
      left: 0,
    },
    title: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
    },
    titleWithBackButton: {
      // No need to adjust since we're using absolute positioning for the button
    },
    spacer: {
      width: 40, // Same width as back button for balance
    },
    subtitle: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    stepIndicator: {
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      marginTop: responsive.spacing.elementSpacing,
      gap: theme.spacing.sm,
    },
    stepDot: {
      width: 8,
      height: 8,
      borderRadius: 4,
      backgroundColor: theme.colors.ui.divider,
    },
    stepDotActive: {
      backgroundColor: theme.colors.primary,
      width: 24,
      height: 8,
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
    button: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md,
      alignItems: "center",
      marginTop: responsive.spacing.elementSpacing / 2, // Reduced from full elementSpacing
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
      minHeight: responsive.spacing.buttonHeight,
      justifyContent: "center",
    },
    buttonDisabled: {
      backgroundColor: theme.colors.ui.disabled,
      opacity: 0.6,
    },
    buttonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
    },
    buttonTextDisabled: {
      color: theme.colors.text.disabled,
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
    footer: {
      textAlign: "center",
      marginTop: responsive.spacing.elementSpacing * 2,
      color: theme.colors.text.secondary,
      fontSize: responsive.fontSizes.body,
    },
    link: {
      color: theme.colors.primary,
      fontWeight: "600",
      fontSize: responsive.fontSizes.headline, // Increased from body to headline
    },
  });
