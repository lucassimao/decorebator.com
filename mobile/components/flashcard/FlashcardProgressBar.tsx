import React from "react";
import { View, StyleSheet } from "react-native";

interface FlashcardProgressBarProps {
  currentIndex: number;
  totalWords: number;
}

const colors = {
  success: "#4CAF50",
  borderGray: "#E0E0E0",
};

export const FlashcardProgressBar: React.FC<FlashcardProgressBarProps> = ({
  currentIndex,
  totalWords,
}) => {
  const progressPercentage = ((currentIndex + 1) / totalWords) * 100;

  return (
    <View style={styles.progressContainer}>
      <View style={styles.progressBar}>
        <View
          style={[styles.progressFill, { width: `${progressPercentage}%` }]}
        />
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
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
});
