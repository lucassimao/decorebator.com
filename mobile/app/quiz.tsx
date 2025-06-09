import * as errorReportingApi from "@/api/errorReporting";
import { ErrorType, ErrorReportRateLimitError } from "@/api/errorReporting";
import * as offlineQuizApi from "@/api/offlineWordlists";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { ErrorReportModal } from "@/components/ErrorReportModal";
import { QuizHeader } from "@/components/quiz/QuizHeader";
import { QuizProgressBar } from "@/components/quiz/QuizProgressBar";
import { QuizModeToggle } from "@/components/quiz/QuizModeToggle";
import { QuizContent } from "@/components/quiz/QuizContent";
import { QuizOptions } from "@/components/quiz/QuizOptions";
import { QuizNextButton } from "@/components/quiz/QuizNextButton";
import { useOffline } from "@/hooks/useOffline";
import { MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useAudioPlayer, useAudioPlayerStatus } from "expo-audio";
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

  // State
  const [selectedAnswer, setSelectedAnswer] = useState<number | null>(null);
  const [showResult, setShowResult] = useState(false);
  const [fastMode, setFastMode] = useState(false);
  const [showReportModal, setShowReportModal] = useState(false);
  const player = useAudioPlayer();
  const { didJustFinish } = useAudioPlayerStatus(player);
  const [quizCount, setQuizCount] = useState(0);
  const [correctCount, setCorrectCount] = useState(0);
  const [userInput, setUserInput] = useState("");
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [isLoadingNext, setIsLoadingNext] = useState(false);
  const quizDisplayedAtRef = useRef(0);

  // Reset player
  useEffect(() => {
    if (didJustFinish) {
      player.seekTo(0);
    }
  }, [didJustFinish, player]);

  // Fetch quiz
  const {
    data: quiz,
    isLoading,
    refetch,
    error,
    isFetching,
  } = useQuery({
    queryKey: ["quiz", wordlistId],
    queryFn: () => offlineQuizApi.newQuiz(Number(wordlistId)),
    retry: isOnline ? 3 : 0,
  });

  useEffect(() => {
    if (quiz?.id) {
      quizDisplayedAtRef.current = Date.now();
      setIsLoadingNext(false);
    }
  }, [quiz?.id]);

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
      if (fastMode) {
        setTimeout(() => {
          handleNextQuiz();
        }, 600);
      }
    },
    onError: console.error,
  });

  // Report error mutation
  const reportMutation = useMutation({
    mutationFn: ({ errorType }: { errorType: errorReportingApi.ErrorType }) => {
      if (!isOnline) {
        throw new Error("Reporting not available in offline mode");
      }
      return errorReportingApi.reportError({
        wordId: quiz!.wordId,
        definitionId: quiz!.definitionId,
        errorType,
      });
    },
    onSuccess: () => {
      Alert.alert(t("common.success"), t("quiz.reportSubmitted"));
      setShowReportModal(false);
      handleNextQuiz();
    },
    onError: (error) => {
      if (error instanceof ErrorReportRateLimitError) {
        let message: string;
        
        if (error.windowType === "cooldown") {
          // Cooldown for specific error on this word
          message = error.retryAfter 
            ? t("quiz.cooldownError", { minutes: Math.ceil(error.retryAfter / 60) })
            : error.message;
        } else {
          // Rate limit (hourly/daily)
          message = error.retryAfter 
            ? t("quiz.rateLimitError", { minutes: Math.ceil(error.retryAfter / 60) })
            : error.message;
        }
        
        Alert.alert(t("common.error"), message);
      } else {
        Alert.alert(t("common.error"), t("offline.featureUnavailable"));
      }
    },
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
    refetch();
  };

  const handleReportError = (errorType: ErrorType) => {
    reportMutation.mutate({ errorType });
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
            onReportPress={() => setShowReportModal(true)}
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
          onReportPress={() => setShowReportModal(true)}
        />

        <QuizProgressBar correctCount={correctCount} quizCount={quizCount} />

        <QuizModeToggle fastMode={fastMode} onToggle={onPressFastModeToggle} />

        <ScrollView
          contentContainerStyle={styles.scrollContent}
          showsVerticalScrollIndicator={false}
        >
          <View style={styles.quizCard}>
            {isLoadingNext || (isFetching && !quiz) ? (
              <View style={styles.quizLoadingContainer}>
                <ActivityIndicator size="large" color="#FF7B54" />
                <Text style={styles.loadingText}>{t("quiz.loadingNextQuestion")}</Text>
              </View>
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
          onClose={() => setShowReportModal(false)}
          onReportError={handleReportError}
          isLoading={reportMutation.isPending}
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
  quizLoadingContainer: {
    minHeight: 300,
    justifyContent: "center",
    alignItems: "center",
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: "#636E72",
  },
});
