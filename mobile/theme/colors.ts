// theme/colors.ts

/**
 * Color System for WordWise Language Learning App
 *
 * This color palette creates a warm, inviting learning environment
 * while maintaining clear visual hierarchy and accessibility.
 */

export const Colors = {
  // ===============================
  // BRAND COLORS
  // ===============================

  /**
   * Primary brand color - Used for main CTAs, primary buttons,
   * active states, and key interactive elements
   */
  primary: "#FF7B54",

  /**
   * Success color - Used for correct answers, progress indicators,
   * completion states, and positive feedback
   */
  success: "#4CAF50",

  /**
   * Error color - Used for incorrect answers, validation errors,
   * delete actions, and warning states
   */
  error: "#FF6B6B",

  /**
   * Premium accent - Used for premium features, subscription highlights,
   * achievements, and special badges
   */
  premium: "#FFD700",

  // ===============================
  // BACKGROUND COLORS
  // ===============================

  background: {
    /**
     * Main app background - The primary warm beige used throughout
     * the app for main screen backgrounds
     */
    primary: "#FDF6E3",

    /**
     * Lighter cream variant - Used in gradients and for subtle
     * background variations
     */
    light: "#FFF9F0",

    /**
     * Peachy beige accent - Used in gradients for warmth and
     * visual interest
     */
    accent: "#FFE8D6",

    /**
     * Soft orange tint - Very subtle background accent for
     * special sections
     */
    soft: "#FFDCC3",

    /**
     * Sage-tinted beige - Alternative background with slight
     * green undertone for variety
     */
    sage: "#F5F0E6",
  },

  // ===============================
  // TEXT COLORS
  // ===============================

  text: {
    /**
     * Primary text - Used for headings, body text, and any
     * high-emphasis content
     */
    primary: "#2D3436",

    /**
     * Secondary text - Used for subtitles, labels, descriptions,
     * and medium-emphasis content
     */
    secondary: "#636E72",

    /**
     * Placeholder text - Used for input placeholders and
     * low-emphasis helper text
     */
    placeholder: "#B2BEC3",

    /**
     * White text - Used on dark backgrounds, colored buttons,
     * and inverted UI elements
     */
    white: "#FFFFFF",
  },

  // ===============================
  // UI ELEMENT COLORS
  // ===============================

  ui: {
    /**
     * Card background - Clean white for content cards, modals,
     * and elevated surfaces
     */
    card: "#FFFFFF",

    /**
     * Input background - Subtle gray for form inputs and
     * text fields
     */
    inputBackground: "#FAFAFA",

    /**
     * Default border - Standard border color for inputs,
     * cards, and dividers
     */
    border: "#E0E0E0",

    /**
     * Disabled state - Used for disabled buttons, inputs,
     * and non-interactive elements
     */
    disabled: "#DFE6E9",

    /**
     * Subtle background - Very light gray for section backgrounds
     * and subtle UI elements
     */
    subtleBackground: "#F5F5F5",

    /**
     * Light divider - Extremely light separator for subtle
     * content divisions
     */
    divider: "#F0F0F0",
  },

  // ===============================
  // SEMANTIC COLORS
  // ===============================

  semantic: {
    /**
     * Information blue - Used for informational messages,
     * links, and neutral highlights
     */
    info: "#2196F3",

    /**
     * Warning orange - Used for warnings, streaks, and
     * attention-grabbing elements
     */
    warning: "#FF6B3D",

    /**
     * Purple accent - Used for special features, unique
     * achievements, or distinctive UI elements
     */
    special: "#9C27B0",
  },

  // ===============================
  // STATE COLORS
  // ===============================

  state: {
    /**
     * Correct answer background - Light green background
     * for correct quiz answers
     */
    correctBackground: "#E8F5E9",

    /**
     * Incorrect answer background - Light red background
     * for incorrect quiz answers
     */
    incorrectBackground: "#FFEBEE",

    /**
     * Information background - Light blue background
     * for informational messages
     */
    infoBackground: "#F0F9FF",

    /**
     * Success text dark - Darker green for text on
     * success backgrounds
     */
    successDark: "#2E7D32",

    /**
     * Error text dark - Darker red for text on
     * error backgrounds
     */
    errorDark: "#C62828",
  },

  // ===============================
  // TRANSPARENCY COLORS
  // ===============================

  overlay: {
    /**
     * Modal backdrop - Dark semi-transparent overlay
     * for modals and dialogs
     */
    backdrop: "rgba(0, 0, 0, 0.4)",

    /**
     * Heavy backdrop - Darker overlay for important
     * modal states
     */
    backdropHeavy: "rgba(0, 0, 0, 0.5)",

    /**
     * Light backdrop - Subtle overlay for minor
     * UI overlays
     */
    backdropLight: "rgba(0, 0, 0, 0.3)",

    /**
     * White overlay - Semi-transparent white for
     * glass-morphism effects
     */
    white90: "rgba(255, 255, 255, 0.9)",
    white70: "rgba(255, 255, 255, 0.7)",
    white60: "rgba(255, 255, 255, 0.6)",
  },
} as const;

// Type for the color system
export type ColorSystem = typeof Colors;

// Helper type to get all color values
export type ColorValue =
  | typeof Colors.primary
  | typeof Colors.success
  | typeof Colors.error
  | typeof Colors.premium
  | (typeof Colors.background)[keyof typeof Colors.background]
  | (typeof Colors.text)[keyof typeof Colors.text]
  | (typeof Colors.ui)[keyof typeof Colors.ui]
  | (typeof Colors.semantic)[keyof typeof Colors.semantic]
  | (typeof Colors.state)[keyof typeof Colors.state]
  | (typeof Colors.overlay)[keyof typeof Colors.overlay];

// Usage examples in comments
/**
 * Usage Examples:
 *
 * // Import the colors
 * import { Colors } from '@/theme/colors';
 *
 * // Use in styles
 * const styles = StyleSheet.create({
 *   button: {
 *     backgroundColor: Colors.primary,
 *   },
 *   text: {
 *     color: Colors.text.primary,
 *   },
 *   card: {
 *     backgroundColor: Colors.ui.card,
 *     borderColor: Colors.ui.border,
 *   },
 * });
 *
 * // Use with styled-components (if applicable)
 * const Button = styled.TouchableOpacity`
 *   background-color: ${Colors.primary};
 * `;
 */
