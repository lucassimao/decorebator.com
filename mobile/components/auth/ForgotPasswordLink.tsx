import React from "react";
import { TouchableOpacity, Text, StyleSheet } from "react-native";
import { useTranslation } from "react-i18next";
import { router } from "expo-router";
import { useResponsive } from "@/hooks/useResponsive";
import type { Theme } from "@/contexts/ThemeContext";

interface ForgotPasswordLinkProps {
  theme: Theme;
  responsive: ReturnType<typeof useResponsive>;
  disabled: boolean;
}

export const ForgotPasswordLink: React.FC<ForgotPasswordLinkProps> = ({
  theme,
  responsive,
  disabled,
}) => {
  const { t } = useTranslation();

  const handleForgotPassword = () => {
    router.push("/forgotPassword");
  };

  const styles = StyleSheet.create({
    forgotPassword: {
      alignSelf: "flex-end",
      marginBottom: responsive.spacing.elementSpacing,
      marginTop: -8,
    },
    forgotPasswordText: {
      color: theme.colors.primary,
      fontSize: responsive.fontSizes.label,
      fontWeight: "500",
    },
  });

  return (
    <TouchableOpacity
      style={styles.forgotPassword}
      onPress={handleForgotPassword}
      disabled={disabled}
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
  );
};
