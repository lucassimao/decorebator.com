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
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

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
        timeoutRef.current = null;
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

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    container: {
      position: "absolute",
      bottom: 0,
      left: 0,
      right: 0,
      paddingHorizontal: 16,
      paddingBottom: Platform.OS === "ios" ? 30 : 16,
      zIndex: 9999,
    },
    snackbar: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between",
      paddingVertical: 14,
      paddingHorizontal: 16,
      borderRadius: 8,
      shadowColor: theme.colors.text.primary,
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.25,
      shadowRadius: 4,
      elevation: 6,
      minHeight: 48,
    },
    message: {
      flex: 1,
      color: theme.colors.text.inverse,
      fontSize: 14,
      lineHeight: 20,
      marginRight: 16,
    },
    actionButton: {
      paddingHorizontal: 8,
      paddingVertical: 4,
    },
    actionText: {
      color: theme.colors.text.inverse,
      fontSize: 14,
      fontWeight: "600",
      letterSpacing: 0.5,
    },
  });
