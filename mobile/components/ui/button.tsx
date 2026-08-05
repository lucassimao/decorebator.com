import React from "react";
import {
  ActivityIndicator,
  Pressable,
  PressableProps,
  StyleProp,
  ViewStyle,
} from "react-native";
import Animated, {
  useAnimatedStyle,
  useReducedMotion,
  useSharedValue,
  withSpring,
} from "react-native-reanimated";

import { useTheme } from "@/contexts/ThemeContext";
import { UiText } from "@/components/ui/text";

export type ButtonVariant = "primary" | "onInk" | "secondary" | "quiet";

export interface ButtonProps
  extends Omit<PressableProps, "style" | "children"> {
  children: string;
  leading?: React.ReactNode;
  trailing?: React.ReactNode;
  variant?: ButtonVariant;
  loading?: boolean;
  loadingLabel?: string;
  style?: StyleProp<ViewStyle>;
}

export const Button = React.forwardRef<
  React.ElementRef<typeof Pressable>,
  ButtonProps
>(
  (
    {
      children,
      variant = "primary",
      loading = false,
      loadingLabel,
      leading,
      trailing,
      disabled = false,
      onPressIn,
      onPressOut,
      onFocus,
      onBlur,
      accessibilityState,
      style,
      ...props
    },
    ref,
  ) => {
    const { theme } = useTheme();
    const reduceMotion = useReducedMotion();
    const scale = useSharedValue(1);
    const translateY = useSharedValue(0);
    const [focused, setFocused] = React.useState(false);
    const isDisabled = disabled || loading;

    const animatedStyle = useAnimatedStyle(() => ({
      transform: [{ scale: scale.value }, { translateY: translateY.value }],
    }));

    const animatePress = (pressed: boolean) => {
      const nextScale = pressed ? theme.motion.pressScale : 1;
      const nextTranslate = pressed ? 3 : 0;
      if (reduceMotion) {
        scale.value = 1;
        translateY.value = 0;
        return;
      }
      scale.value = withSpring(nextScale, { damping: 18, stiffness: 320 });
      translateY.value = withSpring(nextTranslate, {
        damping: 18,
        stiffness: 320,
      });
    };

    const colors = (() => {
      switch (variant) {
        case "onInk":
          return {
            background: theme.colors.roles.brandAccent,
            pressed: theme.colors.roles.brandAccentPressed,
            foreground: theme.colors.roles.onBrandAccent,
            border: "transparent",
            shadow: "rgba(40, 20, 13, 0.55)",
          };
        case "secondary":
          return {
            background: theme.colors.background.surface,
            pressed: theme.colors.background.elevated,
            foreground: theme.colors.text.primary,
            border: theme.colors.border.medium,
            shadow: "transparent",
          };
        case "quiet":
          return {
            background: "transparent",
            pressed: theme.colors.background.subtle,
            foreground: theme.colors.roles.action,
            border: "transparent",
            shadow: "transparent",
          };
        default:
          return {
            background: theme.colors.roles.action,
            pressed: theme.colors.roles.actionPressed,
            foreground: theme.colors.roles.onAction,
            border: "transparent",
            shadow: "rgba(72, 20, 6, 0.62)",
          };
      }
    })();

    return (
      <Animated.View style={[animatedStyle, style]}>
        <Pressable
          ref={ref}
          accessible
          accessibilityRole="button"
          accessibilityState={{
            ...accessibilityState,
            disabled: isDisabled,
            busy: loading,
          }}
          disabled={isDisabled}
          onFocus={(event) => {
            setFocused(true);
            onFocus?.(event);
          }}
          onBlur={(event) => {
            setFocused(false);
            onBlur?.(event);
          }}
          onPressIn={(event) => {
            animatePress(true);
            onPressIn?.(event);
          }}
          onPressOut={(event) => {
            animatePress(false);
            onPressOut?.(event);
          }}
          style={({ pressed }) => [
            {
              minHeight: theme.geometry.controlHeight,
              minWidth: theme.geometry.touchTarget,
              paddingHorizontal: theme.spacing.md,
              borderWidth: focused ? 2 : 1,
              borderColor: focused ? theme.colors.roles.focus : colors.border,
              borderRadius: theme.borderRadius.md,
              borderCurve: "continuous",
              alignItems: "center",
              justifyContent: "center",
              flexDirection: "row",
              gap: theme.spacing.sm,
              backgroundColor: pressed ? colors.pressed : colors.background,
              boxShadow:
                variant === "primary" || variant === "onInk"
                  ? `0 5px 0 ${colors.shadow}`
                  : undefined,
            },
            disabled && {
              backgroundColor: theme.colors.roles.disabledBackground,
              borderColor: "transparent",
              boxShadow: undefined,
            },
          ]}
          {...props}
        >
          {leading}
          {loading && !reduceMotion ? (
            <ActivityIndicator
              testID="button-loading-indicator"
              size="small"
              color={
                disabled ? theme.colors.roles.disabledText : colors.foreground
              }
            />
          ) : null}
          {loading && reduceMotion && !loadingLabel ? (
            <UiText accessibilityElementsHidden>…</UiText>
          ) : null}
          <UiText
            variant="heading"
            style={{
              color: disabled
                ? theme.colors.roles.disabledText
                : colors.foreground,
              textAlign: "center",
            }}
          >
            {loading && loadingLabel ? loadingLabel : children}
          </UiText>
          {trailing}
        </Pressable>
      </Animated.View>
    );
  },
);

Button.displayName = "Button";
