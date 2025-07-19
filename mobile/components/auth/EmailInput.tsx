import React from "react";
import { View, Text, TextInput, StyleSheet } from "react-native";
import { Controller, Control, FieldErrors } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { MaterialIcons } from "@expo/vector-icons";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { useTheme } from "@/contexts/ThemeContext";

interface LoginFormData {
  email: string;
  password: string;
}

interface EmailInputProps {
  control: Control<LoginFormData>;
  errors: FieldErrors<LoginFormData>;
  isPending: boolean;
}

export const EmailInput: React.FC<EmailInputProps> = ({
  control,
  errors,
  isPending,
}) => {
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();

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
            style={[styles.input, errors.email && styles.inputError]}
            placeholder={t("auth.signin.emailPlaceholder")}
            placeholderTextColor={theme.colors.text.placeholder}
            value={value}
            onChangeText={onChange}
            onBlur={onBlur}
            autoCapitalize="none"
            keyboardType="email-address"
            autoComplete="email"
            textContentType="emailAddress"
            autoCorrect={false}
            spellCheck={false}
            importantForAutofill="no"
            returnKeyType="next"
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
