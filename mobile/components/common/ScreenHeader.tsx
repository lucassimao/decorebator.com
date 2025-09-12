import React from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTheme } from "@/contexts/ThemeContext";

type ScreenHeaderProps = {
  title: string;
  subtitle?: string;
  onBackPress: () => void;
  right?: React.ReactNode;
  rightAccessibilityLabel?: string;
  rightAccessibilityHint?: string;
};

export const ScreenHeader: React.FC<ScreenHeaderProps> = ({
  title,
  subtitle,
  onBackPress,
  right,
  rightAccessibilityLabel,
  rightAccessibilityHint,
}) => {
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

  return (
    <View style={styles.header}>
      <TouchableOpacity
        style={styles.iconButton}
        onPress={onBackPress}
        accessibilityRole="button"
        accessibilityLabel="Go back"
        accessibilityHint="Return to previous screen"
        hitSlop={{ top: 8, left: 8, right: 8, bottom: 8 }}
      >
        <Ionicons
          name="arrow-back"
          size={24}
          color={theme.colors.text.primary}
        />
      </TouchableOpacity>

      <View style={styles.center}>
        <Text style={styles.title} accessibilityRole="header">
          {title}
        </Text>
        {subtitle ? <Text style={styles.subtitle}>{subtitle}</Text> : null}
      </View>

      <View style={styles.rightSlot}>
        {right ? (
          <View
            accessibilityLabel={rightAccessibilityLabel}
            accessibilityHint={rightAccessibilityHint}
            style={styles.iconButton}
          >
            {right}
          </View>
        ) : (
          <View style={{ width: 40, height: 40 }} />
        )}
      </View>
    </View>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) =>
  StyleSheet.create({
    header: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingHorizontal: 16,
      paddingVertical: 12,
    },
    iconButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    center: {
      flex: 1,
      alignItems: "center",
    },
    title: {
      fontSize: responsive.getScaledFont("title"),
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    subtitle: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      marginTop: 2,
      textAlign: "center",
    },
    rightSlot: {
      justifyContent: "center",
      alignItems: "center",
    },
  });

export default ScreenHeader;
