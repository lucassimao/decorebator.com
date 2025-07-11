import React from "react";
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
} from "react-native";
import { useTranslation } from "react-i18next";
import { Ionicons } from "@expo/vector-icons";
import { useResponsive } from "@/hooks/useResponsive";
import { useTheme } from "@/contexts/ThemeContext";

interface LoginFooterProps {
  keyboardVisible: boolean;
  keyboardHeight?: number;
  isPending: boolean;
  onSubmit: () => void;
}

export const LoginFooter: React.FC<LoginFooterProps> = ({
  keyboardVisible,
  isPending,
  onSubmit,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const responsive = useResponsive();

  const styles = StyleSheet.create({
    fixedFooter: {
      paddingTop: responsive.spacing.vertical,
    },
    loginButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md + 2,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: 8,
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
  });

  return (
    <View
      style={styles.fixedFooter}
      // Accessibility
      accessible={true}
      accessibilityLabel={t("auth.signin.loginActions")}
    >
      {/* Login Button */}
      <TouchableOpacity
        style={[styles.loginButton, isPending && styles.buttonDisabled]}
        onPress={onSubmit}
        disabled={isPending}
        activeOpacity={0.8}
        // Accessibility
        accessible={true}
        accessibilityLabel={
          isPending ? t("auth.signin.signingIn") : t("auth.signin.signInButton")
        }
        accessibilityRole="button"
        accessibilityHint={t("auth.signin.signInHint")}
        accessibilityState={{ disabled: isPending }}
      >
        {isPending ? (
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
    </View>
  );
};
