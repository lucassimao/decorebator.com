import { Ionicons } from "@expo/vector-icons";
import React from "react";
import { useTranslation } from "react-i18next";
import { StyleSheet, Text, TouchableOpacity } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";

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
  const { theme } = useTheme();
  const styles = createStyles(theme);

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
      <Ionicons
        name="arrow-forward"
        size={20}
        color={theme.colors.text.inverse}
      />
    </TouchableOpacity>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    nextButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: 12,
      paddingVertical: 16,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      gap: 8,
      marginTop: 24,
    },
    nextButtonText: {
      color: theme.colors.text.inverse,
      fontSize: 16,
      fontWeight: "600",
    },
  });
