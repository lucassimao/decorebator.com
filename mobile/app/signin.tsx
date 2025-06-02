import * as usersApi from "@/api/users";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useMutation } from "@tanstack/react-query";
import { LinearGradient } from "expo-linear-gradient";
import { router } from "expo-router";
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
  TextInput,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";

const { width: SCREEN_WIDTH, height: SCREEN_HEIGHT } = Dimensions.get("window");

interface LoginFormData {
  email: string;
  password: string;
}

const LoginScreen: React.FC = () => {
  const [showPassword, setShowPassword] = useState(false);
  const { t } = useTranslation();

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
    onSuccess: () => {
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
            {/* Logo/App Name */}
            <View style={styles.logoContainer}>
              <Text style={styles.appName}>{t("common.appName")}</Text>
              <Text style={styles.tagline}>{t("auth.tagline")}</Text>
            </View>

            {/* Foreground Illustration */}
            <View style={styles.illustrationContainer}>
              <Image
                source={require("@/assets/images/login-fg.png")} // Foreground illustration
                style={styles.illustration}
                resizeMode="contain"
              />
            </View>

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
                    <Text style={styles.errorText}>{errors.email.message}</Text>
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
                        color="#FFFFFF"
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
      </SafeAreaView>
    </ImageBackground>
  );
};

export default LoginScreen;
const styles = StyleSheet.create({
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
    marginTop: 40,
    marginBottom: 20,
  },
  appName: {
    fontSize: 36,
    fontWeight: "700",
    color: "#2D3436",
    marginBottom: 8,
  },
  tagline: {
    fontSize: 16,
    color: "#636E72",
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
    paddingHorizontal: 20,
    flex: 1,
    justifyContent: "center",
  },
  formCard: {
    borderRadius: 24,
    padding: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    elevation: 5,
  },
  welcomeText: {
    fontSize: 24,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 8,
    textAlign: "center",
  },
  subtitleText: {
    fontSize: 16,
    color: "#636E72",
    marginBottom: 24,
    textAlign: "center",
  },
  inputGroup: {
    marginBottom: 20,
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
    color: "#2D3436",
  },
  input: {
    backgroundColor: "#FAFAFA",
    borderWidth: 1,
    borderColor: "#E0E0E0",
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 16,
    color: "#2D3436",
  },
  inputError: {
    borderColor: "#FF6B6B",
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
    color: "#FF6B6B",
    fontSize: 12,
    marginTop: 4,
  },
  forgotPassword: {
    alignSelf: "flex-end",
    marginBottom: 24,
    marginTop: -8,
  },
  forgotPasswordText: {
    color: "#FF7B54",
    fontSize: 14,
    fontWeight: "500",
  },
  loginButton: {
    backgroundColor: "#FF7B54",
    borderRadius: 12,
    paddingVertical: 16,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    marginBottom: 24,
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 5,
  },
  loginButtonText: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  buttonDisabled: {
    opacity: 0.7,
  },
  dividerContainer: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 24,
  },
  divider: {
    flex: 1,
    height: 1,
    backgroundColor: "#E0E0E0",
  },
  dividerText: {
    marginHorizontal: 16,
    color: "#636E72",
    fontSize: 14,
  },
  socialContainer: {
    gap: 12,
    marginBottom: 24,
  },
  socialButton: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "#FFFFFF",
    borderWidth: 1,
    borderColor: "#E0E0E0",
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
    color: "#2D3436",
    fontWeight: "500",
  },
  signUpContainer: {
    flexDirection: "row",
    justifyContent: "center",
    alignItems: "center",
  },
  signUpText: {
    fontSize: 14,
    color: "#636E72",
  },
  signUpLink: {
    fontSize: 14,
    color: "#FF7B54",
    fontWeight: "600",
  },
});
