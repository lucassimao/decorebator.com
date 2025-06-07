import React from "react";
import { useTranslation } from "react-i18next";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";

interface QuizModeToggleProps {
  fastMode: boolean;
  onToggle: () => void;
}

export const QuizModeToggle: React.FC<QuizModeToggleProps> = ({
  fastMode,
  onToggle,
}) => {
  const { t } = useTranslation();

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

const styles = StyleSheet.create({
  modeContainer: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    marginBottom: 16,
    gap: 12,
  },
  modeText: {
    fontSize: 16,
    color: "#2D3436",
    fontWeight: "500",
  },
  modeToggle: {
    width: 50,
    height: 30,
    borderRadius: 15,
    backgroundColor: "#E0E0E0",
    padding: 3,
  },
  modeToggleActive: {
    backgroundColor: "#FF7B54",
  },
  modeToggleCircle: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: "#FFFFFF",
  },
  modeToggleCircleActive: {
    transform: [{ translateX: 20 }],
  },
});