import { DeviceType } from "@/utils/deviceDetection";

/**
 * Responsive design tokens and utilities
 * Provides consistent spacing, typography, and layout values across device types
 */

export interface ResponsiveSpacing {
  xs: number;
  sm: number;
  md: number;
  lg: number;
  xl: number;
  xxl: number;
  // Additional semantic spacing
  section: number;
  card: number;
  container: number;
  modal: number;
}

export interface ResponsiveTypography {
  sizes: {
    caption: number;
    body: number;
    bodyLarge: number;
    subtitle: number;
    title: number;
    headline: number;
    display: number;
  };
  lineHeights: {
    caption: number;
    body: number;
    bodyLarge: number;
    subtitle: number;
    title: number;
    headline: number;
    display: number;
  };
  weights: {
    regular: "400";
    medium: "500";
    semibold: "600";
    bold: "700";
  };
}

export interface ResponsiveLayout {
  maxContentWidth: number;
  gridColumns: number;
  cardMinWidth: number;
  modalMaxWidth: number;
  sidebarWidth?: number;
  headerHeight: number;
  tabBarHeight: number;
}

export interface ResponsiveTouchTargets {
  minimum: number;
  comfortable: number;
  large: number;
}

/**
 * Base responsive tokens that scale by device type
 */
export const RESPONSIVE_TOKENS = {
  // Spacing multipliers for different device types
  spacingMultipliers: {
    phone: 1,
    "tablet-7": 1.2,
    "tablet-10": 1.4,
  },

  // Typography multipliers for different device types
  typographyMultipliers: {
    phone: 1,
    "tablet-7": 1.1,
    "tablet-10": 1.25,
  },

  // Touch target sizes following platform guidelines
  touchTargets: {
    phone: {
      minimum: 44, // iOS minimum
      comfortable: 48, // Android minimum
      large: 56,
    },
    "tablet-7": {
      minimum: 48,
      comfortable: 56,
      large: 64,
    },
    "tablet-10": {
      minimum: 52,
      comfortable: 60,
      large: 68,
    },
  },

  // Content width constraints for optimal readability
  maxContentWidths: {
    phone: "100%",
    "tablet-7": 680,
    "tablet-10": 800,
  },

  // Grid configurations
  gridColumns: {
    phone: 1,
    "tablet-7": 2,
    "tablet-10": 3,
  },
} as const;

/**
 * Generate responsive spacing based on device type
 */
export function getResponsiveSpacing(
  deviceType: DeviceType,
): ResponsiveSpacing {
  const multiplier = RESPONSIVE_TOKENS.spacingMultipliers[deviceType];

  // Base spacing values (8px grid system)
  const base = {
    xs: 4,
    sm: 8,
    md: 16,
    lg: 24,
    xl: 32,
    xxl: 48,
  };

  // Apply multiplier to base values
  const scaled = Object.entries(base).reduce(
    (acc, [key, value]) => ({
      ...acc,
      [key]: Math.round(value * multiplier),
    }),
    {} as typeof base,
  );

  // Add semantic spacing values
  return {
    ...scaled,
    section: scaled.xl,
    card: scaled.lg,
    container: scaled.md,
    modal: scaled.lg,
  };
}

/**
 * Generate responsive typography based on device type
 */
export function getResponsiveTypography(
  deviceType: DeviceType,
): ResponsiveTypography {
  const multiplier = RESPONSIVE_TOKENS.typographyMultipliers[deviceType];

  // Base font sizes following Material Design and iOS HIG
  const baseSizes = {
    caption: 12,
    body: 16,
    bodyLarge: 18,
    subtitle: 20,
    title: 24,
    headline: 32,
    display: 40,
  };

  // Apply multiplier to base sizes
  const scaledSizes = Object.entries(baseSizes).reduce(
    (acc, [key, value]) => ({
      ...acc,
      [key]: Math.round(value * multiplier),
    }),
    {} as typeof baseSizes,
  );

  // Calculate line heights (1.4-1.6x font size for optimal readability)
  const lineHeights = Object.entries(scaledSizes).reduce(
    (acc, [key, value]) => ({
      ...acc,
      [key]: Math.round(value * 1.5),
    }),
    {} as typeof scaledSizes,
  );

  return {
    sizes: scaledSizes,
    lineHeights,
    weights: {
      regular: "400",
      medium: "500",
      semibold: "600",
      bold: "700",
    },
  };
}

