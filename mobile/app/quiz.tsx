import * as offlineQuizApi from "@/api/offlineWordlists";
import { ErrorReportModal } from "@/components/ErrorReportModal";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { QuizContent } from "@/components/quiz/QuizContent";
import { QuizHeader } from "@/components/quiz/QuizHeader";
import { QuizLoadingState } from "@/components/quiz/QuizLoadingState";
import { QuizModeToggle } from "@/components/quiz/QuizModeToggle";
import { QuizNextButton } from "@/components/quiz/QuizNextButton";
import { QuizOptions } from "@/components/quiz/QuizOptions";
import { QuizProgressBar } from "@/components/quiz/QuizProgressBar";
import { useTheme } from "@/contexts/ThemeContext";
import { useErrorReporting } from "@/hooks/useErrorReporting";
import { useInvalidateAnalytics } from "@/hooks/useInvalidateAnalytics";
import { useOffline } from "@/hooks/useOffline";
import { createCommonStyles } from "@/styles/common";
import { MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useLocalSearchParams } from "expo-router";
import React, { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { SafeAreaView, ScrollView, StyleSheet, Text, View } from "react-native";

const QuizScreen: React.FC = () => {
  const navigation = useNavigation();
  const { wordlistId, wordlistName } = useLocalSearchParams();
  const { t } = useTranslation();
  const { isOnline, isOfflineAvailable } = useOffline();
  const { invalidateAllAnalytics } = useInvalidateAnalytics();
  const { theme, responsive } = useTheme();
  const commonStyles = createCommonStyles(theme, responsive);
  const styles = createStyles(theme);

  // State
  const [selectedAnswer, setSelectedAnswer] = useState<number | null>(null);
  const [showResult, setShowResult] = useState(false);
  const [fastMode, setFastMode] = useState(false);
  const [quizCount, setQuizCount] = useState(0);
  const [correctCount, setCorrectCount] = useState(0);
  const [userInput, setUserInput] = useState("");
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [isLoadingNext, setIsLoadingNext] = useState(false);
  const [loadingTimeout, setLoadingTimeout] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const quizDisplayedAtRef = useRef(0);
  const loadingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Fetch quiz
  const {
    data: quiz,
    isLoading,
    refetch,
    error,
    isFetching,
  } = useQuery({
    queryKey: ["quiz", wordlistId, retryCount],
    queryFn: () => offlineQuizApi.newQuiz(Number(wordlistId)),
    retry: (failureCount, error) => {
      // Don't retry timeout errors automatically - let user decide
      if (error?.message?.includes("timeout")) {
        return false;
      }
      return isOnline ? failureCount < 2 : false;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    staleTime: 0, // Always fetch fresh quiz
    refetchOnWindowFocus: false,
    enabled: !!wordlistId, // Only fetch if wordlistId exists
  });

  // Error reporting hook
  const {
    showReportModal,
    isReporting,
    handleReportError,
    openReportModal,
    closeReportModal,
  } = useErrorReporting({
    wordId: quiz?.wordId || 0,
    definitionId: quiz?.definitionId || null,
    isOnline,
    context: "quiz",
    onSuccess: () => {
      handleNextQuiz();
    },
    quizDetails: quiz
      ? {
          quizType: quiz.type,
          value: quiz.value,
          options: quiz.options,
          answerIndex: quiz.answerIndex,
          context: "quiz",
        }
      : undefined,
  });

  // const [currentQuizId, setCurrentQuizId] = useState<number | null>(null); // Removed unused state

  useEffect(() => {
    if (quiz?.id) {
      // Quiz received (even if it's the same ID)
      quizDisplayedAtRef.current = Date.now();
      // setCurrentQuizId(quiz.id); // Removed unused state setter
      setIsLoadingNext(false);
      setLoadingTimeout(false);
      setRetryCount(0);
      // Clear loading timeout
      if (loadingTimeoutRef.current) {
        clearTimeout(loadingTimeoutRef.current);
        loadingTimeoutRef.current = null;
      }
    }
  }, [quiz, retryCount]); // Trigger on quiz object change or retry, not just ID

  // Handle loading timeout
  useEffect(() => {
    if (isLoadingNext || isFetching || !quiz) {
      // Clear any existing timeout
      if (loadingTimeoutRef.current) {
        clearTimeout(loadingTimeoutRef.current);
      }

      // Set timeout for loading state (aligned with API timeout)
      loadingTimeoutRef.current = setTimeout(() => {
        setLoadingTimeout(true);
      }, 16000); // Show timeout options after 16 seconds (1s after API timeout)
    } else {
      // Clear timeout when not loading
      if (loadingTimeoutRef.current) {
        clearTimeout(loadingTimeoutRef.current);
      }
      setLoadingTimeout(false);
    }

    return () => {
      if (loadingTimeoutRef.current) {
        clearTimeout(loadingTimeoutRef.current);
      }
    };
  }, [isLoadingNext, isFetching, quiz]);

  // Answer mutation
  const answerMutation = useMutation<void, Error, { success: boolean }>({
    mutationFn: ({ success }) =>
      offlineQuizApi.answerQuiz({
        definitionID: quiz!.definitionId,
        isCorrect: !!success,
        leitnerSystemTrackingID: quiz!.id,
        quizType: quiz!.type,
        responseTimeMs: Date.now() - quizDisplayedAtRef.current,
        wordID: quiz!.wordId,
        wordlistID: Number(wordlistId),
      }),
    onSuccess: () => {
      // Invalidate all analytics after answering a quiz (real-time updates)
      invalidateAllAnalytics(Number(wordlistId));

      if (fastMode) {
        setTimeout(() => {
          handleNextQuiz();
        }, 600);
      }
    },
    onError: console.error,
  });

  const handleAnswerSelect = (index: number) => {
    if (showResult) return;

    setSelectedAnswer(index);
    setShowResult(true);

    const isCorrect = index === quiz?.answerIndex;
    setQuizCount((prev) => prev + 1);
    if (isCorrect) {
      setCorrectCount((prev) => prev + 1);
    }

    answerMutation.mutate({ success: isCorrect });
  };

  const handleWriteAnswer = () => {
    if (!userInput.trim() || !quiz) return;

    setIsSubmitted(true);
    setShowResult(true);

    const isCorrect =
      userInput.toLowerCase().trim() ===
      quiz.options[quiz.answerIndex]?.toLowerCase();

    setQuizCount((prev) => prev + 1);
    if (isCorrect) {
      setCorrectCount((prev) => prev + 1);
    }

    answerMutation.mutate({ success: isCorrect });
  };

  const handleSkipQuestion = () => {
    if (!quiz) return;

    setIsSubmitted(true);
    setShowResult(true);

    setQuizCount((prev) => prev + 1);
    answerMutation.mutate({ success: false });
  };

  const handleNextQuiz = () => {
    setIsLoadingNext(true);
    setSelectedAnswer(null);
    setShowResult(false);
    setUserInput("");
    setIsSubmitted(false);
    setLoadingTimeout(false);
    // setCurrentQuizId(null); // Reset current quiz ID to ensure loading state shows
    refetch();
  };

  const handleRetryQuiz = () => {
    setRetryCount((prev) => prev + 1);
    setLoadingTimeout(false);
    setIsLoadingNext(true);
    // setCurrentQuizId(null); // Reset current quiz ID to ensure loading state shows
    refetch(); // Explicitly trigger refetch
  };

  const handleGoBack = () => {
    navigation.goBack();
  };

  const onPressFastModeToggle = () => {
    setFastMode((v) => !v);

    if (!fastMode && showResult) {
      handleNextQuiz();
    }
  };

  // Show initial loading screen while fetching first quiz
  if ((isLoading || isFetching || !quiz) && quizCount === 0) {
    return (
      <SafeAreaView style={[commonStyles.safeArea, styles.container]}>
        <View style={[styles.quizCard, { margin: 20 }]}>
          <QuizLoadingState
            isLoading={isLoading || isFetching}
            hasTimeout={loadingTimeout}
            error={error}
            onRetry={handleRetryQuiz}
            onGoBack={handleGoBack}
          />
        </View>
      </SafeAreaView>
    );
  }

  // Handle offline error state
  if (error && !isOnline && !isOfflineAvailable) {
    return (
      <SafeAreaView style={[commonStyles.safeArea, styles.container]}>
        <QuizHeader
          wordlistName={String(wordlistName)}
          correctCount={correctCount}
          quizCount={quizCount}
          isOnline={isOnline}
          onBackPress={() => navigation.goBack()}
          onReportPress={openReportModal}
        />
        <View style={styles.errorContainer}>
          <MaterialIcons
            name="cloud-off"
            size={64}
            color={theme.colors.text.secondary}
          />
          <Text style={styles.errorTitle}>{t("offline.premiumRequired")}</Text>
          <Text style={styles.errorMessage}>
            {t("offline.premiumRequiredMessage")}
          </Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={[commonStyles.safeArea, styles.container]}>
      <OfflineIndicator />

      <QuizHeader
        wordlistName={String(wordlistName)}
        correctCount={correctCount}
        quizCount={quizCount}
        isOnline={isOnline}
        onBackPress={() => navigation.goBack()}
        onReportPress={openReportModal}
      />

      <QuizProgressBar correctCount={correctCount} quizCount={quizCount} />

      <QuizModeToggle fastMode={fastMode} onToggle={onPressFastModeToggle} />

      <ScrollView
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.quizCard}>
          {isLoadingNext || isFetching || !quiz ? (
            <QuizLoadingState
              isLoading={isLoadingNext || isFetching}
              hasTimeout={loadingTimeout}
              error={error}
              onRetry={handleRetryQuiz}
              onGoBack={handleGoBack}
            />
          ) : (
            <>
              {quiz && (
                <QuizContent
                  quiz={quiz}
                  userInput={userInput}
                  setUserInput={setUserInput}
                  isSubmitted={isSubmitted}
                  onSubmitAnswer={handleWriteAnswer}
                  onSkipQuestion={handleSkipQuestion}
                />
              )}

              <QuizOptions
                quiz={quiz!}
                selectedAnswer={selectedAnswer}
                showResult={showResult}
                onAnswerSelect={handleAnswerSelect}
              />

              <QuizNextButton
                showResult={showResult}
                fastMode={fastMode}
                quizType={quiz?.type || ""}
                isSubmitted={isSubmitted}
                onNextQuiz={handleNextQuiz}
              />
            </>
          )}
        </View>
      </ScrollView>

      <ErrorReportModal
        visible={showReportModal}
        onClose={closeReportModal}
        onReportError={handleReportError}
        isLoading={isReporting}
        context="quiz"
      />
    </SafeAreaView>
  );
};

export default QuizScreen;

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    loadingContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      backgroundColor: theme.colors.background.default,
    },
    scrollContent: {
      paddingHorizontal: 20,
      paddingBottom: 20,
    },
    quizCard: {
      backgroundColor:
        theme.mode === "light"
          ? theme.colors.background.surface
          : theme.colors.background.elevated,
      borderRadius: theme.borderRadius.xl,
      padding: theme.spacing.lg,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? theme.colors.ui.border
          : theme.colors.ui.divider,
      ...theme.shadows.md,
      elevation: theme.mode === "light" ? 6 : 10,
    },
    errorContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: 40,
    },
    errorTitle: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginTop: 16,
      marginBottom: 8,
    },
    errorMessage: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: 22,
    },
  });
