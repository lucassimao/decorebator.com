import React, { useState, useRef, useEffect } from "react";
import {
  View,
  Text,
  StyleSheet,
  Animated,
  Dimensions,
  ActivityIndicator,
  SafeAreaView,
  Alert,
  TouchableOpacity,
} from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { MaterialIcons } from "@expo/vector-icons";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { ErrorReportModal } from "@/components/ErrorReportModal";
import { useAudioPlayer, useAudioPlayerStatus } from "expo-audio";

import { useRouter, useLocalSearchParams } from "expo-router";

import * as offlineWordlistsApi from "@/api/offlineWordlists";
import { useOffline } from "@/hooks/useOffline";
import * as errorReportingApi from "@/api/errorReporting";
import { ErrorType, ErrorReportRateLimitError } from "@/api/errorReporting";

import { FlashcardHeader } from "@/components/flashcard/FlashcardHeader";
import { FlashcardProgressBar } from "@/components/flashcard/FlashcardProgressBar";
import { FlashcardContent } from "@/components/flashcard/FlashcardContent";
import { FlashcardNavigation } from "@/components/flashcard/FlashcardNavigation";

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
  const [showReportModal, setShowReportModal] = useState(false);
  const [isNavigating, setIsNavigating] = useState(false);
  const player = useAudioPlayer();
  const { didJustFinish } = useAudioPlayerStatus(player);

  // Reset player
  useEffect(() => {
    if (didJustFinish) {
      player.seekTo(0);
    }
  }, [didJustFinish, player]);

  // Animation values
  const slideAnimation = useRef(new Animated.Value(0)).current;
  const scaleAnimation = useRef(new Animated.Value(1)).current;

  // Fetch words with definitions only to avoid broken flashcards
  const {
    data: words,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["words", wordlistId, "withDefinitions"],
    queryFn: () => offlineWordlistsApi.getWords(Number(wordlistId), true),
    enabled: !!wordlistId,
    retry: isOnline ? 3 : 0, // Don't retry in offline mode
  });

  const currentWord = words?.[currentIndex];

  // Error reporting mutation
  const reportMutation = useMutation({
    mutationFn: ({ errorType }: { errorType: ErrorType }) => {
      if (!isOnline) {
        throw new Error("Reporting not available in offline mode");
      }

      return errorReportingApi.reportError({
        wordId: currentWord!.id,
        definitionId: definitions[0]?.id || 0,
        errorType,
      });
    },
    onSuccess: () => {
      Alert.alert(t("common.success"), t("flashcards.reportSubmitted"));
      setShowReportModal(false);
    },
    onError: (error) => {
      if (error instanceof ErrorReportRateLimitError) {
        let message: string;
        
        if (error.windowType === "cooldown") {
          // Cooldown for specific error on this word
          message = error.retryAfter 
            ? t("flashcards.cooldownError", { minutes: Math.ceil(error.retryAfter / 60) })
            : error.message;
        } else {
          // Rate limit (hourly/daily)
          message = error.retryAfter 
            ? t("flashcards.rateLimitError", { minutes: Math.ceil(error.retryAfter / 60) })
            : error.message;
        }
        
        Alert.alert(t("common.error"), message);
      } else {
        Alert.alert(t("common.error"), t("offline.featureUnavailable"));
      }
    },
  });

  const handleReportError = (errorType: ErrorType) => {
    reportMutation.mutate({ errorType });
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
    retry: isOnline ? 2 : 0, // Don't retry in offline mode
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });

  // Reset player and definitions fetch state when word changes
  useEffect(() => {
    const currentWord = words?.[currentIndex];

    if (currentWord?.audioURL) {
      player.replace(currentWord?.audioURL);
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

    setIsNavigating(true);

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
    ]).start(() => {
      setIsNavigating(false);
    });

    // Reset flip state and definitions fetch state
    if (isFlipped) {
      flipCard();
    }

    setCurrentIndex(newIndex);
    setShouldFetchDefinitions(false); // Reset definitions fetch state for new card
  };

  // Play audio
  const playAudio = async () => {
    if (!currentWord?.audioURL) return;

    try {
      player.play();
    } catch (error) {
      console.error("Error playing audio:", error);
    }
  };

  if (isLoading) {
    return (
      <LinearGradient
        colors={[
          colors.backgroundLight,
          colors.backgroundPeach,
          colors.backgroundSage,
        ]}
        style={styles.container}
      >
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color={colors.primary} />
        </View>
      </LinearGradient>
    );
  }

  // Handle offline error state
  if (error && !isOnline && !isOfflineAvailable) {
    return (
      <LinearGradient
        colors={[
          colors.backgroundLight,
          colors.backgroundPeach,
          colors.backgroundSage,
        ]}
        style={styles.container}
      >
        <SafeAreaView style={styles.safeArea}>
          <View style={styles.errorContainer}>
            <MaterialIcons
              name="cloud-off"
              size={64}
              color={colors.textMedium}
            />
            <Text style={styles.errorText}>{t("offline.premiumRequired")}</Text>
            <Text style={styles.errorSubText}>
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
      </LinearGradient>
    );
  }

  if (error || !words || words.length === 0) {
    return (
      <LinearGradient
        colors={[
          colors.backgroundLight,
          colors.backgroundPeach,
          colors.backgroundSage,
        ]}
        style={styles.container}
      >
        <SafeAreaView style={styles.safeArea}>
          <View style={styles.errorContainer}>
            <MaterialIcons
              name="error-outline"
              size={64}
              color={colors.error}
            />
            <Text style={styles.errorText}>{t("flashcards.noWordsFound")}</Text>
            <TouchableOpacity
              style={styles.backButton}
              onPress={() => router.back()}
            >
              <Text style={styles.backButtonText}>{t("common.goBack")}</Text>
            </TouchableOpacity>
          </View>
        </SafeAreaView>
      </LinearGradient>
    );
  }

  return (
    <LinearGradient
      colors={[
        colors.backgroundLight,
        colors.backgroundPeach,
        colors.backgroundSage,
      ]}
      style={styles.container}
    >
      <SafeAreaView style={styles.safeArea}>
        <OfflineIndicator />

        <FlashcardHeader
          wordlistName={wordlistName || ""}
          currentIndex={currentIndex}
          totalWords={words.length}
          isOnline={isOnline}
          onClose={() => router.back()}
          onReportError={() => setShowReportModal(true)}
        />

        <FlashcardProgressBar
          currentIndex={currentIndex}
          totalWords={words.length}
        />

        {isNavigating ? (
          <View style={styles.navigatingContainer}>
            <ActivityIndicator size="large" color={colors.primary} />
            <Text style={styles.navigatingText}>{t("flashcards.loadingCard")}</Text>
          </View>
        ) : (
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
            onPlayAudio={playAudio}
            onRefetchDefinitions={refetchDefinitions}
          />
        )}

        <FlashcardNavigation
          currentIndex={currentIndex}
          totalWords={words.length}
          onNavigate={navigateCard}
        />

        {/* Error Reporting Modal */}
        <ErrorReportModal
          visible={showReportModal}
          onClose={() => setShowReportModal(false)}
          onReportError={handleReportError}
          isLoading={reportMutation.isPending}
          context="flashcards"
          wordName={currentWord?.name}
          errorTypes={[ErrorType.SoundNotPlaying, ErrorType.UnrelatedMeaning]}
        />
      </SafeAreaView>
    </LinearGradient>
  );
};

export default FlashcardPractice;

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  safeArea: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  errorContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    padding: 20,
  },
  errorText: {
    fontSize: 18,
    color: colors.textMedium,
    marginTop: 16,
    marginBottom: 32,
    textAlign: "center",
  },
  errorSubText: {
    fontSize: 14,
    color: colors.textMedium,
    marginBottom: 32,
    textAlign: "center",
    lineHeight: 20,
  },
  backButton: {
    backgroundColor: colors.primary,
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 25,
  },
  backButtonText: {
    color: colors.white,
    fontSize: 16,
    fontWeight: "600",
  },
  navigatingContainer: {
    minHeight: 400,
    justifyContent: "center",
    alignItems: "center",
  },
  navigatingText: {
    marginTop: 16,
    fontSize: 16,
    color: colors.textMedium,
  },
});
