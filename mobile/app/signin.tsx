import * as usersApi from "@/api/users";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useMutation } from "@tanstack/react-query";
import { LinearGradient } from "expo-linear-gradient";
import { router } from "expo-router";
import { usePostHog } from "posthog-react-native";
import React, { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  Image,
  ImageBackground,
  KeyboardAvoidingView,
  Platform,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from "react-native";
import { authLightTheme } from "@/theme/authTheme";
import type { Theme } from "@/contexts/ThemeContext";
import * as Sentry from "@sentry/react-native";
import { useResponsive } from "@/hooks/useResponsive";
import { createCommonStyles } from "@/styles/common";

const { width: SCREEN_WIDTH, height: SCREEN_HEIGHT } = Dimensions.get("window");

interface LoginFormData {
  email: string;
  password: string;
}

const LoginScreen: React.FC = () => {
  const [showPassword, setShowPassword] = useState(false);
  const { t } = useTranslation();
  const posthog = usePostHog();
  // Always use light theme for auth screens
  const theme = authLightTheme;
  const { isTablet, type: deviceType } = useResponsive();
  const commonStyles = createCommonStyles(theme);
  const styles = createStyles(theme, isTablet, deviceType);

  const {
    control,
    handleSubmit,
    formState: { errors },
    setError,
  } = useForm<LoginFormData>({
    defaultValues: {
      email: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_EMAIL : "",
      password: __DEV__ ? process.env.EXPO_PUBLIC_TEST_USER_PASSWORD : "",
    },
  });

  // Login mutation
  const loginMutation = useMutation({
    mutationFn: usersApi.signin,
    onSuccess: (_, variables) => {
      posthog.capture("user_signed_in", {
        email: variables.email,
      });
      Sentry.setUser({
        // id: variables.id,
        email: variables.email,
        // username: variables.name,
      });
      router.replace("/dashboard");
    },
    onError: (error: Error) => {
      if (error.message === usersApi.SIGN_IN_ERROR) {
        setError("email", { message: t("auth.signin.invalidCredentials") });
        setError("password", { message: t("auth.signin.invalidCredentials") });
      } else {
        Alert.alert(t("common.error"), t("auth.signin.somethingWentWrong"));
      }
    },
  });

  const handleLogin = (data: LoginFormData) => {
    loginMutation.mutate(data);
  };

  const handleForgotPassword = () => {
    router.push("/forgotPassword");
  };

  const handleSignUp = () => {
    router.replace("/signup");
  };

  return (
    <ImageBackground
      source={require("@/assets/images/login-bg.png")} // Background illustration
      style={styles.backgroundImage}
      resizeMode="cover"
    >
      <SafeAreaView style={[commonStyles.safeArea, styles.container]}>
        <View style={commonStyles.responsiveContainer}>
          <KeyboardAvoidingView
            behavior={Platform.OS === "ios" ? "padding" : "height"}
            style={styles.keyboardView}
          >
            <ScrollView
              contentContainerStyle={styles.scrollContent}
              showsVerticalScrollIndicator={false}
              keyboardShouldPersistTaps="handled"
            >
              {/* Logo/App Name */}
              <View style={styles.logoContainer}>
                <Text style={styles.appName}>{t("common.appName")}</Text>
                <Text style={styles.tagline}>{t("auth.tagline")}</Text>
              </View>

              {/* Foreground Illustration - Hidden on tablets to save space */}
              {!isTablet && (
                <View style={styles.illustrationContainer}>
                  <Image
                    source={require("@/assets/images/login-fg.png")} // Foreground illustration
                    style={styles.illustration}
                    resizeMode="contain"
                  />
                </View>
              )}

              {/* Login Form */}
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
                  <Text style={styles.welcomeText}>
                    {t("auth.signin.welcomeBack")}
                  </Text>
                  <Text style={styles.subtitleText}>
                    {t("auth.signin.subtitle")}
                  </Text>

                  {/* Email Input */}
                  <View style={styles.inputGroup}>
                    <View style={styles.inputLabelRow}>
                      <MaterialIcons name="email" size={20} color="#636E72" />
                      <Text style={styles.inputLabel}>
                        {t("auth.signin.email")}
                      </Text>
                    </View>
                    <Controller
                      control={control}
                      name="email"
                      rules={{
                        required: t("errors.emailRequired"),
                        pattern: {
                          value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                          message: t("errors.invalidEmail"),
                        },
                      }}
                      render={({ field: { onChange, onBlur, value } }) => (
                        <TextInput
                          style={[
                            styles.input,
                            errors.email && styles.inputError,
                          ]}
                          placeholder={t("auth.signin.emailPlaceholder")}
                          placeholderTextColor="#B2BEC3"
                          value={value}
                          onChangeText={onChange}
                          onBlur={onBlur}
                          autoCapitalize="none"
                          keyboardType="email-address"
                          autoComplete="email"
                          editable={!loginMutation.isPending}
                        />
                      )}
                    />
                    {errors.email && (
                      <Text style={styles.errorText}>
                        {errors.email.message}
                      </Text>
                    )}
                  </View>

                  {/* Password Input */}
                  <View style={styles.inputGroup}>
                    <View style={styles.inputLabelRow}>
                      <MaterialIcons name="lock" size={20} color="#636E72" />
                      <Text style={styles.inputLabel}>
                        {t("auth.signin.password")}
                      </Text>
                    </View>
                    <View style={styles.passwordContainer}>
                      <Controller
                        control={control}
                        name="password"
                        rules={{
                          required: t("errors.passwordRequired"),
                        }}
                        render={({ field: { onChange, onBlur, value } }) => (
                          <TextInput
                            style={[
                              styles.input,
                              styles.passwordInput,
                              errors.password && styles.inputError,
                            ]}
                            placeholder={t("auth.signin.passwordPlaceholder")}
                            placeholderTextColor="#B2BEC3"
                            value={value}
                            onChangeText={onChange}
                            onBlur={onBlur}
                            secureTextEntry={!showPassword}
                            autoComplete="password"
                            editable={!loginMutation.isPending}
                          />
                        )}
                      />
                      <TouchableOpacity
                        style={styles.passwordToggle}
                        onPress={() => setShowPassword(!showPassword)}
                      >
                        <Ionicons
                          name={showPassword ? "eye-off" : "eye"}
                          size={20}
                          color="#636E72"
                        />
                      </TouchableOpacity>
                    </View>
                    {errors.password && (
                      <Text style={styles.errorText}>
                        {errors.password.message}
                      </Text>
                    )}
                  </View>

                  {/* Forgot Password */}
                  <TouchableOpacity
                    style={styles.forgotPassword}
                    onPress={handleForgotPassword}
                    disabled={loginMutation.isPending}
                  >
                    <Text style={styles.forgotPasswordText}>
                      {t("auth.signin.forgotPassword")}
                    </Text>
                  </TouchableOpacity>

                  {/* Login Button */}
                  <TouchableOpacity
                    style={[
                      styles.loginButton,
                      loginMutation.isPending && styles.buttonDisabled,
                    ]}
                    onPress={handleSubmit(handleLogin)}
                    disabled={loginMutation.isPending}
                    activeOpacity={0.8}
                  >
                    {loginMutation.isPending ? (
                      <ActivityIndicator size="small" color="#FFFFFF" />
                    ) : (
                      <>
                        <Text style={styles.loginButtonText}>
                          {t("auth.signin.signInButton")}
                        </Text>
                        <Ionicons
                          name="arrow-forward"
                          size={20}
                          color={theme.colors.text.inverse}
                        />
                      </>
                    )}
                  </TouchableOpacity>

                  {/* Divider */}
                  <View style={styles.dividerContainer}>
                    <View style={styles.divider} />
                    <Text style={styles.dividerText}>{t("common.or")}</Text>
                    <View style={styles.divider} />
                  </View>

                  {/* Social Login Options */}
                  {/* <View style={styles.socialContainer}>
                  <TouchableOpacity style={styles.socialButton}>
                    <Image
                      source={require('../assets/google-icon.png')}
                      style={styles.socialIcon}
                    />
                    <Text style={styles.socialButtonText}>Continue with Google</Text>
                  </TouchableOpacity>

                  <TouchableOpacity style={styles.socialButton}>
                    <Ionicons name="logo-apple" size={24} color="#000000" />
                    <Text style={styles.socialButtonText}>Continue with Apple</Text>
                  </TouchableOpacity>
                </View> */}

                  {/* Sign Up Link */}
                  <View style={styles.signUpContainer}>
                    <Text style={styles.signUpText}>
                      {t("auth.signin.noAccount")}{" "}
                    </Text>
                    <TouchableOpacity
                      onPress={handleSignUp}
                      disabled={loginMutation.isPending}
                    >
                      <Text style={styles.signUpLink}>
                        {t("auth.signin.signUp")}
                      </Text>
                    </TouchableOpacity>
                  </View>
                </LinearGradient>
              </View>
            </ScrollView>
          </KeyboardAvoidingView>
        </View>
      </SafeAreaView>
    </ImageBackground>
  );
};

export default LoginScreen;

const createStyles = (theme: Theme, isTablet: boolean, deviceType: string) =>
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
    logoContainer: {
      alignItems: "center",
      marginTop: isTablet ? theme.spacing.section : theme.spacing.xl,
      marginBottom: isTablet ? theme.spacing.lg : theme.spacing.md,
    },
    appName: {
      fontSize: theme.typography.sizes.display,
      fontWeight: theme.typography.weights.bold,
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.xs,
    },
    tagline: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.secondary,
    },
    illustrationContainer: {
      height: SCREEN_HEIGHT * 0.15,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: 10,
    },
    illustration: {
      width: SCREEN_WIDTH * 0.8,
      height: "100%",
      maxWidth: 350,
    },
    formContainer: {
      paddingHorizontal: 0, // ResponsiveContainer handles this
      flex: 1,
      justifyContent: isTablet ? "center" : "flex-start",
      alignItems: isTablet ? "center" : "stretch",
    },
    formCard: {
      borderRadius: theme.borderRadius.xl,
      padding: theme.spacing.modal,
      maxWidth: isTablet ? theme.layout.modalMaxWidth : "100%",
      width: isTablet ? "auto" : "100%",
      alignSelf: isTablet ? "center" : "stretch",
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.1,
      shadowRadius: 12,
      elevation: 5,
    },
    welcomeText: {
      fontSize: theme.typography.sizes.title,
      fontWeight: theme.typography.weights.semibold,
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.xs,
      textAlign: "center",
    },
    subtitleText: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.secondary,
      marginBottom: theme.spacing.lg,
      textAlign: "center",
    },
    inputGroup: {
      marginBottom: theme.spacing.md,
    },
    inputLabelRow: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: theme.spacing.xs,
      gap: theme.spacing.xs,
    },
    inputLabel: {
      fontSize: theme.typography.sizes.caption,
      fontWeight: theme.typography.weights.medium,
      color: theme.colors.text.primary,
    },
    input: {
      backgroundColor: theme.colors.ui.inputBackground,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      borderRadius: theme.borderRadius.md,
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.md,
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.primary,
      minHeight: theme.touchTargets.comfortable,
    },
    inputError: {
      borderColor: theme.colors.error,
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
    errorText: {
      color: theme.colors.error,
      fontSize: theme.typography.sizes.caption,
      marginTop: theme.spacing.xs,
    },
    forgotPassword: {
      alignSelf: "flex-end",
      marginBottom: theme.spacing.lg,
      marginTop: -theme.spacing.xs,
      minHeight: theme.touchTargets.minimum,
      justifyContent: "center",
    },
    forgotPasswordText: {
      color: theme.colors.primary,
      fontSize: theme.typography.sizes.caption,
      fontWeight: theme.typography.weights.medium,
    },
    loginButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: theme.spacing.xs,
      marginBottom: theme.spacing.lg,
      minHeight: theme.touchTargets.comfortable,
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.2,
      shadowRadius: 8,
      elevation: 5,
    },
    loginButtonText: {
      color: theme.colors.text.inverse,
      fontSize: theme.typography.sizes.body,
      fontWeight: theme.typography.weights.semibold,
    },
    buttonDisabled: {
      opacity: 0.7,
    },
    dividerContainer: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: theme.spacing.lg,
    },
    divider: {
      flex: 1,
      height: 1,
      backgroundColor: theme.colors.ui.divider,
    },
    dividerText: {
      marginHorizontal: theme.spacing.md,
      color: theme.colors.text.secondary,
      fontSize: theme.typography.sizes.caption,
    },
    socialContainer: {
      gap: 12,
      marginBottom: 24,
    },
    socialButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      backgroundColor: theme.colors.background.surface,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      borderRadius: 12,
      paddingVertical: 14,
      gap: 12,
    },
    socialIcon: {
      width: 24,
      height: 24,
    },
    socialButtonText: {
      fontSize: 16,
      color: theme.colors.text.primary,
      fontWeight: "500",
    },
    signUpContainer: {
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      minHeight: theme.touchTargets.minimum,
    },
    signUpText: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
    },
    signUpLink: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.primary,
      fontWeight: theme.typography.weights.semibold,
    },
  });
