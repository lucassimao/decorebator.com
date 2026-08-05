import React from "react";
import { Text, TextProps, TextStyle } from "react-native";

import { useTheme } from "@/contexts/ThemeContext";

export type UiTextVariant = keyof ReturnType<
  typeof useTheme
>["theme"]["typography"];
export type UiTextTone =
  | "primary"
  | "secondary"
  | "inverse"
  | "danger"
  | "action";

export interface UiTextProps extends TextProps {
  variant?: UiTextVariant;
  tone?: UiTextTone;
}

const toneColor = (
  tone: UiTextTone,
  theme: ReturnType<typeof useTheme>["theme"],
): string => {
  switch (tone) {
    case "secondary":
      return theme.colors.text.secondary;
    case "inverse":
      return theme.colors.roles.onAction;
    case "danger":
      return theme.colors.roles.error;
    case "action":
      return theme.colors.roles.action;
    default:
      return theme.colors.text.primary;
  }
};

export const UiText = React.forwardRef<Text, UiTextProps>(
  (
    {
      variant = "body",
      tone = "primary",
      style,
      allowFontScaling = true,
      ...props
    },
    ref,
  ) => {
    const { theme } = useTheme();
    const colorStyle: TextStyle = { color: toneColor(tone, theme) };

    return (
      <Text
        ref={ref}
        allowFontScaling={allowFontScaling}
        style={[theme.typography[variant], colorStyle, style]}
        {...props}
      />
    );
  },
);

UiText.displayName = "UiText";
