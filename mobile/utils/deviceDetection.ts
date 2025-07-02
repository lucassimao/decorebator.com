import { Dimensions, Platform } from "react-native";

export type DeviceType = "phone" | "tablet-7" | "tablet-10";
export type DeviceOrientation = "portrait" | "landscape";

export interface DeviceInfo {
  type: DeviceType;
  orientation: DeviceOrientation;
  width: number;
  height: number;
  isTablet: boolean;
  isPhone: boolean;
  devicePixelRatio: number;
}

// Breakpoints based on physical screen dimensions
// These are based on common tablet sizes and UX best practices
const BREAKPOINTS = {
  PHONE_MAX: 767, // < 768px = phone
  TABLET_7_MAX: 1023, // 768px - 1023px = 7" tablet
  TABLET_10_MIN: 1024, // >= 1024px = 10" tablet
} as const;

/**
 * Gets the current device type based on screen dimensions
 * Uses the smaller dimension (width in portrait, height in landscape)
 * to ensure consistent classification regardless of orientation
 */
export function getDeviceType(): DeviceType {
  const { width, height } = Dimensions.get("window");
  const smallerDimension = Math.min(width, height);

  if (smallerDimension < BREAKPOINTS.PHONE_MAX) {
    return "phone";
  } else if (smallerDimension <= BREAKPOINTS.TABLET_7_MAX) {
    return "tablet-7";
  } else {
    return "tablet-10";
  }
}

/**
 * Gets the current device orientation
 */
export function getDeviceOrientation(): DeviceOrientation {
  const { width, height } = Dimensions.get("window");
  return width > height ? "landscape" : "portrait";
}

/**
 * Gets comprehensive device information
 */
export function getDeviceInfo(): DeviceInfo {
  const { width, height } = Dimensions.get("window");
  const type = getDeviceType();
  const orientation = getDeviceOrientation();

  return {
    type,
    orientation,
    width,
    height,
    isTablet: type === "tablet-7" || type === "tablet-10",
    isPhone: type === "phone",
    devicePixelRatio: Dimensions.get("window").scale,
  };
}

/**
 * Utility function to check if current device is a tablet
 */
export function isTablet(): boolean {
  const deviceType = getDeviceType();
  return deviceType === "tablet-7" || deviceType === "tablet-10";
}

/**
 * Utility function to check if current device is a phone
 */
export function isPhone(): boolean {
  return getDeviceType() === "phone";
}

/**
 * Utility function to check if current device is a specific tablet size
 */
export function isTablet7(): boolean {
  return getDeviceType() === "tablet-7";
}

export function isTablet10(): boolean {
  return getDeviceType() === "tablet-10";
}

/**
 * Utility function to check if device is in landscape mode
 */
export function isLandscape(): boolean {
  return getDeviceOrientation() === "landscape";
}

/**
 * Utility function to check if device is in portrait mode
 */
export function isPortrait(): boolean {
  return getDeviceOrientation() === "portrait";
}

/**
 * Get platform-specific information for responsive design
 */
export function getPlatformInfo() {
  return {
    isIOS: Platform.OS === "ios",
    isAndroid: Platform.OS === "android",
    platform: Platform.OS,
  };
}

/**
 * Calculate responsive dimensions based on device type
 * Returns object with commonly needed responsive values
 */
export function getResponsiveDimensions() {
  const { width, height, type } = getDeviceInfo();
  const isTabletDevice = type === "tablet-7" || type === "tablet-10";

  // Content width constraints for optimal readability
  // Based on Material Design and Human Interface Guidelines
  const getOptimalContentWidth = () => {
    if (type === "phone") {
      return width - 40; // 20px margins on phones
    } else if (type === "tablet-7") {
      return Math.min(width - 80, 600); // Max 600px content width on 7" tablets
    } else {
      return Math.min(width - 120, 800); // Max 800px content width on 10" tablets
    }
  };

  // Grid columns for list layouts
  const getGridColumns = () => {
    if (type === "phone") return 1;
    if (type === "tablet-7") return 2;
    return 3; // tablet-10
  };

  return {
    screenWidth: width,
    screenHeight: height,
    contentWidth: getOptimalContentWidth(),
    gridColumns: getGridColumns(),
    isTablet: isTabletDevice,
    // Responsive spacing multipliers
    spacingMultiplier: isTabletDevice ? 1.5 : 1,
    // Touch target sizes (minimum 44pt/48dp)
    minTouchTarget: isTabletDevice ? 48 : 44,
  };
}
