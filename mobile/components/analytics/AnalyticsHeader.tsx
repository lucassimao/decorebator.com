import React from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive } from "@/hooks/useResponsive";

interface AnalyticsHeaderProps {
  onBackPress: () => void;
}

export const AnalyticsHeader: React.FC<AnalyticsHeaderProps> = ({
  onBackPress,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType } = useResponsive();
  const styles = createStyles(theme, isTablet, deviceType);

  return (
    <View style={styles.header}>
      <TouchableOpacity style={styles.backButton} onPress={onBackPress}>
        <Ionicons
          name="arrow-back"
          size={24}
          color={theme.colors.text.primary}
        />
      </TouchableOpacity>
      <View style={styles.headerTextContainer}>
        <Text style={styles.headerTitle}>{t("analytics.header.title")}</Text>
        <Text style={styles.headerSubtitle}>
          {t("analytics.header.subtitle")}
        </Text>
        <Text style={styles.headerNote}>
          {t("analytics.header.updateNote")}
        </Text>
      </View>
      <View style={styles.headerSpacer} />
    </View>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  deviceType: string,
) =>
  StyleSheet.create({
    header: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.sm,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.3)"
          : theme.colors.background.elevated,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.divider,
    },
    backButton: {
      width: theme.touchTargets.comfortable,
      height: theme.touchTargets.comfortable,
      borderRadius: theme.touchTargets.comfortable / 2,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    headerTextContainer: {
      flex: 1,
      marginLeft: theme.spacing.sm,
    },
    headerTitle: {
      fontSize: isTablet
        ? theme.typography.sizes.display
        : theme.typography.sizes.title,
      fontWeight: theme.typography.weights.bold,
      color: theme.colors.text.primary,
    },
    headerSubtitle: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
      marginTop: theme.spacing.xs,
    },
    headerNote: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.disabled,
      marginTop: theme.spacing.xs,
      fontStyle: "italic",
    },
    headerSpacer: {
      width: theme.touchTargets.comfortable,
    },
  });
