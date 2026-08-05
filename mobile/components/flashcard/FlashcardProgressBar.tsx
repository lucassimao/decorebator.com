import React from "react";
import { View, StyleSheet } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";
import { useTranslation } from "react-i18next";

interface FlashcardProgressBarProps {
  currentIndex: number;
  totalWords: number;
}

export const FlashcardProgressBar: React.FC<FlashcardProgressBarProps> = ({
  currentIndex,
  totalWords,
}) => {
  const { theme } = useTheme();
  const { t } = useTranslation();
  const styles = createStyles(theme);
  const progressPercentage = ((currentIndex + 1) / totalWords) * 100;

  return (
    <View style={styles.progressContainer}>
      <View
        style={styles.progressBar}
        accessibilityRole="progressbar"
        accessibilityLabel={t("flashcards.progressLabel")}
        accessibilityValue={{
          min: 0,
          max: totalWords,
          now: currentIndex + 1,
        }}
      >
        <View
          style={[styles.progressFill, { width: `${progressPercentage}%` }]}
        />
      </View>
    </View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    progressContainer: {
      paddingHorizontal: 20,
      paddingBottom: 20,
    },
    progressBar: {
      height: 6,
      backgroundColor: theme.colors.ui.border,
      borderRadius: 3,
      overflow: "hidden",
    },
    progressFill: {
      height: "100%",
      backgroundColor: theme.colors.success,
      borderRadius: 3,
    },
  });
