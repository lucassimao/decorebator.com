import React, { useEffect, useId, useState } from "react";
import {
  AccessibilityInfo,
  Platform,
  StyleProp,
  TextInput,
  TextInputProps,
  View,
  ViewStyle,
} from "react-native";

import { useTheme } from "@/contexts/ThemeContext";
import { UiText } from "@/components/ui/text";

export interface InputProps extends TextInputProps {
  label: string;
  hint?: string;
  error?: string;
  leading?: React.ReactNode;
  trailing?: React.ReactNode;
  containerStyle?: StyleProp<ViewStyle>;
}

export const Input = React.forwardRef<TextInput, InputProps>(
  (
    {
      label,
      hint,
      error,
      leading,
      trailing,
      containerStyle,
      style,
      editable = true,
      onFocus,
      onBlur,
      accessibilityHint,
      ...props
    },
    ref,
  ) => {
    const { theme } = useTheme();
    const [focused, setFocused] = useState(false);
    const id = useId().replaceAll(":", "");
    const descriptionId = error
      ? `${id}-error`
      : hint
        ? `${id}-hint`
        : undefined;

    useEffect(() => {
      if (error && Platform.OS === "ios") {
        AccessibilityInfo.announceForAccessibility(error);
      }
    }, [error]);

    return (
      <View style={[{ gap: theme.spacing.sm }, containerStyle]}>
        <UiText variant="label">{label}</UiText>
        <View
          style={{
            minHeight: theme.geometry.controlHeight,
            flexDirection: "row",
            alignItems: "center",
            gap: theme.spacing.sm,
            paddingHorizontal: theme.spacing.compact,
            borderWidth: focused ? 2 : 1,
            borderColor: error
              ? theme.colors.roles.error
              : focused
                ? theme.colors.roles.focus
                : theme.colors.border.medium,
            borderRadius: theme.borderRadius.md,
            borderCurve: "continuous",
            backgroundColor: editable
              ? theme.colors.ui.inputBackground
              : theme.colors.roles.disabledBackground,
          }}
        >
          {leading}
          <TextInput
            ref={ref}
            editable={editable}
            accessibilityLabel={props.accessibilityLabel ?? label}
            accessibilityHint={error ?? accessibilityHint ?? hint}
            aria-invalid={Boolean(error)}
            aria-disabled={!editable}
            accessibilityState={{ disabled: !editable }}
            aria-describedby={descriptionId}
            placeholderTextColor={theme.colors.text.placeholder}
            selectionColor={theme.colors.roles.focus}
            style={[
              theme.typography.body,
              {
                minHeight: theme.geometry.touchTarget,
                flex: 1,
                paddingVertical: 0,
                color: editable
                  ? theme.colors.text.primary
                  : theme.colors.roles.disabledText,
              },
              style,
            ]}
            onFocus={(event) => {
              setFocused(true);
              onFocus?.(event);
            }}
            onBlur={(event) => {
              setFocused(false);
              onBlur?.(event);
            }}
            {...props}
          />
          {trailing}
        </View>
        {error ? (
          <UiText
            nativeID={descriptionId}
            variant="label"
            tone="danger"
            accessibilityLiveRegion="polite"
          >
            {error}
          </UiText>
        ) : hint ? (
          <UiText nativeID={descriptionId} variant="label" tone="secondary">
            {hint}
          </UiText>
        ) : null}
      </View>
    );
  },
);

Input.displayName = "Input";
