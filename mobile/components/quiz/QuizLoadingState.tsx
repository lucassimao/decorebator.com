import React from "react";
import { useTranslation } from "react-i18next";
import { LoadingWithTimeout } from "../LoadingWithTimeout";

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