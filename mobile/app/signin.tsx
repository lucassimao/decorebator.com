import * as usersApi from "@/api/users";
import { LoginHeader } from "@/components/auth/LoginHeader";
import { EmailInput } from "@/components/auth/EmailInput";
import { PasswordInput } from "@/components/auth/PasswordInput";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { router } from "expo-router";
import { usePostHog } from "posthog-react-native";
import React, { useState, useEffect, useRef } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Alert,
  ImageBackground,
  Keyboard,
  KeyboardAvoidingView,
  KeyboardEvent,
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
import { Ionicons } from "@expo/vector-icons";
import { authLightTheme } from "@/theme/authTheme";
import type { Theme } from "@/contexts/ThemeContext";
import { useTheme } from "@/contexts/ThemeContext";
import type { ResponsiveValues } from "@/contexts/ThemeContext";
import { decode } from "@/api/jwt";
import * as Sentry from "@sentry/react-native";
interface LoginFormData {
  email: string;
  password: string;
}

const LoginScreen: React.FC = () => {
  const [showPassword, setShowPassword] = useState(false);
  const [keyboardVisible, setKeyboardVisible] = useState(false);
  const [keyboardHeight, setKeyboardHeight] = useState(0);
  const { t } = useTranslation();
  const posthog = usePostHog();
  const queryClient = useQueryClient();

  // Refs for input focus management
  const passwordInputRef = useRef<TextInput>(null);

  // Always use light theme for auth screens
  const theme = authLightTheme;

  // Get responsive values from the unified theme context
  const { responsive } = useTheme();

  // Memoize styles to prevent recreation on every render
  const styles = React.useMemo(
    () => createStyles(theme, responsive, keyboardVisible, keyboardHeight),
    [theme, responsive, keyboardVisible, keyboardHeight],
  );

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
    onSuccess: async (_, variables) => {
      posthog.capture("user_signed_in", {
        email: variables.email,
      });

      // Extract user info from JWT for Sentry
      try {
        const authorization = usersApi.getAuthorization();
        if (authorization) {
          const decoded = decode(authorization);
          Sentry.setUser({
            id: decoded.payload.sub,
            email: decoded.payload.email,
          });
        }
      } catch (error) {
        console.error("Error setting Sentry user:", error);
        // Fallback to email only if JWT decode fails
        Sentry.setUser({
          email: variables.email,
        });
      }

      // Clear all cache to prevent data leakage between users
      queryClient.clear();

      // Pre-cache user data by invalidating the query - this will trigger a fresh fetch
      await queryClient.invalidateQueries({ queryKey: ["userProfile"] });

      router.replace("/");
    },
    onError: (error: Error) => {
      if (error.message === usersApi.SIGN_IN_ERROR) {
        setError("email", { message: t("auth.signin.invalidCredentials") });
        setError("password", { message: t("auth.signin.invalidCredentials") });
      } else {
        // Log unexpected errors to Sentry for debugging
        Sentry.captureException(error);
        Alert.alert(t("common.error"), t("auth.signin.somethingWentWrong"));
      }
    },
  });

  const handleLogin = (data: LoginFormData) => {
    loginMutation.mutate(data);
  };

  const handleSignUp = () => {
    router.replace("/onboarding");
  };

  const handleForgotPassword = () => {
    router.push("/forgotPassword");
  };

  // Focus handlers for input navigation
  const focusPasswordInput = () => {
    passwordInputRef.current?.focus();
  };

  // Keyboard listeners with dynamic height detection
  useEffect(() => {
    const keyboardWillShow = Keyboard.addListener(
      Platform.OS === "ios" ? "keyboardWillShow" : "keyboardDidShow",
      (event: KeyboardEvent) => {
        setKeyboardVisible(true);
        setKeyboardHeight(event.endCoordinates.height);
      },
    );

    const keyboardWillHide = Keyboard.addListener(
      Platform.OS === "ios" ? "keyboardWillHide" : "keyboardDidHide",
      () => {
        setKeyboardVisible(false);
        setKeyboardHeight(0);
      },
    );

    return () => {
      keyboardWillShow.remove();
      keyboardWillHide.remove();
    };
  }, []);

  return (
    <ImageBackground
      source={require("@/assets/images/signup-bg3.png")} // Background illustration
      style={styles.backgroundImage}
      imageStyle={styles.backgroundImageStyle}
      resizeMode="cover"
    >
      <SafeAreaView style={styles.safeArea}>
        <KeyboardAvoidingView
          style={styles.container}
          behavior={Platform.OS === "ios" ? "padding" : undefined}
          keyboardVerticalOffset={0}
        >
          <TouchableWithoutFeedback onPress={Keyboard.dismiss}>
            <View style={styles.innerContainer}>
              <ScrollView
                contentContainerStyle={styles.scrollContent}
                showsVerticalScrollIndicator={false}
                keyboardShouldPersistTaps="handled"
              >
                {/* Add some space at the top for the background to show */}
                <View style={styles.topSpacer} />

                {/* Header with Logo and Illustration */}
                <LoginHeader keyboardVisible={keyboardVisible} theme={theme} />

                {/* Login Form */}
                <View
                  style={styles.formCard}
                  // Accessibility
                  accessible={true}
                  accessibilityLabel={t("auth.signin.loginForm")}
                >
                  <Text testID="welcome-text" style={styles.welcomeText}>
                    {t("auth.signin.welcomeBack")}
                  </Text>
                  <Text style={styles.subtitleText}>
                    {t("auth.signin.subtitle")}
                  </Text>

                  {/* Email Input */}
                  <EmailInput
                    control={control}
                    errors={errors}
                    isPending={loginMutation.isPending}
                    theme={theme}
                    onSubmitEditing={focusPasswordInput}
                  />

                  {/* Password Input */}
                  <PasswordInput
                    ref={passwordInputRef}
                    control={control}
                    errors={errors}
                    showPassword={showPassword}
                    setShowPassword={setShowPassword}
                    isPending={loginMutation.isPending}
                    theme={theme}
                    onSubmitEditing={handleSubmit(handleLogin)}
                  />

                  {/* Sign In Button */}
                  <TouchableOpacity
                    testID="signin-button"
                    style={[
                      styles.loginButton,
                      loginMutation.isPending && styles.buttonDisabled,
                    ]}
                    onPress={handleSubmit(handleLogin)}
                    disabled={loginMutation.isPending}
                    activeOpacity={0.8}
                    // Accessibility
                    accessible={true}
                    accessibilityLabel={
                      loginMutation.isPending
                        ? t("auth.signin.signingIn")
                        : t("auth.signin.signInButton")
                    }
                    accessibilityRole="button"
                    accessibilityHint={t("auth.signin.signInHint")}
                    accessibilityState={{ disabled: loginMutation.isPending }}
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

                  {/* Bottom Links Row */}
                  <View style={styles.bottomLinksRow}>
                    {/* Sign Up Link - Left */}
                    <TouchableOpacity
                      testID="signup-link"
                      style={styles.bottomLinkLeft}
                      onPress={handleSignUp}
                      disabled={loginMutation.isPending}
                      // Accessibility
                      accessible={true}
                      accessibilityLabel={t("auth.signin.signUp")}
                      accessibilityRole="button"
                      accessibilityHint={t("auth.signin.signUpHint")}
                    >
                      <Text style={styles.signUpLink}>
                        {t("auth.signin.signUp")}
                      </Text>
                    </TouchableOpacity>

                    {/* Forgot Password Link - Right */}
                    <TouchableOpacity
                      testID="forgot-password-link"
                      style={styles.bottomLinkRight}
                      onPress={handleForgotPassword}
                      disabled={loginMutation.isPending}
                      // Accessibility
                      accessible={true}
                      accessibilityLabel={t("auth.signin.forgotPassword")}
                      accessibilityRole="button"
                      accessibilityHint={t("auth.signin.forgotPasswordHint")}
                    >
                      <Text style={styles.forgotPasswordText}>
                        {t("auth.signin.forgotPassword")}
                      </Text>
                    </TouchableOpacity>
                  </View>
                </View>
              </ScrollView>
            </View>
          </TouchableWithoutFeedback>
        </KeyboardAvoidingView>
      </SafeAreaView>
    </ImageBackground>
  );
};

