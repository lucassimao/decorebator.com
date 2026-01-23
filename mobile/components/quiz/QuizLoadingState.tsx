import React from "react";
import { useTranslation } from "react-i18next";
import { LoadingWithTimeout } from "../LoadingWithTimeout";
import { MaterialIcons } from "@expo/vector-icons";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";

interface QuizLoadingStateProps {
  isLoading: boolean;
  hasTimeout: boolean;
  error?: Error | null;
  onRetry: () => void;
  onGoBack: () => void;
}

export const QuizLoadingState: React.FC<QuizLoadingStateProps> = ({
  isLoading,
  hasTimeout,
  error,
  onRetry,
  onGoBack,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

  // Check if the error is due to words still being processed
  const isProcessingError =
    error?.message?.includes("no definitions found in wordlist") ||
    error?.message?.includes("no unlearned words in wordlist");
  const errorStatus = (error as any)?.data?.status;
  const isNoMatchingTypesError =
    errorStatus === "no_matching_quiz_types" ||
    error?.message?.includes("no quiz types available for selection");

  if (isProcessingError && !isLoading) {
    return (
      <View style={styles.processingContainer}>
        <MaterialIcons
          name="hourglass-empty"
          size={48}
          color={theme.colors.text.secondary}
        />
        <Text style={styles.processingTitle}>{t("quiz.wordsProcessing")}</Text>
        <Text style={styles.processingMessage}>
          {t("quiz.wordsProcessingMessage")}
        </Text>
        <View style={styles.actionButtons}>
          <TouchableOpacity
            style={[
              styles.retryButton,
              { backgroundColor: theme.colors.primary },
            ]}
            onPress={onRetry}
          >
            <MaterialIcons name="refresh" size={20} color="#FFFFFF" />
            <Text style={styles.retryButtonText}>{t("common.tryAgain")}</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[
              styles.goBackButton,
              { backgroundColor: theme.colors.background.surface },
            ]}
            onPress={onGoBack}
          >
            <MaterialIcons
              name="arrow-back"
              size={20}
              color={theme.colors.text.primary}
            />
            <Text
              style={[
                styles.goBackButtonText,
                { color: theme.colors.text.primary },
              ]}
            >
              {t("common.goBack")}
            </Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  if (isNoMatchingTypesError && !isLoading) {
    return (
      <View style={styles.processingContainer}>
        <MaterialIcons
          name="tune"
          size={48}
          color={theme.colors.text.secondary}
        />
        <Text style={styles.processingTitle}>
          {t("quiz.quizTypeEmptyTitle", "No matching quiz types")}
        </Text>
        <Text style={styles.processingMessage}>
          {t(
            "quiz.quizTypeEmptyMessage",
            "Try selecting different quiz types or select all.",
          )}
        </Text>
        <View style={styles.actionButtons}>
          <TouchableOpacity
            style={[
              styles.retryButton,
              { backgroundColor: theme.colors.primary },
            ]}
            onPress={onRetry}
          >
            <MaterialIcons name="refresh" size={20} color="#FFFFFF" />
            <Text style={styles.retryButtonText}>{t("common.tryAgain")}</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[
              styles.goBackButton,
              { backgroundColor: theme.colors.background.surface },
            ]}
            onPress={onGoBack}
          >
            <MaterialIcons
              name="arrow-back"
              size={20}
              color={theme.colors.text.primary}
            />
            <Text
              style={[
                styles.goBackButtonText,
                { color: theme.colors.text.primary },
              ]}
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
      loadingMessage={t("quiz.loadingNextQuestion")}
      timeoutMessage={t("quiz.loadingTakingLonger")}
      onRetry={onRetry}
      onGoBack={onGoBack}
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
