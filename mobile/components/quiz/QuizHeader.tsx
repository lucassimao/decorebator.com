import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import React from "react";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";

interface QuizHeaderProps {
  wordlistName: string;
  correctCount: number;
  quizCount: number;
  isOnline: boolean;
  showQuizTypeSelector?: boolean;
  quizTypeCount?: number;
  quizTypeTotal?: number;
  isQuizTypeDisabled?: boolean;
  onQuizTypePress?: () => void;
  onBackPress: () => void;
  onReportPress: () => void;
}

export const QuizHeader: React.FC<QuizHeaderProps> = ({
  wordlistName,
  correctCount,
  quizCount,
  isOnline,
  showQuizTypeSelector = false,
  quizTypeCount,
  quizTypeTotal,
  isQuizTypeDisabled = false,
  onQuizTypePress,
  onBackPress,
  onReportPress,
}) => {
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const showQuizTypeBadge =
    showQuizTypeSelector &&
    quizTypeCount != null &&
    quizTypeTotal != null &&
    quizTypeCount < quizTypeTotal;
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
          size={24}
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

      <View style={styles.headerActions}>
        {showQuizTypeSelector && (
          <TouchableOpacity
            style={[
              styles.settingsButton,
              isQuizTypeDisabled && styles.disabledButton,
            ]}
            onPress={() => !isQuizTypeDisabled && onQuizTypePress?.()}
            disabled={isQuizTypeDisabled}
            accessibilityRole="button"
            accessibilityLabel="Choose quiz types"
            accessibilityHint={
              isQuizTypeDisabled
                ? "Requires internet connection"
                : "Filter the types of questions in this session"
            }
            accessibilityState={{ disabled: isQuizTypeDisabled }}
          >
            <MaterialIcons
              name="tune"
              size={22}
              color={
                isQuizTypeDisabled
                  ? theme.colors.ui.disabled
                  : theme.colors.text.secondary
              }
            />
            {showQuizTypeBadge && (
              <View style={styles.quizTypeBadge}>
                <Text style={styles.quizTypeBadgeText}>{quizTypeCount}</Text>
              </View>
            )}
          </TouchableOpacity>
        )}

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
            size={24}
            color={
              isOnline ? theme.colors.text.secondary : theme.colors.ui.disabled
            }
          />
        </TouchableOpacity>
      </View>
    </View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    header: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingHorizontal: 20,
      paddingVertical: 16,
    },
    backButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    headerCenter: {
      flex: 1,
      alignItems: "center",
      marginHorizontal: 16,
    },
    headerTitle: {
      fontSize: 18,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    headerSubtitle: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginTop: 2,
    },
    settingsButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    headerActions: {
      flexDirection: "row",
      alignItems: "center",
      gap: 10,
    },
    disabledButton: {
      opacity: 0.5,
    },
    quizTypeBadge: {
      position: "absolute",
      top: -4,
      right: -4,
      minWidth: 18,
      height: 18,
      borderRadius: 9,
      backgroundColor: theme.colors.primary,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: 4,
    },
    quizTypeBadgeText: {
      fontSize: 10,
      color: theme.colors.text.inverse,
      fontWeight: "700",
    },
  });
