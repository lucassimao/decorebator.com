import { Platform } from "react-native";

// Screen size breakpoints based on logical pixels
export enum BREAKPOINTS {
  SMALL = 375, // <= 375px logical width (iPhone SE, smaller Android devices)
  MEDIUM = 430, // 376-430px logical width (iPhone 12/13/14, standard Android, iPhone Plus)
  LARGE = 480, // 431-480px logical width (larger Android devices, phablets)
  XLARGE = 999, // >= 481px logical width (tablets, very large devices)
}

// Determine screen size category - requires explicit width parameter
export const getScreenSizeCategory = (width: number) => {
  if (width <= BREAKPOINTS.SMALL) return "small";
  if (width <= BREAKPOINTS.MEDIUM) return "medium";
  if (width <= BREAKPOINTS.LARGE) return "large";
  return "xlarge";
};

// Responsive spacing based on screen size - requires explicit width parameter
export const getResponsiveSpacing = (width: number) => {
  const category = getScreenSizeCategory(width);

  switch (category) {
    case "small":
      return {
        horizontal: 12, // Reduced back to 12px to give form more space
        vertical: 12,
        formPadding: 12, // Keep form internal padding reasonable
        elementSpacing: 12, // Increased from 8dp to 12dp for better touch interaction
        buttonHeight: 44, // Minimum accessibility standard
        minTouchTarget: 44,
      };
    case "medium":
      return {
        horizontal: 24, // Increased for better proportions
        vertical: 16,
        formPadding: 20,
        elementSpacing: 16, // Increased from 12dp to 16dp for better form field separation
        buttonHeight: 48,
        minTouchTarget: 48,
      };
    case "large":
      return {
        horizontal: 32, // Increased for better proportions on larger screens
        vertical: 20,
        formPadding: 24,
        elementSpacing: 16,
        buttonHeight: 52,
        minTouchTarget: 52,
      };
    case "xlarge":
    default:
      return {
        horizontal: 40, // Increased for better proportions on largest screens
        vertical: 24,
        formPadding: 28,
        elementSpacing: 20,
        buttonHeight: 56,
        minTouchTarget: 56,
      };
  }
};

// Base typography scale (Material Design 3 foundation)
const BASE_TYPOGRAPHY_SCALE = {
  display: 34, // Major headings
  title: 20, // Card titles, headers
  headline: 16, // Subheadings, list items
  body: 14, // Default text
  label: 12, // Labels, captions
  micro: 10, // Very small text
} as const;

// Screen size multipliers for typography
const TYPOGRAPHY_MULTIPLIERS = {
  small: 1.0, // 375px and below (iPhone SE, smaller Android)
  medium: 1.1, // 376-430px (iPhone 12/13/14, standard Android, iPhone Plus)
  large: 1.2, // 431-480px (larger Android devices, phablets)
  xlarge: 1.3, // 481px+ (tablets, very large devices)
} as const;

// Get typography multiplier for screen width
const getTypographyMultiplier = (width: number): number => {
  const category = getScreenSizeCategory(width);
  return TYPOGRAPHY_MULTIPLIERS[category];
};

// Get scaled font size for specific typography level
export const getScaledFontSize = (
  fontName: keyof typeof BASE_TYPOGRAPHY_SCALE,
  width: number,
): number => {
  const baseSize = BASE_TYPOGRAPHY_SCALE[fontName];
  const multiplier = getTypographyMultiplier(width);
  return Math.round(baseSize * multiplier);
};

// Responsive font sizes - requires explicit width parameter
export const getResponsiveFontSizes = (width: number) => {
  const multiplier = getTypographyMultiplier(width);

  return {
    display: Math.round(BASE_TYPOGRAPHY_SCALE.display * multiplier),
    title: Math.round(BASE_TYPOGRAPHY_SCALE.title * multiplier),
    headline: Math.round(BASE_TYPOGRAPHY_SCALE.headline * multiplier),
    body: Math.round(BASE_TYPOGRAPHY_SCALE.body * multiplier),
    label: Math.round(BASE_TYPOGRAPHY_SCALE.label * multiplier),
    micro: Math.round(BASE_TYPOGRAPHY_SCALE.micro * multiplier),
    lineHeight: 1.4 + (multiplier - 1.0) * 0.2, // Scale line height slightly
  };
};

// Get keyboard avoiding behavior based on platform
export const getKeyboardBehavior = (): "height" | "position" | "padding" => {
  return Platform.OS === "ios" ? "padding" : "height";
};

// Calculate safe keyboard offset - requires explicit width parameter
export const getKeyboardOffset = (width: number) => {
  const category = getScreenSizeCategory(width);

  // Smaller screens need more offset to ensure form visibility
  switch (category) {
    case "small":
      return Platform.OS === "ios" ? 20 : 0;
    case "medium":
      return Platform.OS === "ios" ? 10 : 0;
    case "large":
      return Platform.OS === "ios" ? 5 : 0;
    case "xlarge":
    default:
      return 0;
  }
};

// Helper to get responsive value based on screen size
export const getResponsiveValue = <T>(
  smallValue: T,
  mediumValue: T,
  largeValue: T,
  xlargeValue: T,
  width: number,
): T => {
  const category = getScreenSizeCategory(width);

  switch (category) {
    case "small":
      return smallValue;
    case "medium":
      return mediumValue;
    case "large":
      return largeValue;
    case "xlarge":
    default:
      return xlargeValue;
  }
};

// Type definitions for responsive values
export type ScreenSizeCategory = "small" | "medium" | "large" | "xlarge";

export interface ResponsiveSpacing {
  horizontal: number;
  vertical: number;
  formPadding: number;
  elementSpacing: number;
  buttonHeight: number;
  minTouchTarget: number;
}

export interface ResponsiveFontSizes {
  display: number;
  title: number;
  headline: number;
  body: number;
  label: number;
  micro: number;
  lineHeight: number;
}
