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
  Platform,
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
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

interface PasswordResetFormData {
  email: string;
}

const ForgotPasswordScreen: React.FC = () => {
  const navigation = useNavigation();
  const [emailSent, setEmailSent] = useState(false);
  const { t } = useTranslation();
  // Always use light theme for auth screens
  const theme = authLightTheme;
  const {
    isTablet,
    contentWidth,
    width: screenWidth,
    height: screenHeight,
  } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(
    theme,
    isTablet,
    contentWidth,
    spacing,
    screenWidth,
    screenHeight,
  );

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
          <View
            style={[
              styles.successContainer,
              isTablet && styles.tabletSuccessContainer,
            ]}
          >
            <LinearGradient
              colors={["rgba(255, 255, 255, 0.95)", "rgba(255, 255, 255, 0.9)"]}
              style={[styles.successCard, isTablet && styles.tabletSuccessCard]}
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

              <Text style={styles.successTitle}>
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
          behavior={Platform.OS === "ios" ? "padding" : "height"}
          style={styles.keyboardView}
        >
          <ScrollView
            contentContainerStyle={styles.scrollContent}
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
          >
            {/* Header */}
            <View style={styles.header}>
              <TouchableOpacity
                style={styles.backButton}
                onPress={handleBackToLogin}
              >
                <Ionicons
                  name="arrow-back"
                  size={24}
                  color={theme.colors.text.primary}
                />
              </TouchableOpacity>
            </View>

            {/* Icon */}
            <View style={styles.iconContainer}>
              <View style={styles.iconBackground}>
                <MaterialIcons
                  name="lock-reset"
                  size={60}
                  color={theme.colors.primary}
                />
              </View>
            </View>

            {/* Form Container */}
            <View
              style={[
                styles.formContainer,
                isTablet && styles.tabletFormContainer,
              ]}
            >
              <LinearGradient
                colors={[
                  "rgba(255, 255, 255, 0.95)",
                  "rgba(255, 255, 255, 0.9)",
                ]}
                style={[styles.formCard, isTablet && styles.tabletFormCard]}
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

const createStyles = (
  theme: Theme,
  isTablet: boolean,
  contentWidth: number,
  spacing: ReturnType<typeof useResponsiveSpacing>,
  screenWidth: number,
  screenHeight: number,
) =>
  StyleSheet.create({
    backgroundImage: {
      flex: 1,
      width: screenWidth,
      height: screenHeight,
    },
    container: {
      flex: 1,
    },
    keyboardView: {
      flex: 1,
    },
    scrollContent: {
      flexGrow: 1,
      paddingBottom: spacing.lg,
      alignItems: isTablet ? "center" : "stretch",
    },
    header: {
      paddingHorizontal: spacing.lg,
      paddingTop: spacing.lg,
      paddingBottom: spacing.lg,
      width: isTablet ? contentWidth : "100%",
      maxWidth: isTablet ? 480 : undefined,
    },
    backButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: "rgba(255, 255, 255, 0.9)",
      justifyContent: "center",
      alignItems: "center",
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.1,
      shadowRadius: 4,
      elevation: 3,
    },
    iconContainer: {
      alignItems: "center",
      marginVertical: spacing.xl,
      width: isTablet ? contentWidth : "100%",
      maxWidth: isTablet ? 480 : undefined,
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
      paddingHorizontal: spacing.lg,
      flex: 1,
      width: isTablet ? contentWidth : "100%",
      maxWidth: isTablet ? 480 : undefined,
    },
    tabletFormContainer: {
      alignItems: "center",
    },
    formCard: {
      borderRadius: theme.borderRadius.xl,
      padding: spacing.lg,
      width: "100%",
      ...theme.shadows.md,
    },
    tabletFormCard: {
      padding: spacing.xl,
      maxWidth: 480,
    },
    title: {
      fontSize: isTablet ? 28 : 24,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: spacing.sm,
      textAlign: "center",
    },
    subtitle: {
      fontSize: isTablet ? 18 : 16,
      color: theme.colors.text.secondary,
      marginBottom: spacing.xl,
      textAlign: "center",
      lineHeight: isTablet ? 28 : 24,
    },
    inputGroup: {
      marginBottom: spacing.lg,
    },
    inputLabelRow: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: spacing.sm,
      gap: spacing.xs,
    },
    inputLabel: {
      fontSize: isTablet ? 16 : 14,
      fontWeight: "500",
      color: theme.colors.text.primary,
    },
    input: {
      backgroundColor: theme.colors.ui.inputBackground,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      borderRadius: theme.borderRadius.md,
      paddingHorizontal: spacing.md,
      paddingVertical: spacing.md,
      fontSize: isTablet ? 18 : 16,
      color: theme.colors.text.primary,
      minHeight: isTablet ? 52 : 48,
    },
    inputError: {
      borderColor: theme.colors.error,
    },
    errorText: {
      color: theme.colors.error,
      fontSize: isTablet ? 14 : 12,
      marginTop: spacing.xs,
    },
    submitButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: spacing.md,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: spacing.sm,
      marginBottom: spacing.lg,
      minHeight: isTablet ? 56 : 48,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
    },
    submitButtonText: {
      color: theme.colors.text.inverse,
      fontSize: isTablet ? 18 : 16,
      fontWeight: "600",
    },
    buttonDisabled: {
      opacity: 0.7,
    },
    infoContainer: {
      flexDirection: "row",
      backgroundColor: theme.colors.state.infoBackground,
      borderRadius: theme.borderRadius.md,
      padding: spacing.md,
      marginBottom: spacing.xl,
      gap: spacing.sm,
    },
    infoText: {
      flex: 1,
      fontSize: isTablet ? 16 : 14,
      color: theme.colors.text.secondary,
      lineHeight: isTablet ? 24 : 20,
    },
    textButton: {
      alignItems: "center",
      paddingVertical: spacing.sm,
      minHeight: isTablet ? 48 : 44,
      justifyContent: "center",
    },
    textButtonText: {
      fontSize: isTablet ? 16 : 14,
      color: theme.colors.text.secondary,
    },
    linkText: {
      color: theme.colors.primary,
      fontWeight: "600",
    },
    bottomSpacer: {
      height: isTablet ? spacing.xxl : spacing.xl,
    },
    // Success state styles
    successContainer: {
      flex: 1,
      justifyContent: "center",
      paddingHorizontal: spacing.lg,
      alignItems: isTablet ? "center" : "stretch",
    },
    tabletSuccessContainer: {
      paddingHorizontal: spacing.xl,
    },
    successCard: {
      borderRadius: theme.borderRadius.xl,
      padding: spacing.xl,
      alignItems: "center",
      width: isTablet ? Math.min(contentWidth * 0.7, 600) : "100%",
      ...theme.shadows.lg,
    },
    tabletSuccessCard: {
      padding: spacing.xxl,
      maxWidth: 600,
    },
    successIconContainer: {
      marginBottom: spacing.xl,
    },
    successTitle: {
      fontSize: isTablet ? 28 : 24,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: spacing.sm,
      textAlign: "center",
    },
    successMessage: {
      fontSize: isTablet ? 18 : 16,
      color: theme.colors.text.secondary,
      marginBottom: spacing.xs,
      textAlign: "center",
      lineHeight: isTablet ? 26 : 24,
    },
    emailText: {
      fontSize: isTablet ? 18 : 16,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: spacing.xl,
      textAlign: "center",
    },
    instructionText: {
      fontSize: isTablet ? 16 : 14,
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: spacing.xl,
      lineHeight: isTablet ? 24 : 20,
    },
    resendButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      paddingVertical: spacing.md,
      paddingHorizontal: spacing.xl,
      borderRadius: theme.borderRadius.md,
      borderWidth: 1,
      borderColor: theme.colors.primary,
      backgroundColor: "transparent",
      gap: spacing.sm,
      marginBottom: spacing.lg,
      minHeight: isTablet ? 48 : 44,
    },
    resendButtonText: {
      color: theme.colors.primary,
      fontSize: isTablet ? 18 : 16,
      fontWeight: "500",
    },
    divider: {
      height: 1,
      backgroundColor: theme.colors.ui.border,
      marginVertical: spacing.lg,
      width: "100%",
    },
    backToLoginButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: spacing.sm,
      paddingVertical: spacing.sm,
      minHeight: isTablet ? 48 : 44,
    },
    backToLoginText: {
      color: theme.colors.text.primary,
      fontSize: isTablet ? 18 : 16,
      fontWeight: "500",
    },
  });
