import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
  useMemo,
} from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { TextStyle, useColorScheme, useWindowDimensions } from "react-native";
import {
  getResponsiveFontSizes,
  getResponsiveSpacing,
  getScreenSizeCategory,
  getKeyboardBehavior,
  getKeyboardOffset,
  getScaledFontSize,
  getResponsiveValue,
} from "@/utils/responsive";

// Theme type definitions
export interface Theme {
  mode: "light" | "dark";
  colors: {
    // Backgrounds - subtle and clean
    background: {
      default: string; // Main screen background
      surface: string; // Card backgrounds
      elevated: string; // Elevated surfaces
      subtle: string; // Very subtle variations
      secondary: string; // Secondary background
    };
    // Brand colors - keep existing vibrant palette
    primary: string; // #FF7B54 - main orange
    success: string; // #4CAF50 - green
    error: string; // #FF6B6B - red
    premium: string; // #FFD700 - gold
    // Border colors
    border: {
      light: string; // Light border
      medium: string; // Medium border
      dark: string; // Dark border
    };
    // Text colors
    text: {
      primary: string; // Main text
      secondary: string; // Secondary text
      placeholder: string; // Placeholder text
      inverse: string; // Text on colored backgrounds
      disabled: string; // Disabled text
      tertiary: string; // Tertiary text
    };
    // UI elements
    ui: {
      card: string; // Card backgrounds
      border: string; // Borders
      divider: string; // Dividers
      disabled: string; // Disabled states
      inputBackground: string; // Input fields
    };
    // Semantic colors
    semantic: {
      info: string; // Information blue
      warning: string; // Warning orange
      special: string; // Purple accent
    };
    // State colors
    state: {
      correctBackground: string; // Light green
      incorrectBackground: string; // Light red
      infoBackground: string; // Light blue
      successDark: string; // Dark green
      errorDark: string; // Dark red
    };
    // Overlay colors
    overlay: {
      backdrop: string; // Modal backdrop
      backdropHeavy: string; // Darker backdrop
      backdropLight: string; // Light backdrop
    };
    roles: {
      action: string;
      actionPressed: string;
      onAction: string;
      brandAccent: string;
      brandAccentPressed: string;
      onBrandAccent: string;
      error: string;
      emphasisSurface: string;
      focus: string;
      disabledBackground: string;
      disabledText: string;
      scrim: string;
    };
  };
  spacing: {
    xs: number;
    sm: number;
    compact: number;
    md: number;
    comfortable: number;
    lg: number;
    xl: number;
    xxl: number;
  };
  borderRadius: {
    sm: number;
    md: number;
    lg: number;
    xl: number;
    full: number;
  };
  typography: Record<
    | "micro"
    | "caption"
    | "label"
    | "small"
    | "body"
    | "heading"
    | "title"
    | "feature"
    | "display",
    TextStyle
  >;
  geometry: {
    touchTarget: number;
    controlHeight: number;
  };
  motion: {
    fast: number;
    standard: number;
    sheet: number;
    pressScale: number;
  };
  shadows: {
    sm: {
      shadowColor: string;
      shadowOffset: { width: number; height: number };
      shadowOpacity: number;
      shadowRadius: number;
      elevation: number;
    };
    md: {
      shadowColor: string;
      shadowOffset: { width: number; height: number };
      shadowOpacity: number;
      shadowRadius: number;
      elevation: number;
    };
    lg: {
      shadowColor: string;
      shadowOffset: { width: number; height: number };
      shadowOpacity: number;
      shadowRadius: number;
      elevation: number;
    };
  };
}

const sharedTypography: Theme["typography"] = {
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
};

const sharedGeometry: Theme["geometry"] = {
  touchTarget: 44,
  controlHeight: 48,
};

const sharedMotion: Theme["motion"] = {
  fast: 130,
  standard: 180,
  sheet: 240,
  pressScale: 0.975,
};

// Light theme definition
const lightTheme: Theme = {
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
  typography: sharedTypography,
  geometry: sharedGeometry,
  motion: sharedMotion,
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

// Dark theme definition
const darkTheme: Theme = {
  mode: "dark",
  colors: {
    background: {
      default: "#0F0F0F", // Rich black
      surface: "#1C1C1E", // Dark surface
      elevated: "#2C2C2E", // Elevated dark surface
      subtle: "#242426", // Subtle variation
      secondary: "#2C2C2E", // Secondary background
    },
    // Keep vibrant colors - they pop on dark backgrounds
    primary: "#FF7B54",
    success: "#4CAF50",
    error: "#FF6B6B",
    premium: "#FFD700",
    border: {
      light: "#38383A",
      medium: "#48484A",
      dark: "#58585A",
    },
    text: {
      primary: "#FFFFFF",
      secondary: "#A0A0A0",
      placeholder: "#606060",
      inverse: "#000000",
      disabled: "#48484A",
      tertiary: "#606060",
    },
    ui: {
      card: "#1C1C1E",
      border: "#38383A",
      divider: "#2C2C2E",
      disabled: "#48484A",
      inputBackground: "#2C2C2E",
    },
    semantic: {
      info: "#2196F3",
      warning: "#FF6B3D",
      special: "#9C27B0",
    },
    state: {
      correctBackground: "rgba(76, 175, 80, 0.15)",
      incorrectBackground: "rgba(255, 107, 107, 0.15)",
      infoBackground: "rgba(33, 150, 243, 0.15)",
      successDark: "#4CAF50",
      errorDark: "#FF6B6B",
    },
    overlay: {
      backdrop: "rgba(0, 0, 0, 0.7)",
      backdropHeavy: "rgba(0, 0, 0, 0.8)",
      backdropLight: "rgba(0, 0, 0, 0.5)",
    },
    roles: {
      action: "#FF8D69",
      actionPressed: "#FFAC91",
      onAction: "#28140D",
      brandAccent: "#FF8D69",
      brandAccentPressed: "#FFAC91",
      onBrandAccent: "#28140D",
      error: "#FF6B6B",
      emphasisSurface: "#3A2720",
      focus: "#89B8DE",
      disabledBackground: "#394245",
      disabledText: "#C6C0B6",
      scrim: "rgba(0, 0, 0, 0.68)",
    },
  },
  spacing: lightTheme.spacing,
  borderRadius: lightTheme.borderRadius,
  typography: sharedTypography,
  geometry: sharedGeometry,
  motion: sharedMotion,
  shadows: {
    sm: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 1 },
      shadowOpacity: 0.3,
      shadowRadius: 3,
      elevation: 2,
    },
    md: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.4,
      shadowRadius: 8,
      elevation: 4,
    },
    lg: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.5,
      shadowRadius: 16,
      elevation: 8,
    },
  },
};