export default LoginScreen;

const createStyles = (
  theme: Theme,
  responsive: ResponsiveValues,
  keyboardVisible: boolean,
  keyboardHeight: number,
) =>
  StyleSheet.create({
    backgroundImage: {
      flex: 1,
      width: "100%",
      height: "100%",
      backgroundColor: theme.colors.background.default, // Fallback warm background color
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
    innerContainer: {
      flex: 1,
    },
    scrollContent: {
      flexGrow: 1,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingBottom: responsive.isMediumPhone
        ? responsive.spacing.vertical * 4 // More padding for medium devices
        : responsive.spacing.vertical * 3, // Keep current for others
    },
    topSpacer: {
      height: keyboardVisible
        ? responsive.isMediumPhone
          ? (responsive.screenHeight - keyboardHeight) * 0.08 // Medium: 8% = more scroll up
          : responsive.isLargePhone || responsive.isExtraLargePhone
            ? (responsive.screenHeight - keyboardHeight) * 0.12 // Large+: keep 15%
            : responsive.spacing.vertical // Small: original behavior
        : responsive.isMediumPhone
          ? responsive.screenHeight * 0.08 // 8% for medium devices (saves ~33dp)
          : responsive.screenHeight * 0.12, // Keep 12% for others
    },
    formCard: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.xl,
      padding:
        responsive.spacing.formPadding + responsive.spacing.elementSpacing / 2, // Increased padding for height
      ...theme.shadows.lg,
      // Add subtle border for definition
      borderWidth: 1,
      borderColor: theme.colors.border.light,
    },
    // Sign In Button styles
    loginButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical:
        responsive.spacing.vertical + responsive.spacing.elementSpacing / 4,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: responsive.spacing.elementSpacing / 2,
      marginTop: responsive.spacing.elementSpacing * 0.8,
      marginBottom: responsive.spacing.elementSpacing,
      ...theme.shadows.lg,
      shadowColor: theme.colors.primary,
      minHeight:
        responsive.spacing.buttonHeight + responsive.spacing.elementSpacing / 3,
      elevation: theme.shadows.md.elevation,
    },
    loginButtonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
    },
    buttonDisabled: {
      opacity: theme.shadows.sm.shadowOpacity * 14, // Approximately 0.7 based on theme opacity
    },
    // Bottom Links Row styles
    bottomLinksRow: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      marginTop: responsive.spacing.elementSpacing * 0.5,
      paddingHorizontal: responsive.spacing.elementSpacing / 3, // Small horizontal padding for better touch targets
    },
    bottomLinkLeft: {
      flexDirection: "row",
      alignItems: "center",
      flex: 1,
      maxWidth: "50%", // Prevent overflow into right side
    },
    bottomLinkRight: {
      alignItems: "flex-end",
      flex: 1,
      maxWidth: "50%", // Prevent overflow into left side
    },
    signUpLink: {
      fontSize: responsive.fontSizes.headline,
      color: theme.colors.primary,
      fontWeight: "600",
      flexShrink: 1,
    },
    forgotPasswordText: {
      color: theme.colors.text.secondary,
      fontSize: responsive.fontSizes.body,
      fontWeight: "400",
      textAlign: "right",
      flexShrink: 1,
      opacity: theme.shadows.md.shadowOpacity * 10, // Approximately 0.8 based on theme opacity
    },
    welcomeText: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing / 1.5,
      textAlign: "center",
    },
    subtitleText: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.secondary,
      marginBottom: responsive.spacing.elementSpacing * 1.5,
      textAlign: "center",
    },
  });
