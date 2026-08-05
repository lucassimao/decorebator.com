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
import { useTheme } from "@/contexts/ThemeContext";

interface LoadingWithTimeoutProps {
  isLoading: boolean;
  hasTimeout: boolean;
  error?: Error | null;
  loadingMessage?: string;
  timeoutMessage?: string;
  onRetry: () => void;
  onGoBack: () => void;
  showTimeoutActions?: boolean;
  timeoutErrorMessage?: string;
  slowConnectionMessage?: string;
  backLabel?: string;
}

export const LoadingWithTimeout: React.FC<LoadingWithTimeoutProps> = ({
  isLoading,
  hasTimeout,
  error,
  loadingMessage,
  timeoutMessage,
  onRetry,
  onGoBack,
  showTimeoutActions = true,
  timeoutErrorMessage,
  slowConnectionMessage,
  backLabel,
}) => {
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

  if (!isLoading) return null;

  return (
    <View style={styles.loadingContainer}>
      <ActivityIndicator
        size="large"
        color={theme.colors.primary}
        accessibilityRole="progressbar"
        accessibilityLabel={hasTimeout ? timeoutMessage : loadingMessage}
      />
      <Text style={styles.loadingText}>
        {hasTimeout ? timeoutMessage : loadingMessage}
      </Text>

      {hasTimeout && showTimeoutActions && (
        <View style={styles.timeoutActions}>
          <Text style={styles.timeoutMessage}>
            {error?.message.includes("timeout")
              ? timeoutErrorMessage ?? t("quiz.requestTimedOut")
              : slowConnectionMessage ?? t("quiz.slowConnection")}
          </Text>
          <View style={styles.actionButtons}>
            <TouchableOpacity
              style={styles.retryButton}
              onPress={onRetry}
              accessibilityRole="button"
              accessibilityLabel={t("common.retry")}
            >
              <MaterialIcons
                name="refresh"
                size={responsive.getValueForSize(18, 20, 22, 24)}
                color={theme.colors.text.inverse}
                accessibilityElementsHidden
                importantForAccessibility="no-hide-descendants"
              />
              <Text style={styles.retryButtonText}>{t("common.retry")}</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={styles.goBackButton}
              onPress={onGoBack}
              accessibilityRole="button"
              accessibilityLabel={backLabel ?? t("common.goBack")}
            >
              <MaterialIcons
                name="arrow-back"
                size={responsive.getValueForSize(18, 20, 22, 24)}
                color={theme.colors.text.secondary}
                accessibilityElementsHidden
                importantForAccessibility="no-hide-descendants"
              />
              <Text style={styles.goBackButtonText}>
                {backLabel ?? t("common.goBack")}
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
  responsive: ReturnType<typeof useTheme>["responsive"],
) => {
  return StyleSheet.create({
    loadingContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
      backgroundColor: theme.colors.background.default,
    },
    loadingText: {
      marginTop: responsive.spacing.elementSpacing,
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    timeoutActions: {
      marginTop: responsive.spacing.vertical,
      alignItems: "center",
    },
    timeoutMessage: {
      fontSize: responsive.getScaledFont("label"),
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginBottom: responsive.spacing.vertical,
      lineHeight: responsive.getScaledFont("label") * 1.4,
    },
    actionButtons: {
      flexDirection: "row",
      flexWrap: "wrap",
      justifyContent: "center",
      gap: responsive.spacing.elementSpacing,
    },
    retryButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
      minHeight: theme.geometry.touchTarget,
      borderRadius: theme.borderRadius.lg,
      backgroundColor: theme.colors.primary,
      gap: responsive.spacing.elementSpacing / 2,
      ...theme.shadows.sm,
    },
    retryButtonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "600",
    },
    goBackButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
      minHeight: theme.geometry.touchTarget,
      borderRadius: theme.borderRadius.lg,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      backgroundColor: theme.colors.background.surface,
      gap: responsive.spacing.elementSpacing / 2,
      ...theme.shadows.sm,
    },
    goBackButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "600",
      color: theme.colors.text.secondary,
    },
  });
};
