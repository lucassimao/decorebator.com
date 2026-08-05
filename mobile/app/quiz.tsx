import * as offlineQuizApi from "@/api/offlineWordlists";
import AppReviewModal from "@/components/common/AppReviewModal";
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
import { useAppReview } from "@/hooks/useAppReview";
import { useErrorReporting } from "@/hooks/useErrorReporting";
import { useInvalidateAnalytics } from "@/hooks/useInvalidateAnalytics";
import { useOffline } from "@/hooks/useOffline";
import { createCommonStyles } from "@/styles/common";
import { MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useLocalSearchParams } from "expo-router";
import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import { usePostHog } from "posthog-react-native";
import {
  ACTIVATION_EVENT_NAMES,
  captureActivationEvent,
  createActivationSessionId,
} from "@/utils/activationEvents";
import {
  Modal,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { useUserSession } from "@/hooks/useUserSession";
import { registerDevicePushToken } from "@/utils/pushNotifications";
import type { QuizType } from "@/api/wordlists";
import {
  claimQuizSubmission,
  prepareQuizTypeSelection,
  resetQuizSubmission,
} from "@/utils/quizSession";

const QuizScreen: React.FC = () => {
  const navigation = useNavigation();
  const { wordlistId, wordlistName } = useLocalSearchParams();
  const { t } = useTranslation();
  const posthog = usePostHog();
  const { user, isPremium } = useUserSession();
  const { isOnline, isOfflineAvailable } = useOffline();
  const { invalidateAllAnalytics } = useInvalidateAnalytics();
  const { theme, responsive } = useTheme();
  const commonStyles = createCommonStyles(theme, responsive);
  const styles = createStyles(theme);
  const notificationPromptedRef = useRef(false);
  const allQuizTypes = useMemo<QuizType[]>(
    () => [
      "GUESS_MEANING",
      "WORD_FROM_MEANING",
      "WORD_FROM_IMAGE",
      "COMPLETE_SENTENCE",
      "WRITE_WORD_FROM_DEFINITION",
      "WORD_FROM_AUDIO",
      "MEANING_FROM_AUDIO",
      "WORD_FROM_MEANING_AUDIO",
      "WORD_FROM_EXAMPLE_AUDIO",
    ],
    [],
  );

  // App review prompt
  const {
    showReviewModal,
    checkAndPromptReview,
    handleLoveIt,
    handleNotReally,
    handleMaybeLater,
    closeReviewModal,
  } = useAppReview();
  const [pendingNavigation, setPendingNavigation] = useState(false);

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
  const quizSessionIdRef = useRef(createActivationSessionId());
  const quizSessionStartedAtRef = useRef<number | null>(null);
  const quizSessionStartedRef = useRef(false);
  const quizSessionCompletedRef = useRef(false);
  const quizCountRef = useRef(0);
  const correctCountRef = useRef(0);
  const submissionLockRef = useRef(false);
  const [showQuizTypeModal, setShowQuizTypeModal] = useState(false);
  const [selectedQuizTypes, setSelectedQuizTypes] =
    useState<QuizType[]>(allQuizTypes);
  const [draftQuizTypes, setDraftQuizTypes] =
    useState<QuizType[]>(allQuizTypes);

  // Fetch quiz
  const quizTypeSelectionKey = useMemo(() => {
    if (!isPremium || selectedQuizTypes.length === allQuizTypes.length) {
      return "all";
    }
    return selectedQuizTypes.slice().sort().join(",");
  }, [allQuizTypes.length, isPremium, selectedQuizTypes]);
  const quizTypeFilter =
    isPremium && selectedQuizTypes.length !== allQuizTypes.length
      ? selectedQuizTypes
      : undefined;
  const {
    data: quiz,
    isLoading,
    refetch,
    error,
    isFetching,
  } = useQuery({
    queryKey: ["quiz", wordlistId, retryCount, quizTypeSelectionKey],
    queryFn: () => offlineQuizApi.newQuiz(Number(wordlistId), quizTypeFilter),
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

  useEffect(() => {
    if (!quiz?.id || quizSessionStartedRef.current) return;

    quizSessionStartedRef.current = true;
    quizSessionStartedAtRef.current = Date.now();
    captureActivationEvent(
      posthog,
      ACTIVATION_EVENT_NAMES.QUIZ_SESSION_STARTED,
      {
        sessionId: quizSessionIdRef.current,
        wordlistId: Number(wordlistId),
        quizMode: fastMode ? "fast" : "standard",
        source: "quiz_screen",
      },
    );
  }, [fastMode, posthog, quiz?.id, wordlistId]);

  useEffect(() => {
    setSelectedQuizTypes(allQuizTypes);
    setDraftQuizTypes(allQuizTypes);
  }, [wordlistId, allQuizTypes]);

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
    onSuccess: (_, variables) => {
      // Invalidate all analytics after answering a quiz (real-time updates)
      invalidateAllAnalytics(Number(wordlistId));
      captureActivationEvent(posthog, ACTIVATION_EVENT_NAMES.QUIZ_ANSWERED, {
        wordlistId: Number(wordlistId),
        correct: variables.success,
        responseTimeMs: Date.now() - quizDisplayedAtRef.current,
        sessionId: quizSessionIdRef.current,
        ...(quiz?.type != null ? { quizType: quiz.type } : {}),
      });

      if (
        quizCount === 1 &&
        !notificationPromptedRef.current &&
        user?.notificationsEnabled
      ) {
        notificationPromptedRef.current = true;
        AsyncStorage.getItem("pushPromptedAfterFirstQuiz")
          .then(async (stored) => {
            if (stored) {
              return;
            }
            await AsyncStorage.setItem("pushPromptedAfterFirstQuiz", "true");
            await registerDevicePushToken({ prompt: true });
          })
          .catch((error) => {
            console.warn("Failed to prompt push registration:", error);
          });
      }

      if (fastMode) {
        setTimeout(() => {
          handleNextQuiz();
        }, 600);
      }
    },
    onError: console.error,
  });

  const handleAnswerSelect = (index: number) => {
    if (!claimQuizSubmission(submissionLockRef)) return;

    setSelectedAnswer(index);
    setShowResult(true);

    const isCorrect = index === quiz?.answerIndex;
    quizCountRef.current += 1;
    setQuizCount((prev) => prev + 1);
    if (isCorrect) {
      correctCountRef.current += 1;
      setCorrectCount((prev) => prev + 1);
    }

    answerMutation.mutate({ success: isCorrect });
  };

  const handleWriteAnswer = () => {
    if (!userInput.trim() || !quiz || !claimQuizSubmission(submissionLockRef))
      return;

    setIsSubmitted(true);
    setShowResult(true);

    const isCorrect =
      userInput.toLowerCase().trim() ===
      quiz.options[quiz.answerIndex]?.toLowerCase();

    quizCountRef.current += 1;
    setQuizCount((prev) => prev + 1);
    if (isCorrect) {
      correctCountRef.current += 1;
      setCorrectCount((prev) => prev + 1);
    }

    answerMutation.mutate({ success: isCorrect });
  };

  const handleSkipQuestion = () => {
    if (!quiz || !claimQuizSubmission(submissionLockRef)) return;

    setIsSubmitted(true);
    setShowResult(true);

    quizCountRef.current += 1;
    setQuizCount((prev) => prev + 1);
    answerMutation.mutate({ success: false });
  };

  const handleNextQuiz = () => {
    resetQuizSubmission(submissionLockRef);
    setIsLoadingNext(true);
    setSelectedAnswer(null);
    setShowResult(false);
    setUserInput("");
    setIsSubmitted(false);
    setLoadingTimeout(false);
    // setCurrentQuizId(null); // Reset current quiz ID to ensure loading state shows
    refetch();
  };

  const completeQuizSession = useCallback(() => {
    if (!quizSessionStartedRef.current || quizSessionCompletedRef.current) {
      navigation.goBack();
      return;
    }

    quizSessionCompletedRef.current = true;
    captureActivationEvent(
      posthog,
      ACTIVATION_EVENT_NAMES.QUIZ_SESSION_COMPLETED,
      {
        sessionId: quizSessionIdRef.current,
        wordlistId: Number(wordlistId),
        answeredCount: quizCountRef.current,
        correctCount: correctCountRef.current,
        durationMs: quizSessionStartedAtRef.current
          ? Date.now() - quizSessionStartedAtRef.current
          : undefined,
        outcome: "success",
      },
    );
    navigation.goBack();
  }, [navigation, posthog, wordlistId]);

  useEffect(() => {
    const sessionId = quizSessionIdRef.current;

    return () => {
      if (!quizSessionStartedRef.current || quizSessionCompletedRef.current) {
        return;
      }

      quizSessionCompletedRef.current = true;
      captureActivationEvent(
        posthog,
        ACTIVATION_EVENT_NAMES.QUIZ_SESSION_COMPLETED,
        {
          sessionId,
          wordlistId: Number(wordlistId),
          answeredCount: quizCountRef.current,
          correctCount: correctCountRef.current,
          durationMs: quizSessionStartedAtRef.current
            ? Date.now() - quizSessionStartedAtRef.current
            : undefined,
          outcome: "cancelled",
        },
      );
    };
  }, [posthog, wordlistId]);

  const handleRetryQuiz = () => {
    setRetryCount((prev) => prev + 1);
    setLoadingTimeout(false);
    setIsLoadingNext(true);
    // setCurrentQuizId(null); // Reset current quiz ID to ensure loading state shows
    refetch(); // Explicitly trigger refetch
  };

  const openQuizTypeModal = () => {
    setDraftQuizTypes(selectedQuizTypes);
    setShowQuizTypeModal(true);
  };

  const toggleQuizType = (quizType: QuizType) => {
    setDraftQuizTypes((prev) =>
      prev.includes(quizType)
        ? prev.filter((type) => type !== quizType)
        : [...prev, quizType],
    );
  };

  const handleSelectAllQuizTypes = () => {
    setDraftQuizTypes(allQuizTypes);
  };

  const handleClearQuizTypes = () => {
    setDraftQuizTypes([]);
  };

  const handleApplyQuizTypes = () => {
    if (!prepareQuizTypeSelection(submissionLockRef, draftQuizTypes)) return;
    setSelectedQuizTypes(draftQuizTypes);
    setShowQuizTypeModal(false);
    setRetryCount(0);
    setIsLoadingNext(true);
    setSelectedAnswer(null);
    setShowResult(false);
    setUserInput("");
    setIsSubmitted(false);
    setLoadingTimeout(false);
  };

  const translateQuizType = (quizType: QuizType) => {
    const key = `analytics.quizTypes.${quizType.toLowerCase()}`;
    return t(key, quizType.replace(/_/g, " ").toLowerCase());
  };

  const handleGoBack = async () => {
    // Check if we should show review prompt before navigating
    const shouldShowReview = await checkAndPromptReview(
      quizCount,
      correctCount,
    );
    if (shouldShowReview) {
      setPendingNavigation(true);
      return;
    }
    completeQuizSession();
  };

  // Handle navigation after review modal closes
  const handleReviewLoveIt = async () => {
    await handleLoveIt();
    if (pendingNavigation) {
      setPendingNavigation(false);
      completeQuizSession();
    }
  };

  const handleReviewNotReally = async () => {
    await handleNotReally();
    if (pendingNavigation) {
      setPendingNavigation(false);
      completeQuizSession();
    }
  };

  const handleReviewMaybeLater = () => {
    handleMaybeLater();
    if (pendingNavigation) {
      setPendingNavigation(false);
      completeQuizSession();
    }
  };

  const handleReviewClose = () => {
    closeReviewModal();
    if (pendingNavigation) {
      setPendingNavigation(false);
      completeQuizSession();
    }
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
          showQuizTypeSelector={isPremium}
          quizTypeCount={selectedQuizTypes.length}
          quizTypeTotal={allQuizTypes.length}
          isQuizTypeDisabled={!isOnline}
          onQuizTypePress={openQuizTypeModal}
          onBackPress={handleGoBack}
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
        showQuizTypeSelector={isPremium}
        quizTypeCount={selectedQuizTypes.length}
        quizTypeTotal={allQuizTypes.length}
        isQuizTypeDisabled={!isOnline}
        onQuizTypePress={openQuizTypeModal}
        onBackPress={handleGoBack}
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

      <Modal
        visible={showQuizTypeModal}
        transparent
        animationType="slide"
        onRequestClose={() => setShowQuizTypeModal(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.modalCard}>
            <Text style={styles.modalTitle}>
              {t("quiz.quizTypeTitle", "Quiz types")}
            </Text>
            <Text style={styles.modalSubtitle}>
              {t(
                "quiz.quizTypeSubtitle",
                "Choose which quiz types to practice in this session.",
              )}
            </Text>

            <View style={styles.modalQuickActions}>
              <TouchableOpacity
                style={styles.modalQuickAction}
                onPress={handleSelectAllQuizTypes}
              >
                <Text style={styles.modalQuickActionText}>
                  {t("quiz.quizTypeSelectAll", "Select all")}
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={styles.modalQuickAction}
                onPress={handleClearQuizTypes}
              >
                <Text style={styles.modalQuickActionText}>
                  {t("quiz.quizTypeClear", "Clear")}
                </Text>
              </TouchableOpacity>
            </View>

            <ScrollView style={styles.quizTypeList}>
              {allQuizTypes.map((quizType) => {
                const isSelected = draftQuizTypes.includes(quizType);
                return (
                  <TouchableOpacity
                    key={quizType}
                    style={styles.quizTypeRow}
                    onPress={() => toggleQuizType(quizType)}
                  >
                    <MaterialIcons
                      name={
                        isSelected ? "check-box" : "check-box-outline-blank"
                      }
                      size={22}
                      color={
                        isSelected
                          ? theme.colors.primary
                          : theme.colors.text.secondary
                      }
                    />
                    <Text style={styles.quizTypeLabel}>
                      {translateQuizType(quizType)}
                    </Text>
                  </TouchableOpacity>
                );
              })}
            </ScrollView>

            <View style={styles.modalActions}>
              <TouchableOpacity
                style={styles.modalCancelButton}
                onPress={() => setShowQuizTypeModal(false)}
              >
                <Text style={styles.modalCancelText}>
                  {t("common.cancel", "Cancel")}
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[
                  styles.modalApplyButton,
                  draftQuizTypes.length === 0 && styles.modalApplyDisabled,
                ]}
                onPress={handleApplyQuizTypes}
                disabled={draftQuizTypes.length === 0}
              >
                <Text style={styles.modalApplyText}>
                  {t("common.apply", "Apply")}
                </Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>

      <ErrorReportModal
        visible={showReportModal}
        onClose={closeReportModal}
        onReportError={handleReportError}
        isLoading={isReporting}
        context="quiz"
      />

      <AppReviewModal
        visible={showReviewModal}
        onLoveIt={handleReviewLoveIt}
        onNotReally={handleReviewNotReally}
        onMaybeLater={handleReviewMaybeLater}
        onClose={handleReviewClose}
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
    modalOverlay: {
      flex: 1,
      backgroundColor: "rgba(0, 0, 0, 0.4)",
      justifyContent: "flex-end",
    },
    modalCard: {
      backgroundColor: theme.colors.background.surface,
      borderTopLeftRadius: 24,
      borderTopRightRadius: 24,
      paddingHorizontal: 20,
      paddingVertical: 16,
      maxHeight: "80%",
    },
    modalTitle: {
      fontSize: 20,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    modalSubtitle: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginTop: 6,
      marginBottom: 12,
    },
    modalQuickActions: {
      flexDirection: "row",
      justifyContent: "space-between",
      marginBottom: 12,
    },
    modalQuickAction: {
      paddingVertical: 6,
      paddingHorizontal: 12,
      borderRadius: 16,
      backgroundColor: theme.colors.background.elevated,
    },
    modalQuickActionText: {
      fontSize: 13,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    quizTypeList: {
      marginBottom: 16,
    },
    quizTypeRow: {
      flexDirection: "row",
      alignItems: "center",
      paddingVertical: 10,
      gap: 10,
    },
    quizTypeLabel: {
      fontSize: 15,
      color: theme.colors.text.primary,
      flex: 1,
    },
    modalActions: {
      flexDirection: "row",
      justifyContent: "space-between",
      gap: 12,
      paddingBottom: 6,
    },
    modalCancelButton: {
      flex: 1,
      paddingVertical: 12,
      borderRadius: 20,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      alignItems: "center",
    },
    modalCancelText: {
      fontSize: 15,
      fontWeight: "600",
      color: theme.colors.text.secondary,
    },
    modalApplyButton: {
      flex: 1,
      paddingVertical: 12,
      borderRadius: 20,
      backgroundColor: theme.colors.primary,
      alignItems: "center",
    },
    modalApplyDisabled: {
      opacity: 0.5,
    },
    modalApplyText: {
      fontSize: 15,
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
  });
