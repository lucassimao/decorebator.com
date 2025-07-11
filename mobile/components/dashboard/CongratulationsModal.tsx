import React from "react";
import {
  Animated,
  Dimensions,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { Wordlist } from "@/api/wordlists";

const { width: SCREEN_WIDTH } = Dimensions.get("window");

interface CongratulationsModalProps {
  visible: boolean;
  onClose: () => void;
  onAddWords: () => void;
  wordlist: Wordlist | null;
}

export const CongratulationsModal: React.FC<CongratulationsModalProps> = ({
  visible,
  onClose,
  onAddWords,
  wordlist,
}) => {
  const { theme } = useTheme();
  const { t } = useTranslation();
  const styles = createStyles(theme);

  const backdropAnim = React.useRef(new Animated.Value(0)).current;
  const modalAnim = React.useRef(new Animated.Value(0)).current;

  React.useEffect(() => {
    if (visible) {
      Animated.parallel([
        Animated.timing(backdropAnim, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.spring(modalAnim, {
          toValue: 1,
          useNativeDriver: true,
          tension: 50,
          friction: 8,
        }),
      ]).start();
    } else {
      Animated.parallel([
        Animated.timing(backdropAnim, {
          toValue: 0,
          duration: 250,
          useNativeDriver: true,
        }),
        Animated.timing(modalAnim, {
          toValue: 0,
          duration: 250,
          useNativeDriver: true,
        }),
      ]).start();
    }
  }, [visible, backdropAnim, modalAnim]);

  const handleAddWords = () => {
    onAddWords();
    onClose();
  };

  if (!visible || !wordlist) return null;

  return (
    <Modal
      transparent
      visible={visible}
      animationType="none"
      onRequestClose={onClose}
    >
      <TouchableWithoutFeedback onPress={onClose}>
        <Animated.View style={[styles.backdrop, { opacity: backdropAnim }]} />
      </TouchableWithoutFeedback>

      <View style={styles.container}>
        <Animated.View
          style={[
            styles.modalContent,
            {
              transform: [
                {
                  scale: modalAnim.interpolate({
                    inputRange: [0, 1],
                    outputRange: [0.8, 1],
                  }),
                },
              ],
              opacity: modalAnim,
            },
          ]}
        >
          {/* Celebration Icon */}
          <View style={styles.iconContainer}>
            <Text style={styles.celebrationIcon}>🎉</Text>
          </View>

          {/* Title */}
          <Text style={styles.title}>
            {t("welcome.firstWordlistSuccess", "Congratulations! 🎉")}
          </Text>

          {/* Message */}
          <Text style={styles.message}>
            {t(
              "welcome.firstWordlistMessage",
              "You've created your first wordlist! Now you can start adding words and begin learning.",
            )}
          </Text>

          {/* Wordlist Info */}
          <View style={styles.wordlistInfo}>
            <Text style={styles.wordlistName}>"{wordlist.name}"</Text>
            <Text style={styles.wordlistLanguage}>
              {wordlist.languageCode.toUpperCase()}
            </Text>
          </View>

          {/* Action Buttons */}
          <View style={styles.buttonContainer}>
            <TouchableOpacity
              style={styles.primaryButton}
              onPress={handleAddWords}
              activeOpacity={0.8}
            >
              <Ionicons
                name="add-circle"
                size={20}
                color={theme.colors.text.inverse}
                style={styles.buttonIcon}
              />
              <Text style={styles.primaryButtonText}>
                {t("welcome.addWordsNow", "Add Words Now")}
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.secondaryButton}
              onPress={onClose}
              activeOpacity={0.8}
            >
              <Text style={styles.secondaryButtonText}>
                {t("welcome.addWordsLater", "I'll do this later")}
              </Text>
            </TouchableOpacity>
          </View>
        </Animated.View>
      </View>
    </Modal>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    container: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: 24,
    },
    backdrop: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor:
        theme.mode === "light" ? "rgba(0, 0, 0, 0.5)" : "rgba(0, 0, 0, 0.7)",
    },
    modalContent: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: 24,
      padding: 32,
      width: Math.min(SCREEN_WIDTH - 48, 400),
      maxWidth: "100%",
      alignItems: "center",
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: 8 },
      shadowOpacity: 0.15,
      shadowRadius: 24,
      elevation: 12,
    },
    iconContainer: {
      width: 80,
      height: 80,
      borderRadius: 40,
      backgroundColor: theme.colors.background.secondary,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: 24,
    },
    celebrationIcon: {
      fontSize: 40,
    },
    title: {
      fontSize: 24,
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: 12,
    },
    message: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: 24,
      marginBottom: 24,
    },
    wordlistInfo: {
      backgroundColor: theme.colors.background.secondary,
      borderRadius: 16,
      padding: 16,
      alignItems: "center",
      marginBottom: 32,
      minWidth: 160,
    },
    wordlistName: {
      fontSize: 18,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 4,
      textAlign: "center",
    },
    wordlistLanguage: {
      fontSize: 14,
      color: theme.colors.primary,
      fontWeight: "500",
    },
    buttonContainer: {
      width: "100%",
      gap: 12,
    },
    primaryButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: 16,
      paddingVertical: 16,
      paddingHorizontal: 24,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.3,
      shadowRadius: 8,
      elevation: 4,
    },
    primaryButtonText: {
      color: theme.colors.text.inverse,
      fontSize: 16,
      fontWeight: "600",
    },
    buttonIcon: {
      marginRight: 8,
    },
    secondaryButton: {
      backgroundColor: "transparent",
      borderRadius: 16,
      paddingVertical: 16,
      paddingHorizontal: 24,
      alignItems: "center",
      justifyContent: "center",
      borderWidth: 1,
      borderColor: theme.colors.border.medium,
    },
    secondaryButtonText: {
      color: theme.colors.text.secondary,
      fontSize: 16,
      fontWeight: "500",
    },
  });
