import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import React from "react";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";

interface QuizHeaderProps {
  wordlistName: string;
  correctCount: number;
  quizCount: number;
  isOnline: boolean;
  onBackPress: () => void;
  onReportPress: () => void;
}

export const QuizHeader: React.FC<QuizHeaderProps> = ({
  wordlistName,
  correctCount,
  quizCount,
  isOnline,
  onBackPress,
  onReportPress,
}) => {
  return (
    <View style={styles.header}>
      <TouchableOpacity 
        style={styles.backButton} 
        onPress={onBackPress}
        accessibilityRole="button"
        accessibilityLabel="Go back"
        accessibilityHint="Return to previous screen"
      >
        <Ionicons name="arrow-back" size={24} color="#2D3436" />
      </TouchableOpacity>

      <View style={styles.headerCenter}>
        <Text 
          style={styles.headerTitle}
          accessibilityRole="header"
          accessibilityLevel={1}
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

      <TouchableOpacity
        style={[styles.settingsButton, !isOnline && styles.disabledButton]}
        onPress={() => isOnline && onReportPress()}
        disabled={!isOnline}
        accessibilityRole="button"
        accessibilityLabel={isOnline ? "Report error" : "Report error (offline)"}
        accessibilityHint={isOnline ? "Report an issue with the current question" : "Requires internet connection"}
        accessibilityState={{ disabled: !isOnline }}
      >
        <MaterialIcons
          name="flag"
          size={24}
          color={isOnline ? "#636E72" : "#DFE6E9"}
        />
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
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
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    justifyContent: "center",
    alignItems: "center",
  },
  headerCenter: {
    flex: 1,
    alignItems: "center",
    marginHorizontal: 16,
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: "600",
    color: "#2D3436",
  },
  headerSubtitle: {
    fontSize: 14,
    color: "#636E72",
    marginTop: 2,
  },
  settingsButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    justifyContent: "center",
    alignItems: "center",
  },
  disabledButton: {
    opacity: 0.5,
  },
});
