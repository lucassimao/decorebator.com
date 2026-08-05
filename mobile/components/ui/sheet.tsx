import React, { useCallback, useEffect, useRef, useState } from "react";
import {
  AccessibilityInfo,
  findNodeHandle,
  KeyboardAvoidingView,
  Modal,
  Pressable,
  ScrollView,
  View,
} from "react-native";
import Animated, {
  cancelAnimation,
  runOnJS,
  useAnimatedStyle,
  useReducedMotion,
  useSharedValue,
  withTiming,
} from "react-native-reanimated";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { useTranslation } from "react-i18next";

import { useTheme } from "@/contexts/ThemeContext";
import { UiText } from "@/components/ui/text";

export interface SheetProps {
  visible: boolean;
  onClose: () => void;
  title?: string;
  eyebrow?: string;
  children: React.ReactNode;
  returnFocusRef?: React.RefObject<View | null>;
  testID?: string;
}

export function Sheet({
  visible,
  onClose,
  title,
  eyebrow,
  children,
  returnFocusRef,
  testID,
}: SheetProps) {
  const { theme } = useTheme();
  const { t } = useTranslation();
  const insets = useSafeAreaInsets();
  const reduceMotion = useReducedMotion();
  const [rendered, setRendered] = useState(visible);
  const opacity = useSharedValue(visible ? 1 : 0);
  const translateY = useSharedValue(visible ? 0 : 48);
  const pendingRestoreRef = useRef(false);

  const restoreFocus = useCallback(() => {
    const node = returnFocusRef?.current
      ? findNodeHandle(returnFocusRef.current)
      : null;
    if (node) {
      AccessibilityInfo.setAccessibilityFocus(node);
    }
  }, [returnFocusRef]);

  const finishDismiss = useCallback(() => {
    pendingRestoreRef.current = true;
    setRendered(false);
  }, []);

  useEffect(() => {
    if (rendered || !pendingRestoreRef.current) return;
    pendingRestoreRef.current = false;
    const frame = requestAnimationFrame(restoreFocus);
    return () => cancelAnimationFrame(frame);
  }, [rendered, restoreFocus]);

  useEffect(() => {
    if (visible) {
      setRendered(true);
    }
  }, [visible]);

  useEffect(() => {
    if (!rendered) return;

    cancelAnimation(opacity);
    cancelAnimation(translateY);
    if (visible) {
      if (reduceMotion) {
        opacity.value = 1;
        translateY.value = 0;
      } else {
        opacity.value = withTiming(1, { duration: theme.motion.standard });
        translateY.value = withTiming(0, { duration: theme.motion.sheet });
      }
      return;
    }

    if (reduceMotion) {
      opacity.value = 0;
      translateY.value = 48;
      finishDismiss();
      return;
    }

    opacity.value = withTiming(0, { duration: theme.motion.standard });
    translateY.value = withTiming(
      48,
      { duration: theme.motion.sheet },
      (finished) => {
        if (finished) runOnJS(finishDismiss)();
      },
    );
  }, [
    finishDismiss,
    opacity,
    reduceMotion,
    rendered,
    theme.motion,
    translateY,
    visible,
  ]);

  useEffect(
    () => () => {
      cancelAnimation(opacity);
      cancelAnimation(translateY);
    },
    [opacity, translateY],
  );

  const backdropStyle = useAnimatedStyle(() => ({ opacity: opacity.value }));
  const sheetStyle = useAnimatedStyle(() => ({
    transform: [{ translateY: translateY.value }],
  }));

  if (!rendered) return null;

  return (
    <Modal
      testID={testID}
      visible
      transparent
      animationType="none"
      statusBarTranslucent
      onRequestClose={onClose}
    >
      <View
        style={{ flex: 1, justifyContent: "flex-end" }}
        accessibilityViewIsModal
        importantForAccessibility="yes"
      >
        <Animated.View
          pointerEvents="none"
          style={[
            {
              position: "absolute",
              inset: 0,
              backgroundColor: theme.colors.roles.scrim,
            },
            backdropStyle,
          ]}
        />
        <Pressable
          testID="sheet-backdrop"
          accessibilityRole="button"
          accessibilityLabel={t("common.close")}
          style={{ position: "absolute", inset: 0 }}
          onPress={onClose}
        />
        <KeyboardAvoidingView
          pointerEvents="box-none"
          behavior={process.env.EXPO_OS === "ios" ? "padding" : undefined}
        >
          <Animated.View
            testID={testID ? `${testID}-panel` : undefined}
            style={[
              {
                maxHeight: "92%",
                paddingTop: theme.spacing.sm,
                paddingBottom: Math.max(insets.bottom, theme.spacing.lg),
                borderTopLeftRadius: theme.borderRadius.xl,
                borderTopRightRadius: theme.borderRadius.xl,
                borderCurve: "continuous",
                backgroundColor: theme.colors.background.surface,
              },
              sheetStyle,
            ]}
          >
            <View
              accessibilityElementsHidden
              importantForAccessibility="no-hide-descendants"
              style={{
                width: 40,
                height: 4,
                alignSelf: "center",
                marginBottom: theme.spacing.md,
                borderRadius: theme.borderRadius.full,
                backgroundColor: theme.colors.border.dark,
              }}
            />
            <ScrollView
              contentInsetAdjustmentBehavior="automatic"
              keyboardShouldPersistTaps="handled"
              contentContainerStyle={{
                paddingHorizontal: theme.spacing.comfortable,
                gap: theme.spacing.md,
              }}
            >
              {eyebrow ? (
                <UiText variant="label" tone="action">
                  {eyebrow.toUpperCase()}
                </UiText>
              ) : null}
              {title ? <UiText variant="title">{title}</UiText> : null}
              {children}
            </ScrollView>
          </Animated.View>
        </KeyboardAvoidingView>
      </View>
    </Modal>
  );
}
