import React from "react";
import { useTranslation } from "react-i18next";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

interface QuizModeToggleProps {
  fastMode: boolean;
  onToggle: () => void;
}

export const QuizModeToggle: React.FC<QuizModeToggleProps> = ({
  fastMode,
  onToggle,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, spacing);

  return (
    <View style={styles.modeContainer}>
      <Text style={styles.modeText}>{t("quiz.fastMode")}</Text>
      <TouchableOpacity
        style={[styles.modeToggle, fastMode && styles.modeToggleActive]}
        onPress={onToggle}
      >
        <View
          style={[
            styles.modeToggleCircle,
            fastMode && styles.modeToggleCircleActive,
          ]}
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
    modeContainer: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      marginBottom: spacing.md,
      gap: spacing.sm,
    },
    modeText: {
      fontSize: isTablet ? 18 : 16,
      color: theme.colors.text.primary,
      fontWeight: "500",
    },
    modeToggle: {
      width: isTablet ? 60 : 50,
      height: isTablet ? 36 : 30,
      borderRadius: isTablet ? 18 : 15,
      backgroundColor: theme.colors.ui.border,
      padding: 3,
    },
    modeToggleActive: {
      backgroundColor: theme.colors.primary,
    },
    modeToggleCircle: {
      width: isTablet ? 30 : 24,
      height: isTablet ? 30 : 24,
      borderRadius: isTablet ? 15 : 12,
      backgroundColor: theme.colors.text.inverse,
    },
    modeToggleCircleActive: {
      transform: [{ translateX: isTablet ? 24 : 20 }],
    },
  });