// Theme context
export interface ResponsiveValues {
  spacing: {
    horizontal: number;
    vertical: number;
    formPadding: number;
    elementSpacing: number;
    buttonHeight: number;
    minTouchTarget: number;
  };
  fontSizes: {
    display: number;
    title: number;
    headline: number;
    body: number;
    label: number;
    micro: number;
    lineHeight: number;
  };
  screenWidth: number;
  screenHeight: number;
  category: "small" | "medium" | "large" | "xlarge";
  keyboardBehavior: "height" | "position" | "padding";
  keyboardOffset: number;
  isSmallPhone: boolean;
  isMediumPhone: boolean;
  isLargePhone: boolean;
  isExtraLargePhone: boolean;
  getValueForSize: <T>(
    smallValue: T,
    mediumValue: T,
    largeValue: T,
    xlargeValue: T,
  ) => T;
  getScaledFont: (
    fontName: "display" | "title" | "headline" | "body" | "label" | "micro",
  ) => number;
}

interface ThemeContextType {
  theme: Theme;
  responsive: ResponsiveValues;
  toggleTheme: () => void;
  setThemeMode: (mode: "light" | "dark" | "system") => void;
  themeMode: "light" | "dark" | "system";
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

const THEME_STORAGE_KEY = "@decorebator_theme_mode";

// Theme provider component
export const ThemeProvider: React.FC<{ children: ReactNode }> = ({
  children,
}) => {
  const systemColorScheme = useColorScheme();
  const { width: screenWidth, height: screenHeight } = useWindowDimensions();
  const [themeMode, setThemeModeState] = useState<"light" | "dark" | "system">(
    "system",
  );
  const [theme, setTheme] = useState<Theme>(lightTheme);

  // Load saved theme preference
  useEffect(() => {
    const loadThemePreference = async () => {
      try {
        const savedMode = await AsyncStorage.getItem(THEME_STORAGE_KEY);
        if (savedMode && ["light", "dark", "system"].includes(savedMode)) {
          setThemeModeState(savedMode as "light" | "dark" | "system");
        }
      } catch (error) {
        console.error("Error loading theme preference:", error);
      }
    };
    loadThemePreference();
  }, []);

  // Update theme based on mode
  useEffect(() => {
    let selectedTheme: Theme;

    if (themeMode === "system") {
      selectedTheme = systemColorScheme === "dark" ? darkTheme : lightTheme;
    } else {
      selectedTheme = themeMode === "dark" ? darkTheme : lightTheme;
    }

    setTheme(selectedTheme);
  }, [themeMode, systemColorScheme]);

  const toggleTheme = () => {
    const newMode = theme.mode === "light" ? "dark" : "light";
    setThemeMode(newMode);
  };

  const setThemeMode = async (mode: "light" | "dark" | "system") => {
    try {
      await AsyncStorage.setItem(THEME_STORAGE_KEY, mode);
      setThemeModeState(mode);
    } catch (error) {
      console.error("Error saving theme preference:", error);
    }
  };

  // Calculate responsive values - memoized for performance
  const responsive: ResponsiveValues = useMemo(() => {
    const category = getScreenSizeCategory(screenWidth);
    const spacing = getResponsiveSpacing(screenWidth);
    const fontSizes = getResponsiveFontSizes(screenWidth);
    const keyboardBehavior = getKeyboardBehavior();
    const keyboardOffset = getKeyboardOffset(screenWidth);

    // Device type detection
    const isSmallPhone = category === "small";
    const isMediumPhone = category === "medium";
    const isLargePhone = category === "large";
    const isExtraLargePhone = category === "xlarge";

    // Helper function to get responsive value based on current screen size
    function getValueForSize<T>(
      smallValue: T,
      mediumValue: T,
      largeValue: T,
      xlargeValue: T,
    ): T {
      return getResponsiveValue(
        smallValue,
        mediumValue,
        largeValue,
        xlargeValue,
        screenWidth,
      );
    }

    // Helper function to get scaled font size for specific typography level
    function getScaledFont(
      fontName: "display" | "title" | "headline" | "body" | "label" | "micro",
    ): number {
      return getScaledFontSize(fontName, screenWidth);
    }

    return {
      screenWidth,
      screenHeight,
      category,
      spacing,
      fontSizes,
      keyboardBehavior,
      keyboardOffset,
      getValueForSize,
      getScaledFont,
      isSmallPhone,
      isMediumPhone,
      isLargePhone,
      isExtraLargePhone,
    };
  }, [screenWidth, screenHeight]);

  return (
    <ThemeContext.Provider
      value={{ theme, responsive, toggleTheme, setThemeMode, themeMode }}
    >
      {children}
    </ThemeContext.Provider>
  );
};

// Hook to use theme
export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }
  return context;
};
