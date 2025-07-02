import React, { useEffect, useRef, useCallback } from "react";
import {
  Animated,
  Platform,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

export type SnackBarProps = {
  message: string;
  type: "success" | "error";
  onDismiss: () => void;
  duration?: number;
  visible: boolean;
};

export default function SnackBar({
  message,
  type,
  onDismiss,
  duration = 2000,
  visible,
}: SnackBarProps) {
  const translateY = useRef(new Animated.Value(100)).current;
  const opacity = useRef(new Animated.Value(0)).current;
  const timeoutRef = useRef<number>(0);
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, spacing);

  const hide = useCallback(() => {
    Animated.parallel([
      Animated.timing(translateY, {
        toValue: 100,
        duration: 200,
        useNativeDriver: true,
      }),
      Animated.timing(opacity, {
        toValue: 0,
        duration: 200,
        useNativeDriver: true,
      }),
    ]).start(() => {
      onDismiss();
    });
  }, [onDismiss, opacity, translateY]);

  useEffect(() => {
    if (visible) {
      // Show animation
      Animated.parallel([
        Animated.timing(translateY, {
          toValue: 0,
          duration: 250,
          useNativeDriver: true,
        }),
        Animated.timing(opacity, {
          toValue: 1,
          duration: 250,
          useNativeDriver: true,
        }),
      ]).start();

      // Auto dismiss after duration
      if (duration > 0) {
        timeoutRef.current = setTimeout(() => {
          hide();
        }, duration);
      }
    } else {
      hide();
    }

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [visible, duration, hide, opacity, translateY]);

  if (!visible) return null;

  const backgroundColor =
    type === "error" ? theme.colors.error : theme.colors.success;

  return (
    <Animated.View
      style={[
        styles.container,
        {
          transform: [{ translateY }],
          opacity,
        },
      ]}
      pointerEvents="box-none"
    >
      <View style={[styles.snackbar, { backgroundColor }]}>
        <Text style={styles.message} numberOfLines={2}>
          {message}
        </Text>
        <TouchableOpacity
          style={styles.actionButton}
          onPress={hide}
          hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
        >
          <Text style={styles.actionText}>{t("common.hide")}</Text>
        </TouchableOpacity>
      </View>
    </Animated.View>
  );
}

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  spacing: ReturnType<typeof useResponsiveSpacing>,
) =>
  StyleSheet.create({
    container: {
      position: "absolute",
      bottom: 0,
      left: 0,
      right: 0,
      paddingHorizontal: spacing.md,
      paddingBottom: Platform.OS === "ios" ? (isTablet ? 40 : 30) : spacing.md,
      zIndex: 9999,
    },
    snackbar: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: isTablet ? spacing.md : spacing.sm,
      paddingHorizontal: spacing.md,
      borderRadius: theme.borderRadius.md,
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.25,
      shadowRadius: 4,
      elevation: 6,
      minHeight: isTablet ? 56 : 48,
    },
    message: {
      flex: 1,
      color: theme.colors.text.inverse,
      fontSize: isTablet ? 16 : 14,
      lineHeight: isTablet ? 24 : 20,
      marginRight: spacing.md,
    },
    actionButton: {
      paddingHorizontal: spacing.xs,
      paddingVertical: spacing.xs,
    },
    actionText: {
      color: theme.colors.text.inverse,
      fontSize: isTablet ? 16 : 14,
      fontWeight: "600",
      letterSpacing: 0.5,
    },
  });
