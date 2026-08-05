import React from "react";
import { StyleProp, View, ViewProps, ViewStyle } from "react-native";

import { useTheme } from "@/contexts/ThemeContext";

export type CardVariant = "surface" | "elevated" | "emphasis";

export interface CardProps extends Omit<ViewProps, "style"> {
  variant?: CardVariant;
  padding?: "none" | "compact" | "default" | "comfortable";
  style?: StyleProp<ViewStyle>;
}

export const Card = React.forwardRef<View, CardProps>(
  ({ variant = "surface", padding = "default", style, ...props }, ref) => {
    const { theme } = useTheme();
    const paddingValue = {
      none: 0,
      compact: theme.spacing.compact,
      default: theme.spacing.md,
      comfortable: theme.spacing.comfortable,
    }[padding];
    const backgroundColor =
      variant === "elevated"
        ? theme.colors.background.elevated
        : variant === "emphasis"
          ? theme.colors.roles.emphasisSurface
          : theme.colors.background.surface;

    return (
      <View
        ref={ref}
        style={[
          {
            padding: paddingValue,
            borderWidth: 1,
            borderColor: theme.colors.border.light,
            borderRadius: theme.borderRadius.lg,
            borderCurve: "continuous",
            backgroundColor,
            ...(variant === "elevated" ? theme.shadows.lg : undefined),
          },
          style,
        ]}
        {...props}
      />
    );
  },
);

Card.displayName = "Card";
