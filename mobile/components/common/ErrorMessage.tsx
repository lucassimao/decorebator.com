import React from "react";
import { Text, StyleProp, TextStyle } from "react-native";
import { FieldError } from "react-hook-form";

interface ErrorMessageProps {
  error?: FieldError | null;
  style?: StyleProp<TextStyle>;
  errorStyle?: StyleProp<TextStyle>;
}

export const ErrorMessage: React.FC<ErrorMessageProps> = React.memo(
  ({ error, style, errorStyle }) => {
    if (!error) return null;

    return (
      <Text style={[style, errorStyle]}>{error?.message || "Invalid"}</Text>
    );
  },
);

ErrorMessage.displayName = "ErrorMessage";
