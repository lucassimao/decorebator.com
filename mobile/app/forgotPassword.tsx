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
  Dimensions,
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

const { width: SCREEN_WIDTH, height: SCREEN_HEIGHT } = Dimensions.get("window");

interface PasswordResetFormData {
  email: string;
}

const ForgotPasswordScreen: React.FC = () => {
  const navigation = useNavigation();
  const [emailSent, setEmailSent] = useState(false);
  const { t } = useTranslation();
  // Always use light theme for auth screens
  const theme = authLightTheme;
  const styles = createStyles(theme);

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

const createStyles = (theme: Theme) =>
  StyleSheet.create({
    backgroundImage: {
      flex: 1,
      width: SCREEN_WIDTH,
      height: SCREEN_HEIGHT,
    },
    container: {
      flex: 1,
    },
    keyboardView: {
      flex: 1,
    },
    scrollContent: {
      flexGrow: 1,
      paddingBottom: 20,
    },
    header: {
      paddingHorizontal: 20,
      paddingTop: 20,
      paddingBottom: 20,
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
      marginVertical: 30,
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
      paddingHorizontal: 20,
      flex: 1,
    },
    formCard: {
      borderRadius: theme.borderRadius.xl,
      padding: theme.spacing.lg,
      ...theme.shadows.md,
    },
    title: {
      fontSize: 24,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 12,
      textAlign: "center",
    },
    subtitle: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      marginBottom: 32,
      textAlign: "center",
      lineHeight: 24,
    },
    inputGroup: {
      marginBottom: 24,
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
      borderRadius: 12,
      paddingHorizontal: 16,
      paddingVertical: 14,
      fontSize: 16,
      color: theme.colors.text.primary,
    },
    inputError: {
      borderColor: theme.colors.error,
    },
    errorText: {
      color: theme.colors.error,
      fontSize: 12,
      marginTop: 4,
    },
    submitButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: 12,
      paddingVertical: 16,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: 8,
      marginBottom: 20,
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.2,
      shadowRadius: 8,
      elevation: 5,
    },
    submitButtonText: {
      color: theme.colors.text.inverse,
      fontSize: 16,
      fontWeight: "600",
    },
    buttonDisabled: {
      opacity: 0.7,
    },
    infoContainer: {
      flexDirection: "row",
      backgroundColor: theme.colors.state.infoBackground,
      borderRadius: 12,
      padding: 16,
      marginBottom: 24,
      gap: 8,
    },
    infoText: {
      flex: 1,
      fontSize: 14,
      color: theme.colors.text.secondary,
      lineHeight: 20,
    },
    textButton: {
      alignItems: "center",
      paddingVertical: 8,
    },
    textButtonText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
    linkText: {
      color: theme.colors.primary,
      fontWeight: "600",
    },
    bottomSpacer: {
      height: 40,
    },
    // Success state styles
    successContainer: {
      flex: 1,
      justifyContent: "center",
      paddingHorizontal: 20,
    },
    successCard: {
      borderRadius: 24,
      padding: 32,
      alignItems: "center",
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.1,
      shadowRadius: 12,
      elevation: 5,
    },
    successIconContainer: {
      marginBottom: 24,
    },
    successTitle: {
      fontSize: 24,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 12,
    },
    successMessage: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      marginBottom: 8,
      textAlign: "center",
    },
    emailText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 24,
    },
    instructionText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: 24,
      lineHeight: 20,
    },
    resendButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      paddingVertical: 12,
      paddingHorizontal: 24,
      borderRadius: 12,
      borderWidth: 1,
      borderColor: theme.colors.primary,
      backgroundColor: "transparent",
      gap: 8,
      marginBottom: 20,
    },
    resendButtonText: {
      color: theme.colors.primary,
      fontSize: 16,
      fontWeight: "500",
    },
    divider: {
      height: 1,
      backgroundColor: theme.colors.ui.border,
      marginVertical: 20,
      width: "100%",
    },
    backToLoginButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: 8,
    },
    backToLoginText: {
      color: theme.colors.text.primary,
      fontSize: 16,
      fontWeight: "500",
    },
  });
