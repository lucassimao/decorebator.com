import { ErrorReportModal } from "@/components/ErrorReportModal";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { useTheme } from "@/contexts/ThemeContext";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { useQuery } from "@tanstack/react-query";
import { useAudioPlayer } from "expo-audio";
import React, { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { ScrollView, StyleSheet } from "react-native";

import { useLocalSearchParams, useRouter } from "expo-router";

import { ErrorType } from "@/api/errorReporting";
import * as offlineWordlistsApi from "@/api/offlineWordlists";
import { useErrorReporting } from "@/hooks/useErrorReporting";
import { useOffline } from "@/hooks/useOffline";
import {
  flashcardPositionKey,
  flashcardSavePositionKey,
  resetSavedFlashcardPosition,
  resolveFlashcardStartIndex,
  visitFlashcard,
} from "@/utils/flashcardSession";

import { FlashcardContent } from "@/components/flashcard/FlashcardContent";
import { FlashcardCompletion } from "@/components/flashcard/FlashcardCompletion";
import { FlashcardHeader } from "@/components/flashcard/FlashcardHeader";
import { FlashcardLoadingState } from "@/components/flashcard/FlashcardLoadingState";
import { FlashcardNavigation } from "@/components/flashcard/FlashcardNavigation";
import { FlashcardProgressBar } from "@/components/flashcard/FlashcardProgressBar";
import { FlashcardStatusState } from "@/components/flashcard/FlashcardStatusState";
import { SafeAreaView } from "react-native-safe-area-context";
import {
  boundFlashcardIndex,
  resolveFlashcardLoadState,
} from "@/utils/flashcardPresentation";

const FlashcardPractice: React.FC = () => {
  const { t } = useTranslation();
  const router = useRouter();
  const { isOnline, isOfflineAvailable } = useOffline();
  const { wordlistId, wordlistName } = useLocalSearchParams<{
    wordlistId: string;
    wordlistName: string;
  }>();

  const [currentIndex, setCurrentIndex] = useState(0);
  const [isFlipped, setIsFlipped] = useState(false);
  const [shouldFetchDefinitions, setShouldFetchDefinitions] = useState(false);
  const [savePosition, setSavePosition] = useState(false);
  const [isLoadingPosition, setIsLoadingPosition] = useState(true);
  const [loadingTimeout, setLoadingTimeout] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [isRetrying, setIsRetrying] = useState(false);
  const [showCompletion, setShowCompletion] = useState(false);
  const [animateCompletion, setAnimateCompletion] = useState(false);
  const [navigationDirection, setNavigationDirection] = useState<
    "next" | "prev"
  >("next");
  const [flipFocusRequest, setFlipFocusRequest] = useState(0);
  const player = useAudioPlayer();
  const loadingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const visitedIndicesRef = useRef(new Set([0]));
  const hasCelebratedPassRef = useRef(false);
  const skipNextPositionSaveRef = useRef(false);
  const { theme } = useTheme();
  // const commonStyles = createCommonStyles(theme); // Remove if not defined
  const styles = createStyles(theme);

  // Fetch words with definitions only to avoid broken flashcards
  const {
    data: words,
    isLoading,
    error,
    // refetch: refetchWords, // Removed unused variable
    isFetching,
  } = useQuery({
    queryKey: ["words", wordlistId, "withDefinitions", retryCount],
    queryFn: () => offlineWordlistsApi.getWords(Number(wordlistId), true),
    enabled: !!wordlistId,
    retry: (failureCount, error) => {
      // Don't retry timeout errors automatically - let user decide
      if (error.message.includes("timeout")) {
        return false;
      }
      return isOnline ? failureCount < 2 : false;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });

  const boundedCurrentIndex = boundFlashcardIndex(
    currentIndex,
    words?.length ?? 0,
  );
  const currentWord = words?.[boundedCurrentIndex];

  // Load saved position on mount
  useEffect(() => {
    const loadSavedPosition = async () => {
      let restoredIndex = 0;
      try {
        const [savedPositionValue, savedIndex] = await Promise.all([
          AsyncStorage.getItem(flashcardSavePositionKey(wordlistId)),
          AsyncStorage.getItem(flashcardPositionKey(wordlistId)),
        ]);

        if (savedPositionValue !== null) {
          setSavePosition(savedPositionValue === "true");

          restoredIndex = resolveFlashcardStartIndex(
            savedPositionValue,
            savedIndex,
            words?.length ?? 0,
          );
        }
      } catch (error) {
        console.error("Error loading saved position:", error);
      } finally {
        visitedIndicesRef.current = new Set([restoredIndex]);
        setCurrentIndex(restoredIndex);
        setIsLoadingPosition(false);
      }
    };

    if (wordlistId && words) {
      loadSavedPosition();
    } else {
      setIsLoadingPosition(false);
    }
  }, [wordlistId, words]);

  // Save position when it changes
  useEffect(() => {
    const saveCurrentPosition = async () => {
      if (!savePosition || isLoadingPosition) return;
      if (skipNextPositionSaveRef.current) {
        skipNextPositionSaveRef.current = false;
        return;
      }

      try {
        await AsyncStorage.setItem(
          flashcardPositionKey(wordlistId),
          currentIndex.toString(),
        );
      } catch (error) {
        console.error("Error saving position:", error);
      }
    };

    saveCurrentPosition();
  }, [currentIndex, savePosition, wordlistId, isLoadingPosition]);

  useEffect(() => {
    if (!isLoadingPosition) {
      visitedIndicesRef.current = visitFlashcard(
        visitedIndicesRef.current,
        currentIndex,
      );
    }
  }, [currentIndex, isLoadingPosition]);

  // Handle save position toggle
  const handleToggleSavePosition = async () => {
    const newValue = !savePosition;
    setSavePosition(newValue);

    try {
      await AsyncStorage.setItem(
        flashcardSavePositionKey(wordlistId),
        newValue.toString(),
      );

      if (!newValue) {
        // Clear saved position when disabled
        await AsyncStorage.removeItem(flashcardPositionKey(wordlistId));
      }
    } catch (error) {
      console.error("Error toggling save position:", error);
    }
  };

  // Fetch definitions using React Query with offline support
  const {
    data: definitions = [],
    isLoading: loadingDefinitions,
    error: definitionsError,
    refetch: refetchDefinitions,
  } = useQuery({
    queryKey: ["definitions", wordlistId, currentWord?.id],
    queryFn: () =>
      offlineWordlistsApi.getWordDefinitions(
        Number(wordlistId),
        currentWord!.id,
      ),
    enabled: !!wordlistId && !!currentWord?.id && shouldFetchDefinitions,
    staleTime: 5 * 60 * 1000, // 5 minutes - definitions don't change often
    retry: (failureCount, error) => {
      // Don't retry timeout errors automatically - let user decide
      if (error.message.includes("timeout")) {
        return false;
      }
      return isOnline ? failureCount < 1 : false;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });

  // Error reporting hook
  const {
    showReportModal,
    isReporting,
    handleReportError,
    openReportModal,
    closeReportModal,
  } = useErrorReporting({
    wordId: currentWord?.id || 0,
    definitionId: definitions[0]?.id || null,
    isOnline,
    context: "flashcards",
    quizDetails: currentWord
      ? {
          quizType: "FLASHCARD_VIEW", // Custom type for flashcard context
          value: currentWord.name,
          options: [],
          answerIndex: 0, // Always 0 for flashcards since there's no option
          context: "flashcard",
        }
      : undefined,
  });

  // Reset player and definitions fetch state when word changes
  useEffect(() => {
    const currentWord = words?.[currentIndex];

    if (currentWord?.audioURL) {
      player.replace(currentWord?.audioURL);
      player.pause();
      player.seekTo(0);
    }

    // Reset definitions fetch state when word changes
    setShouldFetchDefinitions(false);
  }, [currentIndex, words, player]);

  // Reset flip state when word changes
  useEffect(() => {
    setIsFlipped(false);
  }, [currentIndex]);

  // Handle card flip
  const flipCard = () => {
    // If flipping to back, enable definitions fetching
    if (!isFlipped && currentWord) {
      setShouldFetchDefinitions(true);
    }

    setIsFlipped(!isFlipped);
    setFlipFocusRequest((request) => request + 1);
  };

  // Handle navigation
  const navigateCard = (direction: "next" | "prev") => {
    const newIndex =
      direction === "next"
        ? Math.min(boundedCurrentIndex + 1, (words?.length || 1) - 1)
        : Math.max(boundedCurrentIndex - 1, 0);

    if (newIndex === boundedCurrentIndex) return;

    // Reset flip state and definitions fetch state
    if (isFlipped) {
      setIsFlipped(false);
    }

    setNavigationDirection(direction);
    setCurrentIndex(newIndex);
    setShouldFetchDefinitions(false); // Reset definitions fetch state for new card
  };

  const handleFinish = () => {
    setAnimateCompletion(!hasCelebratedPassRef.current);
    hasCelebratedPassRef.current = true;
    setShowCompletion(true);
  };

  const handleReturnToLastCard = () => {
    setAnimateCompletion(false);
    setShowCompletion(false);
  };

  const handleReviewAgain = async () => {
    skipNextPositionSaveRef.current = savePosition && currentIndex !== 0;
    visitedIndicesRef.current = new Set([0]);
    hasCelebratedPassRef.current = false;
    setAnimateCompletion(false);
    setShowCompletion(false);
    setCurrentIndex(0);
    setIsFlipped(false);
    setShouldFetchDefinitions(false);

    try {
      await resetSavedFlashcardPosition(AsyncStorage, wordlistId, savePosition);
    } catch (error) {
      console.error("Error resetting saved flashcard position:", error);
    }
  };

  // Handle loading timeout for words
  useEffect(() => {
    if (isLoading || (isRetrying && isFetching)) {
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
      setIsRetrying(false); // Clear retry state when data loads
    }

    return () => {
      if (loadingTimeoutRef.current) {
        clearTimeout(loadingTimeoutRef.current);
      }
    };
  }, [isLoading, isRetrying, isFetching]);

  const handleRetryWords = () => {
    setRetryCount((prev) => prev + 1);
    setLoadingTimeout(false);
    setIsRetrying(true);
  };

  const loadState = resolveFlashcardLoadState({
    isLoading,
    isLoadingPosition,
    isRetrying,
    isFetching,
    error,
    words,
    isOnline,
    isOfflineAvailable,
  });

  if (loadState === "loading") {
    return (
      <SafeAreaView style={styles.container}>
        <FlashcardLoadingState
          isLoading={true}
          hasTimeout={loadingTimeout}
          error={error}
          isLoadingPosition={isLoadingPosition}
          onRetry={handleRetryWords}
          onGoBack={() => router.back()}
        />
      </SafeAreaView>
    );
  }

  if (loadState === "processing") {
    return (
      <SafeAreaView style={styles.container}>
        <FlashcardLoadingState
          isLoading={false}
          hasTimeout={false}
          error={error}
          onRetry={handleRetryWords}
          onGoBack={() => router.back()}
        />
      </SafeAreaView>
    );
  }

  if (loadState === "offline-unavailable") {
    return (
      <SafeAreaView style={styles.container}>
        <FlashcardStatusState
          icon="cloud-off"
          title={t("flashcards.offlineUnavailableTitle")}
          message={t("flashcards.offlineUnavailableMessage")}
          onBack={() => router.back()}
          assertive
        />
      </SafeAreaView>
    );
  }

  if (loadState === "error") {
    return (
      <SafeAreaView style={styles.container}>
        <FlashcardStatusState
          icon="error-outline"
          title={t("flashcards.loadErrorTitle")}
          message={t("flashcards.loadErrorMessage")}
          onRetry={handleRetryWords}
          onBack={() => router.back()}
          assertive
        />
      </SafeAreaView>
    );
  }

  if (loadState === "empty") {
    return (
      <SafeAreaView style={styles.container}>
        <FlashcardStatusState
          icon="style"
          title={t("flashcards.emptyTitle")}
          message={t("flashcards.emptyMessage")}
          onBack={() => router.back()}
        />
      </SafeAreaView>
    );
  }

  if (showCompletion) {
    return (
      <SafeAreaView style={styles.container}>
        <OfflineIndicator />
        <FlashcardCompletion
          wordlistName={wordlistName || ""}
          viewedCount={visitedIndicesRef.current.size}
          shouldAnimate={animateCompletion}
          onReviewAgain={handleReviewAgain}
          onBackToWordlists={() => router.back()}
          onReturnToLastCard={handleReturnToLastCard}
        />
      </SafeAreaView>
    );
  }

  if (!currentWord) {
    return (
      <SafeAreaView style={styles.container}>
        <FlashcardStatusState
          icon="style"
          title={t("flashcards.emptyTitle")}
          message={t("flashcards.emptyMessage")}
          onBack={() => router.back()}
        />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <OfflineIndicator />
      <ScrollView
        nestedScrollEnabled
        contentInsetAdjustmentBehavior="automatic"
        contentContainerStyle={styles.practiceContent}
      >
        <FlashcardHeader
          wordlistName={wordlistName || ""}
          currentIndex={boundedCurrentIndex}
          totalWords={words!.length}
          isOnline={isOnline}
          onClose={() => router.back()}
          onReportError={openReportModal}
          savePosition={savePosition}
          onToggleSavePosition={handleToggleSavePosition}
        />

        <FlashcardProgressBar
          currentIndex={boundedCurrentIndex}
          totalWords={words!.length}
        />

        <FlashcardContent
          currentWord={currentWord}
          definitions={definitions}
          isFlipped={isFlipped}
          focusRequestKey={flipFocusRequest}
          navigationDirection={navigationDirection}
          shouldFetchDefinitions={shouldFetchDefinitions}
          loadingDefinitions={loadingDefinitions}
          definitionsError={definitionsError}
          onFlip={flipCard}
          onRefetchDefinitions={refetchDefinitions}
        />

        <FlashcardNavigation
          currentIndex={boundedCurrentIndex}
          totalWords={words!.length}
          onNavigate={navigateCard}
          onFinish={handleFinish}
        />
      </ScrollView>

      {/* Error Reporting Modal */}
      <ErrorReportModal
        visible={showReportModal}
        onClose={closeReportModal}
        onReportError={handleReportError}
        isLoading={isReporting}
        context="flashcards"
        wordName={currentWord?.name}
        errorTypes={[ErrorType.SoundNotPlaying, ErrorType.UnrelatedMeaning]}
      />
    </SafeAreaView>
  );
};

export default FlashcardPractice;

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    practiceContent: {
      flexGrow: 1,
    },
  });
