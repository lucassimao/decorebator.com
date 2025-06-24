import { Theme } from "@/contexts/ThemeContext";

// Chart colors that work well in both light and dark themes
export const getChartColors = (theme: Theme) => {
  const hexToRgba = (hex: string, opacity: number = 1) => {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    if (result) {
      const r = parseInt(result[1], 16);
      const g = parseInt(result[2], 16);
      const b = parseInt(result[3], 16);
      return `rgba(${r}, ${g}, ${b}, ${opacity})`;
    }
    return hex;
  };

  return {
    primary: (opacity = 1) => hexToRgba(theme.colors.primary, opacity),
    secondary: (opacity = 1) => hexToRgba("#4A90E2", opacity), // Blue
    success: (opacity = 1) => hexToRgba(theme.colors.success, opacity),
    warning: (opacity = 1) => hexToRgba("#FFD700", opacity), // Gold
    info: (opacity = 1) => hexToRgba("#9C27B0", opacity), // Purple
    error: (opacity = 1) => hexToRgba("#FF6B3D", opacity), // Orange
    colors: [
      theme.colors.primary,
      theme.colors.success,
      "#FFD700", // gold
      "#9C27B0", // purple
      "#2196F3", // blue
      "#FF6B3D", // orange
    ],
  };
};

export const getChartConfig = (theme: Theme) => ({
  backgroundColor: theme.colors.background.surface,
  backgroundGradientFrom: theme.colors.background.surface,
  backgroundGradientTo: theme.colors.background.surface,
  decimalPlaces: 0,
  color: (opacity = 1) => {
    const rgb = theme.mode === "light" ? "255, 123, 84" : "255, 139, 100";
    return `rgba(${rgb}, ${opacity})`;
  },
  labelColor: (opacity = 1) => {
    const rgb = theme.mode === "light" ? "45, 52, 54" : "200, 200, 200";
    return `rgba(${rgb}, ${opacity})`;
  },
  style: {
    borderRadius: 16,
  },
  propsForBackgroundLines: {
    strokeDasharray: "",
    stroke: theme.colors.ui.divider,
  },
});
