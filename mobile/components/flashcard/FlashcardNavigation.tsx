import React from "react";
import { View, TouchableOpacity, StyleSheet } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";
import { useTheme } from "@/contexts/ThemeContext";

interface FlashcardNavigationProps {
  currentIndex: number;
  totalWords: number;
  onNavigate: (direction: "next" | "prev") => void;
  onFinish: () => void;
}

export const FlashcardNavigation: React.FC<FlashcardNavigationProps> = ({
  currentIndex,
  totalWords,
  onNavigate,
  onFinish,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);
  const isFirstCard = currentIndex === 0;
  const isLastCard = currentIndex === totalWords - 1;

  return (
    <View style={styles.navigationContainer}>
      <TouchableOpacity
        style={[styles.previousButton, isFirstCard && styles.disabledButton]}
        onPress={() => onNavigate("prev")}
        disabled={isFirstCard}
        accessibilityRole="button"
        accessibilityLabel={t("flashcards.previousCard")}
        accessibilityState={{ disabled: isFirstCard }}
      >
        <Ionicons
          name="chevron-back"
          size={28}
          color={
            isFirstCard
              ? theme.colors.roles.disabledText
              : theme.colors.text.primary
          }
        />
      </TouchableOpacity>

      <Button
        style={styles.actionButton}
        controlStyle={styles.actionControl}
        onPress={isLastCard ? onFinish : () => onNavigate("next")}
        accessibilityLabel={
          isLastCard ? t("flashcards.finishWordlist") : t("flashcards.nextCard")
        }
        trailing={
          <Ionicons
            name={isLastCard ? "flag-outline" : "arrow-forward"}
            size={20}
            color={theme.colors.roles.onAction}
            accessibilityElementsHidden
            importantForAccessibility="no-hide-descendants"
          />
        }
      >
        {isLastCard ? t("flashcards.finishWordlist") : t("flashcards.nextCard")}
      </Button>
    </View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    navigationContainer: {
      flexDirection: "row",
      alignItems: "center",
      gap: theme.spacing.compact,
      paddingHorizontal: theme.spacing.comfortable,
      paddingTop: theme.spacing.comfortable,
      paddingBottom: theme.spacing.lg,
    },
    previousButton: {
      width: 56,
      height: 56,
      borderRadius: theme.borderRadius.full,
      borderWidth: 1,
      borderColor: theme.colors.roles.controlBorder,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    disabledButton: {
      opacity: 0.45,
      backgroundColor: theme.colors.roles.disabledBackground,
      shadowOpacity: 0,
      elevation: 0,
    },
    actionButton: {
      flex: 1,
      minHeight: 56,
    },
    actionControl: { minHeight: 56 },
  });
