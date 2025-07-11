import React from "react";
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Keyboard,
} from "react-native";
import { Controller, Control, FieldErrors } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { useResponsive } from "@/hooks/useResponsive";
import { useTheme } from "@/contexts/ThemeContext";

interface LoginFormData {
  email: string;
  password: string;
}

interface PasswordInputProps {
  control: Control<LoginFormData>;
  errors: FieldErrors<LoginFormData>;
  passwordInputRef: React.RefObject<TextInput | null>;
  showPassword: boolean;
  setShowPassword: (show: boolean) => void;
  onFocus: () => void;
  isPending: boolean;
}

export const PasswordInput: React.FC<PasswordInputProps> = ({
  control,
  errors,
  passwordInputRef,
  showPassword,
  setShowPassword,
  onFocus,
  isPending,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const responsive = useResponsive();

  const styles = StyleSheet.create({
    inputGroup: {
      marginBottom: responsive.spacing.elementSpacing * 0.8, // Reduced spacing
    },
    inputLabelRow: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: 6, // Reduced from 8 to 6
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
              ref={passwordInputRef}
              style={[
                styles.input,
                styles.passwordInput,
                errors.password && styles.inputError,
              ]}
              placeholder={t("auth.signin.passwordPlaceholder")}
              placeholderTextColor={theme.colors.text.placeholder}
              value={value}
              onChangeText={onChange}
              onFocus={onFocus}
              onBlur={onBlur}
              secureTextEntry={!showPassword}
              autoComplete="password"
              textContentType="password"
              autoCorrect={false}
              spellCheck={false}
              importantForAutofill="no"
              passwordRules=""
              returnKeyType="done"
              onSubmitEditing={() => Keyboard.dismiss()}
              editable={!isPending}
              // Accessibility
              accessible={true}
              accessibilityLabel={t("auth.signin.password")}
              accessibilityHint={t("auth.signin.passwordPlaceholder")}
            />
          )}
        />
        <TouchableOpacity
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
};
