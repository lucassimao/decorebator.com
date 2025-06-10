import React from "react";
import { useTranslation } from "react-i18next";
import { LoadingWithTimeout } from "../LoadingWithTimeout";

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