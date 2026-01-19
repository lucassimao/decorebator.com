import React, { useEffect, useRef } from "react";
import {
  Animated,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";

type Props = {
  visible: boolean;
  onLoveIt: () => void;
  onNotReally: () => void;
  onMaybeLater: () => void;
  onClose: () => void;
};

export default function AppReviewModal({
  visible,
  onLoveIt,
  onNotReally,
  onMaybeLater,
  onClose,
}: Props) {
  const { theme, responsive } = useTheme();
  const { t } = useTranslation();

  const opacity = useRef(new Animated.Value(0)).current;
  const scale = useRef(new Animated.Value(0.96)).current;

  useEffect(() => {
    if (visible) {
      Animated.parallel([
        Animated.timing(opacity, {
          toValue: 1,
          duration: 160,
          useNativeDriver: true,
        }),
        Animated.spring(scale, {
          toValue: 1,
          friction: 10,
          tension: 70,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      Animated.parallel([
        Animated.timing(opacity, {
          toValue: 0,
          duration: 140,
          useNativeDriver: true,
        }),
        Animated.timing(scale, {
          toValue: 0.96,
          duration: 140,
          useNativeDriver: true,
        }),
      ]).start();
    }
  }, [visible, opacity, scale]);

  const styles = createStyles(theme, responsive);

  return (
    <Modal
      visible={visible}
      transparent
      animationType="none"
      onRequestClose={onClose}
    >
      <TouchableWithoutFeedback onPress={onClose}>
        <Animated.View style={[styles.overlay, { opacity }]}>
          <TouchableWithoutFeedback>
            <Animated.View style={[styles.card, { transform: [{ scale }] }]}>
              <View style={styles.header}>
                <View style={styles.iconBadge}>
                  <MaterialIcons
                    name="star"
                    size={28}
                    color={theme.colors.text.inverse}
                  />
                </View>
                <Text style={styles.title}>{t("review.title")}</Text>
                <Text style={styles.subtitle}>{t("review.subtitle")}</Text>
              </View>

              <View style={styles.actions}>
                <TouchableOpacity
                  style={styles.primaryButton}
                  onPress={onLoveIt}
                  accessibilityLabel={t("review.loveIt")}
                >
                  <MaterialIcons
                    name="favorite"
                    size={18}
                    color={theme.colors.text.inverse}
                    style={styles.buttonIcon}
                  />
                  <Text style={styles.primaryButtonText}>
                    {t("review.loveIt")}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  style={styles.secondaryButton}
                  onPress={onNotReally}
                  accessibilityLabel={t("review.notReally")}
                >
                  <Text style={styles.secondaryButtonText}>
                    {t("review.notReally")}
                  </Text>
                </TouchableOpacity>
              </View>

              <TouchableOpacity
                style={styles.maybeLaterButton}
                onPress={onMaybeLater}
                accessibilityLabel={t("review.maybeLater")}
              >
                <Text style={styles.maybeLaterText}>
                  {t("review.maybeLater")}
                </Text>
              </TouchableOpacity>
            </Animated.View>
          </TouchableWithoutFeedback>
        </Animated.View>
      </TouchableWithoutFeedback>
    </Modal>
  );
}

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) =>
  StyleSheet.create({
    overlay: {
      flex: 1,
      backgroundColor:
        theme.mode === "light" ? "rgba(0,0,0,0.5)" : "rgba(0,0,0,0.7)",
      justifyContent: "center",
      alignItems: "center",
      padding: responsive.spacing.horizontal,
    },
    card: {
      width: "100%",
      maxWidth: responsive.getValueForSize(320, 360, 380, 400),
      borderRadius: theme.borderRadius.xl,
      backgroundColor: theme.colors.background.surface,
      padding: responsive.spacing.formPadding,
      ...theme.shadows.lg,
    },
    header: {
      alignItems: "center",
    },
    iconBadge: {
      width: 56,
      height: 56,
      borderRadius: 28,
      backgroundColor: theme.colors.premium,
      alignItems: "center",
      justifyContent: "center",
      marginBottom: 12,
    },
    title: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
    },
    subtitle: {
      marginTop: 8,
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: responsive.getScaledFont("body") * 1.4,
      paddingHorizontal: 8,
    },
    actions: {
      marginTop: responsive.spacing.elementSpacing * 1.5,
      gap: 12,
    },
    primaryButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      borderRadius: 12,
      paddingVertical: 14,
      backgroundColor: theme.colors.success,
    },
    buttonIcon: {
      marginRight: 8,
    },
    primaryButtonText: {
      color: theme.colors.text.inverse,
      fontWeight: "700",
      fontSize: responsive.getScaledFont("body"),
    },
    secondaryButton: {
      borderRadius: 12,
      paddingVertical: 14,
      backgroundColor: theme.colors.background.default,
      alignItems: "center",
    },
    secondaryButtonText: {
      color: theme.colors.text.secondary,
      fontWeight: "600",
      fontSize: responsive.getScaledFont("body"),
    },
    maybeLaterButton: {
      marginTop: 16,
      alignItems: "center",
      paddingVertical: 8,
    },
    maybeLaterText: {
      color: theme.colors.text.tertiary,
      fontSize: responsive.getScaledFont("label"),
    },
  });
