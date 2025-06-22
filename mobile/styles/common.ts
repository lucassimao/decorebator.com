import { Theme } from '@/contexts/ThemeContext';
import { ViewStyle, TextStyle, ImageStyle } from 'react-native';

type NamedStyles<T> = { [P in keyof T]: ViewStyle | TextStyle | ImageStyle };

// Common style creator functions
export const createCommonStyles = <T extends NamedStyles<T>>(theme: Theme) => ({
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
    padding: theme.spacing.md,
    ...theme.shadows.sm,
  },
  cardElevated: {
    backgroundColor: theme.colors.background.elevated,
    borderRadius: theme.borderRadius.lg,
    padding: theme.spacing.md,
    ...theme.shadows.md,
  },
  
  // Buttons
  buttonPrimary: {
    backgroundColor: theme.colors.primary,
    paddingVertical: theme.spacing.md,
    paddingHorizontal: theme.spacing.lg,
    borderRadius: theme.borderRadius.md,
    alignItems: 'center',
    justifyContent: 'center',
    ...theme.shadows.sm,
  },
  buttonSecondary: {
    backgroundColor: theme.colors.background.elevated,
    paddingVertical: theme.spacing.md,
    paddingHorizontal: theme.spacing.lg,
    borderRadius: theme.borderRadius.md,
    borderWidth: 1,
    borderColor: theme.colors.ui.border,
    alignItems: 'center',
    justifyContent: 'center',
  },
  buttonText: {
    color: theme.colors.text.inverse,
    fontSize: 16,
    fontWeight: '600',
  },
  buttonTextSecondary: {
    color: theme.colors.text.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  
  // Text styles
  heading1: {
    fontSize: 32,
    fontWeight: '700',
    color: theme.colors.text.primary,
    marginBottom: theme.spacing.sm,
  },
  heading2: {
    fontSize: 24,
    fontWeight: '600',
    color: theme.colors.text.primary,
    marginBottom: theme.spacing.sm,
  },
  heading3: {
    fontSize: 20,
    fontWeight: '600',
    color: theme.colors.text.primary,
    marginBottom: theme.spacing.xs,
  },
  bodyText: {
    fontSize: 16,
    color: theme.colors.text.primary,
    lineHeight: 24,
  },
  secondaryText: {
    fontSize: 14,
    color: theme.colors.text.secondary,
    lineHeight: 20,
  },
  caption: {
    fontSize: 12,
    color: theme.colors.text.secondary,
    lineHeight: 16,
  },
  
  // Form elements
  input: {
    backgroundColor: theme.colors.ui.inputBackground,
    borderWidth: 1,
    borderColor: theme.colors.ui.border,
    borderRadius: theme.borderRadius.md,
    paddingVertical: theme.spacing.md,
    paddingHorizontal: theme.spacing.md,
    fontSize: 16,
    color: theme.colors.text.primary,
  },
  inputError: {
    borderColor: theme.colors.error,
  },
  inputLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: theme.colors.text.primary,
    marginBottom: theme.spacing.xs,
  },
  
  // Layout helpers
  row: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  spaceBetween: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  center: {
    alignItems: 'center',
    justifyContent: 'center',
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
    alignItems: 'center',
    justifyContent: 'center',
  },
  
  // Modal styles
  modalBackdrop: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: theme.colors.overlay.backdrop,
  },
  modalContent: {
    backgroundColor: theme.colors.background.surface,
    borderRadius: theme.borderRadius.xl,
    padding: theme.spacing.lg,
    ...theme.shadows.lg,
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
  if (theme.mode === 'light') {
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