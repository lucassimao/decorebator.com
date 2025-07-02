import React from "react";
import { useTranslation } from "react-i18next";
import { LoadingWithTimeout } from "../LoadingWithTimeout";
import { MaterialIcons } from "@expo/vector-icons";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

interface FlashcardLoadingStateProps {
  isLoading: boolean;
  hasTimeout: boolean;
  error?: Error | null;
  isLoadingPosition?: boolean;
  onRetry: () => void;
  onGoBack: () => void;
  colors?: {
    primary: string;
    white: string;
    textMedium: string;
  };
}

export const FlashcardLoadingState: React.FC<FlashcardLoadingStateProps> = ({
  isLoading,
  hasTimeout,
  error,
  isLoadingPosition = false,
  onRetry,
  onGoBack,
  colors = {
    primary: "#FF7B54",
    white: "#FFFFFF",
    textMedium: "#636E72",
  },
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, spacing);

  // Check if this is a "no words with definitions" error
  const isProcessingError =
    error?.message?.includes("no words found") ||
    error?.message?.includes("no definitions");

  if (isProcessingError && !isLoading && !isLoadingPosition) {
    return (
      <View style={styles.processingContainer}>
        <MaterialIcons
          name="hourglass-empty"
          size={48}
          color={theme.colors.text.secondary}
        />
        <Text style={styles.processingTitle}>
          {t("flashcards.wordsProcessing")}
        </Text>
        <Text style={styles.processingMessage}>
          {t("flashcards.wordsProcessingMessage")}
        </Text>
        <View style={styles.actionButtons}>
          <TouchableOpacity
            style={[styles.retryButton, { backgroundColor: colors.primary }]}
            onPress={onRetry}
          >
            <MaterialIcons name="refresh" size={20} color="#FFFFFF" />
            <Text style={styles.retryButtonText}>{t("common.tryAgain")}</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.goBackButton, { backgroundColor: colors.white }]}
            onPress={onGoBack}
          >
            <MaterialIcons
              name="arrow-back"
              size={20}
              color={colors.textMedium}
            />
            <Text
              style={[styles.goBackButtonText, { color: colors.textMedium }]}
            >
              {t("common.goBack")}
            </Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  return (
    <LoadingWithTimeout
      isLoading={isLoading}
      hasTimeout={hasTimeout}
      error={error}
      primaryColor={colors.primary}
      backgroundColor={colors.white}
      textColor={colors.textMedium}
      loadingMessage={t("flashcards.loadingWords")}
      timeoutMessage={t("flashcards.loadingTakingLonger")}
      onRetry={onRetry}
      onGoBack={onGoBack}
      showTimeoutActions={!isLoadingPosition}
    />
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  spacing: ReturnType<typeof useResponsiveSpacing>,
) =>
  StyleSheet.create({
    processingContainer: {
      minHeight: isTablet ? 400 : 300,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: spacing.lg,
    },
    processingTitle: {
      fontSize: isTablet ? 24 : 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginTop: spacing.md,
      marginBottom: spacing.xs,
      textAlign: "center",
    },
    processingMessage: {
      fontSize: isTablet ? 18 : 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: isTablet ? 26 : 22,
      marginBottom: spacing.xl,
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
