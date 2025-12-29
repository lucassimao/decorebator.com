import { ErrorReportModal } from "@/components/ErrorReportModal";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { useTheme } from "@/contexts/ThemeContext";
import { MaterialIcons } from "@expo/vector-icons";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { useQuery } from "@tanstack/react-query";
import { useAudioPlayer } from "expo-audio";
import React, { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Animated,
  Dimensions,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";

import { useLocalSearchParams, useRouter } from "expo-router";

import { ErrorType } from "@/api/errorReporting";
import * as offlineWordlistsApi from "@/api/offlineWordlists";
import { useErrorReporting } from "@/hooks/useErrorReporting";
import { useOffline } from "@/hooks/useOffline";

import { FlashcardContent } from "@/components/flashcard/FlashcardContent";
import { FlashcardHeader } from "@/components/flashcard/FlashcardHeader";
import { FlashcardLoadingState } from "@/components/flashcard/FlashcardLoadingState";
import { FlashcardNavigation } from "@/components/flashcard/FlashcardNavigation";
import { FlashcardProgressBar } from "@/components/flashcard/FlashcardProgressBar";
import { SafeAreaView } from "react-native-safe-area-context";

const { width: screenWidth } = Dimensions.get("window");

// Color palette
const colors = {
  primary: "#FF7B54",
  success: "#4CAF50",
  error: "#FF6B6B",
  gold: "#FFD700",
  background: "#FDF6E3",
  backgroundLight: "#FFF9F0",
  backgroundPeach: "#FFE8D6",
  backgroundOrange: "#FFDCC3",
  backgroundSage: "#F5F0E6",
  textDark: "#2D3436",
  textMedium: "#636E72",
  textLight: "#B2BEC3",
  white: "#FFFFFF",
  lightBackground: "#FAFAFA",
  borderGray: "#E0E0E0",
  divider: "#F0F0F0",
};

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
  const player = useAudioPlayer();
  const loadingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
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

  const currentWord = words?.[currentIndex];

  // Load saved position on mount
  useEffect(() => {
    const loadSavedPosition = async () => {
      try {
        const [savedPositionValue, savedIndex] = await Promise.all([
          AsyncStorage.getItem(`flashcard_save_position_${wordlistId}`),
          AsyncStorage.getItem(`flashcard_position_${wordlistId}`),
        ]);

        if (savedPositionValue !== null) {
          setSavePosition(savedPositionValue === "true");

          if (savedPositionValue === "true" && savedIndex !== null) {
            const index = parseInt(savedIndex, 10);
            if (!isNaN(index) && index >= 0 && words && index < words.length) {
              setCurrentIndex(index);
            }
          }
        }
      } catch (error) {
        console.error("Error loading saved position:", error);
      } finally {
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

      try {
        await AsyncStorage.setItem(
          `flashcard_position_${wordlistId}`,
          currentIndex.toString(),
        );
      } catch (error) {
        console.error("Error saving position:", error);
      }
    };

    saveCurrentPosition();
  }, [currentIndex, savePosition, wordlistId, isLoadingPosition]);

  // Handle save position toggle
  const handleToggleSavePosition = async () => {
    const newValue = !savePosition;
    setSavePosition(newValue);

    try {
      await AsyncStorage.setItem(
        `flashcard_save_position_${wordlistId}`,
        newValue.toString(),
      );

      if (!newValue) {
        // Clear saved position when disabled
        await AsyncStorage.removeItem(`flashcard_position_${wordlistId}`);
      }
    } catch (error) {
      console.error("Error toggling save position:", error);
    }
  };

  // Animation values
  const slideAnimation = useRef(new Animated.Value(0)).current;
  const scaleAnimation = useRef(new Animated.Value(1)).current;

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
    scaleAnimation.setValue(1);
    slideAnimation.setValue(0);
  }, [currentIndex, scaleAnimation, slideAnimation]);

  // Handle card flip
  const flipCard = () => {
    // If flipping to back, enable definitions fetching
    if (!isFlipped && currentWord) {
      setShouldFetchDefinitions(true);
    }

    Animated.parallel([
      Animated.sequence([
        Animated.timing(scaleAnimation, {
          toValue: 0.95,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.timing(scaleAnimation, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
      ]),
    ]).start();

    setIsFlipped(!isFlipped);
  };

  // Handle navigation
  const navigateCard = (direction: "next" | "prev") => {
    const newIndex =
      direction === "next"
        ? Math.min(currentIndex + 1, (words?.length || 1) - 1)
        : Math.max(currentIndex - 1, 0);

    if (newIndex === currentIndex) return;

    // Slide animation
    Animated.sequence([
      Animated.timing(slideAnimation, {
        toValue: direction === "next" ? -screenWidth : screenWidth,
        duration: 200,
        useNativeDriver: true,
      }),
      Animated.timing(slideAnimation, {
        toValue: 0,
        duration: 0,
        useNativeDriver: true,
      }),
    ]).start();

    // Reset flip state and definitions fetch state
    if (isFlipped) {
      flipCard();
    }

    setCurrentIndex(newIndex);
    setShouldFetchDefinitions(false); // Reset definitions fetch state for new card
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

  if (isLoading || isLoadingPosition || (isRetrying && isFetching)) {
    return (
      <SafeAreaView style={styles.container}>
        <FlashcardLoadingState
          isLoading={true}
          hasTimeout={loadingTimeout}
          error={error}
          isLoadingPosition={isLoadingPosition}
          onRetry={handleRetryWords}
          onGoBack={() => router.back()}
          colors={colors}
        />
      </SafeAreaView>
    );
  }

  // Handle offline error state
  if (error && !isOnline && !isOfflineAvailable) {
    return (
      <SafeAreaView style={styles.container}>
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
          <TouchableOpacity
            style={styles.backButton}
            onPress={() => router.back()}
          >
            <Text style={styles.backButtonText}>{t("common.goBack")}</Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  if (error || !words || words.length === 0) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.errorContainer}>
          <MaterialIcons
            name="error-outline"
            size={64}
            color={theme.colors.error}
          />
          <Text style={styles.errorTitle}>{t("flashcards.noWordsFound")}</Text>
          <TouchableOpacity
            style={styles.backButton}
            onPress={() => router.back()}
          >
            <Text style={styles.backButtonText}>{t("common.goBack")}</Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <OfflineIndicator />

      <FlashcardHeader
        wordlistName={wordlistName || ""}
        currentIndex={currentIndex}
        totalWords={words.length}
        isOnline={isOnline}
        onClose={() => router.back()}
        onReportError={openReportModal}
        savePosition={savePosition}
        onToggleSavePosition={handleToggleSavePosition}
      />

      <FlashcardProgressBar
        currentIndex={currentIndex}
        totalWords={words.length}
      />

      <FlashcardContent
        currentWord={currentWord!}
        definitions={definitions}
        isFlipped={isFlipped}
        shouldFetchDefinitions={shouldFetchDefinitions}
        loadingDefinitions={loadingDefinitions}
        definitionsError={definitionsError}
        slideAnimation={slideAnimation}
        scaleAnimation={scaleAnimation}
        onFlip={flipCard}
        onRefetchDefinitions={refetchDefinitions}
      />

      <FlashcardNavigation
        currentIndex={currentIndex}
        totalWords={words.length}
        onNavigate={navigateCard}
      />

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
      textAlign: "center",
    },
    errorMessage: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: 22,
      marginBottom: 24,
    },
    backButton: {
      backgroundColor: theme.colors.primary,
      paddingHorizontal: 24,
      paddingVertical: 12,
      borderRadius: 25,
    },
    backButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
  });
