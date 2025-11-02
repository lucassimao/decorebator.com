import React from "react";
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  StyleProp,
  ViewStyle,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { LinearGradient } from "expo-linear-gradient";
import { StatusBar } from "expo-status-bar";
import { useTheme } from "@/contexts/ThemeContext";

type OnboardingLayoutProps = {
  step: number;
  totalSteps: number;
  children: React.ReactNode;
  contentStyle?: StyleProp<ViewStyle>;
  footer?: React.ReactNode;
  showSkip?: boolean;
  skipLabel?: string;
  onSkip?: () => void;
  showBack?: boolean;
  backLabel?: string;
  onBack?: () => void;
  stepLabel?: string;
};

export const OnboardingLayout: React.FC<OnboardingLayoutProps> = ({
  step,
  totalSteps,
  children,
  contentStyle,
  footer,
  showSkip = false,
  skipLabel,
  onSkip,
  showBack = false,
  backLabel,
  onBack,
  stepLabel,
}) => {
  const { theme } = useTheme();
  const gradients = {
    light: ["#FFF4E7", "#FFF8F2", "#FFEDE1"] as const,
    dark: ["#1C1A20", "#1E1C24", "#15141B"] as const,
  };
  const gradientColors =
    theme.mode === "dark" ? gradients.dark : gradients.light;

  return (
    <LinearGradient colors={gradientColors} style={styles.gradient}>
      <StatusBar style={theme.mode === "dark" ? "light" : "dark"} />
      <SafeAreaView style={styles.safeArea} edges={["top", "bottom"]}>
        <View style={styles.header}>
          {showBack ? (
            <TouchableOpacity
              onPress={onBack}
              style={styles.headerButton}
              accessibilityRole="button"
            >
              <Text
                style={[
                  styles.headerText,
                  { color: theme.colors.text.secondary },
                ]}
              >
                {backLabel}
              </Text>
            </TouchableOpacity>
          ) : (
            <View style={styles.headerSpacer} />
          )}
          <View style={styles.progressContainer}>
            <View style={styles.progressSteps}>
              {Array.from({ length: totalSteps }).map((_, index) => (
                <View
                  key={index}
                  style={[
                    styles.progressStep,
                    {
                      backgroundColor:
                        index < step
                          ? theme.colors.primary
                          : theme.mode === "dark"
                            ? "rgba(255, 255, 255, 0.12)"
                            : "rgba(255, 123, 84, 0.18)",
                    },
                  ]}
                />
              ))}
            </View>
            {stepLabel ? (
              <Text
                style={[
                  styles.stepLabel,
                  { color: theme.colors.text.secondary },
                ]}
              >
                {stepLabel}
              </Text>
            ) : null}
          </View>
          {showSkip ? (
            <TouchableOpacity
              onPress={onSkip}
              style={styles.headerButton}
              accessibilityRole="button"
            >
              <Text
                style={[
                  styles.headerText,
                  { color: theme.colors.text.secondary },
                ]}
              >
                {skipLabel}
              </Text>
            </TouchableOpacity>
          ) : (
            <View style={styles.headerSpacer} />
          )}
        </View>

        <View style={[styles.content, contentStyle]}>{children}</View>
        {footer ? <View style={styles.footer}>{footer}</View> : null}
      </SafeAreaView>
    </LinearGradient>
  );
};

const styles = StyleSheet.create({
  gradient: {
    flex: 1,
  },
  safeArea: {
    flex: 1,
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 24,
    paddingTop: 8,
    paddingBottom: 12,
  },
  headerButton: {
    paddingHorizontal: 8,
    paddingVertical: 8,
  },
  headerText: {
    fontSize: 14,
    fontWeight: "600",
  },
  headerSpacer: {
    width: 64,
  },
  progressContainer: {
    alignItems: "center",
    flex: 1,
  },
  progressSteps: {
    flexDirection: "row",
    gap: 6,
    marginBottom: 4,
  },
  progressStep: {
    width: 36,
    height: 4,
    borderRadius: 4,
  },
  stepLabel: {
    fontSize: 12,
    fontWeight: "600",
    textTransform: "uppercase",
    letterSpacing: 0.8,
  },
  content: {
    flex: 1,
    paddingHorizontal: 24,
  },
  footer: {
    paddingHorizontal: 24,
    paddingBottom: 28,
  },
});
