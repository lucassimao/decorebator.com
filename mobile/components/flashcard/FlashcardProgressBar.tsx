import React from "react";
import { View, StyleSheet } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

interface FlashcardProgressBarProps {
  currentIndex: number;
  totalWords: number;
}

export const FlashcardProgressBar: React.FC<FlashcardProgressBarProps> = ({
  currentIndex,
  totalWords,
}) => {
  const { theme } = useTheme();
  const { isTablet } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, spacing);
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

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  spacing: ReturnType<typeof useResponsiveSpacing>,
) =>
  StyleSheet.create({
    progressContainer: {
      paddingHorizontal: spacing.lg,
      paddingBottom: spacing.lg,
    },
    progressBar: {
      height: isTablet ? 8 : 6,
      backgroundColor: theme.colors.ui.border,
      borderRadius: isTablet ? 4 : 3,
      overflow: "hidden",
    },
    progressFill: {
      height: "100%",
      backgroundColor: theme.colors.success,
      borderRadius: isTablet ? 4 : 3,
    },
  });
