import { ErrorType } from "@/api/errorReporting";
import * as offlineQuizApi from "@/api/offlineWordlists";
import { ErrorReportModal } from "@/components/ErrorReportModal";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { QuizContent } from "@/components/quiz/QuizContent";
import { QuizHeader } from "@/components/quiz/QuizHeader";
import { QuizModeToggle } from "@/components/quiz/QuizModeToggle";
import { QuizNextButton } from "@/components/quiz/QuizNextButton";
import { QuizOptions } from "@/components/quiz/QuizOptions";
import { QuizProgressBar } from "@/components/quiz/QuizProgressBar";
import { QuizLoadingState } from "@/components/quiz/QuizLoadingState";
import { useOffline } from "@/hooks/useOffline";
import { useErrorReporting } from "@/hooks/useErrorReporting";
import { useInvalidateAnalytics } from "@/hooks/useInvalidateAnalytics";
import { MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useLocalSearchParams } from "expo-router";
import React, { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  ImageBackground,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";

const { width: SCREEN_WIDTH, height: SCREEN_HEIGHT } = Dimensions.get("window");

interface Quiz {
  id: number;
  type: string;
  value: string;
  options: string[];
  answerIndex: number;
  pos?: string;
  pronunciation?: string;
  audioURL?: string;
  imageDescription?: string;
  wordId: number;
  definitionId: number;
}

const QuizScreen: React.FC = () => {
  const navigation = useNavigation();
  const { wordlistId, wordlistName } = useLocalSearchParams();
  const { t } = useTranslation();
  const { isOnline, isOfflineAvailable } = useOffline();
  const { invalidateBoxDistribution } = useInvalidateAnalytics();

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
      if (error.message.includes('timeout')) {
        return false;
      }
      return isOnline ? failureCount < 2 : false;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    staleTime: 0, // Always fetch fresh quiz
    refetchOnWindowFocus: false,
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
    definitionId: quiz?.definitionId || 0,
    isOnline,
    context: "quiz",
    onSuccess: () => {
      handleNextQuiz();
    },
  });

  useEffect(() => {
    if (quiz?.id) {
      quizDisplayedAtRef.current = Date.now();
      setIsLoadingNext(false);
      setLoadingTimeout(false);
      setRetryCount(0);
      // Clear loading timeout
      if (loadingTimeoutRef.current) {
        clearTimeout(loadingTimeoutRef.current);
      }
    }
  }, [quiz?.id]);

  // Handle loading timeout
  useEffect(() => {
    if (isLoadingNext || (isFetching && !quiz)) {
      // Clear any existing timeout
      if (loadingTimeoutRef.current) {
        clearTimeout(loadingTimeoutRef.current);
      }
      
      // Set timeout for loading state
      loadingTimeoutRef.current = setTimeout(() => {
        setLoadingTimeout(true);
      }, 10000); // Show timeout options after 10 seconds
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
      // Invalidate box distribution after answering a quiz
      invalidateBoxDistribution(Number(wordlistId));
      
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
    refetch();
  };

  const handleRetryQuiz = () => {
    setRetryCount(prev => prev + 1);
    setLoadingTimeout(false);
    setIsLoadingNext(true);
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

  if (isLoading && quizCount === 0) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#FF7B54" />
      </View>
    );
  }

  // Handle offline error state
  if (error && !isOnline && !isOfflineAvailable) {
    return (
      <ImageBackground
        source={require("@/assets/images/dashboard-bg.png")}
        style={styles.backgroundImage}
        resizeMode="cover"
      >
        <SafeAreaView style={styles.container}>
          <QuizHeader
            wordlistName={String(wordlistName)}
            correctCount={correctCount}
            quizCount={quizCount}
            isOnline={isOnline}
            onBackPress={() => navigation.goBack()}
            onReportPress={openReportModal}
          />
          <View style={styles.errorContainer}>
            <MaterialIcons name="cloud-off" size={64} color="#636E72" />
            <Text style={styles.errorTitle}>
              {t("offline.premiumRequired")}
            </Text>
            <Text style={styles.errorMessage}>
              {t("offline.premiumRequiredMessage")}
            </Text>
          </View>
        </SafeAreaView>
      </ImageBackground>
    );
  }

  return (
    <ImageBackground
      source={require("@/assets/images/dashboard-bg.png")}
      style={styles.backgroundImage}
      resizeMode="cover"
    >
      <SafeAreaView style={styles.container}>
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
            {isLoadingNext || (isFetching && !quiz) ? (
              <QuizLoadingState
                isLoading={true}
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
    </ImageBackground>
  );
};

export default QuizScreen;

const styles = StyleSheet.create({
  backgroundImage: {
    flex: 1,
    width: SCREEN_WIDTH,
    height: SCREEN_HEIGHT,
  },
  container: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: "#FDF6E3",
  },
  scrollContent: {
    paddingHorizontal: 20,
    paddingBottom: 20,
  },
  quizCard: {
    backgroundColor: "rgba(255, 255, 255, 0.95)",
    borderRadius: 24,
    padding: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    elevation: 5,
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
    color: "#2D3436",
    marginTop: 16,
    marginBottom: 8,
  },
  errorMessage: {
    fontSize: 16,
    color: "#636E72",
    textAlign: "center",
    lineHeight: 22,
  },
});
