import React, { useState, useEffect, useRef } from "react";
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  TouchableWithoutFeedback,
  TextInput,
  Animated,
  Dimensions,
  KeyboardAvoidingView,
  Platform,
  ActivityIndicator,
  Alert,
  ScrollView,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm, Controller } from "react-hook-form";
import * as userApi from "@/api/users";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { router } from "expo-router";
import {
  isPasswordTooLong,
  isPasswordTooShort,
  passwordCodePointLength,
} from "@/utils/passwordPolicy";
import { completePasswordChange } from "@/utils/completePasswordChange";

const { height: SCREEN_HEIGHT } = Dimensions.get("window");

// Types
interface ChangePasswordFormData {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}

interface ChangePasswordModalProps {
  visible: boolean;
  onClose: () => void;
}

export const ChangePasswordModal: React.FC<ChangePasswordModalProps> = ({
  visible,
  onClose,
}) => {
  const slideAnim = useRef(new Animated.Value(SCREEN_HEIGHT)).current;
  const backdropAnim = useRef(new Animated.Value(0)).current;
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const queryClient = useQueryClient();

  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
    watch,
  } = useForm<ChangePasswordFormData>({
    defaultValues: {
      currentPassword: "",
      newPassword: "",
      confirmPassword: "",
    },
  });

  const newPassword = watch("newPassword");

  // Change password mutation
  const changePasswordMutation = useMutation<
    void,
    Error,
    userApi.UpdatePasswordPayload
  >({
    mutationFn: async (data) => {
      return userApi.update({
        updatePassword: data,
      });
    },
    onSuccess: async () => {
      await completePasswordChange({
        clearCredentials: userApi.clearSessionCredentials,
        presentSuccess: (onConfirm) => {
          Alert.alert(
            t("common.success"),
            t("profile.changePassword.passwordChanged"),
            [{ text: t("common.ok"), onPress: onConfirm }],
          );
        },
        resetForm: reset,
        close: onClose,
        clearInMemoryState: () => queryClient.clear(),
        redirectToSignIn: () => {
          router.dismissAll();
          router.replace("/signin");
        },
      });
    },
    onError: (error: Error) => {
      Alert.alert(
        t("common.error"),
        error.message || t("profile.changePassword.changeError"),
      );
    },
  });

  // Animation
  useEffect(() => {
    if (visible) {
      Animated.parallel([
        Animated.timing(slideAnim, {
          toValue: 0,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.timing(backdropAnim, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      Animated.parallel([
        Animated.timing(slideAnim, {
          toValue: SCREEN_HEIGHT,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.timing(backdropAnim, {
          toValue: 0,
          duration: 300,
          useNativeDriver: true,
        }),
      ]).start(() => {
        reset();
        setShowCurrentPassword(false);
        setShowNewPassword(false);
        setShowConfirmPassword(false);
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visible]);

  const handlePasswordChange = (data: ChangePasswordFormData) => {
    changePasswordMutation.mutate({
      currentPassword: data.currentPassword,
      newPassword: data.newPassword,
    });
  };

  const getPasswordStrength = (
    password: string,
  ): { text: string; color: string; percentage: number } => {
    if (!password)
      return { text: "", color: theme.colors.ui.border, percentage: 0 };

    let strength = 0;
    if (passwordCodePointLength(password) >= 8) strength++;
    if (passwordCodePointLength(password) >= 12) strength++;
    if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength++;
    if (/\d/.test(password)) strength++;
    if (/[^a-zA-Z\d]/.test(password)) strength++;

    if (strength <= 2)
      return {
        text: t("profile.changePassword.passwordStrength.weak"),
        color: theme.colors.error,
        percentage: 33,
      };
    if (strength <= 4)
      return {
        text: t("profile.changePassword.passwordStrength.good"),
        color: theme.colors.primary,
        percentage: 66,
      };
    return {
      text: t("profile.changePassword.passwordStrength.strong"),
      color: theme.colors.success,
      percentage: 100,
    };
  };

  const passwordStrength = getPasswordStrength(newPassword);

  if (!visible) return null;

  return (
    <Modal
      transparent
      visible={visible}
      animationType="none"
      onRequestClose={onClose}
    >
      <TouchableWithoutFeedback onPress={onClose}>
        <Animated.View style={[styles.backdrop, { opacity: backdropAnim }]} />
      </TouchableWithoutFeedback>

      <KeyboardAvoidingView
        behavior={Platform.OS === "ios" ? "padding" : "height"}
        style={styles.container}
      >
        <Animated.View
          style={[
            styles.modalContent,
            { transform: [{ translateY: slideAnim }] },
          ]}
        >
          {/* Header */}
          <View style={styles.header}>
            <View style={styles.handle} />
            <View style={styles.titleRow}>
              <Text style={styles.title}>
                {t("profile.changePassword.title")}
              </Text>
              <TouchableOpacity style={styles.closeButton} onPress={onClose}>
                <Ionicons
                  name="close"
                  size={24}
                  color={theme.colors.text.secondary}
                />
              </TouchableOpacity>
            </View>
          </View>

          <ScrollView
            style={styles.scrollContent}
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
          >
            {/* Current Password */}
            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>
                {t("profile.changePassword.currentPassword")}
              </Text>
              <Controller
                control={control}
                name="currentPassword"
                rules={{
                  required: t("profile.changePassword.currentPasswordRequired"),
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <View style={styles.passwordContainer}>
                    <TextInput
                      style={[
                        styles.input,
                        errors.currentPassword && styles.inputError,
                      ]}
                      placeholder={t(
                        "profile.changePassword.enterCurrentPassword",
                      )}
                      placeholderTextColor={theme.colors.text.placeholder}
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                      secureTextEntry={!showCurrentPassword}
                      autoComplete="password"
                      textContentType="password"
                    />
                    <TouchableOpacity
                      style={styles.passwordToggle}
                      onPress={() =>
                        setShowCurrentPassword(!showCurrentPassword)
                      }
                    >
                      <Ionicons
                        name={showCurrentPassword ? "eye-off" : "eye"}
                        size={20}
                        color={theme.colors.text.secondary}
                      />
                    </TouchableOpacity>
                  </View>
                )}
              />
              {errors.currentPassword && (
                <Text style={styles.errorText}>
                  {errors.currentPassword.message}
                </Text>
              )}
            </View>

            {/* New Password */}
            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>
                {t("profile.changePassword.newPassword")}
              </Text>
              <Controller
                control={control}
                name="newPassword"
                rules={{
                  required: t("profile.changePassword.newPasswordRequired"),
                  validate: (value) => {
                    if (isPasswordTooShort(value)) {
                      return t("profile.changePassword.minLength", { min: 8 });
                    }
                    if (isPasswordTooLong(value)) {
                      return t("errors.longPassword");
                    }
                    if (value === watch("currentPassword")) {
                      return t("profile.changePassword.mustBeDifferent");
                    }
                    return true;
                  },
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <View style={styles.passwordContainer}>
                    <TextInput
                      style={[
                        styles.input,
                        errors.newPassword && styles.inputError,
                      ]}
                      placeholder={t("profile.changePassword.enterNewPassword")}
                      placeholderTextColor={theme.colors.text.placeholder}
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                      secureTextEntry={!showNewPassword}
                      autoComplete="password-new"
                      textContentType="newPassword"
                    />
                    <TouchableOpacity
                      style={styles.passwordToggle}
                      onPress={() => setShowNewPassword(!showNewPassword)}
                    >
                      <Ionicons
                        name={showNewPassword ? "eye-off" : "eye"}
                        size={20}
                        color={theme.colors.text.secondary}
                      />
                    </TouchableOpacity>
                  </View>
                )}
              />
              {errors.newPassword && (
                <Text style={styles.errorText}>
                  {errors.newPassword.message}
                </Text>
              )}

              {/* Password Strength Indicator */}
              {newPassword && (
                <View style={styles.strengthContainer}>
                  <View style={styles.strengthBar}>
                    <Animated.View
                      style={[
                        styles.strengthFill,
                        {
                          width: `${passwordStrength.percentage}%`,
                          backgroundColor: passwordStrength.color,
                        },
                      ]}
                    />
                  </View>
                  <Text
                    style={[
                      styles.strengthText,
                      { color: passwordStrength.color },
                    ]}
                  >
                    {passwordStrength.text}
                  </Text>
                </View>
              )}
            </View>

            {/* Confirm Password */}
            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>
                {t("profile.changePassword.confirmPassword")}
              </Text>
              <Controller
                control={control}
                name="confirmPassword"
                rules={{
                  required: t("profile.changePassword.confirmPasswordRequired"),
                  validate: (value) =>
                    value === newPassword ||
                    t("profile.changePassword.passwordMismatch"),
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <View style={styles.passwordContainer}>
                    <TextInput
                      style={[
                        styles.input,
                        errors.confirmPassword && styles.inputError,
                      ]}
                      placeholder={t(
                        "profile.changePassword.confirmNewPassword",
                      )}
                      placeholderTextColor={theme.colors.text.placeholder}
                      value={value}
                      onChangeText={onChange}
                      onBlur={onBlur}
                      secureTextEntry={!showConfirmPassword}
                      autoComplete="password-new"
                      textContentType="newPassword"
                    />
                    <TouchableOpacity
                      style={styles.passwordToggle}
                      onPress={() =>
                        setShowConfirmPassword(!showConfirmPassword)
                      }
                    >
                      <Ionicons
                        name={showConfirmPassword ? "eye-off" : "eye"}
                        size={20}
                        color={theme.colors.text.secondary}
                      />
                    </TouchableOpacity>
                  </View>
                )}
              />
              {errors.confirmPassword && (
                <Text style={styles.errorText}>
                  {errors.confirmPassword.message}
                </Text>
              )}
            </View>

            {/* Enforced password requirements */}
            <View style={styles.requirementsContainer}>
              <Text style={styles.requirementsTitle}>
                {t("profile.changePassword.passwordRequirements")}
              </Text>
              <View style={styles.requirement}>
                <Text style={styles.requirementText}>
                  {t("profile.changePassword.atLeast8Characters")}
                </Text>
              </View>
              <View style={styles.requirement}>
                <Text style={styles.requirementText}>
                  {t("errors.longPassword")}
                </Text>
              </View>
            </View>
          </ScrollView>

          {/* Action Buttons */}
          <View style={styles.actions}>
            <TouchableOpacity style={styles.cancelButton} onPress={onClose}>
              <Text style={styles.cancelButtonText}>{t("common.cancel")}</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[
                styles.submitButton,
                changePasswordMutation.isPending && styles.buttonDisabled,
              ]}
              onPress={handleSubmit(handlePasswordChange)}
              disabled={changePasswordMutation.isPending}
            >
              {changePasswordMutation.isPending ? (
                <ActivityIndicator
                  size="small"
                  color={theme.colors.text.inverse}
                />
              ) : (
                <Text style={styles.submitButtonText}>
                  {t("profile.changePassword.title")}
                </Text>
              )}
            </TouchableOpacity>
          </View>
        </Animated.View>
      </KeyboardAvoidingView>
    </Modal>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    container: {
      flex: 1,
      justifyContent: "flex-end",
    },
    backdrop: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: theme.colors.overlay.backdrop,
    },
    modalContent: {
      backgroundColor: theme.colors.background.surface,
      borderTopLeftRadius: theme.borderRadius.xl,
      borderTopRightRadius: theme.borderRadius.xl,
      maxHeight: SCREEN_HEIGHT * 0.9,
      ...theme.shadows.lg,
    },
    header: {
      paddingTop: theme.spacing.sm,
      paddingBottom: theme.spacing.lg,
      paddingHorizontal: theme.spacing.lg,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.divider,
    },
    handle: {
      width: 40,
      height: 4,
      backgroundColor: theme.colors.ui.disabled,
      borderRadius: 2,
      alignSelf: "center",
      marginBottom: theme.spacing.md,
    },
    titleRow: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
    },
    title: {
      fontSize: 24,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    closeButton: {
      width: 32,
      height: 32,
      borderRadius: 16,
      backgroundColor: theme.colors.background.subtle,
      justifyContent: "center",
      alignItems: "center",
    },
    scrollContent: {
      padding: theme.spacing.lg,
      maxHeight: SCREEN_HEIGHT * 0.6,
    },
    inputGroup: {
      marginBottom: theme.spacing.lg,
    },
    inputLabel: {
      fontSize: 14,
      fontWeight: "500",
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.sm,
    },
    passwordContainer: {
      position: "relative",
    },
    input: {
      backgroundColor: theme.colors.ui.inputBackground,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      borderRadius: theme.borderRadius.md,
      paddingHorizontal: theme.spacing.md,
      paddingVertical: 14,
      paddingRight: 48,
      fontSize: 16,
      color: theme.colors.text.primary,
    },
    inputError: {
      borderColor: theme.colors.error,
    },
    passwordToggle: {
      position: "absolute",
      right: 16,
      top: "50%",
      transform: [{ translateY: -10 }],
    },
    errorText: {
      color: theme.colors.error,
      fontSize: 12,
      marginTop: 4,
    },
    strengthContainer: {
      flexDirection: "row",
      alignItems: "center",
      marginTop: theme.spacing.sm,
      gap: 12,
    },
    strengthBar: {
      flex: 1,
      height: 4,
      backgroundColor: theme.colors.ui.divider,
      borderRadius: 2,
      overflow: "hidden",
    },
    strengthFill: {
      height: "100%",
      borderRadius: 2,
    },
    strengthText: {
      fontSize: 12,
      fontWeight: "500",
    },
    requirementsContainer: {
      backgroundColor: theme.colors.background.subtle,
      borderRadius: theme.borderRadius.md,
      padding: theme.spacing.md,
      marginTop: theme.spacing.sm,
    },
    requirementsTitle: {
      fontSize: 14,
      fontWeight: "500",
      color: theme.colors.text.primary,
      marginBottom: theme.spacing.sm,
    },
    requirement: {
      flexDirection: "row",
      alignItems: "center",
      gap: 8,
      marginBottom: theme.spacing.sm,
    },
    requirementText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
    actions: {
      flexDirection: "row",
      paddingHorizontal: theme.spacing.lg,
      paddingTop: theme.spacing.md,
      paddingBottom: Platform.OS === "ios" ? 34 : theme.spacing.lg,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
      gap: 12,
    },
    cancelButton: {
      flex: 1,
      paddingVertical: theme.spacing.md,
      borderRadius: theme.borderRadius.md,
      backgroundColor: theme.colors.background.subtle,
      alignItems: "center",
    },
    cancelButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.secondary,
    },
    submitButton: {
      flex: 1,
      paddingVertical: theme.spacing.md,
      borderRadius: theme.borderRadius.md,
      backgroundColor: theme.colors.primary,
      alignItems: "center",
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
    },
    submitButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    buttonDisabled: {
      opacity: 0.7,
    },
  });
