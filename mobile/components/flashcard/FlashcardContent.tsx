import React, { useEffect, useRef } from "react";
import {
  AccessibilityInfo,
  findNodeHandle,
  View,
  Text,
  TouchableOpacity,
  ScrollView,
  ActivityIndicator,
  StyleSheet,
  Pressable,
} from "react-native";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import {
  setAudioModeAsync,
  useAudioPlayer,
  useAudioPlayerStatus,
} from "expo-audio";
import type { Definition, Word } from "../../api/wordlists";
import { useTheme } from "@/contexts/ThemeContext";
import Animated, {
  cancelAnimation,
  Easing,
  interpolate,
  useAnimatedStyle,
  useReducedMotion,
  useSharedValue,
  withSpring,
  withTiming,
} from "react-native-reanimated";
import type { SharedValue } from "react-native-reanimated";
import { flashcardSettleOffset } from "@/utils/flashcardPresentation";

interface FlashcardContentProps {
  currentWord: Word;
  definitions: Definition[];
  isFlipped: boolean;
  focusRequestKey: number;
  shouldFetchDefinitions: boolean;
  loadingDefinitions: boolean;
  definitionsError: any;
  navigationDirection: "next" | "prev";
  onFlip: () => void;
  onRefetchDefinitions: () => void;
}

