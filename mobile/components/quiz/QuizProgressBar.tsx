import React from "react";
import { StyleSheet, View } from "react-native";

interface QuizProgressBarProps {
  correctCount: number;
  quizCount: number;
}

export const QuizProgressBar: React.FC<QuizProgressBarProps> = ({
  correctCount,
  quizCount,
}) => {
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

const styles = StyleSheet.create({
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
});
