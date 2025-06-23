import React from "react";
import { useTranslation } from "react-i18next";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";

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
  const styles = createStyles(theme);

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

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    modeContainer: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      marginBottom: 16,
      gap: 12,
    },
    modeText: {
      fontSize: 16,
      color: theme.colors.text.primary,
      fontWeight: "500",
    },
    modeToggle: {
      width: 50,
      height: 30,
      borderRadius: 15,
      backgroundColor: theme.colors.ui.border,
      padding: 3,
    },
    modeToggleActive: {
      backgroundColor: theme.colors.primary,
    },
    modeToggleCircle: {
      width: 24,
      height: 24,
      borderRadius: 12,
      backgroundColor: theme.colors.text.inverse,
    },
    modeToggleCircleActive: {
      transform: [{ translateX: 20 }],
    },
  });
