import React from "react";
import { useTranslation } from "react-i18next";
import { LoadingWithTimeout } from "../LoadingWithTimeout";
import { MaterialIcons } from "@expo/vector-icons";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";
import { isFlashcardProcessingError } from "@/utils/flashcardPresentation";

interface FlashcardLoadingStateProps {
  isLoading: boolean;
  hasTimeout: boolean;
  error?: Error | null;
  isLoadingPosition?: boolean;
  onRetry: () => void;
  onGoBack: () => void;
}

export const FlashcardLoadingState: React.FC<FlashcardLoadingStateProps> = ({
  isLoading,
  hasTimeout,
  error,
  isLoadingPosition = false,
  onRetry,
  onGoBack,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

  // Check if this is a "no words with definitions" error
  const isProcessingError = isFlashcardProcessingError(error);

  if (isProcessingError && !isLoading && !isLoadingPosition) {
    return (
      <View style={styles.processingContainer}>
        <MaterialIcons
          name="hourglass-empty"
          size={48}
          color={theme.colors.text.secondary}
          accessibilityElementsHidden
          importantForAccessibility="no-hide-descendants"
        />
        <Text style={styles.processingTitle}>
          {t("flashcards.wordsProcessing")}
        </Text>
        <Text style={styles.processingMessage}>
          {t("flashcards.wordsProcessingMessage")}
        </Text>
        <View style={styles.actionButtons}>
          <TouchableOpacity
            style={styles.retryButton}
            onPress={onRetry}
            accessibilityRole="button"
            accessibilityLabel={t("common.tryAgain")}
          >
            <MaterialIcons
              name="refresh"
              size={20}
              color={theme.colors.roles.onAction}
              accessibilityElementsHidden
              importantForAccessibility="no-hide-descendants"
            />
            <Text style={styles.retryButtonText}>{t("common.tryAgain")}</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={styles.goBackButton}
            onPress={onGoBack}
            accessibilityRole="button"
            accessibilityLabel={t("flashcards.backToWordlists")}
          >
            <MaterialIcons
              name="arrow-back"
              size={20}
              color={theme.colors.text.secondary}
              accessibilityElementsHidden
              importantForAccessibility="no-hide-descendants"
            />
            <Text style={styles.goBackButtonText}>
              {t("flashcards.backToWordlists")}
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
      loadingMessage={t("flashcards.loadingWords")}
      timeoutMessage={t("flashcards.loadingTakingLonger")}
      onRetry={onRetry}
      onGoBack={onGoBack}
      showTimeoutActions={!isLoadingPosition}
      timeoutErrorMessage={t("flashcards.requestTimedOut")}
      slowConnectionMessage={t("flashcards.slowConnection")}
      backLabel={t("flashcards.backToWordlists")}
    />
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    processingContainer: {
      minHeight: 300,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: 20,
    },
    processingTitle: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginTop: 16,
      marginBottom: 8,
      textAlign: "center",
    },
    processingMessage: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: 22,
      marginBottom: 24,
    },
    actionButtons: {
      flexDirection: "row",
      flexWrap: "wrap",
      justifyContent: "center",
      gap: 12,
    },
    retryButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 20,
      paddingVertical: 12,
      minHeight: theme.geometry.touchTarget,
      borderRadius: 25,
      gap: 8,
      backgroundColor: theme.colors.roles.action,
    },
    retryButtonText: {
      color: theme.colors.roles.onAction,
      fontSize: 16,
      fontWeight: "600",
    },
    goBackButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 20,
      paddingVertical: 12,
      minHeight: theme.geometry.touchTarget,
      borderRadius: 25,
      borderWidth: 1,
      borderColor: theme.colors.border.light,
      gap: 8,
      backgroundColor: theme.colors.background.surface,
    },
    goBackButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.secondary,
    },
  });
