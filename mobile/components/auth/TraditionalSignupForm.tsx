import React from "react";
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
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

interface TraditionalSignupFormProps {
  form: UseFormReturn<SignupFormData>;
  onSubmit: (data: SignupFormData) => void;
  theme: Theme;
  responsive: ResponsiveValues;
}

export const TraditionalSignupForm: React.FC<TraditionalSignupFormProps> = ({
  form,
  onSubmit,
  theme,
  responsive,
}) => {
  const { t } = useTranslation();
  const [secureTextEntry, setSecureTextEntry] = React.useState(true);

  const {
    control,
    handleSubmit,
    formState: { errors, isValid },
    watch,
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

  const isFormFilled = email && fullName && password;

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.headerSection}>
        <Text style={styles.title}>{t("auth.signup.title")}</Text>
        <Text style={styles.subtitle}>
          {t("auth.signup.subtitle", "Start your language learning journey")}
        </Text>
      </View>

      {/* Email Input */}
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
              <Text style={styles.inputLabel}>{t("auth.signup.email")}</Text>
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
              onSubmitEditing={() => fullNameInputRef.current?.focus()}
            />
            <ErrorMessage error={errors.email} />
          </View>
        )}
      />

      {/* Full Name Input */}
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
              <Text style={styles.inputLabel}>{t("auth.signup.fullName")}</Text>
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
              onSubmitEditing={() => passwordInputRef.current?.focus()}
            />
            <ErrorMessage error={errors.fullName} />
          </View>
        )}
      />

      {/* Password Input */}
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
              <Text style={styles.inputLabel}>{t("auth.signup.password")}</Text>
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

      {/* Submit Button */}
      <TouchableOpacity
        testID="signup-submit-button"
        style={[
          styles.button,
          (!isFormFilled || !isValid) && styles.buttonDisabled,
        ]}
        onPress={handleSubmit(onSubmit)}
        activeOpacity={0.8}
        disabled={!isFormFilled || !isValid}
      >
        <Text
          style={[
            styles.buttonText,
            (!isFormFilled || !isValid) && styles.buttonTextDisabled,
          ]}
        >
          {t("auth.signup.signUpButton")}
        </Text>
      </TouchableOpacity>

      {/* Terms */}
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
      fontSize: responsive.fontSizes.body,
    },
  });
