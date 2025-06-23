import React from "react";
import { StyleSheet, View } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";

interface QuizProgressBarProps {
  correctCount: number;
  quizCount: number;
}

export const QuizProgressBar: React.FC<QuizProgressBarProps> = ({
  correctCount,
  quizCount,
}) => {
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const progressPercentage =
    quizCount > 0 ? (correctCount / quizCount) * 100 : 0;

  return (
    <View style={styles.progressContainer}>
      <View style={styles.progressBar}>
        <View
          style={[
            styles.progressFill,
            {
              width: `${progressPercentage}%`,
            },
          ]}
        />
      </View>
    </View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    progressContainer: {
      paddingHorizontal: 20,
      marginBottom: 16,
    },
    progressBar: {
      height: 8,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.1)"
          : "rgba(255, 255, 255, 0.2)",
      borderRadius: 4,
      overflow: "hidden",
    },
    progressFill: {
      height: "100%",
      backgroundColor: theme.colors.success,
      borderRadius: 4,
    },
  });
