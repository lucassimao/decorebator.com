import React from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";

interface AnalyticsHeaderProps {
  onBackPress: () => void;
}

export const AnalyticsHeader: React.FC<AnalyticsHeaderProps> = ({
  onBackPress,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

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

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    header: {
      flexDirection: "row",
      alignItems: "center",
      paddingHorizontal: 16,
      paddingVertical: 12,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.3)"
          : theme.colors.background.elevated,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.divider,
    },
    backButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      shadowColor: theme.colors.text.primary,
      shadowOffset: {
        width: 0,
        height: 2,
      },
      shadowOpacity: 0.1,
      shadowRadius: 3.84,
      elevation: 3,
    },
    headerTextContainer: {
      flex: 1,
      marginLeft: 12,
    },
    headerTitle: {
      fontSize: 24,
      fontWeight: "bold",
      color: theme.colors.text.primary,
    },
    headerSubtitle: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginTop: 2,
    },
    headerNote: {
      fontSize: 12,
      color: theme.colors.text.disabled,
      marginTop: 2,
      fontStyle: "italic",
    },
    headerSpacer: {
      width: 40,
    },
  });
