import { Dimensions, Platform } from "react-native";

// Screen size breakpoints based on common mobile dimensions
export const BREAKPOINTS = {
  SMALL: 359, // <= 359px width (small phones, iPhone SE)
  MEDIUM: 389, // 360-389px width (standard Android)
  LARGE: 390, // >= 390px width (modern iPhones, large Android)
} as const;

// Get current screen dimensions
export const getScreenDimensions = () => {
  const { width, height } = Dimensions.get("window");
  return { width, height };
};

// Determine screen size category - requires explicit width parameter
export const getScreenSizeCategory = (width: number) => {
  if (width <= BREAKPOINTS.SMALL) return "small";
  if (width <= BREAKPOINTS.MEDIUM) return "medium";
  return "large";
};

// Responsive spacing based on screen size - requires explicit width parameter
export const getResponsiveSpacing = (width: number) => {
  const category = getScreenSizeCategory(width);

  switch (category) {
    case "small":
      return {
        horizontal: 16,
        vertical: 12,
        formPadding: 16,
        elementSpacing: 8,
        buttonHeight: 48,
      };
    case "medium":
      return {
        horizontal: 20,
        vertical: 16,
        formPadding: 20,
        elementSpacing: 12,
        buttonHeight: 52,
      };
    case "large":
    default:
      return {
        horizontal: 24,
        vertical: 20,
        formPadding: 24,
        elementSpacing: 16,
        buttonHeight: 56,
      };
  }
};

// Responsive font sizes - requires explicit width parameter
export const getResponsiveFontSizes = (width: number) => {
  const category = getScreenSizeCategory(width);

  switch (category) {
    case "small":
      return {
        title: 24,
        body: 16,
        label: 14,
        caption: 12,
        button: 16,
      };
    case "medium":
      return {
        title: 26,
        body: 16,
        label: 14,
        caption: 12,
        button: 16,
      };
    case "large":
    default:
      return {
        title: 28,
        body: 17,
        label: 15,
        caption: 13,
        button: 17,
      };
  }
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
    default:
      return 0;
  }
};

// Responsive image dimensions for background - requires explicit width and height parameters
export const getResponsiveImageDimensions = (width: number, height: number) => {
  // Calculate maximum height for background image based on screen size
  const category = getScreenSizeCategory(width);

  switch (category) {
    case "small":
      return {
        maxHeight: height * 0.35, // Use less height on small screens
        opacity: 0.12, // Lighter opacity for better contrast
      };
    case "medium":
      return {
        maxHeight: height * 0.4,
        opacity: 0.15,
      };
    case "large":
    default:
      return {
        maxHeight: height * 0.45,
        opacity: 0.18,
      };
  }
};

// Helper to create responsive styles
export const createResponsiveStyle = <T extends Record<string, any>>(
  smallStyle: T,
  mediumStyle: T,
  largeStyle: T,
  width: number,
): T => {
  const category = getScreenSizeCategory(width);

  switch (category) {
    case "small":
      return smallStyle;
    case "medium":
      return mediumStyle;
    case "large":
    default:
      return largeStyle;
  }
};
