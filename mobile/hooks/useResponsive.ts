import { useState, useEffect } from "react";
import { Dimensions } from "react-native";
import {
  DeviceInfo,
  getDeviceInfo,
  getResponsiveDimensions,
  getPlatformInfo,
} from "@/utils/deviceDetection";

export interface ResponsiveInfo extends DeviceInfo {
  contentWidth: number;
  gridColumns: number;
  spacingMultiplier: number;
  minTouchTarget: number;
  platform: {
    isIOS: boolean;
    isAndroid: boolean;
    platform: string;
  };
}

/**
 * Hook that provides reactive device information and responsive utilities
 * Updates automatically when device orientation changes
 */
export function useResponsive(): ResponsiveInfo {
  const [deviceInfo, setDeviceInfo] = useState<DeviceInfo>(getDeviceInfo());

  useEffect(() => {
    const subscription = Dimensions.addEventListener("change", () => {
      setDeviceInfo(getDeviceInfo());
    });

    return () => subscription?.remove();
  }, []);

  const responsiveDimensions = getResponsiveDimensions();
  const platformInfo = getPlatformInfo();

  return {
    ...deviceInfo,
    ...responsiveDimensions,
    platform: platformInfo,
  };
}

/**
 * Hook that provides device type specific values
 * Useful for conditional rendering based on device type
 */
export function useDeviceType() {
  const { type, isTablet, isPhone } = useResponsive();

  return {
    type,
    isTablet,
    isPhone,
    isTablet7: type === "tablet-7",
    isTablet10: type === "tablet-10",
  };
}

/**
 * Hook that provides responsive spacing values
 * Returns spacing values that scale appropriately for device type
 */
export function useResponsiveSpacing() {
  const { spacingMultiplier, type } = useResponsive();

  // Base spacing values (matching theme spacing)
  const baseSpacing = {
    xs: 4,
    sm: 8,
    md: 16,
    lg: 24,
    xl: 32,
    xxl: 48,
  };

  // Apply responsive multiplier
  const responsiveSpacing = Object.entries(baseSpacing).reduce(
    (acc, [key, value]) => ({
      ...acc,
      [key]: Math.round(value * spacingMultiplier),
    }),
    {} as typeof baseSpacing,
  );

  // Additional tablet-specific spacing
  const getAdditionalSpacing = () => {
    if (type === "phone") {
      return {
        section: responsiveSpacing.lg,
        card: responsiveSpacing.md,
        container: responsiveSpacing.md,
      };
    } else if (type === "tablet-7") {
      return {
        section: responsiveSpacing.xl,
        card: responsiveSpacing.lg,
        container: responsiveSpacing.lg,
      };
    } else {
      return {
        section: responsiveSpacing.xxl,
        card: responsiveSpacing.xl,
        container: responsiveSpacing.xl,
      };
    }
  };

  return {
    ...responsiveSpacing,
    ...getAdditionalSpacing(),
    multiplier: spacingMultiplier,
  };
}

/**
 * Hook that provides responsive typography values
 * Returns font sizes that scale appropriately for device type and screen size
 */
export function useResponsiveTypography() {
  const { type } = useResponsive();

  // Base typography scale
  const baseSizes = {
    caption: 12,
    body: 16,
    bodyLarge: 18,
    subtitle: 20,
    title: 24,
    headline: 32,
    display: 40,
  };

  // Typography multipliers based on device type
  const getMultiplier = () => {
    if (type === "phone") return 1;
    if (type === "tablet-7") return 1.1;
    return 1.2; // tablet-10
  };

  const multiplier = getMultiplier();

  const responsiveSizes = Object.entries(baseSizes).reduce(
    (acc, [key, value]) => ({
      ...acc,
      [key]: Math.round(value * multiplier),
    }),
    {} as typeof baseSizes,
  );

  // Line heights (1.5x font size for readability)
  const lineHeights = Object.entries(responsiveSizes).reduce(
    (acc, [key, value]) => ({
      ...acc,
      [key]: Math.round(value * 1.5),
    }),
    {} as Record<keyof typeof baseSizes, number>,
  );

  return {
    sizes: responsiveSizes,
    lineHeights,
    multiplier,
  };
}

/**
 * Hook that provides breakpoint utilities
 * Useful for conditional styles and layouts
 */
export function useBreakpoints() {
  const { width, type, orientation } = useResponsive();

  const breakpoints = {
    isPhone: type === "phone",
    isTablet7: type === "tablet-7",
    isTablet10: type === "tablet-10",
    isTablet: type === "tablet-7" || type === "tablet-10",
    isPortrait: orientation === "portrait",
    isLandscape: orientation === "landscape",
    isSmallScreen: width < 600,
    isMediumScreen: width >= 600 && width < 1024,
    isLargeScreen: width >= 1024,
  };

  return breakpoints;
}

/**
 * Hook that provides responsive grid utilities
 * Returns grid configurations for different layouts
 */
export function useResponsiveGrid() {
  const { gridColumns, contentWidth, type } = useResponsive();
  const spacing = useResponsiveSpacing();

  const getItemWidth = (columns: number = gridColumns) => {
    const gaps = (columns - 1) * spacing.md;
    return (contentWidth - gaps) / columns;
  };

  const getCardSpacing = () => {
    if (type === "phone") return spacing.sm;
    if (type === "tablet-7") return spacing.md;
    return spacing.lg;
  };

  return {
    columns: gridColumns,
    itemWidth: getItemWidth(),
    cardSpacing: getCardSpacing(),
    getItemWidth,
    // Predefined grid configurations
    configs: {
      single: { columns: 1, itemWidth: getItemWidth(1) },
      double: { columns: 2, itemWidth: getItemWidth(2) },
      triple: { columns: 3, itemWidth: getItemWidth(3) },
      auto: { columns: gridColumns, itemWidth: getItemWidth() },
    },
  };
}

/**
 * Utility hook for responsive values
 * Allows defining different values for different device types
 */
export function useResponsiveValue<T>(values: {
  phone: T;
  tablet7?: T;
  tablet10?: T;
  tablet?: T; // fallback for both tablet sizes
}): T {
  const { type } = useResponsive();

  if (type === "phone") {
    return values.phone;
  } else if (type === "tablet-7") {
    return values.tablet7 ?? values.tablet ?? values.phone;
  } else {
    // tablet-10
    return values.tablet10 ?? values.tablet ?? values.phone;
  }
}
