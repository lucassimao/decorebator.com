import * as errorReportingApi from "@/api/errorReporting";
import { ErrorType } from "@/api/errorReporting";
import * as quizApi from "@/api/wordlists";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useNavigation, useRoute } from "@react-navigation/native";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useAudioPlayer, useAudioPlayerStatus } from "expo-audio";
import { useLocalSearchParams } from "expo-router";
import React, { useEffect, useRef, useState } from "react";
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  Image,
  ImageBackground,
  Modal,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from "react-native";

const { width: SCREEN_WIDTH, height: SCREEN_HEIGHT } = Dimensions.get("window");

const QuizScreen: React.FC = () => {
  const navigation = useNavigation();
  const { wordlistId, wordlistName } = useLocalSearchParams();

  // State
  const [selectedAnswer, setSelectedAnswer] = useState<number | null>(null);
  const [showResult, setShowResult] = useState(false);
  const [fastMode, setFastMode] = useState(false);
  const [showReportModal, setShowReportModal] = useState(false);
  const player = useAudioPlayer();
  const { playing: isPlaying, didJustFinish } = useAudioPlayerStatus(player);
  const [quizCount, setQuizCount] = useState(0);
  const [correctCount, setCorrectCount] = useState(0);
  const [imageLoading, setImageLoading] = useState(false);
  const [currentImageUrl, setCurrentImageUrl] = useState<string | null>(null);
  const [userInput, setUserInput] = useState("");
  const [isSubmitted, setIsSubmitted] = useState(false);

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
  } = useQuery({
    queryKey: ["quiz", wordlistId],
    queryFn: () => quizApi.newQuiz(Number(wordlistId)),
  });

  // handle image changes
  useEffect(() => {
    if (quiz?.type === "WORD_FROM_IMAGE" && quiz.value !== currentImageUrl) {
      setImageLoading(true);
      setCurrentImageUrl(quiz.value);
    }
  }, [quiz?.value, quiz?.type]);

  // Answer mutation
  const answerMutation = useMutation<void, Error, { success: boolean }>({
    mutationFn: ({ success }) =>
      quizApi.answerQuiz(Number(wordlistId), quiz!.id, success),
    onSuccess: () => {
      if (fastMode) {
        // Auto advance in fast mode
        setTimeout(handleNextQuiz, 600);
      }
    },
    onError: console.error,
  });

  // Report error mutation
  const reportMutation = useMutation({
    mutationFn: ({ errorType }: { errorType: errorReportingApi.ErrorType }) =>
      errorReportingApi.reportError(errorType, quiz!),
    onSuccess: () => {
      Alert.alert("Thank you", "Your report has been submitted.");
      setShowReportModal(false);
      handleNextQuiz();
    },
  });

  // Audio setup
  useEffect(() => {
    if (quiz?.audioURL) {
      player.replace(quiz.audioURL);
      player.seekTo(0);
    }
  }, [quiz?.audioURL, player]);

  const playAudio = () => player.play();

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

  // Add this function
  const handleWriteAnswer = () => {
    if (!userInput.trim() || !quiz) return;

    setIsSubmitted(true);
    setShowResult(true);

    const isCorrect = userInput.toLowerCase() === quiz.value?.toLowerCase();
    setQuizCount((prev) => prev + 1);
    if (isCorrect) {
      setCorrectCount((prev) => prev + 1);
    }

    answerMutation.mutate({ success: isCorrect });
  };

  const handleNextQuiz = () => {
    setSelectedAnswer(null);
    setShowResult(false);
    setUserInput(""); // Add this
    setIsSubmitted(false); // Add this
    refetch();
  };

  const handleReportError = (errorType: ErrorType) => {
    reportMutation.mutate({ errorType });
  };

  const onPressFastModeToggle = () => {
    setFastMode((v) => !v);

    // If toggled on while seem the quiz result, jump to the next quiz
    if (!fastMode && showResult) {
      handleNextQuiz();
    }
  };

  const getQuizTitle = () => {
    switch (quiz?.type) {
      case "WRITE_WORD_FROM_DEFINITION":
        return "Write the word for this definition";
      case "GUESS_MEANING":
        return "What does this word mean?";
      case "COMPLETE_SENTENCE":
        return "Complete the sentence";
      case "WORD_FROM_MEANING":
        return "Which word matches this meaning?";
      case "WORD_FROM_IMAGE":
        return "What word describes this image?";
      case "WORD_FROM_AUDIO":
        return "Which word did you hear?";
      case "MEANING_FROM_AUDIO":
        return "What does the word you heard mean?";
      default:
        return "Quiz";
    }
  };

  const hideSquareBracketContent = (text: string): string => {
    // Replace content between square brackets with underscores
    return text.replace(/\[([^\]]+)\]/g, "_____");
  };

  const renderQuizContent = () => {
    if (!quiz) return null;

    switch (quiz.type) {
      case "GUESS_MEANING":
        return (
          <View style={styles.questionContainer}>
            <Text style={styles.wordText}>{quiz.value}</Text>
            {quiz.pos && <Text style={styles.posText}>({quiz.pos})</Text>}
            {quiz.audioURL && (
              <TouchableOpacity style={styles.audioButton} onPress={playAudio}>
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
            <Text style={styles.sentenceText}>
              {hideSquareBracketContent(quiz.value)}
            </Text>
          </View>
        );

      case "WORD_FROM_MEANING":
        return (
          <View style={styles.questionContainer}>
            <Text style={styles.meaningText}>{quiz.value}</Text>
          </View>
        );

      case "WORD_FROM_IMAGE":
        return (
          <View style={styles.questionContainer}>
            <View style={styles.imageContainer}>
              {imageLoading && (
                <View style={styles.imageLoadingContainer}>
                  <ActivityIndicator size="large" color="#FF7B54" />
                  <Text style={styles.imageLoadingText}>Loading image...</Text>
                </View>
              )}
              <Image
                source={{ uri: quiz.value }}
                style={[
                  styles.questionImage,
                  imageLoading && styles.hiddenImage,
                ]}
                resizeMode="contain"
                onLoadStart={() => setImageLoading(true)}
                onLoadEnd={() => setImageLoading(false)}
                onError={() => {
                  setImageLoading(false);
                  Alert.alert("Error", "Failed to load image");
                }}
              />
            </View>
            {quiz.imageDescription && !imageLoading && (
              <Text style={styles.imageDescription}>
                {hideSquareBracketContent(quiz.imageDescription)}
              </Text>
            )}
          </View>
        );

      case "WORD_FROM_AUDIO":
      case "MEANING_FROM_AUDIO":
        return (
          <View style={styles.questionContainer}>
            <TouchableOpacity
              style={styles.largeAudioButton}
              onPress={playAudio}
            >
              <Ionicons
                name={isPlaying ? "pause-circle" : "play-circle"}
                size={80}
                color="#FF7B54"
              />
              <Text style={styles.audioText}>
                {isPlaying ? "Playing..." : "Tap to play"}
              </Text>
            </TouchableOpacity>
          </View>
        );
      // Add this case to renderQuizContent
      case "WRITE_WORD_FROM_DEFINITION":{
        const correctAnswer = quiz.options[quiz.answerIndex]
        return (
          <View style={styles.questionContainer}>
            <Text style={styles.meaningText}>{quiz.value}</Text>
            {quiz.pos && <Text style={styles.posText}>({quiz.pos})</Text>}
            <View style={styles.writeInputContainer}>
              <TextInput
                style={[
                  styles.writeInput,
                  isSubmitted &&
                    userInput.toLowerCase() ===
                      correctAnswer.toLowerCase() &&
                    styles.correctInput,
                  isSubmitted &&
                    userInput.toLowerCase() !==
                      correctAnswer.toLowerCase() &&
                    styles.incorrectInput,
                ]}
                placeholder="Type your answer here"
                placeholderTextColor="#B2BEC3"
                value={userInput}
                onChangeText={setUserInput}
                autoCapitalize="none"
                autoCorrect={false}
                editable={!isSubmitted}
                onSubmitEditing={() => handleWriteAnswer()}
              />
              {!isSubmitted && (
                <TouchableOpacity
                  style={[
                    styles.submitAnswerButton,
                    !userInput.trim() && styles.submitButtonDisabled,
                  ]}
                  onPress={handleWriteAnswer}
                  disabled={!userInput.trim()}
                >
                  <Text style={styles.submitAnswerText}>Submit</Text>
                </TouchableOpacity>
              )}
            </View>
            {isSubmitted && (
              <View style={styles.answerFeedback}>
                {userInput.toLowerCase() ===
                correctAnswer.toLowerCase() ? (
                  <>
                    <Ionicons
                      name="checkmark-circle"
                      size={24}
                      color="#4CAF50"
                    />
                    <Text style={styles.correctFeedback}>Correct!</Text>
                  </>
                ) : (
                  <>
                    <Ionicons name="close-circle" size={24} color="#FF6B6B" />
                    <Text style={styles.incorrectFeedback}>
                      Incorrect. The answer is: {correctAnswer}
                    </Text>
                  </>
                )}
              </View>
            )}
          </View>
        );
      }
    }
  };

  const getOptionStyle = (index: number) => {
    if (!showResult) return styles.optionButton;

    const isSelected = selectedAnswer === index;
    const isCorrect = index === quiz?.answerIndex;

    if (isCorrect) {
      return [styles.optionButton, styles.correctOption];
    } else if (isSelected && !isCorrect) {
      return [styles.optionButton, styles.incorrectOption];
    }

    return [styles.optionButton, styles.disabledOption];
  };

  const getOptionTextStyle = (index: number) => {
    if (!showResult) return styles.optionText;

    const isCorrect = index === quiz?.answerIndex;
    const isSelected = selectedAnswer === index;

    if (isCorrect) {
      return [styles.optionText, styles.correctOptionText];
    } else if (isSelected && !isCorrect) {
      return [styles.optionText, styles.incorrectOptionText];
    }

    return [styles.optionText, styles.disabledOptionText];
  };

  if (isLoading && quizCount === 0) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#FF7B54" />
      </View>
    );
  }

  return (
    <ImageBackground
      source={require("@/assets/images/dashboard-bg.png")}
      style={styles.backgroundImage}
      resizeMode="cover"
    >
      <SafeAreaView style={styles.container}>
        {/* Header */}
        <View style={styles.header}>
          <TouchableOpacity
            style={styles.backButton}
            onPress={() => navigation.goBack()}
          >
            <Ionicons name="arrow-back" size={24} color="#2D3436" />
          </TouchableOpacity>

          <View style={styles.headerCenter}>
            <Text style={styles.headerTitle}>{wordlistName || "Quiz"}</Text>
            <Text style={styles.headerSubtitle}>
              {correctCount}/{quizCount} correct
            </Text>
          </View>

          <TouchableOpacity
            style={styles.settingsButton}
            onPress={() => setShowReportModal(true)}
          >
            <MaterialIcons name="flag" size={24} color="#636E72" />
          </TouchableOpacity>
        </View>

        {/* Progress Bar */}
        <View style={styles.progressContainer}>
          <View style={styles.progressBar}>
            <View
              style={[
                styles.progressFill,
                {
                  width: `${quizCount > 0 ? (correctCount / quizCount) * 100 : 0}%`,
                },
              ]}
            />
          </View>
        </View>

        {/* Fast Mode Toggle */}
        <View style={styles.modeContainer}>
          <Text style={styles.modeText}>Fast Mode</Text>
          <TouchableOpacity
            style={[styles.modeToggle, fastMode && styles.modeToggleActive]}
            onPress={onPressFastModeToggle}
          >
            <View
              style={[
                styles.modeToggleCircle,
                fastMode && styles.modeToggleCircleActive,
              ]}
            />
          </TouchableOpacity>
        </View>

        {/* Quiz Content */}
        <ScrollView
          contentContainerStyle={styles.scrollContent}
          showsVerticalScrollIndicator={false}
        >
          <View style={styles.quizCard}>
            <Text style={styles.quizTitle}>{getQuizTitle()}</Text>

            {renderQuizContent()}

            {/* Options */}
            {quiz?.type !== "WRITE_WORD_FROM_DEFINITION" && (
              <View style={styles.optionsContainer}>
                {quiz?.options.map((option, index) => (
                  <TouchableOpacity
                    key={index}
                    style={getOptionStyle(index)}
                    onPress={() => handleAnswerSelect(index)}
                    disabled={showResult}
                    activeOpacity={0.7}
                  >
                    <Text style={getOptionTextStyle(index)}>{option}</Text>
                    {showResult && index === quiz.answerIndex && (
                      <Ionicons
                        name="checkmark-circle"
                        size={24}
                        color="#4CAF50"
                      />
                    )}
                    {showResult &&
                      selectedAnswer === index &&
                      index !== quiz.answerIndex && (
                        <Ionicons
                          name="close-circle"
                          size={24}
                          color="#FF6B6B"
                        />
                      )}
                  </TouchableOpacity>
                ))}
              </View>
            )}

            {/* Next Button */}
            {showResult &&
              !fastMode &&
              (quiz?.type === "WRITE_WORD_FROM_DEFINITION"
                ? isSubmitted
                : true) && (
                <TouchableOpacity
                  style={styles.nextButton}
                  onPress={handleNextQuiz}
                >
                  <Text style={styles.nextButtonText}>Next Question</Text>
                  <Ionicons name="arrow-forward" size={20} color="#FFFFFF" />
                </TouchableOpacity>
              )}
          </View>
        </ScrollView>

        {/* Report Modal */}
        <Modal
          visible={showReportModal}
          transparent
          animationType="slide"
          onRequestClose={() => setShowReportModal(false)}
        >
          <View style={styles.modalOverlay}>
            <View style={styles.modalContent}>
              <Text style={styles.modalTitle}>Report an Issue</Text>

              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => handleReportError(ErrorType.UnrelatedImage)}
              >
                <MaterialIcons name="image" size={24} color="#636E72" />
                <Text style={styles.reportOptionText}>Image doesn't match</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => handleReportError(ErrorType.MissingImage)}
              >
                <MaterialIcons name="broken-image" size={24} color="#636E72" />
                <Text style={styles.reportOptionText}>Image not loading</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => handleReportError(ErrorType.UnrelatedMeaning)}
              >
                <MaterialIcons name="help-outline" size={24} color="#636E72" />
                <Text style={styles.reportOptionText}>Wrong meaning</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => handleReportError(ErrorType.UnrelatedExample)}
              >
                <MaterialIcons name="format-quote" size={24} color="#636E72" />
                <Text style={styles.reportOptionText}>Wrong example</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={styles.reportOption}
                onPress={() => handleReportError(ErrorType.SoundNotPlaying)}
              >
                <MaterialIcons name="volume-off" size={24} color="#636E72" />
                <Text style={styles.reportOptionText}>Sound not playing</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={styles.cancelButton}
                onPress={() => setShowReportModal(false)}
              >
                <Text style={styles.cancelButtonText}>Cancel</Text>
              </TouchableOpacity>
            </View>
          </View>
        </Modal>
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
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 20,
    paddingVertical: 16,
  },
  backButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    justifyContent: "center",
    alignItems: "center",
  },
  headerCenter: {
    flex: 1,
    alignItems: "center",
    marginHorizontal: 16,
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: "600",
    color: "#2D3436",
  },
  headerSubtitle: {
    fontSize: 14,
    color: "#636E72",
    marginTop: 2,
  },
  settingsButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    justifyContent: "center",
    alignItems: "center",
  },
  progressContainer: {
    paddingHorizontal: 20,
    marginBottom: 16,
  },
  progressBar: {
    height: 8,
    backgroundColor: "rgba(255, 255, 255, 0.5)",
    borderRadius: 4,
    overflow: "hidden",
  },
  progressFill: {
    height: "100%",
    backgroundColor: "#4CAF50",
    borderRadius: 4,
  },
  modeContainer: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    marginBottom: 16,
    gap: 12,
  },
  modeText: {
    fontSize: 16,
    color: "#2D3436",
    fontWeight: "500",
  },
  modeToggle: {
    width: 50,
    height: 30,
    borderRadius: 15,
    backgroundColor: "#E0E0E0",
    padding: 3,
  },
  modeToggleActive: {
    backgroundColor: "#FF7B54",
  },
  modeToggleCircle: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: "#FFFFFF",
  },
  modeToggleCircleActive: {
    transform: [{ translateX: 20 }],
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
  optionsContainer: {
    gap: 12,
  },
  optionButton: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    backgroundColor: "#FAFAFA",
    borderWidth: 2,
    borderColor: "#E0E0E0",
    borderRadius: 12,
    paddingVertical: 16,
    paddingHorizontal: 20,
  },
  correctOption: {
    borderColor: "#4CAF50",
    backgroundColor: "#E8F5E9",
  },
  incorrectOption: {
    borderColor: "#FF6B6B",
    backgroundColor: "#FFEBEE",
  },
  disabledOption: {
    opacity: 0.6,
  },
  optionText: {
    fontSize: 16,
    color: "#2D3436",
    flex: 1,
  },
  correctOptionText: {
    color: "#2E7D32",
    fontWeight: "600",
  },
  incorrectOptionText: {
    color: "#C62828",
    fontWeight: "600",
  },
  disabledOptionText: {
    color: "#636E72",
  },
  nextButton: {
    backgroundColor: "#FF7B54",
    borderRadius: 12,
    paddingVertical: 16,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    marginTop: 24,
  },
  nextButtonText: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: "rgba(0, 0, 0, 0.5)",
    justifyContent: "flex-end",
  },
  modalContent: {
    backgroundColor: "#FFFFFF",
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    padding: 24,
    paddingBottom: 40,
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: "600",
    color: "#2D3436",
    textAlign: "center",
    marginBottom: 24,
  },
  reportOption: {
    flexDirection: "row",
    alignItems: "center",
    paddingVertical: 16,
    gap: 16,
  },
  reportOptionText: {
    fontSize: 16,
    color: "#2D3436",
    flex: 1,
  },
  cancelButton: {
    alignItems: "center",
    paddingVertical: 16,
    marginTop: 16,
  },
  cancelButtonText: {
    fontSize: 16,
    color: "#636E72",
    fontWeight: "500",
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
