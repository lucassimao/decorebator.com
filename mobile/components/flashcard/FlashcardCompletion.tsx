import React, { useEffect, useRef } from "react";
import {
  AccessibilityInfo,
  findNodeHandle,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import Animated, {
  useAnimatedStyle,
  useReducedMotion,
  useSharedValue,
  withDelay,
  withSpring,
  withTiming,
} from "react-native-reanimated";

import { Button } from "@/components/ui/button";
import { UiText } from "@/components/ui/text";
import { useTheme } from "@/contexts/ThemeContext";

interface FlashcardCompletionProps {
  wordlistName: string;
  viewedCount: number;
  shouldAnimate: boolean;
  onReviewAgain: () => void;
  onBackToWordlists: () => void;
  onReturnToLastCard: () => void;
}

export const FlashcardCompletion: React.FC<FlashcardCompletionProps> = ({
  wordlistName,
  viewedCount,
  shouldAnimate,
  onReviewAgain,
  onBackToWordlists,
  onReturnToLastCard,
}) => {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const reduceMotion = useReducedMotion();
  const styles = createStyles(theme);
  const headingRef = useRef<Text>(null);
  const backCardProgress = useSharedValue(
    shouldAnimate && !reduceMotion ? 0 : 1,
  );
  const frontCardProgress = useSharedValue(
    shouldAnimate && !reduceMotion ? 0 : 1,
  );
  const copyProgress = useSharedValue(shouldAnimate && !reduceMotion ? 0 : 1);

  useEffect(() => {
    const focusTimer = setTimeout(() => {
      const node = findNodeHandle(headingRef.current);
      if (node) AccessibilityInfo.setAccessibilityFocus(node);
    }, 0);

    if (!shouldAnimate || reduceMotion) {
      backCardProgress.value = 1;
      frontCardProgress.value = 1;
      copyProgress.value = 1;
    } else {
      backCardProgress.value = withTiming(1, { duration: 420 });
      frontCardProgress.value = withDelay(
        80,
        withSpring(1, { damping: 14, stiffness: 180 }),
      );
      copyProgress.value = withDelay(210, withTiming(1, { duration: 260 }));
    }

    return () => clearTimeout(focusTimer);
  }, [
    backCardProgress,
    copyProgress,
    frontCardProgress,
    reduceMotion,
    shouldAnimate,
  ]);

  const leftCardStyle = useAnimatedStyle(() => ({
    opacity: backCardProgress.value,
    transform: [
      { translateY: (1 - backCardProgress.value) * 32 },
      { rotate: `${-10 * backCardProgress.value}deg` },
    ],
  }));
  const rightCardStyle = useAnimatedStyle(() => ({
    opacity: backCardProgress.value,
    transform: [
      { translateY: (1 - backCardProgress.value) * 32 },
      { rotate: `${9 * backCardProgress.value}deg` },
    ],
  }));
  const frontCardStyle = useAnimatedStyle(() => ({
    opacity: frontCardProgress.value,
    transform: [
      { translateY: (1 - frontCardProgress.value) * 38 },
      { scale: 0.84 + frontCardProgress.value * 0.16 },
    ],
  }));
  const copyStyle = useAnimatedStyle(() => ({
    opacity: copyProgress.value,
    transform: [{ translateY: (1 - copyProgress.value) * 12 }],
  }));

  return (
    <ScrollView
      style={styles.scrollView}
      contentContainerStyle={styles.contentContainer}
      showsVerticalScrollIndicator={false}
    >
      <View style={styles.closeRow}>
        <TouchableOpacity
          style={styles.closeButton}
          onPress={onBackToWordlists}
          accessibilityRole="button"
          accessibilityLabel={t("flashcards.backToWordlists")}
        >
          <Ionicons name="close" size={24} color={theme.colors.text.primary} />
        </TouchableOpacity>
      </View>

      <View
        style={styles.deck}
        accessible={false}
        accessibilityElementsHidden
        importantForAccessibility="no-hide-descendants"
      >
        <Animated.View
          style={[styles.deckCard, styles.leftCard, leftCardStyle]}
        />
        <Animated.View
          style={[styles.deckCard, styles.rightCard, rightCardStyle]}
        />
        <Animated.View
          style={[styles.deckCard, styles.frontCard, frontCardStyle]}
        >
          <View style={styles.bookmark} />
        </Animated.View>
      </View>

      <Animated.View style={[styles.copy, copyStyle]}>
        {wordlistName ? (
          <UiText variant="label" tone="action" style={styles.wordlistName}>
            {wordlistName}
          </UiText>
        ) : null}
        <UiText
          ref={headingRef}
          variant="feature"
          accessibilityRole="header"
          style={styles.heading}
        >
          {t("flashcards.wordlistComplete")}
        </UiText>
        <UiText tone="secondary" style={styles.message}>
          {t("flashcards.completionMessage")}
        </UiText>

        <View style={styles.receipt}>
          <UiText variant="feature">{viewedCount}</UiText>
          <UiText tone="secondary" style={styles.receiptLabel}>
            {t("flashcards.cardsViewedThisVisit", { count: viewedCount })}
          </UiText>
        </View>

        <View style={styles.actions}>
          <Button
            onPress={onReviewAgain}
            style={styles.actionButton}
            controlStyle={styles.actionControl}
          >
            {t("flashcards.reviewAgain")}
          </Button>
          <Button
            variant="secondary"
            onPress={onBackToWordlists}
            style={styles.actionButton}
            controlStyle={styles.actionControl}
          >
            {t("flashcards.backToWordlists")}
          </Button>
          <Button variant="quiet" onPress={onReturnToLastCard}>
            {t("flashcards.returnToLastCard")}
          </Button>
        </View>

        <UiText variant="caption" tone="secondary" style={styles.note}>
          {t("flashcards.completionNote")}
        </UiText>
      </Animated.View>
    </ScrollView>
  );
};

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    scrollView: { flex: 1 },
    contentContainer: {
      flexGrow: 1,
      justifyContent: "center",
      paddingHorizontal: theme.spacing.lg,
      paddingTop: theme.spacing.md,
      paddingBottom: theme.spacing.xl,
    },
    closeRow: {
      minHeight: theme.geometry.touchTarget,
      alignItems: "flex-end",
      marginBottom: theme.spacing.md,
    },
    closeButton: {
      width: theme.geometry.touchTarget,
      height: theme.geometry.touchTarget,
      alignItems: "center",
      justifyContent: "center",
      borderWidth: 1,
      borderColor: theme.colors.roles.controlBorder,
      borderRadius: theme.borderRadius.full,
      backgroundColor: theme.colors.background.surface,
    },
    deck: {
      width: 132,
      height: 118,
      alignSelf: "center",
      marginBottom: theme.spacing.lg,
    },
    deckCard: {
      position: "absolute",
      top: 14,
      left: 12,
      right: 12,
      bottom: 0,
      borderWidth: 2,
      borderColor: theme.colors.roles.controlBorder,
      borderRadius: theme.borderRadius.lg,
      backgroundColor: theme.colors.background.surface,
      transformOrigin: "50% 90%",
      ...theme.shadows.sm,
    },
    leftCard: { transform: [{ rotate: "-10deg" }] },
    rightCard: { transform: [{ rotate: "9deg" }] },
    frontCard: { overflow: "visible" },
    bookmark: {
      position: "absolute",
      width: 16,
      height: 32,
      right: 18,
      bottom: 0,
      borderTopLeftRadius: 4,
      borderTopRightRadius: 4,
      backgroundColor: theme.colors.roles.action,
    },
    copy: { alignItems: "stretch" },
    wordlistName: {
      alignSelf: "center",
      marginBottom: theme.spacing.sm,
      textAlign: "center",
    },
    heading: { textAlign: "center" },
    message: {
      marginTop: theme.spacing.compact,
      textAlign: "center",
    },
    receipt: {
      marginVertical: theme.spacing.lg,
      paddingVertical: theme.spacing.md,
      borderTopWidth: 1,
      borderBottomWidth: 1,
      borderColor: theme.colors.border.medium,
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "baseline",
      gap: theme.spacing.compact,
    },
    receiptLabel: { flexShrink: 1 },
    actions: { gap: theme.spacing.compact },
    actionButton: { minHeight: 56 },
    actionControl: { minHeight: 56 },
    note: {
      marginTop: theme.spacing.md,
      textAlign: "center",
    },
  });
