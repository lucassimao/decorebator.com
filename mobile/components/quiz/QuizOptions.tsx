import { Ionicons } from "@expo/vector-icons";
import React from "react";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";

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

  return (
    <View style={styles.optionsContainer}>
      {quiz?.options.map((option, index) => (
        <TouchableOpacity
          key={index}
          style={getOptionStyle(index)}
          onPress={() => onAnswerSelect(index)}
          disabled={showResult}
          activeOpacity={0.7}
        >
          <Text style={getOptionTextStyle(index)}>{option}</Text>
          {showResult && index === quiz.answerIndex && (
            <Ionicons name="checkmark-circle" size={24} color="#4CAF50" />
          )}
          {showResult &&
            selectedAnswer === index &&
            index !== quiz.answerIndex && (
              <Ionicons name="close-circle" size={24} color="#FF6B6B" />
            )}
        </TouchableOpacity>
      ))}
    </View>
  );
};

const styles = StyleSheet.create({
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
});
