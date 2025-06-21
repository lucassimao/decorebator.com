import React from "react";
import { View, Text, TouchableOpacity, StyleSheet, Switch } from "react-native";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";

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

const colors = {
  textDark: "#2D3436",
  textMedium: "#636E72",
  borderGray: "#E0E0E0",
  primary: "#FF7B54",
  lightBackground: "#FFF9F0",
  white: "#FFFFFF",
};

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

  return (
    <View>
      <View style={styles.header}>
        <TouchableOpacity 
          style={styles.closeButton} 
          onPress={onClose}
          accessibilityRole="button"
          accessibilityLabel="Close flashcards"
          accessibilityHint="Return to previous screen"
        >
          <Ionicons name="close" size={28} color={colors.textDark} />
        </TouchableOpacity>

        <View style={styles.headerCenter}>
          <Text 
            style={styles.headerTitle}
            accessibilityRole="header"
            accessibilityLevel={1}
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
          accessibilityLabel={isOnline ? "Report error" : "Report error (offline)"}
          accessibilityHint={isOnline ? "Report an issue with the current flashcard" : "Requires internet connection"}
          accessibilityState={{ disabled: !isOnline }}
        >
          <MaterialIcons
            name="flag"
            size={24}
            color={isOnline ? colors.textMedium : colors.borderGray}
          />
        </TouchableOpacity>
      </View>

      {onToggleSavePosition && (
        <View style={styles.savePositionContainer}>
          <View style={styles.savePositionTextContainer}>
            <MaterialIcons
              name="bookmark"
              size={20}
              color={savePosition ? colors.primary : colors.textMedium}
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
            thumbColor={colors.white}
            trackColor={{ false: colors.borderGray, true: colors.primary }}
            ios_backgroundColor={colors.borderGray}
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

const styles = StyleSheet.create({
  header: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  closeButton: {
    width: 44,
    height: 44,
    justifyContent: "center",
    alignItems: "center",
  },
  headerCenter: {
    flex: 1,
    alignItems: "center",
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: "600",
    color: colors.textDark,
  },
  headerSubtitle: {
    fontSize: 14,
    color: colors.textMedium,
    marginTop: 2,
  },
  reportButton: {
    width: 44,
    height: 44,
    justifyContent: "center",
    alignItems: "center",
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
    color: colors.textDark,
    fontWeight: "600",
    marginLeft: 8,
    textShadowColor: "rgba(255, 255, 255, 0.8)",
    textShadowOffset: { width: 0, height: 1 },
    textShadowRadius: 2,
  },
  savePositionTextActive: {
    color: colors.primary,
    textShadowColor: "rgba(255, 255, 255, 0.8)",
    textShadowOffset: { width: 0, height: 1 },
    textShadowRadius: 2,
  },
});
