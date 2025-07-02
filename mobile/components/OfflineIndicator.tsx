import React, { useEffect, useRef } from "react";
import { Text, StyleSheet, Animated } from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useOffline } from "@/hooks/useOffline";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

export const OfflineIndicator: React.FC = () => {
  const { t } = useTranslation();
  const { isOnline, isOfflineAvailable } = useOffline();
  const { theme } = useTheme();
  const { isTablet } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, spacing);
  const slideAnim = useRef(new Animated.Value(isTablet ? -80 : -60)).current;

  useEffect(() => {
    if (!isOnline) {
      // Slide in
      Animated.timing(slideAnim, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }).start();
    } else {
      // Slide out
      Animated.timing(slideAnim, {
        toValue: isTablet ? -80 : -60,
        duration: 300,
        useNativeDriver: true,
      }).start();
    }
  }, [isOnline, slideAnim, isTablet]);

  if (isOnline) return null;

  return (
    <Animated.View
      style={[
        styles.container,
        {
          transform: [{ translateY: slideAnim }],
          backgroundColor: isOfflineAvailable
            ? theme.colors.primary
            : theme.colors.text.secondary,
        },
      ]}
    >
      <MaterialIcons
        name={isOfflineAvailable ? "offline-bolt" : "wifi-off"}
        size={isTablet ? 24 : 20}
        color={theme.colors.text.inverse}
      />
      <Text style={styles.text}>
        {isOfflineAvailable
          ? t("offline.availableMessage")
          : t("offline.unavailableMessage")}
      </Text>
    </Animated.View>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  spacing: ReturnType<typeof useResponsiveSpacing>,
) =>
  StyleSheet.create({
    container: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      paddingVertical: isTablet ? spacing.sm : spacing.xs,
      paddingHorizontal: spacing.md,
      zIndex: 1000,
    },
    text: {
      color: theme.colors.text.inverse,
      fontSize: isTablet ? 16 : 14,
      fontWeight: "500",
      marginLeft: spacing.xs,
    },
  });
