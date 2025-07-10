import { useMemo } from "react";
import {
  getScreenDimensions,
  getScreenSizeCategory,
  getResponsiveSpacing,
  getResponsiveFontSizes,
  getResponsiveImageDimensions,
  getKeyboardBehavior,
  getKeyboardOffset,
} from "@/utils/responsive";

/**
 * Hook that provides memoized responsive values based on screen dimensions
 * Reduces redundant calculations and improves performance
 */
export const useResponsive = () => {
  const { width: screenWidth, height: screenHeight } = getScreenDimensions();

  return useMemo(() => {
    const category = getScreenSizeCategory(screenWidth);
    const spacing = getResponsiveSpacing(screenWidth);
    const fontSizes = getResponsiveFontSizes(screenWidth);
    const imageConfig = getResponsiveImageDimensions(screenWidth, screenHeight);
    const keyboardBehavior = getKeyboardBehavior();
    const keyboardOffset = getKeyboardOffset(screenWidth);

    return {
      screenWidth,
      screenHeight,
      category,
      spacing,
      fontSizes,
      imageConfig,
      keyboardBehavior,
      keyboardOffset,
    };
  }, [screenWidth, screenHeight]);
};
