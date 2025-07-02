import React from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

interface FlashcardNavigationProps {
  currentIndex: number;
  totalWords: number;
  onNavigate: (direction: "next" | "prev") => void;
}

export const FlashcardNavigation: React.FC<FlashcardNavigationProps> = ({
  currentIndex,
  totalWords,
  onNavigate,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, spacing);

  const isFirstCard = currentIndex === 0;
  const isLastCard = currentIndex === totalWords - 1;

  return (
    <View style={styles.navigationContainer}>
      <TouchableOpacity
        style={[styles.navButton, isFirstCard && styles.navButtonDisabled]}
        onPress={() => onNavigate("prev")}
        disabled={isFirstCard}
      >
        <Ionicons
          name="chevron-back"
          size={isTablet ? 40 : 32}
          color={
            isFirstCard ? theme.colors.ui.border : theme.colors.text.primary
          }
        />
      </TouchableOpacity>

      <View style={styles.centerControls}>
        <Text style={styles.swipeHint}>{t("flashcards.navigationHint")}</Text>
      </View>

      <TouchableOpacity
        style={[styles.navButton, isLastCard && styles.navButtonDisabled]}
        onPress={() => onNavigate("next")}
        disabled={isLastCard}
      >
        <Ionicons
          name="chevron-forward"
          size={isTablet ? 40 : 32}
          color={
            isLastCard ? theme.colors.ui.border : theme.colors.text.primary
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
    navigationContainer: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: spacing.lg,
      paddingVertical: spacing.lg,
    },
    navButton: {
      width: isTablet ? 72 : 60,
      height: isTablet ? 72 : 60,
      borderRadius: isTablet ? 36 : 30,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.md,
    },
    navButtonDisabled: {
      opacity: 0.5,
    },
    centerControls: {
      flex: 1,
      alignItems: "center",
    },
    swipeHint: {
      fontSize: isTablet ? 16 : 14,
      color: theme.colors.text.secondary,
      fontStyle: "italic",
    },
  });
