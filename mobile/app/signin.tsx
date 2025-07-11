import * as usersApi from "@/api/users";
import { useResponsive } from "@/hooks/useResponsive";
import { LoginHeader } from "@/components/auth/LoginHeader";
import { EmailInput } from "@/components/auth/EmailInput";
import { PasswordInput } from "@/components/auth/PasswordInput";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { router } from "expo-router";
import { usePostHog } from "posthog-react-native";
import React, { useState, useEffect, useRef, useCallback } from "react";
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
import * as Sentry from "@sentry/react-native";
import { decode } from "@/api/jwt";

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

  // Always use light theme for auth screens
  const theme = authLightTheme;

  // Get responsive values using the optimized hook
  const responsive = useResponsive();

  // Memoize styles to prevent recreation on every render
  const styles = React.useMemo(
    () => createStyles(theme, responsive, keyboardVisible),
    [theme, responsive, keyboardVisible],
  );

  // Refs for form inputs
  const scrollViewRef = useRef<ScrollView>(null);
  const emailInputRef = useRef<TextInput>(null);
  const passwordInputRef = useRef<TextInput>(null);

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
    router.replace("/signup");
  };

  const handleForgotPassword = () => {
    router.push("/forgotPassword");
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

  // Function to handle scrolling to input with dynamic keyboard height
  const scrollToInput = useCallback(
    (inputRef: React.RefObject<TextInput | null>) => {
      if (!inputRef.current || !scrollViewRef.current) return;

      // Improved manual implementation with dynamic keyboard height
      const scrollView = scrollViewRef.current;
      const input = inputRef.current;

      if (!scrollView || !input) return;

      // Use measureInWindow to get input position
      input.measureInWindow((x, y, width, height) => {
        // Only scroll if keyboard is visible and input might be covered
        if (!keyboardVisible || keyboardHeight === 0) {
          return; // Don't scroll when keyboard isn't visible or height unknown
        }

        // Use actual keyboard height from event
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
    },
    [keyboardVisible, keyboardHeight, responsive.screenHeight],
  );

  return (
    <ImageBackground
      source={require("@/assets/images/login-bg.png")} // Background illustration
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
                ref={scrollViewRef}
                contentContainerStyle={styles.scrollContent}
                showsVerticalScrollIndicator={false}
                keyboardShouldPersistTaps="handled"
              >
                {/* Add some space at the top for the background to show */}
                <View style={styles.topSpacer} />

                {/* Header with Logo and Illustration */}
                <LoginHeader keyboardVisible={keyboardVisible} />

                {/* Login Form */}
                <View
                  style={styles.formCard}
                  // Accessibility
                  accessible={true}
                  accessibilityLabel={t("auth.signin.loginForm")}
                >
                  <Text style={styles.welcomeText}>
                    {t("auth.signin.welcomeBack")}
                  </Text>
                  <Text style={styles.subtitleText}>
                    {t("auth.signin.subtitle")}
                  </Text>

                  {/* Email Input */}
                  <EmailInput
                    control={control}
                    errors={errors}
                    emailInputRef={emailInputRef}
                    passwordInputRef={passwordInputRef}
                    onFocus={() => scrollToInput(emailInputRef)}
                    isPending={loginMutation.isPending}
                  />

                  {/* Password Input */}
                  <PasswordInput
                    control={control}
                    errors={errors}
                    passwordInputRef={passwordInputRef}
                    showPassword={showPassword}
                    setShowPassword={setShowPassword}
                    onFocus={() => scrollToInput(passwordInputRef)}
                    isPending={loginMutation.isPending}
                  />

                  {/* Sign In Button */}
                  <TouchableOpacity
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
  responsive: ReturnType<typeof useResponsive>,
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
    innerContainer: {
      flex: 1,
    },
    scrollContent: {
      flexGrow: 1,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingBottom: 50, // Ensure content is not cut off by safe area
    },
    topSpacer: {
      height: keyboardVisible
        ? responsive.spacing.vertical
        : responsive.screenHeight * 0.12, // Balanced spacing
    },
    formCard: {
      backgroundColor: "rgba(255, 255, 255, 0.95)",
      borderRadius: theme.borderRadius.xl,
      padding: responsive.spacing.formPadding + 8, // Increased padding for height
      ...theme.shadows.lg,
      // Add subtle border for definition
      borderWidth: 1,
      borderColor: "rgba(255, 255, 255, 0.3)",
    },
    // Sign In Button styles
    loginButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md + 2,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: 8,
      marginTop: responsive.spacing.elementSpacing * 0.8,
      marginBottom: responsive.spacing.elementSpacing,
      ...theme.shadows.lg,
      shadowColor: theme.colors.primary,
      minHeight: responsive.spacing.buttonHeight + 4,
      elevation: 4,
    },
    loginButtonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.fontSizes.button,
      fontWeight: "600",
    },
    buttonDisabled: {
      opacity: 0.7,
    },
    // Bottom Links Row styles
    bottomLinksRow: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      marginTop: responsive.spacing.elementSpacing * 0.5,
      paddingHorizontal: 4, // Small horizontal padding for better touch targets
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
      fontSize: responsive.fontSizes.body,
      color: theme.colors.primary,
      fontWeight: "600",
      textDecorationLine: "underline",
      flexShrink: 1,
    },
    forgotPasswordText: {
      color: theme.colors.text.secondary,
      fontSize: responsive.fontSizes.label,
      fontWeight: "400",
      textAlign: "right",
      flexShrink: 1,
      opacity: 0.8,
    },
    welcomeText: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 8,
      textAlign: "center",
    },
    subtitleText: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.secondary,
      marginBottom: responsive.spacing.elementSpacing * 1.5,
      textAlign: "center",
    },
  });
