import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { useColorScheme } from "react-native";
import {
  getResponsiveSpacing,
  getResponsiveTypography,
  getResponsiveLayout,
  getResponsiveTouchTargets,
  getResponsiveBorderRadius,
  getResponsiveShadows,
  ResponsiveSpacing,
  ResponsiveTypography,
  ResponsiveLayout,
  ResponsiveTouchTargets,
} from "@/theme/responsive";
import { getDeviceType, DeviceType } from "@/utils/deviceDetection";
import { Dimensions } from "react-native";

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
  };
  // Responsive design tokens
  spacing: ResponsiveSpacing;
  typography: ResponsiveTypography;
  layout: ResponsiveLayout;
  touchTargets: ResponsiveTouchTargets;
  borderRadius: {
    sm: number;
    md: number;
    lg: number;
    xl: number;
    full: number;
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

// Base color definitions (same for all devices)
const lightColors = {
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
};

const darkColors = {
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
};

// Helper function to create responsive theme
function createResponsiveTheme(
  mode: "light" | "dark",
  deviceType: DeviceType,
): Theme {
  const colors = mode === "light" ? lightColors : darkColors;

  return {
    mode,
    colors,
    spacing: getResponsiveSpacing(deviceType),
    typography: getResponsiveTypography(deviceType),
    layout: getResponsiveLayout(deviceType),
    touchTargets: getResponsiveTouchTargets(deviceType),
    borderRadius: getResponsiveBorderRadius(deviceType),
    shadows: getResponsiveShadows(deviceType),
  };
}

// Default themes for fallback (using phone settings)
const lightTheme: Theme = createResponsiveTheme("light", "phone");
// const darkTheme: Theme = createResponsiveTheme("dark", "phone");

// Theme context
interface ThemeContextType {
  theme: Theme;
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
  const [themeMode, setThemeModeState] = useState<"light" | "dark" | "system">(
    "system",
  );
  const [theme, setTheme] = useState<Theme>(lightTheme);
  const [deviceType, setDeviceType] = useState<DeviceType>(getDeviceType());

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

  // Track device type changes
  useEffect(() => {
    const subscription = Dimensions.addEventListener("change", () => {
      setDeviceType(getDeviceType());
    });

    return () => subscription?.remove();
  }, []);

  // Update theme based on mode and device type
  useEffect(() => {
    let mode: "light" | "dark";

    if (themeMode === "system") {
      mode = systemColorScheme === "dark" ? "dark" : "light";
    } else {
      mode = themeMode;
    }

    // Create responsive theme based on current device type
    const responsiveTheme = createResponsiveTheme(mode, deviceType);
    setTheme(responsiveTheme);
  }, [themeMode, systemColorScheme, deviceType]);

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

  return (
    <ThemeContext.Provider
      value={{ theme, toggleTheme, setThemeMode, themeMode }}
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
