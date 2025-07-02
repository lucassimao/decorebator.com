import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import React from "react";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

interface QuizHeaderProps {
  wordlistName: string;
  correctCount: number;
  quizCount: number;
  isOnline: boolean;
  onBackPress: () => void;
  onReportPress: () => void;
}

export const QuizHeader: React.FC<QuizHeaderProps> = ({
  wordlistName,
  correctCount,
  quizCount,
  isOnline,
  onBackPress,
  onReportPress,
}) => {
  const { theme } = useTheme();
  const { isTablet } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, spacing);
  return (
    <View style={styles.header}>
      <TouchableOpacity
        style={styles.backButton}
        onPress={onBackPress}
        accessibilityRole="button"
        accessibilityLabel="Go back"
        accessibilityHint="Return to previous screen"
      >
        <Ionicons
          name="arrow-back"
          size={isTablet ? 28 : 24}
          color={theme.colors.text.primary}
        />
      </TouchableOpacity>

      <View style={styles.headerCenter}>
        <Text
          style={styles.headerTitle}
          accessibilityRole="header"
          // accessibilityLevel={1} // Removed deprecated prop
        >
          {wordlistName || "Quiz"}
        </Text>
        <Text
          style={styles.headerSubtitle}
          accessibilityRole="text"
          accessibilityLabel={`Score: ${correctCount} out of ${quizCount} questions correct`}
        >
          {correctCount}/{quizCount} correct
        </Text>
      </View>

      <TouchableOpacity
        style={[styles.settingsButton, !isOnline && styles.disabledButton]}
        onPress={() => isOnline && onReportPress()}
        disabled={!isOnline}
        accessibilityRole="button"
        accessibilityLabel={
          isOnline ? "Report error" : "Report error (offline)"
        }
        accessibilityHint={
          isOnline
            ? "Report an issue with the current question"
            : "Requires internet connection"
        }
        accessibilityState={{ disabled: !isOnline }}
      >
        <MaterialIcons
          name="flag"
          size={isTablet ? 28 : 24}
          color={
            isOnline ? theme.colors.text.secondary : theme.colors.ui.disabled
          }
        />
      </TouchableOpacity>
    </View>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  spacing: ReturnType<typeof useResponsiveSpacing>,
) =>
  StyleSheet.create({
    header: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingHorizontal: spacing.lg,
      paddingVertical: spacing.md,
    },
    backButton: {
      width: isTablet ? 48 : 40,
      height: isTablet ? 48 : 40,
      borderRadius: isTablet ? 24 : 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    headerCenter: {
      flex: 1,
      alignItems: "center",
      marginHorizontal: spacing.md,
    },
    headerTitle: {
      fontSize: isTablet ? 22 : 18,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    headerSubtitle: {
      fontSize: isTablet ? 16 : 14,
      color: theme.colors.text.secondary,
      marginTop: 2,
    },
    settingsButton: {
      width: isTablet ? 48 : 40,
      height: isTablet ? 48 : 40,
      borderRadius: isTablet ? 24 : 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    disabledButton: {
      opacity: 0.5,
    },
  });
