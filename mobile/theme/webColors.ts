// theme/webColors.ts

/**
 * Web Color System
 *
 * These colors are imported from the web app to maintain
 * visual consistency across platforms. Use sparingly in
 * the mobile app to create subtle connections to the web experience.
 */

export const WebColors = {
  /**
   * Warm beige background from web
   * Used as the main background color on the web landing page
   */
  background: "#FDF6E3",

  /**
   * Dark gray text color from web
   * Primary text color used throughout the web app
   */
  foreground: "#2D3436",

  /**
   * Primary orange from web
   * Main brand color (already matches mobile)
   */
  primaryOrange: "#FF7B54",

  /**
   * Secondary gray from web
   * Used for secondary text and subtle UI elements
   */
  secondaryGray: "#636E72",
} as const;

// Helper function to apply web background with opacity
export const getWebBackgroundWithOpacity = (opacity: number) => {
  // Convert hex to RGB
  const r = parseInt(WebColors.background.slice(1, 3), 16);
  const g = parseInt(WebColors.background.slice(3, 5), 16);
  const b = parseInt(WebColors.background.slice(5, 7), 16);

  return `rgba(${r}, ${g}, ${b}, ${opacity})`;
};
