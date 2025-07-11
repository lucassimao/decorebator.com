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
  keyboardHeight?: number;
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
      paddingVertical: responsive.spacing.vertical,
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
    signUpContainer: {
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      marginTop: responsive.spacing.elementSpacing * 0.1, // Even closer to sign-in button
      paddingBottom: responsive.spacing.vertical,
    },
    signUpText: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.primary,
      opacity: 0.7,
    },
    signUpLink: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.primary,
      fontWeight: "600",
      textDecorationLine: "underline",
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
