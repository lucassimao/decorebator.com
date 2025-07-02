import React from "react";
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
} from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

interface ErrorStateProps {
  type: "offline" | "noData" | "general";
  title: string;
  subtitle?: string;
  onGoBack: () => void;
  colors?: {
    backgroundLight: string;
    backgroundPeach: string;
    backgroundSage: string;
    textMedium: string;
    error: string;
    primary: string;
    white: string;
  };
}

export const ErrorState: React.FC<ErrorStateProps> = ({
  type,
  title,
  subtitle,
  onGoBack,
  colors = {
    backgroundLight: "#FFF9F0",
    backgroundPeach: "#FFE8D6",
    backgroundSage: "#F5F0E6",
    textMedium: "#636E72",
    error: "#FF6B6B",
    primary: "#FF7B54",
    white: "#FFFFFF",
  },
}) => {
  const { t } = useTranslation();
  const { isTablet } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(isTablet, spacing);

  const getIconName = () => {
    switch (type) {
      case "offline":
        return "cloud-off";
      case "noData":
        return "error-outline";
      default:
        return "error-outline";
    }
  };

  const getIconColor = () => {
    switch (type) {
      case "offline":
        return colors.textMedium;
      case "noData":
        return colors.error;
      default:
        return colors.error;
    }
  };

  return (
    <LinearGradient
      colors={[
        colors.backgroundLight,
        colors.backgroundPeach,
        colors.backgroundSage,
      ]}
      style={styles.container}
    >
      <SafeAreaView style={styles.safeArea}>
        <View style={styles.errorContainer}>
          <MaterialIcons
            name={getIconName() as any}
            size={isTablet ? 80 : 64}
            color={getIconColor()}
          />
          <Text style={[styles.errorText, { color: colors.textMedium }]}>
            {title}
          </Text>
          {subtitle && (
            <Text style={[styles.errorSubText, { color: colors.textMedium }]}>
              {subtitle}
            </Text>
          )}
          <TouchableOpacity
            style={[styles.backButton, { backgroundColor: colors.primary }]}
            onPress={onGoBack}
          >
            <Text style={[styles.backButtonText, { color: colors.white }]}>
              {t("common.goBack")}
            </Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    </LinearGradient>
  );
};

const createStyles = (
  isTablet: boolean,
  spacing: ReturnType<typeof useResponsiveSpacing>,
) =>
  StyleSheet.create({
    container: {
      flex: 1,
    },
    safeArea: {
      flex: 1,
    },
    errorContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      padding: spacing.lg,
    },
    errorText: {
      fontSize: isTablet ? 22 : 18,
      marginTop: spacing.md,
      marginBottom: spacing.xl,
      textAlign: "center",
    },
    errorSubText: {
      fontSize: isTablet ? 16 : 14,
      marginBottom: spacing.xl,
      textAlign: "center",
      lineHeight: isTablet ? 24 : 20,
    },
    backButton: {
      paddingHorizontal: spacing.xl,
      paddingVertical: spacing.md,
      borderRadius: 25,
      minHeight: isTablet ? 48 : 44,
      justifyContent: "center",
    },
    backButtonText: {
      fontSize: isTablet ? 18 : 16,
      fontWeight: "600",
    },
  });