export const FlashcardContent: React.FC<FlashcardContentProps> = ({
  currentWord,
  definitions,
  isFlipped,
  focusRequestKey,
  shouldFetchDefinitions,
  loadingDefinitions,
  definitionsError,
  navigationDirection,
  onFlip,
  onRefetchDefinitions,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const reduceMotion = useReducedMotion();
  const player = useAudioPlayer();
  const { playing: isPlaying, didJustFinish } = useAudioPlayerStatus(player);
  const flipProgress = useSharedValue(isFlipped ? 1 : 0);
  const settleX = useSharedValue(0);
  const flipPressScale = useSharedValue(1);
  const audioPressScale = useSharedValue(1);
  const backPressScale = useSharedValue(1);
  const frontActionRef = useRef<React.ElementRef<typeof Pressable>>(null);
  const backActionRef = useRef<React.ElementRef<typeof TouchableOpacity>>(null);
  const lastFocusRequestRef = useRef(focusRequestKey);

  // Reset player when audio finishes
  useEffect(() => {
    if (didJustFinish) {
      player.pause();
      player.seekTo(0);
    }
  }, [didJustFinish, player]);

  useEffect(() => {
    cancelAnimation(flipProgress);
    cancelAnimation(settleX);
    flipProgress.value = 0;
    if (reduceMotion) {
      settleX.value = 0;
      return;
    }
    settleX.value = flashcardSettleOffset(navigationDirection);
    settleX.value = withTiming(0, {
      duration: 160,
      easing: Easing.out(Easing.cubic),
    });
  }, [
    currentWord?.id,
    flipProgress,
    navigationDirection,
    reduceMotion,
    settleX,
  ]);

  // Audio setup - replace audio when word changes
  useEffect(() => {
    if (currentWord?.audioURL) {
      player.replace(currentWord.audioURL);
      player.pause();
      player.seekTo(0);
    }
  }, [currentWord?.audioURL, player]);

  useEffect(() => {
    cancelAnimation(flipProgress);
    if (reduceMotion) {
      flipProgress.value = isFlipped ? 1 : 0;
      return;
    }
    flipProgress.value = withTiming(isFlipped ? 1 : 0, {
      duration: 420,
      easing: Easing.out(Easing.cubic),
    });
  }, [flipProgress, isFlipped, reduceMotion]);

  useEffect(
    () => () => {
      cancelAnimation(flipProgress);
      cancelAnimation(settleX);
      cancelAnimation(flipPressScale);
      cancelAnimation(audioPressScale);
      cancelAnimation(backPressScale);
    },
    [audioPressScale, backPressScale, flipPressScale, flipProgress, settleX],
  );

  useEffect(() => {
    if (lastFocusRequestRef.current === focusRequestKey) return;
    lastFocusRequestRef.current = focusRequestKey;

    let cancelled = false;
    let frame: number | undefined;
    void AccessibilityInfo.isScreenReaderEnabled()
      .then((enabled) => {
        if (!enabled || cancelled) return;
        frame = requestAnimationFrame(() => {
          const target = isFlipped
            ? backActionRef.current
            : frontActionRef.current;
          const node = target ? findNodeHandle(target) : null;
          if (node) AccessibilityInfo.setAccessibilityFocus(node);
        });
      })
      .catch(() => {
        // Focus management is an enhancement; the card remains operable.
      });

    return () => {
      cancelled = true;
      if (frame !== undefined) cancelAnimationFrame(frame);
    };
  }, [focusRequestKey, isFlipped]);

  const animatePress = (value: SharedValue<number>, pressed: boolean) => {
    cancelAnimation(value);
    if (reduceMotion) {
      value.value = 1;
      return;
    }
    value.value = withSpring(pressed ? 0.97 : 1, {
      damping: 18,
      stiffness: 320,
    });
  };

  // Play audio function
  const playAudio = async () => {
    if (!currentWord?.audioURL) return;

    await setAudioModeAsync({ allowsRecording: false });

    try {
      if (isPlaying) {
        player.pause();
        player.seekTo(0);
      } else {
        player.play();
      }
    } catch (error) {
      console.error("Error playing audio:", error);
    }
  };

  const settleStyle = useAnimatedStyle(() => ({
    opacity: interpolate(Math.abs(settleX.value), [0, 12], [1, 0.72]),
    transform: [{ translateX: settleX.value }],
  }));
  const frontAnimatedStyle = useAnimatedStyle(() => ({
    transform: [
      { perspective: 900 },
      { rotateY: `${interpolate(flipProgress.value, [0, 1], [0, 180])}deg` },
    ],
  }));
  const backAnimatedStyle = useAnimatedStyle(() => ({
    transform: [
      { perspective: 900 },
      {
        rotateY: `${interpolate(flipProgress.value, [0, 1], [180, 360])}deg`,
      },
    ],
  }));
  const flipPressStyle = useAnimatedStyle(() => ({
    transform: [{ scale: flipPressScale.value }],
  }));
  const audioPressStyle = useAnimatedStyle(() => ({
    transform: [{ scale: audioPressScale.value }],
  }));
  const backPressStyle = useAnimatedStyle(() => ({
    transform: [{ scale: backPressScale.value }],
  }));

  const renderDefinitionsContent = () => {
    if (loadingDefinitions) {
      return (
        <View style={styles.loadingDefinitionsContainer}>
          <ActivityIndicator size="large" color={theme.colors.primary} />
          <Text style={styles.loadingDefinitionsText}>
            {t("flashcards.loadingDefinitions")}
          </Text>
        </View>
      );
    }

    if (definitionsError) {
      return (
        <View style={styles.errorDefinitionsContainer}>
          <MaterialIcons
            name="error-outline"
            size={48}
            color={theme.colors.error}
          />
          <Text style={styles.errorDefinitionsText}>
            {t("flashcards.definitionsLoadError")}
          </Text>
          <TouchableOpacity
            style={styles.retryButton}
            onPress={onRefetchDefinitions}
            accessibilityRole="button"
            accessibilityLabel={t("common.retry")}
          >
            <Text style={styles.retryButtonText}>{t("common.retry")}</Text>
          </TouchableOpacity>
        </View>
      );
    }

    if (definitions.length > 0) {
      return definitions.map((definition, index) => (
        <View key={definition.id} style={styles.definitionBlock}>
          {index > 0 && <View style={styles.definitionDivider} />}

          {definition.partOfSpeech && (
            <View style={styles.partOfSpeechBadge}>
              <Text style={styles.partOfSpeechText}>
                {definition.partOfSpeech}
              </Text>
            </View>
          )}

          <Text style={styles.meaningText}>{definition.meaning}</Text>

          {(() => {
            const isVerb = definition.isVerbType || false;
            const hasInflections =
              definition.inflections && definition.inflections.length > 0;
            const hasExamples =
              definition.examples && definition.examples.length > 0;

            if (isVerb && hasInflections) {
              const allInflectionExamples = definition.inflections!.flatMap(
                (inf) =>
                  inf.examples.map((ex) => ({
                    example: ex,
                    tense: inf.tense,
                  })),
              );

              return (
                allInflectionExamples.length > 0 && (
                  <View style={styles.examplesContainer}>
                    <Text style={styles.examplesTitle}>
                      {t("flashcards.examples")}:
                    </Text>
                    {allInflectionExamples.map((item, idx) => (
                      <View key={idx} style={styles.inflectionExampleContainer}>
                        <Text style={styles.exampleText}>
                          • {item.example.replace(/\[|\]/g, "")}
                        </Text>
                        <Text style={styles.tenseLabel}>({item.tense})</Text>
                      </View>
                    ))}
                  </View>
                )
              );
            } else if (hasExamples) {
              return (
                <View style={styles.examplesContainer}>
                  <Text style={styles.examplesTitle}>
                    {t("flashcards.examples")}:
                  </Text>
                  {definition.examples!.map((example, idx) => (
                    <Text key={idx} style={styles.exampleText}>
                      • {example.replace(/\[|\]/g, "")}
                    </Text>
                  ))}
                </View>
              );
            }

            return null;
          })()}
        </View>
      ));
    }

    if (shouldFetchDefinitions) {
      return (
        <View style={styles.noDefinitionsContainer}>
          <Text style={styles.noDefinitionsText}>
            {t("flashcards.noDefinitions")}
          </Text>
        </View>
      );
    }

    return (
      <View style={styles.flipPromptContainer}>
        <Text style={styles.flipPromptText}>{t("flashcards.tapToFlip")}</Text>
      </View>
    );
  };

  return (
    <View style={styles.cardContainer}>
      <Animated.View
        testID="flashcard-card-region"
        style={[styles.cardTouchable, settleStyle]}
      >
        {/* Front of card */}
        <Animated.View
          testID="flashcard-front"
          accessibilityElementsHidden={isFlipped}
          importantForAccessibility={isFlipped ? "no-hide-descendants" : "auto"}
          pointerEvents={isFlipped ? "none" : "auto"}
          style={[
            styles.card,
            frontAnimatedStyle,
            !isFlipped ? styles.cardFront : null,
          ]}
        >
          <Animated.View style={[styles.cardContent, flipPressStyle]}>
            <Pressable
              ref={frontActionRef}
              style={styles.flipAction}
              onPress={onFlip}
              onPressIn={() => animatePress(flipPressScale, true)}
              onPressOut={() => animatePress(flipPressScale, false)}
              accessibilityRole="button"
              accessibilityLabel={t("flashcards.showDefinitionsLabel", {
                word: currentWord.name,
              })}
            >
              <Text
                style={styles.wordText}
                allowFontScaling
                maxFontSizeMultiplier={2}
              >
                {currentWord.name}
              </Text>
              {currentWord.pronunciation && (
                <Text
                  style={styles.phoneticTextFront}
                  maxFontSizeMultiplier={2}
                >
                  /{currentWord.pronunciation}/
                </Text>
              )}
              <Text style={styles.flipHint} maxFontSizeMultiplier={2}>
                {t("flashcards.tapToFlip")}
              </Text>
            </Pressable>
          </Animated.View>
          {currentWord?.audioURL && (
            <Animated.View style={audioPressStyle}>
              <TouchableOpacity
                style={styles.audioButton}
                onPress={playAudio}
                onPressIn={() => animatePress(audioPressScale, true)}
                onPressOut={() => animatePress(audioPressScale, false)}
                accessibilityRole="button"
                accessibilityLabel={t("flashcards.playPronunciationLabel", {
                  word: currentWord.name,
                })}
              >
                <Ionicons
                  name={isPlaying ? "pause-circle" : "play-circle"}
                  size={48}
                  color={theme.colors.primary}
                  accessibilityElementsHidden
                  importantForAccessibility="no-hide-descendants"
                />
              </TouchableOpacity>
            </Animated.View>
          )}
        </Animated.View>

        {/* Back of card */}
        <Animated.View
          testID="flashcard-back"
          accessibilityElementsHidden={!isFlipped}
          importantForAccessibility={
            !isFlipped ? "no-hide-descendants" : "auto"
          }
          pointerEvents={!isFlipped ? "none" : "auto"}
          style={[
            styles.card,
            styles.cardBack,
            backAnimatedStyle,
            isFlipped ? styles.cardBackVisible : null,
          ]}
        >
          <ScrollView
            style={styles.definitionsScroll}
            contentContainerStyle={styles.definitionsScrollContent}
            showsVerticalScrollIndicator
            bounces
            alwaysBounceVertical={false}
            nestedScrollEnabled
            accessibilityLabel={t("flashcards.definitionsLabel", {
              word: currentWord.name,
            })}
          >
            {renderDefinitionsContent()}
          </ScrollView>
          <Animated.View style={[styles.flipHintTouchable, backPressStyle]}>
            <TouchableOpacity
              ref={backActionRef}
              style={styles.backActionButton}
              onPress={onFlip}
              onPressIn={() => animatePress(backPressScale, true)}
              onPressOut={() => animatePress(backPressScale, false)}
              accessibilityRole="button"
              accessibilityLabel={t("flashcards.showWordLabel", {
                word: currentWord.name,
              })}
            >
              <Text style={styles.flipHint}>
                {t("flashcards.tapToFlipBack")}
              </Text>
            </TouchableOpacity>
          </Animated.View>
        </Animated.View>
      </Animated.View>
    </View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    cardContainer: {
      flex: 1,
      paddingHorizontal: 20,
      justifyContent: "center",
      minHeight: 220,
    },
    cardTouchable: {
      flex: 1,
      minHeight: 220,
      position: "relative",
    },
    card: {
      position: "absolute",
      width: "100%",
      height: "100%",
      backgroundColor: theme.colors.background.surface,
      borderRadius: 20,
      padding: 24,
      shadowColor: theme.colors.text.primary,
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
    },
    flipAction: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      width: "100%",
    },
    wordText: {
      fontSize: 48,
      fontWeight: "bold",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: 20,
    },
    audioButton: {
      width: 52,
      height: 52,
      alignSelf: "center",
      alignItems: "center",
      justifyContent: "center",
      marginBottom: 12,
    },
    flipHint: {
      fontSize: 14,
      color: theme.colors.ui.disabled,
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
    backActionButton: {
      width: "100%",
      minHeight: 44,
      alignItems: "center",
      justifyContent: "center",
    },
    definitionsScroll: {
      flex: 1,
    },
    definitionsScrollContent: {
      flexGrow: 1,
      paddingBottom: 72,
    },
    definitionBlock: {
      marginBottom: 20,
    },
    definitionDivider: {
      height: 1,
      backgroundColor: theme.colors.ui.divider,
      marginBottom: 16,
    },
    partOfSpeechBadge: {
      backgroundColor:
        theme.mode === "light" ? "#FFF5F0" : theme.colors.background.secondary,
      paddingHorizontal: 12,
      paddingVertical: 4,
      borderRadius: 12,
      alignSelf: "flex-start",
      marginBottom: 8,
    },
    partOfSpeechText: {
      fontSize: 12,
      color: theme.colors.primary,
      fontWeight: "600",
    },
    meaningText: {
      fontSize: 20,
      color: theme.colors.text.primary,
      lineHeight: 28,
      marginBottom: 8,
    },
    phoneticText: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      fontStyle: "italic",
      marginBottom: 12,
    },
    phoneticTextFront: {
      fontSize: 18,
      color: theme.colors.text.secondary,
      fontStyle: "italic",
      marginTop: 8,
      marginBottom: 16,
      textAlign: "center",
    },
    examplesContainer: {
      marginTop: 12,
    },
    examplesTitle: {
      fontSize: 14,
      fontWeight: "600",
      color: theme.colors.text.secondary,
      marginBottom: 8,
    },
    exampleText: {
      fontSize: 16,
      color: theme.colors.text.primary,
      lineHeight: 24,
      marginBottom: 6,
      paddingLeft: 8,
    },
    inflectionExampleContainer: {
      flexDirection: "row",
      alignItems: "flex-start",
      marginBottom: 6,
      paddingLeft: 8,
    },
    tenseLabel: {
      fontSize: 12,
      color: theme.colors.text.secondary,
      fontStyle: "italic",
      marginLeft: 8,
      marginTop: 2,
    },
    loadingDefinitionsContainer: {
      minHeight: 200,
      justifyContent: "center",
      alignItems: "center",
      paddingVertical: 40,
    },
    loadingDefinitionsText: {
      fontSize: 16,
      color: theme.colors.text.secondary,
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
      color: theme.colors.text.secondary,
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
      color: theme.colors.error,
      textAlign: "center",
      marginTop: 16,
      marginBottom: 20,
    },
    retryButton: {
      backgroundColor: theme.colors.primary,
      paddingHorizontal: 20,
      paddingVertical: 10,
      borderRadius: 20,
    },
    retryButtonText: {
      color: theme.colors.text.inverse,
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
      color: theme.colors.ui.disabled,
      textAlign: "center",
      fontStyle: "italic",
    },
    scrollContent: {
      flex: 1,
    },
  });
