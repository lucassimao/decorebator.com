import React from "react";
import { View, Text, TextInput, StyleSheet } from "react-native";
import { Controller, Control, FieldErrors } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { MaterialIcons } from "@expo/vector-icons";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { useResponsive } from "@/hooks/useResponsive";
import { useTheme } from "@/contexts/ThemeContext";

interface LoginFormData {
  email: string;
  password: string;
}

interface EmailInputProps {
  control: Control<LoginFormData>;
  errors: FieldErrors<LoginFormData>;
  emailInputRef: React.RefObject<TextInput | null>;
  passwordInputRef: React.RefObject<TextInput | null>;
  onFocus: () => void;
  isPending: boolean;
}

export const EmailInput: React.FC<EmailInputProps> = ({
  control,
  errors,
  emailInputRef,
  passwordInputRef,
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
  });

  return (
    <View style={styles.inputGroup}>
      <View style={styles.inputLabelRow}>
        <MaterialIcons
          name="email"
          size={20}
          color={theme.colors.text.secondary}
        />
        <Text style={styles.inputLabel}>{t("auth.signin.email")}</Text>
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
            ref={emailInputRef}
            style={[styles.input, errors.email && styles.inputError]}
            placeholder={t("auth.signin.emailPlaceholder")}
            placeholderTextColor={theme.colors.text.placeholder}
            value={value}
            onChangeText={onChange}
            onFocus={onFocus}
            onBlur={onBlur}
            autoCapitalize="none"
            keyboardType="email-address"
            autoComplete="email"
            textContentType="emailAddress"
            autoCorrect={false}
            spellCheck={false}
            importantForAutofill="no"
            returnKeyType="next"
            onSubmitEditing={() => passwordInputRef.current?.focus()}
            editable={!isPending}
            // Accessibility
            accessible={true}
            accessibilityLabel={t("auth.signin.email")}
            accessibilityHint={t("auth.signin.emailPlaceholder")}
          />
        )}
      />
      <ErrorMessage error={errors.email} />
    </View>
  );
};
