import React from "react";
import { Text, StyleProp, TextStyle } from "react-native";
import { FieldError } from "react-hook-form";
import { useTheme } from "@/contexts/ThemeContext";

interface ErrorMessageProps {
  error?: FieldError | null;
  style?: StyleProp<TextStyle>;
}

export const ErrorMessage: React.FC<ErrorMessageProps> = React.memo(
  ({ error, style }) => {
    const { theme, responsive } = useTheme();

    if (!error) return null;

    const defaultErrorStyle = {
      fontSize: responsive.fontSizes.micro,
      color: theme.colors.error,
      marginTop: responsive.spacing.elementSpacing / 2, // Material Design 8px standard
    };

    return (
      <Text style={[defaultErrorStyle, style]}>
        {error?.message || "Invalid"}
      </Text>
    );
  },
);

ErrorMessage.displayName = "ErrorMessage";
