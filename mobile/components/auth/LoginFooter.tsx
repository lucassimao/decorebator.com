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
import { router } from "expo-router";
import { useResponsive } from "@/hooks/useResponsive";
import type { Theme } from "@/contexts/ThemeContext";

interface LoginFooterProps {
  theme: Theme;
  responsive: ReturnType<typeof useResponsive>;
  keyboardVisible: boolean;
  isPending: boolean;
  onSubmit: () => void;
}

export const LoginFooter: React.FC<LoginFooterProps> = ({
  theme,
  responsive,
  keyboardVisible,
  isPending,
  onSubmit,
}) => {
  const { t } = useTranslation();

  const handleSignUp = () => {
    router.replace("/signup");
  };

  const styles = StyleSheet.create({
    fixedFooter: {
      position: "absolute",
      bottom: 0,
      left: 0,
      right: 0,
      backgroundColor: "rgba(255, 255, 255, 0.98)",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.vertical,
      paddingBottom: responsive.spacing.vertical + (keyboardVisible ? 0 : 20),
      borderTopWidth: 1,
      borderTopColor: "rgba(0, 0, 0, 0.05)",
      ...theme.shadows.lg,
    },
    loginButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: 8,
      marginBottom: responsive.spacing.elementSpacing / 2,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
      minHeight: responsive.spacing.buttonHeight,
    },
    loginButtonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.fontSizes.button,
      fontWeight: "600",
    },
    buttonDisabled: {
      opacity: 0.7,
    },
    signUpContainer: {
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      marginTop: responsive.spacing.elementSpacing,
    },
    signUpText: {
      fontSize: responsive.fontSizes.label,
      color: theme.colors.text.secondary,
    },
    signUpLink: {
      fontSize: responsive.fontSizes.label,
      color: theme.colors.primary,
      fontWeight: "600",
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

      {/* Sign Up Link */}
      <View style={styles.signUpContainer}>
        <Text style={styles.signUpText}>{t("auth.signin.noAccount")} </Text>
        <TouchableOpacity
          onPress={handleSignUp}
          disabled={isPending}
          // Accessibility
          accessible={true}
          accessibilityLabel={t("auth.signin.signUp")}
          accessibilityRole="button"
          accessibilityHint={t("auth.signin.signUpHint")}
        >
          <Text style={styles.signUpLink}>{t("auth.signin.signUp")}</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
};
