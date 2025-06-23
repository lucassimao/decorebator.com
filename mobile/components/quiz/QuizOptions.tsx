import { Ionicons } from "@expo/vector-icons";
import React from "react";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";

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

interface QuizOptionsProps {
  quiz: Quiz;
  selectedAnswer: number | null;
  showResult: boolean;
  onAnswerSelect: (index: number) => void;
}

export const QuizOptions: React.FC<QuizOptionsProps> = ({
  quiz,
  selectedAnswer,
  showResult,
  onAnswerSelect,
}) => {
  const { theme } = useTheme();
  const styles = createStyles(theme);
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

  if (quiz?.type === "WRITE_WORD_FROM_DEFINITION") {
    return null; // This quiz type doesn't use options
  }

  const getAccessibilityLabel = (option: string, index: number) => {
    if (!showResult) {
      return `Option ${index + 1}: ${option}`;
    }

    const isCorrect = index === quiz?.answerIndex;
    const isSelected = selectedAnswer === index;

    if (isCorrect) {
      return `Correct answer: ${option}`;
    } else if (isSelected && !isCorrect) {
      return `Your incorrect selection: ${option}`;
    }

    return `Option ${index + 1}: ${option}`;
  };

  const getAccessibilityHint = (index: number) => {
    if (showResult) return undefined;
    return "Double tap to select this answer";
  };

  const getAccessibilityState = (index: number) => {
    if (!showResult) {
      return {
        selected: selectedAnswer === index,
        disabled: false,
      };
    }

    const isCorrect = index === quiz?.answerIndex;
    const isSelected = selectedAnswer === index;

    return {
      selected: isSelected,
      disabled: true,
      checked: isCorrect,
    };
  };

  return (
    <View
      style={styles.optionsContainer}
      accessibilityRole="radiogroup"
      accessibilityLabel="Answer options"
    >
      {quiz?.options.map((option, index) => (
        <TouchableOpacity
          key={index}
          style={getOptionStyle(index)}
          onPress={() => onAnswerSelect(index)}
          disabled={showResult}
          activeOpacity={0.7}
          accessibilityRole="radio"
          accessibilityLabel={getAccessibilityLabel(option, index)}
          accessibilityHint={getAccessibilityHint(index)}
          accessibilityState={getAccessibilityState(index)}
        >
          <Text style={getOptionTextStyle(index)}>{option}</Text>
          {showResult && index === quiz.answerIndex && (
            <Ionicons
              name="checkmark-circle"
              size={24}
              color={theme.colors.success}
              accessibilityLabel="Correct answer indicator"
            />
          )}
          {showResult &&
            selectedAnswer === index &&
            index !== quiz.answerIndex && (
              <Ionicons
                name="close-circle"
                size={24}
                color={theme.colors.error}
                accessibilityLabel="Incorrect answer indicator"
              />
            )}
        </TouchableOpacity>
      ))}
    </View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    optionsContainer: {
      gap: 12,
    },
    optionButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      backgroundColor: theme.colors.background.surface,
      borderWidth: 2,
      borderColor: theme.colors.ui.border,
      borderRadius: 12,
      paddingVertical: 16,
      paddingHorizontal: 20,
    },
    correctOption: {
      borderColor: theme.colors.success,
      backgroundColor:
        theme.mode === "light" ? "#E8F5E9" : theme.colors.background.elevated,
    },
    incorrectOption: {
      borderColor: theme.colors.error,
      backgroundColor:
        theme.mode === "light" ? "#FFEBEE" : theme.colors.background.elevated,
    },
    disabledOption: {
      opacity: 0.6,
    },
    optionText: {
      fontSize: 16,
      color: theme.colors.text.primary,
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
  });
