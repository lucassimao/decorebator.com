import React, { forwardRef } from "react";
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
} from "react-native";
import { Controller, Control, FieldErrors } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { useTheme, Theme } from "@/contexts/ThemeContext";

interface LoginFormData {
  email: string;
  password: string;
}

interface PasswordInputProps {
  control: Control<LoginFormData>;
  errors: FieldErrors<LoginFormData>;
  showPassword: boolean;
  setShowPassword: (show: boolean) => void;
  isPending: boolean;
  theme?: Theme; // Optional theme prop for auth screens
  onSubmitEditing?: () => void; // Callback for form submission
}

export const PasswordInput = forwardRef<TextInput, PasswordInputProps>(
  (
    {
      control,
      errors,
      showPassword,
      setShowPassword,
      isPending,
      theme: propTheme,
      onSubmitEditing,
    },
    ref,
  ) => {
    const { t } = useTranslation();
    const { theme: contextTheme, responsive } = useTheme();

    // Use passed theme prop if available, otherwise use context theme
    const theme = propTheme || contextTheme;

    const styles = StyleSheet.create({
      inputGroup: {
        marginBottom: responsive.spacing.elementSpacing,
      },
      inputLabelRow: {
        flexDirection: "row",
        alignItems: "center",
        marginBottom: responsive.spacing.elementSpacing / 2, // Material Design 8px standard
        gap: responsive.spacing.elementSpacing / 2,
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
        paddingHorizontal: responsive.spacing.horizontal,
        paddingVertical: responsive.spacing.vertical,
        fontSize: responsive.fontSizes.body,
        color: theme.colors.text.primary,
        minHeight: responsive.spacing.minTouchTarget,
      },
      inputError: {
        borderColor: theme.colors.error,
        backgroundColor: theme.colors.state.incorrectBackground,
      },
      passwordContainer: {
        position: "relative",
      },
      passwordInput: {
        paddingRight: responsive.spacing.minTouchTarget,
      },
      passwordToggle: {
        position: "absolute",
        right: responsive.spacing.horizontal,
        top: "50%",
        transform: [{ translateY: -responsive.spacing.elementSpacing / 1.6 }],
        minHeight: responsive.spacing.minTouchTarget / 2,
        minWidth: responsive.spacing.minTouchTarget / 2,
        justifyContent: "center",
        alignItems: "center",
      },
    });

    return (
      <View style={styles.inputGroup}>
        <View style={styles.inputLabelRow}>
          <MaterialIcons
            name="lock"
            size={20}
            color={theme.colors.text.secondary}
          />
          <Text style={styles.inputLabel}>{t("auth.signin.password")}</Text>
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
                testID="password-input"
                ref={ref}
                style={[
                  styles.input,
                  styles.passwordInput,
                  errors.password && styles.inputError,
                ]}
                placeholder={t("auth.signin.passwordPlaceholder")}
                placeholderTextColor={theme.colors.text.placeholder}
                value={value}
                onChangeText={onChange}
                onBlur={onBlur}
                secureTextEntry={!showPassword}
                autoComplete="password"
                textContentType="password"
                autoCorrect={false}
                spellCheck={false}
                importantForAutofill="no"
                passwordRules=""
                returnKeyType="done"
                onSubmitEditing={onSubmitEditing}
                editable={!isPending}
                // Accessibility
                accessible={true}
                accessibilityLabel={t("auth.signin.password")}
                accessibilityHint={t("auth.signin.passwordPlaceholder")}
              />
            )}
          />
          <TouchableOpacity
            testID="password-toggle"
            style={styles.passwordToggle}
            onPress={() => setShowPassword(!showPassword)}
            hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
            // Accessibility
            accessible={true}
            accessibilityLabel={
              showPassword
                ? t("auth.signin.hidePassword")
                : t("auth.signin.showPassword")
            }
            accessibilityRole="button"
            accessibilityHint={t("auth.signin.togglePasswordVisibility")}
          >
            <Ionicons
              name={showPassword ? "eye-off" : "eye"}
              size={20}
              color={theme.colors.text.secondary}
            />
          </TouchableOpacity>
        </View>
        <ErrorMessage error={errors.password} />
      </View>
    );
  },
);

PasswordInput.displayName = "PasswordInput";
