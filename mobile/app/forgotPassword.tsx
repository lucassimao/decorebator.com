import * as usersApi from "@/api/users";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useMutation } from "@tanstack/react-query";
import { LinearGradient } from "expo-linear-gradient";
import React, { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Alert,
  ImageBackground,
  KeyboardAvoidingView,
  SafeAreaView,
  TextInput,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { authLightTheme } from "@/theme/authTheme";
import type { Theme } from "@/contexts/ThemeContext";
import { useTheme } from "@/contexts/ThemeContext";
import type { ResponsiveValues } from "@/contexts/ThemeContext";

interface PasswordResetFormData {
  email: string;
}

const ForgotPasswordScreen: React.FC = () => {
  const navigation = useNavigation();
  const [emailSent, setEmailSent] = useState(false);
  const { t } = useTranslation();
  // Always use light theme for auth screens
  const theme = authLightTheme;
  // Get responsive values from theme context
  const { responsive } = useTheme();
  const styles = createStyles(theme, responsive);

  const {
    control,
    handleSubmit,
    formState: { errors },
    watch,
  } = useForm<PasswordResetFormData>({
    defaultValues: {
      email: "",
    },
  });

  const watchedEmail = watch("email");

  // Password reset mutation
  const resetMutation = useMutation({
    mutationFn: usersApi.requestResetEmailPassword,
    onSuccess: () => {
      setEmailSent(true);
    },
    onError: (error: Error) => {
      Alert.alert(
        t("auth.forgotPassword.errorTitle"),
        error.message || t("auth.forgotPassword.errorMessage"),
        [{ text: t("common.ok") }],
      );
    },
  });

  const handlePasswordReset = (data: PasswordResetFormData) => {
    resetMutation.mutate(data.email);
  };

  const handleBackToLogin = () => {
    navigation.goBack();
  };

  const handleResendEmail = () => {
    if (watchedEmail) {
      resetMutation.mutate(watchedEmail);
    }
  };

  // Success state
  if (emailSent) {
    return (
      <ImageBackground
        source={require("@/assets/images/login-bg.png")}
        style={styles.backgroundImage}
        resizeMode="cover"
      >
        <SafeAreaView style={styles.container}>
          <View style={styles.successContainer}>
            <LinearGradient
              colors={["rgba(255, 255, 255, 0.95)", "rgba(255, 255, 255, 0.9)"]}
              style={styles.successCard}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
            >
              <View style={styles.successIconContainer}>
                <MaterialIcons
                  name="mark-email-read"
                  size={80}
                  color="#4CAF50"
                />
              </View>

              <Text testID="success-title" style={styles.successTitle}>
                {t("auth.forgotPassword.successTitle")}
              </Text>
              <Text style={styles.successMessage}>
                {t("auth.forgotPassword.successSubtitle")}
              </Text>
              <Text style={styles.emailText}>{watchedEmail}</Text>

              <Text style={styles.instructionText}>
                {t("auth.forgotPassword.instructionText")}
              </Text>

              <TouchableOpacity
                style={[
                  styles.resendButton,
                  resetMutation.isPending && styles.buttonDisabled,
                ]}
                onPress={handleResendEmail}
                disabled={resetMutation.isPending}
              >
                {resetMutation.isPending ? (
                  <ActivityIndicator
                    size="small"
                    color={theme.colors.primary}
                  />
                ) : (
                  <>
                    <MaterialIcons
                      name="refresh"
                      size={20}
                      color={theme.colors.primary}
                    />
                    <Text style={styles.resendButtonText}>
                      {t("auth.forgotPassword.resendButton")}
                    </Text>
                  </>
                )}
              </TouchableOpacity>

              <View style={styles.divider} />

              <TouchableOpacity
                style={styles.backToLoginButton}
                onPress={handleBackToLogin}
              >
                <Ionicons
                  name="arrow-back"
                  size={20}
                  color={theme.colors.text.primary}
                />
                <Text style={styles.backToLoginText}>
                  {t("auth.forgotPassword.backToSignIn")}
                </Text>
              </TouchableOpacity>
            </LinearGradient>
          </View>
        </SafeAreaView>
      </ImageBackground>
    );
  }

  // Reset form
  return (
    <ImageBackground
      source={require("@/assets/images/login-bg.png")}
      style={styles.backgroundImage}
      resizeMode="cover"
    >
      <SafeAreaView style={styles.container}>
        <KeyboardAvoidingView
          behavior={responsive.keyboardBehavior}
          style={styles.keyboardView}
          keyboardVerticalOffset={responsive.keyboardOffset}
        >
          <ScrollView
            contentContainerStyle={styles.scrollContent}
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
          >
            {/* Header */}
            <View style={styles.header}>
              <TouchableOpacity
                testID="back-button"
                style={styles.backButton}
                onPress={handleBackToLogin}
                activeOpacity={0.7}
                accessibilityRole="button"
                accessibilityLabel={t("common.back")}
                accessibilityHint={t(
                  "auth.forgotPassword.backHint",
                  "Return to sign in",
                )}
                hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
              >
                <Ionicons
                  name="arrow-back"
                  size={24}
                  color={theme.colors.text.primary}
                />
              </TouchableOpacity>
            </View>

            {/* Icon - Hidden on small devices */}
            {responsive.category !== "small" && (
              <View style={styles.iconContainer}>
                <View style={styles.iconBackground}>
                  <MaterialIcons
                    name="lock-reset"
                    size={60}
                    color={theme.colors.primary}
                  />
                </View>
              </View>
            )}

            {/* Form Container */}
            <View style={styles.formContainer}>
              <LinearGradient
                colors={[
                  "rgba(255, 255, 255, 0.95)",
                  "rgba(255, 255, 255, 0.9)",
                ]}
                style={styles.formCard}
                start={{ x: 0, y: 0 }}
                end={{ x: 1, y: 1 }}
              >
                <Text style={styles.title}>
                  {t("auth.forgotPassword.title")}
                </Text>
                <Text style={styles.subtitle}>
                  {t("auth.forgotPassword.subtitle")}
                </Text>

                {/* Email Input */}
                <View style={styles.inputGroup}>
                  <View style={styles.inputLabelRow}>
                    <MaterialIcons
                      name="email"
                      size={20}
                      color={theme.colors.text.secondary}
                    />
                    <Text style={styles.inputLabel}>
                      {t("auth.forgotPassword.email")}
                    </Text>
                  </View>
                  <Controller
                    control={control}
                    name="email"
                    rules={{
                      required: t("auth.forgotPassword.emailRequired"),
                      pattern: {
                        value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                        message: t("auth.forgotPassword.emailInvalid"),
                      },
                    }}
                    render={({ field: { onChange, onBlur, value } }) => (
                      <TextInput
                        testID="email-input"
                        style={[
                          styles.input,
                          errors.email && styles.inputError,
                        ]}
                        placeholder={t("auth.forgotPassword.emailPlaceholder")}
                        placeholderTextColor={theme.colors.text.placeholder}
                        value={value}
                        onChangeText={onChange}
                        onBlur={onBlur}
                        autoCapitalize="none"
                        keyboardType="email-address"
                        autoComplete="email"
                        editable={!resetMutation.isPending}
                      />
                    )}
                  />
                  {errors.email && (
                    <Text style={styles.errorText}>{errors.email.message}</Text>
                  )}
                </View>

                {/* Submit Button */}
                <TouchableOpacity
                  testID="send-reset-button"
                  style={[
                    styles.submitButton,
                    resetMutation.isPending && styles.buttonDisabled,
                  ]}
                  onPress={handleSubmit(handlePasswordReset)}
                  disabled={resetMutation.isPending}
                  activeOpacity={0.8}
                >
                  {resetMutation.isPending ? (
                    <ActivityIndicator
                      size="small"
                      color={theme.colors.text.inverse}
                    />
                  ) : (
                    <>
                      <Text style={styles.submitButtonText}>
                        {t("auth.forgotPassword.sendButton")}
                      </Text>
                      <MaterialIcons
                        name="send"
                        size={20}
                        color={theme.colors.text.inverse}
                      />
                    </>
                  )}
                </TouchableOpacity>

                {/* Info Message */}
                <View style={styles.infoContainer}>
                  <MaterialIcons
                    name="info-outline"
                    size={16}
                    color={theme.colors.text.secondary}
                  />
                  <Text style={styles.infoText}>
                    {t("auth.forgotPassword.securityNotice")}
                  </Text>
                </View>

                {/* Back to Login */}
                <TouchableOpacity
                  testID="remember-password-link"
                  style={styles.textButton}
                  onPress={handleBackToLogin}
                  disabled={resetMutation.isPending}
                >
                  <Text style={styles.textButtonText}>
                    {t("auth.forgotPassword.rememberPassword")}{" "}
                    <Text style={styles.linkText}>
                      {t("auth.forgotPassword.signIn")}
                    </Text>
                  </Text>
                </TouchableOpacity>
              </LinearGradient>
            </View>

            {/* Bottom Spacing */}
            <View style={styles.bottomSpacer} />
          </ScrollView>
        </KeyboardAvoidingView>
      </SafeAreaView>
    </ImageBackground>
  );
};
export default ForgotPasswordScreen;

const createStyles = (theme: Theme, responsive: ResponsiveValues) =>
  StyleSheet.create({
    backgroundImage: {
      flex: 1,
    },
    container: {
      flex: 1,
    },
    keyboardView: {
      flex: 1,
    },
    scrollContent: {
      flexGrow: 1,
      paddingBottom: responsive.spacing.vertical,
    },
    header: {
      paddingHorizontal: responsive.spacing.horizontal,
      paddingTop: responsive.spacing.vertical,
      paddingBottom: responsive.spacing.vertical,
    },
    backButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    iconContainer: {
      alignItems: "center",
      marginVertical: responsive.spacing.elementSpacing * 2,
    },
    iconBackground: {
      width: 120,
      height: 120,
      borderRadius: 60,
      backgroundColor: "rgba(255, 123, 84, 0.1)",
      justifyContent: "center",
      alignItems: "center",
    },
    formContainer: {
      paddingHorizontal: responsive.spacing.horizontal,
      flex: 1,
    },
    formCard: {
      borderRadius: theme.borderRadius.xl,
      padding: responsive.spacing.formPadding,
      ...theme.shadows.md,
    },
    title: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "700",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing / 2,
      textAlign: "center",
    },
    subtitle: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.secondary,
      marginBottom: responsive.spacing.elementSpacing * 2,
      textAlign: "center",
      lineHeight: responsive.fontSizes.body * 1.5,
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
    errorText: {
      color: theme.colors.error,
      fontSize: responsive.fontSizes.label,
      marginTop: 4,
    },
    submitButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: 8,
      marginTop: responsive.spacing.elementSpacing / 2,
      marginBottom: responsive.spacing.elementSpacing,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
      minHeight: responsive.spacing.buttonHeight,
    },
    submitButtonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
    },
    buttonDisabled: {
      backgroundColor: theme.colors.ui.disabled,
      opacity: 0.6,
    },
    infoContainer: {
      flexDirection: "row",
      backgroundColor: theme.colors.state.infoBackground,
      borderRadius: theme.borderRadius.md,
      padding: theme.spacing.md,
      marginBottom: responsive.spacing.elementSpacing,
      gap: 8,
    },
    infoText: {
      flex: 1,
      fontSize: responsive.fontSizes.label,
      color: theme.colors.text.secondary,
      lineHeight: responsive.fontSizes.label * 1.5,
    },
    textButton: {
      alignItems: "center",
      paddingVertical: theme.spacing.sm,
    },
    textButtonText: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.secondary,
    },
    linkText: {
      color: theme.colors.primary,
      fontWeight: "600",
      fontSize: responsive.fontSizes.headline,
    },
    bottomSpacer: {
      height: responsive.spacing.elementSpacing * 2,
    },
    // Success state styles
    successContainer: {
      flex: 1,
      justifyContent: "center",
      paddingHorizontal: responsive.spacing.horizontal,
    },
    successCard: {
      borderRadius: theme.borderRadius.xl,
      padding: responsive.spacing.formPadding,
      alignItems: "center",
      ...theme.shadows.lg,
    },
    successIconContainer: {
      marginBottom: responsive.spacing.elementSpacing * 1.5,
    },
    successTitle: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "700",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing / 2,
    },
    successMessage: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.secondary,
      marginBottom: theme.spacing.xs,
      textAlign: "center",
    },
    emailText: {
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing * 1.5,
    },
    instructionText: {
      fontSize: responsive.fontSizes.label,
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: responsive.spacing.elementSpacing * 1.5,
      lineHeight: responsive.fontSizes.label * 1.5,
    },
    resendButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      paddingVertical: theme.spacing.sm,
      paddingHorizontal: theme.spacing.lg,
      borderRadius: theme.borderRadius.md,
      borderWidth: 1,
      borderColor: theme.colors.primary,
      backgroundColor: "transparent",
      gap: 8,
      marginBottom: responsive.spacing.elementSpacing,
      minHeight: responsive.spacing.minTouchTarget,
    },
    resendButtonText: {
      color: theme.colors.primary,
      fontSize: responsive.fontSizes.body,
      fontWeight: "600",
    },
    divider: {
      height: 1,
      backgroundColor: theme.colors.ui.divider,
      marginVertical: responsive.spacing.elementSpacing,
      width: "100%",
    },
    backToLoginButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: 8,
      paddingVertical: theme.spacing.sm,
    },
    backToLoginText: {
      color: theme.colors.text.primary,
      fontSize: responsive.fontSizes.body,
      fontWeight: "600",
    },
  });
