import React from "react";
import {
  Animated,
  Dimensions,
  Image,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";

const { width } = Dimensions.get("window");

interface EmptyStateProps {
  fadeAnim: Animated.Value;
  slideAnim: Animated.Value;
  lockNudgeAnim: Animated.Value;
  lockTemporarilyUnlocked: boolean;
  onCreateWordlist: () => void;
  onTriggerLockNudge: () => void;
  onTriggerLockUnlock: () => void;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  fadeAnim,
  slideAnim,
  lockNudgeAnim,
  lockTemporarilyUnlocked,
  onCreateWordlist,
  onTriggerLockNudge,
  onTriggerLockUnlock,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

  return (
    <Animated.View
      style={[
        styles.emptyStateContainer,
        {
          opacity: fadeAnim,
          transform: [{ translateY: slideAnim }],
        },
      ]}
    >
      <View style={styles.emptyStateCard}>
        <Text style={styles.emptyStateTitle}>
          {t(
            "dashboard.wordlists.emptyTitle",
            "Your learning journey starts here",
          )}
        </Text>
        <Text style={styles.emptyStateSubtitle}>
          {t(
            "dashboard.wordlists.emptySubtitle",
            "Add a few words to unlock your progress",
          )}
        </Text>

        <Image
          source={require("../../assets/images/empty-dashboard-rocket.png")}
          style={styles.emptyStateIllustration}
          resizeMode="contain"
        />

        <View style={styles.lockedProgressBar}>
          <View style={styles.lockedProgressFill} />
          <View style={styles.lockedProgressKnob}>
            <Animated.View style={{ transform: [{ scale: lockNudgeAnim }] }}>
              <Ionicons
                name={lockTemporarilyUnlocked ? "lock-open" : "lock-closed"}
                size={15}
                color={theme.colors.text.secondary}
              />
            </Animated.View>
          </View>
        </View>

        <TouchableOpacity
          style={styles.ctaButton}
          onPress={() => {
            onTriggerLockNudge();
            onTriggerLockUnlock();
            onCreateWordlist();
          }}
          activeOpacity={0.8}
        >
          <Ionicons name="add" size={20} color={theme.colors.text.inverse} />
          <Text style={styles.ctaButtonText}>
            {t(
              "dashboard.wordlists.createFirstWordlist",
              "Add your first words",
            )}
          </Text>
        </TouchableOpacity>

        <View style={styles.emptyStateFootnote}>
          <Ionicons
            name="lock-closed"
            size={14}
            color={theme.colors.text.tertiary}
          />
          <Text style={styles.emptyStateFootnoteText}>
            {t(
              "dashboard.wordlists.emptyFootnote",
              "Progress unlocks after adding your first words",
            )}
          </Text>
        </View>
      </View>
    </Animated.View>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    emptyStateContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: 20,
    },
    emptyStateCard: {
      width: "100%",
      maxWidth: 420,
      backgroundColor: theme.colors.background.surface,
      borderRadius: 28,
      paddingHorizontal: 24,
      paddingVertical: 28,
      alignItems: "center",
      ...theme.shadows.lg,
    },
    emptyStateTitle: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
      textAlign: "center",
    },
    emptyStateSubtitle: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginTop: 8,
      marginBottom: 20,
    },
    emptyStateIllustration: {
      width: width * 0.62,
      height: width * 0.43,
      maxWidth: 240,
      maxHeight: 190,
      marginBottom: 2,
    },
    lockedProgressBar: {
      width: "100%",
      height: 12,
      borderRadius: 999,
      backgroundColor: "#EEEFF3",
      justifyContent: "center",
      marginBottom: 12,
    },
    lockedProgressFill: {
      position: "absolute",
      left: 6,
      right: 6,
      height: 6,
      borderRadius: 999,
      backgroundColor: "#E3E4E8",
    },
    lockedProgressKnob: {
      width: 30,
      height: 30,
      borderRadius: 15,
      backgroundColor: theme.colors.background.surface,
      borderWidth: 1,
      borderColor: "#E5E7EB",
      alignItems: "center",
      justifyContent: "center",
      alignSelf: "center",
      transform: [{ translateY: -8 }],
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.12,
      shadowRadius: 8,
      elevation: 6,
    },
    ctaButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: 999,
      paddingVertical: 14,
      paddingHorizontal: theme.spacing.lg,
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      gap: 8,
      width: "100%",
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 8 },
      shadowOpacity: 0.28,
      shadowRadius: 12,
      elevation: 10,
    },
    ctaButtonText: {
      color: theme.colors.text.inverse,
      fontSize: 16,
      fontWeight: "600",
    },
    emptyStateFootnote: {
      flexDirection: "row",
      alignItems: "center",
      gap: 6,
      marginTop: 16,
    },
    emptyStateFootnoteText: {
      fontSize: 13,
      color: theme.colors.text.tertiary,
    },
  });
