import { Theme, ResponsiveValues } from "@/contexts/ThemeContext";
import { ViewStyle, TextStyle, ImageStyle } from "react-native";

type NamedStyles<T> = { [P in keyof T]: ViewStyle | TextStyle | ImageStyle };

// Common style creator functions with responsive support
export const createCommonStyles = <T extends NamedStyles<T>>(
  theme: Theme,
  responsive: ResponsiveValues,
) => {
  return {
    // Containers
    screenContainer: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    safeArea: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    scrollContainer: {
      flexGrow: 1,
      backgroundColor: theme.colors.background.default,
    },

    // Cards
    card: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      padding: responsive.spacing.formPadding,
      ...theme.shadows.sm,
    },
    cardElevated: {
      backgroundColor: theme.colors.background.elevated,
      borderRadius: theme.borderRadius.lg,
      padding: responsive.spacing.formPadding,
      ...theme.shadows.md,
    },

    // Buttons
    buttonPrimary: {
      backgroundColor: theme.colors.primary,
      paddingVertical: responsive.spacing.vertical,
      paddingHorizontal: responsive.spacing.horizontal,
      borderRadius: theme.borderRadius.md,
      alignItems: "center",
      justifyContent: "center",
      minHeight: responsive.spacing.minTouchTarget,
      ...theme.shadows.sm,
    },
    buttonSecondary: {
      backgroundColor: theme.colors.background.elevated,
      paddingVertical: responsive.spacing.vertical,
      paddingHorizontal: responsive.spacing.horizontal,
      borderRadius: theme.borderRadius.md,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      alignItems: "center",
      justifyContent: "center",
      minHeight: responsive.spacing.minTouchTarget,
    },
    buttonText: {
      color: theme.colors.text.inverse,
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
    },
    buttonTextSecondary: {
      color: theme.colors.text.primary,
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
    },

    // Text styles
    heading1: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "700",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing,
      lineHeight: responsive.fontSizes.title * responsive.fontSizes.lineHeight,
    },
    heading2: {
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing,
      lineHeight:
        responsive.fontSizes.headline * responsive.fontSizes.lineHeight,
    },
    heading3: {
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing / 2,
      lineHeight:
        responsive.fontSizes.headline * responsive.fontSizes.lineHeight,
    },
    bodyText: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.primary,
      lineHeight: responsive.fontSizes.body * responsive.fontSizes.lineHeight,
    },
    secondaryText: {
      fontSize: responsive.fontSizes.label,
      color: theme.colors.text.secondary,
      lineHeight: responsive.fontSizes.label * responsive.fontSizes.lineHeight,
    },
    caption: {
      fontSize: responsive.fontSizes.micro,
      color: theme.colors.text.secondary,
      lineHeight: responsive.fontSizes.micro * responsive.fontSizes.lineHeight,
    },

    // Form elements
    input: {
      backgroundColor: theme.colors.ui.inputBackground,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      borderRadius: theme.borderRadius.md,
      paddingVertical: responsive.spacing.vertical,
      paddingHorizontal: responsive.spacing.horizontal,
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.primary,
      minHeight: responsive.spacing.minTouchTarget,
    },
    inputError: {
      borderColor: theme.colors.error,
    },
    inputLabel: {
      fontSize: responsive.fontSizes.label,
      fontWeight: "500",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing / 2, // Material Design 8px standard
    },

    // Layout helpers
    row: {
      flexDirection: "row",
      alignItems: "center",
    },
    spaceBetween: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
    },
    center: {
      alignItems: "center",
      justifyContent: "center",
    },

    // Dividers
    divider: {
      height: 1,
      backgroundColor: theme.colors.ui.divider,
      marginVertical: theme.spacing.md,
    },

    // Status styles
    successBadge: {
      backgroundColor: theme.colors.state.correctBackground,
      paddingHorizontal: responsive.spacing.elementSpacing,
      paddingVertical: responsive.spacing.elementSpacing / 2,
      borderRadius: theme.borderRadius.sm,
    },
    errorBadge: {
      backgroundColor: theme.colors.state.incorrectBackground,
      paddingHorizontal: responsive.spacing.elementSpacing,
      paddingVertical: responsive.spacing.elementSpacing / 2,
      borderRadius: theme.borderRadius.sm,
    },

    // Loading states
    loadingContainer: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
      alignItems: "center",
      justifyContent: "center",
    },

    // Modal styles
    modalBackdrop: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: theme.colors.overlay.backdrop,
    },
    modalContent: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.xl,
      padding: responsive.spacing.formPadding,
      ...theme.shadows.lg,
    },
  };
};

// Helper function to create gradient colors for charts and special UI
export const getChartColors = (theme: Theme) => [
  theme.colors.primary,
  theme.colors.success,
  theme.colors.premium,
  theme.colors.semantic.special,
  theme.colors.semantic.info,
  theme.colors.semantic.warning,
];

// Helper function to create subtle gradients for backgrounds
export const getSubtleGradient = (theme: Theme) => {
  if (theme.mode === "light") {
    return [
      theme.colors.background.default,
      theme.colors.background.subtle,
      theme.colors.background.default,
    ];
  } else {
    return [
      theme.colors.background.default,
      theme.colors.background.subtle,
      theme.colors.background.default,
    ];
  }
};

// Type helper for creating themed styles
export type ThemedStyles<T extends NamedStyles<T>> = (theme: Theme) => T;
