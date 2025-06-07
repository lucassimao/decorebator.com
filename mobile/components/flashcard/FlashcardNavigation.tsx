import React from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";

interface FlashcardNavigationProps {
  currentIndex: number;
  totalWords: number;
  onNavigate: (direction: "next" | "prev") => void;
}

const colors = {
  textDark: "#2D3436",
  textMedium: "#636E72",
  borderGray: "#E0E0E0",
  white: "#FFFFFF",
};

export const FlashcardNavigation: React.FC<FlashcardNavigationProps> = ({
  currentIndex,
  totalWords,
  onNavigate,
}) => {
  const { t } = useTranslation();

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
          size={32}
          color={isFirstCard ? colors.borderGray : colors.textDark}
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
          size={32}
          color={isLastCard ? colors.borderGray : colors.textDark}
        />
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  navigationContainer: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 20,
    paddingVertical: 20,
  },
  navButton: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: colors.white,
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 3,
  },
  navButtonDisabled: {
    opacity: 0.5,
  },
  centerControls: {
    flex: 1,
    alignItems: "center",
  },
  swipeHint: {
    fontSize: 14,
    color: colors.textMedium,
    fontStyle: "italic",
  },
});
