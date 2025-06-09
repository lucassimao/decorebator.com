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
      <TouchableOpacity style={styles.closeButton} onPress={onClose}>
        <Ionicons name="close" size={28} color={colors.textDark} />
      </TouchableOpacity>

      <View style={styles.headerCenter}>
        <Text style={styles.headerTitle}>{wordlistName}</Text>
        <Text style={styles.headerSubtitle}>
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
          <Text style={[styles.savePositionText, savePosition && styles.savePositionTextActive]}>
            {t("flashcards.savePosition")}
          </Text>
        </View>
        <Switch
          value={savePosition}
          onValueChange={onToggleSavePosition}
          thumbColor={colors.white}
          trackColor={{ false: colors.borderGray, true: colors.primary }}
          ios_backgroundColor={colors.borderGray}
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
    backgroundColor: colors.lightBackground,
    borderTopWidth: 1,
    borderTopColor: colors.borderGray,
    borderBottomWidth: 1,
    borderBottomColor: colors.borderGray,
  },
  savePositionTextContainer: {
    flexDirection: "row",
    alignItems: "center",
  },
  savePositionText: {
    fontSize: 15,
    color: colors.textMedium,
    fontWeight: "500",
    marginLeft: 8,
  },
  savePositionTextActive: {
    color: colors.primary,
  },
});
