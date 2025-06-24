import React from "react";
import { View, Text, TouchableOpacity, StyleSheet, Switch } from "react-native";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";

interface FlashcardHeaderProps {
  wordlistName: string;
  currentIndex: number;
  totalWords: number;
  isOnline: boolean;
  onClose: () => void;
  onReportError: () => void;
  savePosition?: boolean;
  onToggleSavePosition?: () => void;
}

export const FlashcardHeader: React.FC<FlashcardHeaderProps> = ({
  wordlistName,
  currentIndex,
  totalWords,
  isOnline,
  onClose,
  onReportError,
  savePosition = false,
  onToggleSavePosition,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

  return (
    <View>
      <View style={styles.header}>
        <TouchableOpacity
          style={styles.backButton}
          onPress={onClose}
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
            accessibilityLabel={wordlistName}
          >
            {wordlistName}
          </Text>
          <Text
            style={styles.headerSubtitle}
            accessibilityRole="text"
            accessibilityLabel={`Card ${currentIndex + 1} of ${totalWords}`}
          >
            {t("flashcards.cardCounter", {
              current: currentIndex + 1,
              total: totalWords,
            })}
          </Text>
        </View>

        <TouchableOpacity
          style={styles.reportButton}
          onPress={onReportError}
          disabled={!isOnline}
          accessibilityRole="button"
          accessibilityLabel={
            isOnline ? "Report error" : "Report error (offline)"
          }
          accessibilityHint={
            isOnline
              ? "Report an issue with the current flashcard"
              : "Requires internet connection"
          }
          accessibilityState={{ disabled: !isOnline }}
        >
          <MaterialIcons
            name="flag"
            size={24}
            color={
              isOnline ? theme.colors.text.secondary : theme.colors.ui.border
            }
          />
        </TouchableOpacity>
      </View>

      {onToggleSavePosition && (
        <View style={styles.savePositionContainer}>
          <View style={styles.savePositionTextContainer}>
            <MaterialIcons
              name="bookmark"
              size={20}
              color={
                savePosition
                  ? theme.colors.primary
                  : theme.colors.text.secondary
              }
            />
            <Text
              style={[
                styles.savePositionText,
                savePosition && styles.savePositionTextActive,
              ]}
            >
              {t("flashcards.savePosition")}
            </Text>
          </View>
          <Switch
            value={savePosition}
            onValueChange={onToggleSavePosition}
            thumbColor={theme.colors.text.inverse}
            trackColor={{
              false: theme.colors.ui.border,
              true: theme.colors.primary,
            }}
            ios_backgroundColor={theme.colors.ui.border}
            accessibilityRole="switch"
            accessibilityLabel={t("flashcards.savePosition")}
            accessibilityHint="Remember your position when you return to this wordlist"
            accessibilityValue={{ text: savePosition ? "On" : "Off" }}
          />
        </View>
      )}
    </View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    header: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 16,
      paddingVertical: 12,
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
    reportButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    savePositionContainer: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingHorizontal: 16,
      paddingVertical: 12,
      backgroundColor: "transparent",
    },
    savePositionTextContainer: {
      flexDirection: "row",
      alignItems: "center",
    },
    savePositionText: {
      fontSize: 15,
      color: theme.colors.text.primary,
      fontWeight: "600",
      marginLeft: 8,
      textShadowColor: "rgba(255, 255, 255, 0.8)",
      textShadowOffset: { width: 0, height: 1 },
      textShadowRadius: 2,
    },
    savePositionTextActive: {
      color: theme.colors.primary,
      textShadowColor: "rgba(255, 255, 255, 0.8)",
      textShadowOffset: { width: 0, height: 1 },
      textShadowRadius: 2,
    },
  });