/**
 * Generate responsive layout values based on device type
 */
export function getResponsiveLayout(deviceType: DeviceType): ResponsiveLayout {
  const baseLayout = {
    phone: {
      maxContentWidth: 0, // Use full width
      gridColumns: 1,
      cardMinWidth: 280,
      modalMaxWidth: 0, // Use full width
      headerHeight: 56,
      tabBarHeight: 49, // iOS standard
    },
    "tablet-7": {
      maxContentWidth: 680,
      gridColumns: 2,
      cardMinWidth: 300,
      modalMaxWidth: 500,
      headerHeight: 64,
      tabBarHeight: 56,
    },
    "tablet-10": {
      maxContentWidth: 800,
      gridColumns: 3,
      cardMinWidth: 320,
      modalMaxWidth: 600,
      sidebarWidth: 280,
      headerHeight: 72,
      tabBarHeight: 64,
    },
  };

  return baseLayout[deviceType];
}

/**
 * Generate responsive touch targets based on device type
 */
export function getResponsiveTouchTargets(
  deviceType: DeviceType,
): ResponsiveTouchTargets {
  return RESPONSIVE_TOKENS.touchTargets[deviceType];
}

/**
 * Responsive breakpoint utilities
 */
export const BREAKPOINTS = {
  phone: 0,
  tablet7: 768,
  tablet10: 1024,
} as const;

/**
 * Helper function to create responsive styles
 * Usage: createResponsiveStyle({ phone: 16, tablet: 20 })
 */
export function createResponsiveValue<T>(
  values: Partial<Record<DeviceType, T>>,
  defaultValue: T,
) {
  return (deviceType: DeviceType): T => {
    return values[deviceType] ?? defaultValue;
  };
}

/**
 * Common responsive patterns for layouts
 */
export const RESPONSIVE_PATTERNS = {
  // Card layouts
  cardGrid: {
    phone: { flexDirection: "column" as const, gap: 12 },
    "tablet-7": {
      flexDirection: "row" as const,
      flexWrap: "wrap" as const,
      gap: 16,
    },
    "tablet-10": {
      flexDirection: "row" as const,
      flexWrap: "wrap" as const,
      gap: 20,
    },
  },

  // Modal sizes
  modal: {
    phone: { width: "100%", height: "100%" },
    "tablet-7": { width: "80%", maxWidth: 500, height: "auto" },
    "tablet-10": { width: "70%", maxWidth: 600, height: "auto" },
  },

  // Content containers
  container: {
    phone: { paddingHorizontal: 16 },
    "tablet-7": { paddingHorizontal: 24, maxWidth: 680 },
    "tablet-10": { paddingHorizontal: 32, maxWidth: 800 },
  },

  // Form layouts
  form: {
    phone: { width: "100%" },
    "tablet-7": { width: "100%", maxWidth: 480 },
    "tablet-10": { width: "100%", maxWidth: 520 },
  },
} as const;

/**
 * Responsive border radius values
 */
export function getResponsiveBorderRadius(deviceType: DeviceType) {
  const multiplier = RESPONSIVE_TOKENS.spacingMultipliers[deviceType];

  const base = {
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
    full: 9999,
  };

  return Object.entries(base).reduce(
    (acc, [key, value]) => ({
      ...acc,
      [key]: key === "full" ? value : Math.round(value * multiplier),
    }),
    {} as typeof base,
  );
}

/**
 * Responsive shadow values
 */
export function getResponsiveShadows(deviceType: DeviceType) {
  const intensity =
    deviceType === "phone" ? 1 : deviceType === "tablet-7" ? 1.2 : 1.4;

  return {
    sm: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: Math.round(1 * intensity) },
      shadowOpacity: 0.05,
      shadowRadius: Math.round(3 * intensity),
      elevation: Math.round(2 * intensity),
    },
    md: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: Math.round(2 * intensity) },
      shadowOpacity: 0.08,
      shadowRadius: Math.round(8 * intensity),
      elevation: Math.round(4 * intensity),
    },
    lg: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: Math.round(4 * intensity) },
      shadowOpacity: 0.12,
      shadowRadius: Math.round(16 * intensity),
      elevation: Math.round(8 * intensity),
    },
  };
}
