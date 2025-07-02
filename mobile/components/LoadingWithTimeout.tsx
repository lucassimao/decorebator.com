import React from "react";
import {
  View,
  Text,
  ActivityIndicator,
  TouchableOpacity,
  StyleSheet,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";
import { useTheme } from "@/contexts/ThemeContext";

interface LoadingWithTimeoutProps {
  isLoading: boolean;
  hasTimeout: boolean;
  error?: Error | null;
  primaryColor?: string;
  backgroundColor?: string;
  textColor?: string;
  loadingMessage?: string;
  timeoutMessage?: string;
  onRetry: () => void;
  onGoBack: () => void;
  showTimeoutActions?: boolean;
}

export const LoadingWithTimeout: React.FC<LoadingWithTimeoutProps> = ({
  isLoading,
  hasTimeout,
  error,
  primaryColor = "#FF7B54",
  backgroundColor = "#FFFFFF",
  textColor = "#636E72",
  loadingMessage,
  timeoutMessage,
  onRetry,
  onGoBack,
  showTimeoutActions = true,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, spacing);

  if (!isLoading) return null;

  return (
    <View style={styles.loadingContainer}>
      <ActivityIndicator size="large" color={primaryColor} />
      <Text style={[styles.loadingText, { color: textColor }]}>
        {hasTimeout ? timeoutMessage : loadingMessage}
      </Text>

      {hasTimeout && showTimeoutActions && (
        <View style={styles.timeoutActions}>
          <Text style={[styles.timeoutMessage, { color: textColor }]}>
            {error?.message.includes("timeout")
              ? t("quiz.requestTimedOut")
              : t("quiz.slowConnection")}
          </Text>
          <View style={styles.actionButtons}>
            <TouchableOpacity
              style={[styles.retryButton, { backgroundColor: primaryColor }]}
              onPress={onRetry}
            >
              <MaterialIcons
                name="refresh"
                size={isTablet ? 24 : 20}
                color={theme.colors.text.inverse}
              />
              <Text style={styles.retryButtonText}>{t("common.retry")}</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.goBackButton, { backgroundColor }]}
              onPress={onGoBack}
            >
              <MaterialIcons
                name="arrow-back"
                size={isTablet ? 24 : 20}
                color={textColor}
              />
              <Text style={[styles.goBackButtonText, { color: textColor }]}>
                {t("common.goBack")}
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      )}
    </View>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  spacing: ReturnType<typeof useResponsiveSpacing>,
) =>
  StyleSheet.create({
    loadingContainer: {
      minHeight: isTablet ? 400 : 300,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: spacing.lg,
    },
    loadingText: {
      marginTop: spacing.md,
      fontSize: isTablet ? 18 : 16,
      textAlign: "center",
    },
    timeoutActions: {
      marginTop: spacing.xl,
      alignItems: "center",
    },
    timeoutMessage: {
      fontSize: isTablet ? 16 : 14,
      textAlign: "center",
      marginBottom: spacing.lg,
      lineHeight: isTablet ? 24 : 20,
    },
    actionButtons: {
      flexDirection: "row",
      gap: spacing.sm,
    },
    retryButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: spacing.lg,
      paddingVertical: spacing.sm,
      borderRadius: theme.borderRadius.lg,
      gap: spacing.xs,
      minHeight: isTablet ? 48 : 44,
    },
    retryButtonText: {
      color: theme.colors.text.inverse,
      fontSize: isTablet ? 18 : 16,
      fontWeight: "600",
    },
    goBackButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: spacing.lg,
      paddingVertical: spacing.sm,
      borderRadius: theme.borderRadius.lg,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      gap: spacing.xs,
      minHeight: isTablet ? 48 : 44,
    },
    goBackButtonText: {
      fontSize: isTablet ? 18 : 16,
      fontWeight: "600",
    },
  });
