import { Theme } from "@/contexts/ThemeContext";
import { ViewStyle, TextStyle, ImageStyle } from "react-native";

type NamedStyles<T> = { [P in keyof T]: ViewStyle | TextStyle | ImageStyle };

// Common style creator functions
export const createCommonStyles = <T extends NamedStyles<T>>(theme: Theme) => ({
  // Basic Containers
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

  // Responsive Containers
  responsiveContainer: {
    flex: 1,
    backgroundColor: theme.colors.background.default,
    paddingHorizontal: theme.spacing.container,
    maxWidth: theme.layout.maxContentWidth || "100%",
    alignSelf: "center",
    width: "100%",
  },
  contentContainer: {
    paddingHorizontal: theme.spacing.container,
    maxWidth: theme.layout.maxContentWidth || "100%",
    alignSelf: "center",
    width: "100%",
  },
  centeredContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: theme.spacing.container,
    backgroundColor: theme.colors.background.default,
  },

  // Cards
  card: {
    backgroundColor: theme.colors.background.surface,
    borderRadius: theme.borderRadius.lg,
    padding: theme.spacing.md,
    ...theme.shadows.sm,
  },
  cardElevated: {
    backgroundColor: theme.colors.background.elevated,
    borderRadius: theme.borderRadius.lg,
    padding: theme.spacing.md,
    ...theme.shadows.md,
  },

  // Responsive Grid System
  gridContainer: {
    flexDirection: "row" as const,
    flexWrap: "wrap" as const,
    justifyContent: "space-between",
    gap: theme.spacing.card,
  },
  gridItem: {
    minWidth: theme.layout.cardMinWidth,
    flex: theme.layout.gridColumns === 1 ? 1 : 0,
    flexBasis:
      theme.layout.gridColumns === 1
        ? "100%"
        : theme.layout.gridColumns === 2
          ? "48%"
          : "31%",
  },
  gridItemSingle: {
    width: "100%",
  },
  gridItemDouble: {
    width: "48%",
    minWidth: 280,
  },
  gridItemTriple: {
    width: "31%",
    minWidth: 200,
  },

  // Layout Utilities
  row: {
    flexDirection: "row" as const,
    alignItems: "center",
  },
  rowWrap: {
    flexDirection: "row" as const,
    flexWrap: "wrap" as const,
    alignItems: "center",
  },
  spaceBetween: {
    flexDirection: "row" as const,
    justifyContent: "space-between",
    alignItems: "center",
  },
  spaceAround: {
    flexDirection: "row" as const,
    justifyContent: "space-around",
    alignItems: "center",
  },
  center: {
    alignItems: "center",
    justifyContent: "center",
  },
  flexStart: {
    alignItems: "flex-start",
    justifyContent: "flex-start",
  },
  flexEnd: {
    alignItems: "flex-end",
    justifyContent: "flex-end",
  },

  // Responsive Buttons
  buttonPrimary: {
    backgroundColor: theme.colors.primary,
    paddingVertical: theme.spacing.md,
    paddingHorizontal: theme.spacing.lg,
    borderRadius: theme.borderRadius.md,
    alignItems: "center" as const,
    justifyContent: "center" as const,
    minHeight: theme.touchTargets.comfortable,
    ...theme.shadows.sm,
  },
  buttonSecondary: {
    backgroundColor: theme.colors.background.elevated,
    paddingVertical: theme.spacing.md,
    paddingHorizontal: theme.spacing.lg,
    borderRadius: theme.borderRadius.md,
    borderWidth: 1,
    borderColor: theme.colors.ui.border,
    alignItems: "center" as const,
    justifyContent: "center" as const,
    minHeight: theme.touchTargets.comfortable,
  },
  buttonLarge: {
    backgroundColor: theme.colors.primary,
    paddingVertical: theme.spacing.lg,
    paddingHorizontal: theme.spacing.xl,
    borderRadius: theme.borderRadius.lg,
    alignItems: "center" as const,
    justifyContent: "center" as const,
    minHeight: theme.touchTargets.large,
    ...theme.shadows.md,
  },
  buttonText: {
    color: theme.colors.text.inverse,
    fontSize: theme.typography.sizes.body,
    fontWeight: theme.typography.weights.semibold,
    lineHeight: theme.typography.lineHeights.body,
  },
  buttonTextSecondary: {
    color: theme.colors.text.primary,
    fontSize: theme.typography.sizes.body,
    fontWeight: theme.typography.weights.semibold,
    lineHeight: theme.typography.lineHeights.body,
  },
  buttonTextLarge: {
    color: theme.colors.text.inverse,
    fontSize: theme.typography.sizes.bodyLarge,
    fontWeight: theme.typography.weights.semibold,
    lineHeight: theme.typography.lineHeights.bodyLarge,
  },

  // Responsive Text Styles
  heading1: {
    fontSize: theme.typography.sizes.display,
    fontWeight: theme.typography.weights.bold,
    color: theme.colors.text.primary,
    lineHeight: theme.typography.lineHeights.display,
    marginBottom: theme.spacing.sm,
  },
  heading2: {
    fontSize: theme.typography.sizes.headline,
    fontWeight: theme.typography.weights.bold,
    color: theme.colors.text.primary,
    lineHeight: theme.typography.lineHeights.headline,
    marginBottom: theme.spacing.sm,
  },
  heading3: {
    fontSize: theme.typography.sizes.title,
    fontWeight: theme.typography.weights.semibold,
    color: theme.colors.text.primary,
    lineHeight: theme.typography.lineHeights.title,
    marginBottom: theme.spacing.xs,
  },
  subtitle: {
    fontSize: theme.typography.sizes.subtitle,
    fontWeight: theme.typography.weights.medium,
    color: theme.colors.text.primary,
    lineHeight: theme.typography.lineHeights.subtitle,
    marginBottom: theme.spacing.xs,
  },
  bodyText: {
    fontSize: theme.typography.sizes.body,
    fontWeight: theme.typography.weights.regular,
    color: theme.colors.text.primary,
    lineHeight: theme.typography.lineHeights.body,
  },
  bodyLarge: {
    fontSize: theme.typography.sizes.bodyLarge,
    fontWeight: theme.typography.weights.regular,
    color: theme.colors.text.primary,
    lineHeight: theme.typography.lineHeights.bodyLarge,
  },
  secondaryText: {
    fontSize: theme.typography.sizes.body,
    fontWeight: theme.typography.weights.regular,
    color: theme.colors.text.secondary,
    lineHeight: theme.typography.lineHeights.body,
  },
  caption: {
    fontSize: theme.typography.sizes.caption,
    fontWeight: theme.typography.weights.regular,
    color: theme.colors.text.secondary,
    lineHeight: theme.typography.lineHeights.caption,
  },

  // Responsive Form Elements
  input: {
    backgroundColor: theme.colors.ui.inputBackground,
    borderWidth: 1,
    borderColor: theme.colors.ui.border,
    borderRadius: theme.borderRadius.md,
    paddingVertical: theme.spacing.md,
    paddingHorizontal: theme.spacing.md,
    fontSize: theme.typography.sizes.body,
    color: theme.colors.text.primary,
    minHeight: theme.touchTargets.comfortable,
  },
  inputError: {
    borderColor: theme.colors.error,
  },
  inputLabel: {
    fontSize: theme.typography.sizes.body,
    fontWeight: theme.typography.weights.medium,
    color: theme.colors.text.primary,
    marginBottom: theme.spacing.xs,
  },
  formContainer: {
    maxWidth: theme.layout.modalMaxWidth || "100%",
    alignSelf: "center",
    width: "100%",
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
    paddingHorizontal: theme.spacing.sm,
    paddingVertical: theme.spacing.xs,
    borderRadius: theme.borderRadius.sm,
  },
  errorBadge: {
    backgroundColor: theme.colors.state.incorrectBackground,
    paddingHorizontal: theme.spacing.sm,
    paddingVertical: theme.spacing.xs,
    borderRadius: theme.borderRadius.sm,
  },

  // Loading states
  loadingContainer: {
    flex: 1,
    backgroundColor: theme.colors.background.default,
    alignItems: "center",
    justifyContent: "center",
  },

  // Responsive Modal Styles
  modalBackdrop: {
    position: "absolute" as const,
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: theme.colors.overlay.backdrop,
    justifyContent: "center" as const,
    alignItems: "center" as const,
    padding: theme.spacing.modal,
  },
  modalContent: {
    backgroundColor: theme.colors.background.surface,
    borderRadius: theme.borderRadius.xl,
    padding: theme.spacing.modal,
    maxWidth: theme.layout.modalMaxWidth || "100%",
    width: theme.layout.modalMaxWidth ? "auto" : "100%",
    maxHeight: "90%",
    ...theme.shadows.lg,
  },
  modalContentFullScreen: {
    backgroundColor: theme.colors.background.surface,
    borderRadius: 0,
    padding: theme.spacing.lg,
    width: "100%",
    height: "100%",
    maxWidth: "100%",
    maxHeight: "100%",
  },

  // Responsive Spacing Utilities
  paddingHorizontalResponsive: {
    paddingHorizontal: theme.spacing.container,
  },
  paddingVerticalResponsive: {
    paddingVertical: theme.spacing.section,
  },
  marginBottomResponsive: {
    marginBottom: theme.spacing.section,
  },
  gapResponsive: {
    gap: theme.spacing.card,
  },

  // Touch Target Helpers
  touchableArea: {
    minHeight: theme.touchTargets.minimum,
    minWidth: theme.touchTargets.minimum,
    justifyContent: "center" as const,
    alignItems: "center" as const,
  },
  touchableAreaComfortable: {
    minHeight: theme.touchTargets.comfortable,
    minWidth: theme.touchTargets.comfortable,
    justifyContent: "center" as const,
    alignItems: "center" as const,
  },
  touchableAreaLarge: {
    minHeight: theme.touchTargets.large,
    minWidth: theme.touchTargets.large,
    justifyContent: "center" as const,
    alignItems: "center" as const,
  },
});

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

