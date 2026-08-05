// theme/authTheme.ts

/**
 * Light theme specifically for authentication screens
 * These screens always use light theme regardless of user preferences
 */

import { Theme } from "@/contexts/ThemeContext";

export const authLightTheme: Theme = {
  mode: "light",
  colors: {
    background: {
      default: "#FDF6E3", // Warm beige from web for main background
      surface: "#FFFFFF", // White for cards
      elevated: "#F8F9FA", // Very light gray for elevation
      subtle: "#F5F6F8", // Subtle gray for sections
      secondary: "#F0F0F0", // Secondary background
    },
    // Keep existing vibrant brand colors
    primary: "#FF7B54",
    success: "#4CAF50",
    error: "#FF6B6B",
    premium: "#FFD700",
    border: {
      light: "#E0E0E0",
      medium: "#D0D0D0",
      dark: "#B0B0B0",
    },
    text: {
      primary: "#2D3436",
      secondary: "#636E72",
      placeholder: "#B2BEC3",
      inverse: "#FFFFFF",
      disabled: "#DFE6E9",
      tertiary: "#B2BEC3",
    },
    ui: {
      card: "#FFFFFF",
      border: "#E0E0E0",
      divider: "#F0F0F0",
      disabled: "#DFE6E9",
      inputBackground: "#FFFFFF",
    },
    semantic: {
      info: "#2196F3",
      warning: "#FF6B3D",
      special: "#9C27B0",
    },
    state: {
      correctBackground: "#E8F5E9",
      incorrectBackground: "#FFEBEE",
      infoBackground: "#F0F9FF",
      successDark: "#2E7D32",
      errorDark: "#C62828",
    },
    overlay: {
      backdrop: "rgba(0, 0, 0, 0.4)",
      backdropHeavy: "rgba(0, 0, 0, 0.5)",
      backdropLight: "rgba(0, 0, 0, 0.3)",
    },
    roles: {
      action: "#A93312",
      actionPressed: "#81260C",
      onAction: "#FFFFFF",
      brandAccent: "#FF7B54",
      brandAccentPressed: "#E85F3A",
      onBrandAccent: "#28140D",
      error: "#B3261E",
      emphasisSurface: "#FFF0E6",
      focus: "#315B7D",
      controlBorder: "#8F887C",
      disabledBackground: "#E6D8CA",
      disabledText: "#675D55",
      scrim: "rgba(15, 18, 18, 0.5)",
    },
  },
  spacing: {
    xs: 4,
    sm: 8,
    compact: 12,
    md: 16,
    comfortable: 20,
    lg: 24,
    xl: 32,
    xxl: 48,
  },
  borderRadius: {
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
    full: 9999,
  },
  typography: {
    micro: { fontSize: 10, lineHeight: 14 },
    caption: { fontSize: 11, lineHeight: 15 },
    label: { fontSize: 12, lineHeight: 16, fontWeight: "700" },
    small: { fontSize: 13, lineHeight: 18 },
    body: { fontSize: 15, lineHeight: 22 },
    heading: { fontSize: 16, lineHeight: 22, fontWeight: "700" },
    title: { fontSize: 20, lineHeight: 26, fontWeight: "700" },
    feature: {
      fontFamily: "Space Mono",
      fontSize: 24,
      lineHeight: 30,
      fontWeight: "400",
    },
    display: {
      fontFamily: "Space Mono",
      fontSize: 34,
      lineHeight: 40,
      fontWeight: "400",
    },
  },
  geometry: { touchTarget: 44, controlHeight: 48 },
  motion: { fast: 130, standard: 180, sheet: 240, pressScale: 0.975 },
  shadows: {
    sm: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 1 },
      shadowOpacity: 0.05,
      shadowRadius: 3,
      elevation: 2,
    },
    md: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.08,
      shadowRadius: 8,
      elevation: 4,
    },
    lg: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.12,
      shadowRadius: 16,
      elevation: 8,
    },
  },
};
