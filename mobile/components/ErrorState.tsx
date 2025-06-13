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
            size={64}
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

const styles = StyleSheet.create({
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
    padding: 20,
  },
  errorText: {
    fontSize: 18,
    marginTop: 16,
    marginBottom: 32,
    textAlign: "center",
  },
  errorSubText: {
    fontSize: 14,
    marginBottom: 32,
    textAlign: "center",
    lineHeight: 20,
  },
  backButton: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 25,
  },
  backButtonText: {
    fontSize: 16,
    fontWeight: "600",
  },
});