// Responsive component helpers
export const createResponsiveContainer = (theme: Theme) => ({
  width: "100%",
  maxWidth: theme.layout.maxContentWidth || "100%",
  alignSelf: "center" as const,
  paddingHorizontal: theme.spacing.container,
});

export const createGridLayout = (theme: Theme, columns?: number) => {
  const gridColumns = columns || theme.layout.gridColumns;
  return {
    flexDirection: "row" as const,
    flexWrap: "wrap" as const,
    justifyContent: "space-between" as const,
    gap: theme.spacing.card,
    itemWidth: gridColumns === 1 ? "100%" : gridColumns === 2 ? "48%" : "31%",
  };
};

export const createModalLayout = (theme: Theme, fullScreen = false) => {
  if (fullScreen || !theme.layout.modalMaxWidth) {
    return {
      width: "100%",
      height: "100%",
      maxWidth: "100%",
      maxHeight: "100%",
    };
  }

  return {
    maxWidth: theme.layout.modalMaxWidth,
    width: "auto",
    maxHeight: "90%",
  };
};

// Helper function to get responsive dimensions
export const getResponsiveDimensions = (theme: Theme) => ({
  contentWidth: theme.layout.maxContentWidth || "100%",
  gridColumns: theme.layout.gridColumns,
  spacing: {
    container: theme.spacing.container,
    section: theme.spacing.section,
    card: theme.spacing.card,
  },
  touchTargets: theme.touchTargets,
});

// Helper function for responsive font sizes
export const getResponsiveFontSize = (
  theme: Theme,
  size: keyof typeof theme.typography.sizes,
) => ({
  fontSize: theme.typography.sizes[size],
  lineHeight: theme.typography.lineHeights[size],
});
