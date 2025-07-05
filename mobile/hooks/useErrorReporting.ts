import { useState } from "react";
import { Alert } from "react-native";
import { useMutation } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import * as errorReportingApi from "@/api/errorReporting";
import { ErrorType, ErrorReportRateLimitError } from "@/api/errorReporting";

interface UseErrorReportingProps {
  wordId: number;
  definitionId: number | null;
  isOnline: boolean;
  context: "quiz" | "flashcards";
  onSuccess?: () => void;
}

export const useErrorReporting = ({
  wordId,
  definitionId,
  isOnline,
  context,
  onSuccess,
}: UseErrorReportingProps) => {
  const { t } = useTranslation();
  const [showReportModal, setShowReportModal] = useState(false);

  const reportMutation = useMutation({
    mutationFn: ({ errorType }: { errorType: ErrorType }) => {
      if (!isOnline) {
        throw new Error("Reporting not available in offline mode");
      }

      return errorReportingApi.reportError({
        wordId,
        definitionId,
        errorType,
      });
    },
    onSuccess: () => {
      Alert.alert(t("common.success"), t(`${context}.reportSubmitted`));
      setShowReportModal(false);
      onSuccess?.();
    },
    onError: (error) => {
      if (error instanceof ErrorReportRateLimitError) {
        let message: string;

        if (error.windowType === "cooldown") {
          // Cooldown for specific error on this word
          message = error.retryAfter
            ? t(`${context}.cooldownError`, {
                minutes: Math.ceil(error.retryAfter / 60),
              })
            : error.message;
        } else {
          // Rate limit (hourly/daily)
          message = error.retryAfter
            ? t(`${context}.rateLimitError`, {
                minutes: Math.ceil(error.retryAfter / 60),
              })
            : error.message;
        }

        Alert.alert(t("common.error"), message);
      } else {
        // Check if it's actually a network/connectivity error
        const errorMessage = error?.message || "";
        const isNetworkError =
          errorMessage.includes("timeout") ||
          errorMessage.includes("fetch") ||
          errorMessage.includes("network") ||
          errorMessage.includes("offline") ||
          !isOnline;

        if (isNetworkError) {
          Alert.alert(t("common.error"), t("offline.featureUnavailable"));
        } else {
          // For debugging: log the full error to console
          console.warn("Error reporting failed:", {
            message: errorMessage,
            error: error,
            wordId,
            definitionId,
          });

          // Show the actual error message for API/server errors
          let displayMessage = errorMessage || t("errors.general");

          // Handle common validation errors more specifically
          if (errorMessage.includes("validation failed")) {
            displayMessage =
              "Unable to report error for this word. Please try again.";
          } else if (errorMessage.includes("Please try again later")) {
            displayMessage =
              "Error reporting is temporarily unavailable. Please try again in a few minutes.";
          }

          Alert.alert(t("common.error"), displayMessage);
        }
      }
    },
  });

  const handleReportError = (errorType: ErrorType) => {
    reportMutation.mutate({ errorType });
  };

  const openReportModal = () => {
    setShowReportModal(true);
  };

  const closeReportModal = () => {
    setShowReportModal(false);
  };

  return {
    showReportModal,
    isReporting: reportMutation.isPending,
    handleReportError,
    openReportModal,
    closeReportModal,
  };
};
