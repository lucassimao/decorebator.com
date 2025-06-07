import { Ionicons } from "@expo/vector-icons";
import React from "react";
import { useTranslation } from "react-i18next";
import { StyleSheet, Text, TouchableOpacity } from "react-native";

interface QuizNextButtonProps {
  showResult: boolean;
  fastMode: boolean;
  quizType: string;
  isSubmitted: boolean;
  onNextQuiz: () => void;
}

export const QuizNextButton: React.FC<QuizNextButtonProps> = ({
  showResult,
  fastMode,
  quizType,
  isSubmitted,
  onNextQuiz,
}) => {
  const { t } = useTranslation();

  const shouldShowButton =
    showResult &&
    !fastMode &&
    (quizType === "WRITE_WORD_FROM_DEFINITION" ? isSubmitted : true);

  if (!shouldShowButton) {
    return null;
  }

  return (
    <TouchableOpacity style={styles.nextButton} onPress={onNextQuiz}>
      <Text style={styles.nextButtonText}>{t("quiz.nextQuestion")}</Text>
      <Ionicons name="arrow-forward" size={20} color="#FFFFFF" />
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
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
});
