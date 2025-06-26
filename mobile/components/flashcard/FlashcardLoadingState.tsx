import React from "react";
import { useTranslation } from "react-i18next";
import { LoadingWithTimeout } from "../LoadingWithTimeout";
import { MaterialIcons } from "@expo/vector-icons";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";

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
  const styles = createStyles(theme);

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
      gap: 12,
    },
    retryButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 20,
      paddingVertical: 12,
      borderRadius: 25,
      gap: 8,
    },
    retryButtonText: {
      color: "#FFFFFF",
      fontSize: 16,
      fontWeight: "600",
    },
    goBackButton: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 20,
      paddingVertical: 12,
      borderRadius: 25,
      borderWidth: 1,
      borderColor: theme.colors.border.light,
      gap: 8,
    },
    goBackButtonText: {
      fontSize: 16,
      fontWeight: "600",
    },
  });
