import React from "react";
import { Text, StyleProp, TextStyle } from "react-native";
import { FieldError } from "react-hook-form";
import { useTheme } from "@/contexts/ThemeContext";
import { useResponsive } from "@/hooks/useResponsive";

interface ErrorMessageProps {
  error?: FieldError | null;
  style?: StyleProp<TextStyle>;
}

export const ErrorMessage: React.FC<ErrorMessageProps> = React.memo(
  ({ error, style }) => {
    const { theme } = useTheme();
    const responsive = useResponsive();
    
    if (!error) return null;

    const defaultErrorStyle = {
      fontSize: responsive.fontSizes.caption,
      color: theme.colors.error,
      marginTop: 4,
    };

    return (
      <Text style={[defaultErrorStyle, style]}>{error?.message || "Invalid"}</Text>
    );
  },
);

ErrorMessage.displayName = "ErrorMessage";
