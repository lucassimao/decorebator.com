import React, { useState, useRef, useEffect } from "react";
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Animated,
  Dimensions,
  ScrollView,
  ActivityIndicator,
  SafeAreaView,
} from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { MaterialIcons, Ionicons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { useAudioPlayer, useAudioPlayerStatus } from "expo-audio";

import { useRouter, useLocalSearchParams } from "expo-router";

import * as wordlistsApi from "@/api/wordlists";

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

interface Word {
  id: number;
  name: string;
  wordlistId: number;
  learned: boolean;
  pronunciation?: string;
  notes?: string;
  audioURL?: string;
}

const FlashcardPractice: React.FC = () => {
  const { t } = useTranslation();
  const router = useRouter();
  const { wordlistId, wordlistName } = useLocalSearchParams<{
    wordlistId: string;
    wordlistName: string;
  }>();

  const [currentIndex, setCurrentIndex] = useState(0);
  const [isFlipped, setIsFlipped] = useState(false);
  const [shouldFetchDefinitions, setShouldFetchDefinitions] = useState(false);
  const player = useAudioPlayer();
  const { playing: isPlaying, didJustFinish } = useAudioPlayerStatus(player);

  // Reset player
  useEffect(() => {
    if (didJustFinish) {
      player.seekTo(0);
    }
  }, [didJustFinish, player]);

  // Animation values
  const flipAnimation = useRef(new Animated.Value(0)).current;
  const slideAnimation = useRef(new Animated.Value(0)).current;
  const scaleAnimation = useRef(new Animated.Value(1)).current;

  // Fetch words
  const {
    data: words,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["words", wordlistId],
    queryFn: () => wordlistsApi.getWords(Number(wordlistId)),
    enabled: !!wordlistId,
  });

  const currentWord = words?.[currentIndex];

  // Fetch definitions using React Query
  const {
    data: definitions = [],
    isLoading: loadingDefinitions,
    error: definitionsError,
    refetch: refetchDefinitions,
  } = useQuery({
    queryKey: ["definitions", wordlistId, currentWord?.id],
    queryFn: () =>
      wordlistsApi.getWordDefinitions(Number(wordlistId), currentWord!.id),
    enabled: !!wordlistId && !!currentWord?.id && shouldFetchDefinitions,
    staleTime: 5 * 60 * 1000, // 5 minutes - definitions don't change often
    retry: 2, // Retry failed requests up to 2 times
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
  }, [currentIndex, words]);

  // Reset flip state when word changes
  useEffect(() => {
    setIsFlipped(false);
    flipAnimation.setValue(0);
    scaleAnimation.setValue(1);
    slideAnimation.setValue(0);
  }, [currentIndex]);

  // Flip animation interpolation
  const frontInterpolate = flipAnimation.interpolate({
    inputRange: [0, 180],
    outputRange: ["0deg", "180deg"],
  });

  const backInterpolate = flipAnimation.interpolate({
    inputRange: [0, 180],
    outputRange: ["180deg", "360deg"],
  });

  const frontAnimatedStyle = {
    transform: [{ rotateY: frontInterpolate }],
  };

  const backAnimatedStyle = {
    transform: [{ rotateY: backInterpolate }],
  };

  // Handle card flip
  const flipCard = () => {
    // If flipping to back, enable definitions fetching
    if (!isFlipped && currentWord) {
      setShouldFetchDefinitions(true);
    }

    Animated.parallel([
      Animated.timing(flipAnimation, {
        toValue: isFlipped ? 0 : 180,
        duration: 600,
        useNativeDriver: true,
      }),
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
        {/* Header */}
        <View style={styles.header}>
          <TouchableOpacity
            style={styles.closeButton}
            onPress={() => router.back()}
          >
            <Ionicons name="close" size={28} color={colors.textDark} />
          </TouchableOpacity>

          <View style={styles.headerCenter}>
            <Text style={styles.headerTitle}>{wordlistName}</Text>
            <Text style={styles.headerSubtitle}>
              {t("flashcards.cardCounter", {
                current: currentIndex + 1,
                total: words.length,
              })}
            </Text>
          </View>

          <View style={styles.headerRight} />
        </View>

        {/* Progress Bar */}
        <View style={styles.progressContainer}>
          <View style={styles.progressBar}>
            <View
              style={[
                styles.progressFill,
                { width: `${((currentIndex + 1) / words.length) * 100}%` },
              ]}
            />
          </View>
        </View>

        {/* Flashcard */}
        <View style={styles.cardContainer}>
          <View style={styles.cardTouchable}>
            <TouchableOpacity
              activeOpacity={0.9}
              onPress={flipCard}
              style={[
                styles.card,
                frontAnimatedStyle,
                {
                  transform: [
                    { translateX: slideAnimation },
                    { scale: scaleAnimation },
                    { rotateY: frontInterpolate },
                  ],
                },
                !isFlipped ? styles.cardFront : null,
              ]}
            >
              <View style={styles.cardContent}>
                <Text style={styles.wordText}>{currentWord?.name}</Text>
                {currentWord?.audioURL && (
                  <TouchableOpacity
                    style={styles.audioButton}
                    onPress={playAudio}
                  >
                    <Ionicons
                      name="volume-high"
                      size={32}
                      color={colors.primary}
                    />
                  </TouchableOpacity>
                )}
              </View>
              <Text style={styles.flipHint}>{t("flashcards.tapToFlip")}</Text>
            </TouchableOpacity>

            <Animated.View
              style={[
                styles.card,
                styles.cardBack,
                backAnimatedStyle,
                {
                  transform: [
                    { translateX: slideAnimation },
                    { scale: scaleAnimation },
                    { rotateY: backInterpolate },
                  ],
                },
                isFlipped ? styles.cardBackVisible : null,
              ]}
            >
              <ScrollView
                style={styles.definitionsScroll}
                contentContainerStyle={styles.definitionsScrollContent}
                showsVerticalScrollIndicator={true}
                bounces={true}
                alwaysBounceVertical={false}
                nestedScrollEnabled={true}
              >
                {loadingDefinitions ? (
                  <View style={styles.loadingDefinitionsContainer}>
                    <ActivityIndicator size="large" color={colors.primary} />
                    <Text style={styles.loadingDefinitionsText}>
                      Loading definitions...
                    </Text>
                  </View>
                ) : definitionsError ? (
                  <View style={styles.errorDefinitionsContainer}>
                    <MaterialIcons
                      name="error-outline"
                      size={48}
                      color={colors.error}
                    />
                    <Text style={styles.errorDefinitionsText}>
                      Failed to load definitions
                    </Text>
                    <TouchableOpacity
                      style={styles.retryButton}
                      onPress={() => refetchDefinitions()}
                    >
                      <Text style={styles.retryButtonText}>Retry</Text>
                    </TouchableOpacity>
                  </View>
                ) : definitions.length > 0 ? (
                  definitions.map((definition, index) => (
                    <View key={definition.id} style={styles.definitionBlock}>
                      {index > 0 && <View style={styles.definitionDivider} />}

                      {definition.partOfSpeech && (
                        <View style={styles.partOfSpeechBadge}>
                          <Text style={styles.partOfSpeechText}>
                            {definition.partOfSpeech}
                          </Text>
                        </View>
                      )}

                      <Text style={styles.meaningText}>
                        {definition.meaning}
                      </Text>

                      {definition.phoneticNotations &&
                        definition.phoneticNotations.length > 0 && (
                          <Text style={styles.phoneticText}>
                            /{definition.phoneticNotations[0].ipa}/
                          </Text>
                        )}

                      {definition.examples &&
                        definition.examples.length > 0 && (
                          <View style={styles.examplesContainer}>
                            <Text style={styles.examplesTitle}>
                              {t("flashcards.examples")}:
                            </Text>
                            {definition.examples.map((example, idx) => (
                              <Text key={idx} style={styles.exampleText}>
                                • {example.replace(/\[|\]/g, "")}
                              </Text>
                            ))}
                          </View>
                        )}
                    </View>
                  ))
                ) : shouldFetchDefinitions ? (
                  <View style={styles.noDefinitionsContainer}>
                    <Text style={styles.noDefinitionsText}>
                      No definitions available for this word.
                    </Text>
                  </View>
                ) : (
                  <View style={styles.flipPromptContainer}>
                    <Text style={styles.flipPromptText}>
                      {t("flashcards.tapToFlip")}
                    </Text>
                  </View>
                )}
              </ScrollView>
              <TouchableOpacity
                style={styles.flipHintTouchable}
                onPress={flipCard}
                activeOpacity={0.7}
              >
                <Text style={styles.flipHint}>
                  {t("flashcards.tapToFlipBack")}
                </Text>
              </TouchableOpacity>
            </Animated.View>
          </View>
        </View>

        {/* Navigation Controls */}
        <View style={styles.navigationContainer}>
          <TouchableOpacity
            style={[
              styles.navButton,
              currentIndex === 0 && styles.navButtonDisabled,
            ]}
            onPress={() => navigateCard("prev")}
            disabled={currentIndex === 0}
          >
            <Ionicons
              name="chevron-back"
              size={32}
              color={currentIndex === 0 ? colors.borderGray : colors.textDark}
            />
          </TouchableOpacity>

          <View style={styles.centerControls}>
            <Text style={styles.swipeHint}>
              {t("flashcards.navigationHint")}
            </Text>
          </View>

          <TouchableOpacity
            style={[
              styles.navButton,
              currentIndex === words.length - 1 && styles.navButtonDisabled,
            ]}
            onPress={() => navigateCard("next")}
            disabled={currentIndex === words.length - 1}
          >
            <Ionicons
              name="chevron-forward"
              size={32}
              color={
                currentIndex === words.length - 1
                  ? colors.borderGray
                  : colors.textDark
              }
            />
          </TouchableOpacity>
        </View>
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
  header: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  closeButton: {
    width: 44,
    height: 44,
    justifyContent: "center",
    alignItems: "center",
  },
  headerCenter: {
    flex: 1,
    alignItems: "center",
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: "600",
    color: colors.textDark,
  },
  headerSubtitle: {
    fontSize: 14,
    color: colors.textMedium,
    marginTop: 2,
  },
  headerRight: {
    width: 44,
  },
  progressContainer: {
    paddingHorizontal: 20,
    paddingBottom: 20,
  },
  progressBar: {
    height: 6,
    backgroundColor: colors.borderGray,
    borderRadius: 3,
    overflow: "hidden",
  },
  progressFill: {
    height: "100%",
    backgroundColor: colors.success,
    borderRadius: 3,
  },
  cardContainer: {
    flex: 1,
    paddingHorizontal: 20,
    justifyContent: "center",
  },
  cardTouchable: {
    height: 400,
    position: "relative",
  },
  card: {
    position: "absolute",
    width: "100%",
    height: "100%",
    backgroundColor: colors.white,
    borderRadius: 20,
    padding: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    elevation: 5,
    backfaceVisibility: "hidden",
  },
  cardFront: {
    zIndex: 2,
  },
  cardBack: {
    zIndex: 1,
  },
  cardBackVisible: {
    zIndex: 3,
  },
  cardContent: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  wordText: {
    fontSize: 48,
    fontWeight: "bold",
    color: colors.textDark,
    textAlign: "center",
    marginBottom: 20,
  },
  audioButton: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: colors.backgroundPeach,
    justifyContent: "center",
    alignItems: "center",
  },
  flipHint: {
    fontSize: 14,
    color: colors.textLight,
    fontStyle: "italic",
    textAlign: "center",
  },
  flipHintTouchable: {
    position: "absolute",
    bottom: 10,
    left: 0,
    right: 0,
    paddingVertical: 10,
    paddingHorizontal: 20,
    alignItems: "center",
  },
  definitionsScroll: {
    flex: 1,
    marginBottom: 40,
  },
  definitionsScrollContent: {
    flexGrow: 1,
    paddingBottom: 20,
  },
  definitionBlock: {
    marginBottom: 20,
  },
  definitionDivider: {
    height: 1,
    backgroundColor: colors.divider,
    marginBottom: 16,
  },
  partOfSpeechBadge: {
    backgroundColor: colors.backgroundPeach,
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 12,
    alignSelf: "flex-start",
    marginBottom: 8,
  },
  partOfSpeechText: {
    fontSize: 12,
    color: colors.primary,
    fontWeight: "600",
  },
  meaningText: {
    fontSize: 20,
    color: colors.textDark,
    lineHeight: 28,
    marginBottom: 8,
  },
  phoneticText: {
    fontSize: 16,
    color: colors.textMedium,
    fontStyle: "italic",
    marginBottom: 12,
  },
  examplesContainer: {
    marginTop: 12,
  },
  examplesTitle: {
    fontSize: 14,
    fontWeight: "600",
    color: colors.textMedium,
    marginBottom: 8,
  },
  exampleText: {
    fontSize: 16,
    color: colors.textDark,
    lineHeight: 24,
    marginBottom: 6,
    paddingLeft: 8,
  },
  navigationContainer: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 20,
    paddingVertical: 20,
  },
  navButton: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: colors.white,
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 3,
  },
  navButtonDisabled: {
    opacity: 0.5,
  },
  centerControls: {
    flex: 1,
    alignItems: "center",
  },
  swipeHint: {
    fontSize: 14,
    color: colors.textMedium,
    fontStyle: "italic",
  },
  loadingDefinitionsContainer: {
    minHeight: 200,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 40,
  },
  loadingDefinitionsText: {
    fontSize: 16,
    color: colors.textMedium,
    marginTop: 16,
    textAlign: "center",
  },
  noDefinitionsContainer: {
    minHeight: 200,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 40,
  },
  noDefinitionsText: {
    fontSize: 16,
    color: colors.textMedium,
    textAlign: "center",
    fontStyle: "italic",
  },
  errorDefinitionsContainer: {
    minHeight: 200,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 40,
  },
  errorDefinitionsText: {
    fontSize: 16,
    color: colors.error,
    textAlign: "center",
    marginTop: 16,
    marginBottom: 20,
  },
  retryButton: {
    backgroundColor: colors.primary,
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 20,
  },
  retryButtonText: {
    color: colors.white,
    fontSize: 14,
    fontWeight: "600",
  },
  flipPromptContainer: {
    minHeight: 200,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 40,
  },
  flipPromptText: {
    fontSize: 18,
    color: colors.textLight,
    textAlign: "center",
    fontStyle: "italic",
  },
});
