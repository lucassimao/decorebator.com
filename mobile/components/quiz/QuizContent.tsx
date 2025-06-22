import { Ionicons } from "@expo/vector-icons";
import { useAudioPlayer, useAudioPlayerStatus } from "expo-audio";
import React, { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  Image,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from "react-native";
import { Quiz } from "../../api/wordlists";
import { useTheme } from "@/contexts/ThemeContext";

const { width: SCREEN_WIDTH } = Dimensions.get("window");

interface QuizContentProps {
  quiz: Quiz;
  userInput: string;
  setUserInput: (input: string) => void;
  isSubmitted: boolean;
  onSubmitAnswer: () => void;
  onSkipQuestion: () => void;
}

export const QuizContent: React.FC<QuizContentProps> = ({
  quiz,
  userInput,
  setUserInput,
  isSubmitted,
  onSubmitAnswer,
  onSkipQuestion,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const player = useAudioPlayer();
  const { playing: isPlaying, didJustFinish } = useAudioPlayerStatus(player);
  const [imageLoading, setImageLoading] = useState(false);
  const [currentImageUrl, setCurrentImageUrl] = useState<string | null>(null);
  const [imageError, setImageError] = useState(false);
  const [imageRetryCount, setImageRetryCount] = useState(0);
  const [showLoading, setShowLoading] = useState(false);
  const imageLoadTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(
    null,
  );
  const loadingDelayRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Reset player when audio finishes
  useEffect(() => {
    if (didJustFinish) {
      player.pause();
      player.seekTo(0);
    }
  }, [didJustFinish, player]);

  // Audio setup
  useEffect(() => {
    if (quiz?.audioURL) {
      player.replace(quiz.audioURL);
      player.pause();
      player.seekTo(0);
    }
  }, [quiz?.audioURL, player]);

  // Handle image changes
  useEffect(() => {
    if (quiz?.type === "WORD_FROM_IMAGE" && quiz.value) {
      // Only reset loading state if the image URL actually changed
      if (quiz.value !== currentImageUrl) {
        setImageLoading(true);
        setImageError(false);
        setImageRetryCount(0);
        setShowLoading(false);
        setCurrentImageUrl(quiz.value);

        // Clear any existing loading delay
        if (loadingDelayRef.current) {
          clearTimeout(loadingDelayRef.current);
        }
      }
    }
  }, [quiz?.value, quiz?.type]); // Remove currentImageUrl from dependencies to avoid loops

  // Cleanup timeouts on unmount
  useEffect(() => {
    return () => {
      if (imageLoadTimeoutRef.current) {
        clearTimeout(imageLoadTimeoutRef.current);
      }
      if (loadingDelayRef.current) {
        clearTimeout(loadingDelayRef.current);
      }
    };
  }, []);

  const playAudio = async () => {
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

  const retryImageLoad = () => {
    setImageError(false);
    setImageLoading(true);
    setShowLoading(false);
    setImageRetryCount((prev) => prev + 1);

    // Clear any existing timeouts
    if (imageLoadTimeoutRef.current) {
      clearTimeout(imageLoadTimeoutRef.current);
    }
    if (loadingDelayRef.current) {
      clearTimeout(loadingDelayRef.current);
    }

    // Force reload by adding timestamp query parameter
    const timestamp = Date.now();
    const separator = quiz.value.includes("?") ? "&" : "?";
    setCurrentImageUrl(`${quiz.value}${separator}retry=${timestamp}`);
  };

  const hideSquareBracketContent = (text: string): string => {
    return text.replace(/\[([^\]]+)\]/g, "_____");
  };

  const getQuizTitle = () => {
    switch (quiz?.type) {
      case "WRITE_WORD_FROM_DEFINITION":
        return t("quiz.writeWordFromDefinition");
      case "GUESS_MEANING":
        return t("quiz.whatDoesThisWordMean");
      case "COMPLETE_SENTENCE":
        return t("quiz.completeSentence");
      case "WORD_FROM_MEANING":
        return t("quiz.whichWordMatchesMeaning");
      case "WORD_FROM_IMAGE":
        return t("quiz.whatWordDescribesImage");
      case "WORD_FROM_AUDIO":
        return t("quiz.whichWordDidYouHear");
      case "MEANING_FROM_AUDIO":
        return t("quiz.whatDoesWordYouHeardMean");
      case "WORD_FROM_EXAMPLE_AUDIO":
        return t("quiz.whichWordFromExample");
      default:
        return t("quiz.title");
    }
  };

  const renderQuizContent = () => {
    if (!quiz) return null;

    switch (quiz.type) {
      case "GUESS_MEANING":
        return (
          <View style={styles.questionContainer}>
            <Text 
              style={styles.wordText}
              accessibilityRole="text"
              accessibilityLabel={`Word to define: ${quiz.value}`}
            >
              {quiz.value}
            </Text>
            {quiz.pos && (
              <Text 
                style={styles.posText}
                accessibilityRole="text"
                accessibilityLabel={`Part of speech: ${quiz.pos}`}
              >
                ({quiz.pos})
              </Text>
            )}
            {quiz.pronunciation && (
              <Text 
                style={styles.pronunciationText}
                accessibilityRole="text"
                accessibilityLabel={`Pronunciation: ${quiz.pronunciation}`}
              >
                /{quiz.pronunciation}/
              </Text>
            )}
            {quiz.audioURL && (
              <TouchableOpacity 
                style={styles.audioButton} 
                onPress={playAudio}
                accessibilityRole="button"
                accessibilityLabel={isPlaying ? "Pause pronunciation" : "Play pronunciation"}
                accessibilityHint="Plays the pronunciation of the current word"
                accessibilityState={{ selected: isPlaying }}
              >
                <Ionicons
                  name={isPlaying ? "pause-circle" : "play-circle"}
                  size={48}
                  color="#FF7B54"
                />
              </TouchableOpacity>
            )}
          </View>
        );

      case "COMPLETE_SENTENCE":
        return (
          <View style={styles.questionContainer}>
            <Text 
              style={styles.sentenceText}
              accessibilityRole="text"
              accessibilityLabel={`Complete this sentence: ${hideSquareBracketContent(quiz.value)}`}
            >
              {hideSquareBracketContent(quiz.value)}
            </Text>
            {quiz.pos && (
              <Text 
                style={styles.posText}
                accessibilityRole="text"
                accessibilityLabel={`Part of speech: ${quiz.pos}${quiz.isVerbType ? ", using inflections" : ""}`}
              >
                ({quiz.pos}
                {quiz.isVerbType && " - using inflections"})
              </Text>
            )}
          </View>
        );

      case "WORD_FROM_MEANING":
        return (
          <View style={styles.questionContainer}>
            <Text 
              style={styles.meaningText}
              accessibilityRole="text"
              accessibilityLabel={`Find the word that matches this meaning: ${quiz.value}`}
            >
              {quiz.value}
            </Text>
          </View>
        );

      case "WORD_FROM_IMAGE":
        return (
          <View style={styles.questionContainer}>
            <View style={styles.imageContainer}>
              {showLoading && !imageError && (
                <View style={styles.imageLoadingContainer}>
                  <ActivityIndicator size="large" color="#FF7B54" />
                  <Text style={styles.imageLoadingText}>
                    {t("quiz.loadingImage")}
                  </Text>
                </View>
              )}
              {imageError ? (
                <View 
                  style={styles.imageErrorContainer}
                  accessibilityRole="alert"
                  accessibilityLabel="Image failed to load"
                >
                  <Ionicons name="image-outline" size={48} color="#B2BEC3" />
                  <Text 
                    style={styles.imageErrorText}
                    accessibilityRole="text"
                  >
                    {t("quiz.imageLoadError")}
                  </Text>
                  <View style={styles.imageErrorActions}>
                    <TouchableOpacity
                      style={styles.retryButton}
                      onPress={retryImageLoad}
                      accessibilityRole="button"
                      accessibilityLabel="Retry loading image"
                      accessibilityHint="Attempts to reload the image"
                    >
                      <Ionicons name="refresh" size={20} color="#FFFFFF" />
                      <Text style={styles.retryButtonText}>
                        {t("common.retry")}
                      </Text>
                    </TouchableOpacity>
                    <TouchableOpacity
                      style={styles.skipImageButton}
                      onPress={onSkipQuestion}
                      accessibilityRole="button"
                      accessibilityLabel="Skip this question"
                      accessibilityHint="Moves to the next question"
                    >
                      <Text style={styles.skipImageButtonText}>
                        {t("quiz.skipQuestion")}
                      </Text>
                    </TouchableOpacity>
                  </View>
                  {imageRetryCount > 0 && (
                    <Text 
                      style={styles.retryCountText}
                      accessibilityRole="text"
                      accessibilityLabel={`Retry attempts: ${imageRetryCount}`}
                    >
                      {t("quiz.retryAttempts", { count: imageRetryCount })}
                    </Text>
                  )}
                </View>
              ) : (
                <Image
                  source={{ uri: currentImageUrl || quiz.value }}
                  style={[
                    styles.questionImage,
                    (imageLoading || imageError) && styles.hiddenImage,
                  ]}
                  resizeMode="contain"
                  accessible={true}
                  accessibilityRole="image"
                  accessibilityLabel={quiz.imageDescription ? 
                    `Quiz image: ${hideSquareBracketContent(quiz.imageDescription)}` : 
                    "Quiz image for word identification"}
                  accessibilityHint="This image represents the word you need to identify"
                  onLoadStart={() => {
                    // Only set loading if not already loading to prevent rapid state changes
                    if (!imageLoading) {
                      setImageLoading(true);
                      setImageError(false);
                      setShowLoading(false);

                      // Delay showing loading indicator to prevent flashing for quick loads
                      loadingDelayRef.current = setTimeout(() => {
                        setShowLoading(true);
                      }, 300); // Show loading after 300ms
                    }

                    // Set a timeout for image loading (10 seconds)
                    if (imageLoadTimeoutRef.current) {
                      clearTimeout(imageLoadTimeoutRef.current);
                    }
                    imageLoadTimeoutRef.current = setTimeout(() => {
                      setImageLoading(false);
                      setShowLoading(false);
                      setImageError(true);
                    }, 10000);
                  }}
                  onLoad={() => {
                    // Clear all timeouts
                    if (imageLoadTimeoutRef.current) {
                      clearTimeout(imageLoadTimeoutRef.current);
                      imageLoadTimeoutRef.current = null;
                    }
                    if (loadingDelayRef.current) {
                      clearTimeout(loadingDelayRef.current);
                      loadingDelayRef.current = null;
                    }

                    setImageLoading(false);
                    setShowLoading(false);
                    setImageError(false);
                  }}
                  onError={() => {
                    // Clear all timeouts
                    if (imageLoadTimeoutRef.current) {
                      clearTimeout(imageLoadTimeoutRef.current);
                      imageLoadTimeoutRef.current = null;
                    }
                    if (loadingDelayRef.current) {
                      clearTimeout(loadingDelayRef.current);
                      loadingDelayRef.current = null;
                    }

                    setImageLoading(false);
                    setShowLoading(false);
                    setImageError(true);
                  }}
                />
              )}
            </View>
            {quiz.imageDescription && !imageLoading && !imageError && (
              <Text style={styles.imageDescription}>
                {hideSquareBracketContent(quiz.imageDescription)}
              </Text>
            )}
          </View>
        );

      case "WORD_FROM_AUDIO":
      case "MEANING_FROM_AUDIO":
      case "WORD_FROM_EXAMPLE_AUDIO":
        const getAudioQuizLabel = () => {
          switch (quiz.type) {
            case "WORD_FROM_AUDIO":
              return isPlaying ? "Audio is playing, listen for the word" : "Play audio to hear the word";
            case "MEANING_FROM_AUDIO":
              return isPlaying ? "Audio is playing, listen and select the meaning" : "Play audio to hear the word, then select its meaning";
            case "WORD_FROM_EXAMPLE_AUDIO":
              return isPlaying ? "Audio is playing, listen for the example sentence" : "Play audio to hear an example sentence";
            default:
              return isPlaying ? "Audio is playing" : "Tap to play audio";
          }
        };

        return (
          <View style={styles.questionContainer}>
            <TouchableOpacity
              style={styles.largeAudioButton}
              onPress={playAudio}
              accessibilityRole="button"
              accessibilityLabel={getAudioQuizLabel()}
              accessibilityHint="Double tap to play or pause the audio"
              accessibilityState={{ selected: isPlaying }}
            >
              <Ionicons
                name={isPlaying ? "pause-circle" : "play-circle"}
                size={80}
                color="#FF7B54"
              />
              <Text 
                style={styles.audioText}
                accessibilityRole="text"
              >
                {isPlaying ? t("quiz.audioPlaying") : t("quiz.audioTapToPlay")}
              </Text>
            </TouchableOpacity>
          </View>
        );

      case "WRITE_WORD_FROM_DEFINITION": {
        const correctAnswer = quiz.options[quiz.answerIndex];
        return (
          <View style={styles.questionContainer}>
            <Text 
              style={styles.meaningText}
              accessibilityRole="text"
              accessibilityLabel={`Type the word that matches this definition: ${quiz.value}`}
            >
              {quiz.value}
            </Text>
            {quiz.pos && (
              <Text 
                style={styles.posText}
                accessibilityRole="text"
                accessibilityLabel={`Part of speech: ${quiz.pos}`}
              >
                ({quiz.pos})
              </Text>
            )}
            {quiz.pronunciation && (
              <Text 
                style={styles.pronunciationText}
                accessibilityRole="text"
                accessibilityLabel={`Pronunciation: ${quiz.pronunciation}`}
              >
                /{quiz.pronunciation}/
              </Text>
            )}
            <View style={styles.writeInputContainer}>
              <TextInput
                style={[
                  styles.writeInput,
                  isSubmitted &&
                    userInput.toLowerCase().trim() ===
                      correctAnswer.toLowerCase() &&
                    styles.correctInput,
                  isSubmitted &&
                    userInput.toLowerCase().trim() !==
                      correctAnswer.toLowerCase() &&
                    styles.incorrectInput,
                ]}
                placeholder={t("quiz.typeAnswerPlaceholder")}
                placeholderTextColor="#B2BEC3"
                value={userInput}
                onChangeText={setUserInput}
                autoCapitalize="none"
                autoCorrect={false}
                editable={!isSubmitted}
                onSubmitEditing={onSubmitAnswer}
                accessibilityRole="search"
                accessibilityLabel="Type your answer here"
                accessibilityHint="Enter the word that matches the given definition"
                accessibilityState={{ 
                  disabled: isSubmitted,
                  selected: isSubmitted && userInput.toLowerCase().trim() === correctAnswer.toLowerCase()
                }}
              />
              {!isSubmitted && (
                <View style={styles.buttonContainer}>
                  <TouchableOpacity
                    style={[
                      styles.submitAnswerButton,
                      !userInput.trim() && styles.submitButtonDisabled,
                    ]}
                    onPress={onSubmitAnswer}
                    disabled={!userInput.trim()}
                    accessibilityRole="button"
                    accessibilityLabel="Submit answer"
                    accessibilityHint="Submit your typed answer for evaluation"
                    accessibilityState={{ disabled: !userInput.trim() }}
                  >
                    <Text style={styles.submitAnswerText}>
                      {t("quiz.submit")}
                    </Text>
                  </TouchableOpacity>
                  <TouchableOpacity
                    style={styles.skipButton}
                    onPress={onSkipQuestion}
                    accessibilityRole="button"
                    accessibilityLabel="Skip this question"
                    accessibilityHint="Move to the next question without answering"
                  >
                    <Text style={styles.skipButtonText}>
                      {t("quiz.skipQuestion")}
                    </Text>
                  </TouchableOpacity>
                </View>
              )}
            </View>
            {isSubmitted && (
              <View 
                style={styles.answerFeedback}
                accessibilityRole="alert"
                accessibilityLiveRegion="assertive"
              >
                {userInput.toLowerCase().trim() ===
                correctAnswer.toLowerCase() ? (
                  <>
                    <Ionicons
                      name="checkmark-circle"
                      size={24}
                      color="#4CAF50"
                    />
                    <Text 
                      style={styles.correctFeedback}
                      accessibilityRole="text"
                      accessibilityLabel={`Correct! ${t("quiz.correctAnswer")}`}
                    >
                      {t("quiz.correctAnswer")}
                    </Text>
                  </>
                ) : (
                  <>
                    <Ionicons name="close-circle" size={24} color="#FF6B6B" />
                    <Text 
                      style={styles.incorrectFeedback}
                      accessibilityRole="text"
                      accessibilityLabel={`Incorrect. The correct answer is: ${correctAnswer}`}
                    >
                      {t("quiz.incorrectAnswer", { answer: correctAnswer })}
                    </Text>
                  </>
                )}
              </View>
            )}
          </View>
        );
      }

      default:
        return null;
    }
  };

  return (
    <View style={styles.container}>
      <Text 
        style={styles.quizTitle}
        accessibilityRole="header"
        accessibilityLevel={1}
        accessibilityLabel={`Quiz question: ${getQuizTitle()}`}
      >
        {getQuizTitle()}
      </Text>
      {renderQuizContent()}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  quizTitle: {
    fontSize: 20,
    fontWeight: "600",
    color: "#2D3436",
    textAlign: "center",
    marginBottom: 24,
  },
  questionContainer: {
    alignItems: "center",
    marginBottom: 32,
  },
  wordText: {
    fontSize: 32,
    fontWeight: "700",
    color: "#FF7B54",
    marginBottom: 8,
  },
  posText: {
    fontSize: 16,
    color: "#636E72",
    fontStyle: "italic",
    marginBottom: 16,
  },
  pronunciationText: {
    fontSize: 14,
    color: "#636E72",
    fontStyle: "italic",
    marginBottom: 16,
    textAlign: "center",
  },
  sentenceText: {
    fontSize: 18,
    color: "#2D3436",
    lineHeight: 28,
    textAlign: "center",
  },
  meaningText: {
    fontSize: 18,
    color: "#2D3436",
    lineHeight: 26,
    textAlign: "center",
  },
  questionImage: {
    width: SCREEN_WIDTH - 80,
    height: 200,
    borderRadius: 12,
    marginBottom: 16,
  },
  imageDescription: {
    fontSize: 14,
    color: "#636E72",
    fontStyle: "italic",
  },
  audioButton: {
    marginTop: 16,
  },
  largeAudioButton: {
    alignItems: "center",
    gap: 12,
  },
  audioText: {
    fontSize: 16,
    color: "#636E72",
  },
  imageContainer: {
    width: SCREEN_WIDTH - 80,
    height: 200,
    position: "relative",
    alignItems: "center",
    justifyContent: "center",
  },
  imageLoadingContainer: {
    position: "absolute",
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: "#F5F5F5",
    borderRadius: 12,
    justifyContent: "center",
    alignItems: "center",
    zIndex: 1,
  },
  imageLoadingText: {
    marginTop: 12,
    fontSize: 14,
    color: "#636E72",
  },
  hiddenImage: {
    opacity: 0,
  },
  imageErrorContainer: {
    width: SCREEN_WIDTH - 80,
    height: 200,
    backgroundColor: "#F5F5F5",
    borderRadius: 12,
    justifyContent: "center",
    alignItems: "center",
    padding: 20,
  },
  imageErrorText: {
    fontSize: 16,
    color: "#636E72",
    marginTop: 12,
    marginBottom: 20,
    textAlign: "center",
  },
  imageErrorActions: {
    flexDirection: "row",
    gap: 12,
  },
  retryButton: {
    flexDirection: "row",
    backgroundColor: "#FF7B54",
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 8,
    alignItems: "center",
    gap: 8,
  },
  retryButtonText: {
    color: "#FFFFFF",
    fontSize: 14,
    fontWeight: "600",
  },
  skipImageButton: {
    backgroundColor: "transparent",
    borderWidth: 2,
    borderColor: "#E0E0E0",
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 8,
  },
  skipImageButtonText: {
    color: "#636E72",
    fontSize: 14,
    fontWeight: "500",
  },
  retryCountText: {
    fontSize: 12,
    color: "#B2BEC3",
    marginTop: 12,
  },
  writeInputContainer: {
    marginTop: 24,
    width: "100%",
  },
  writeInput: {
    backgroundColor: "#FAFAFA",
    borderWidth: 2,
    borderColor: "#E0E0E0",
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 18,
    color: "#2D3436",
    textAlign: "center",
    marginBottom: 16,
  },
  correctInput: {
    borderColor: "#4CAF50",
    backgroundColor: "#E8F5E9",
  },
  incorrectInput: {
    borderColor: "#FF6B6B",
    backgroundColor: "#FFEBEE",
  },
  buttonContainer: {
    gap: 12,
  },
  submitAnswerButton: {
    backgroundColor: "#FF7B54",
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: "center",
  },
  submitButtonDisabled: {
    backgroundColor: "#DFE6E9",
    opacity: 0.6,
  },
  submitAnswerText: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  skipButton: {
    backgroundColor: "transparent",
    borderWidth: 2,
    borderColor: "#E0E0E0",
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: "center",
  },
  skipButtonText: {
    color: "#636E72",
    fontSize: 16,
    fontWeight: "500",
  },
  answerFeedback: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    marginTop: 16,
    gap: 8,
  },
  correctFeedback: {
    color: "#4CAF50",
    fontSize: 16,
    fontWeight: "600",
  },
  incorrectFeedback: {
    color: "#FF6B6B",
    fontSize: 16,
    textAlign: "center",
  },
});
