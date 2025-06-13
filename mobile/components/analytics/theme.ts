// Shared color palette for analytics components
export const colors = {
  primary: "#FF7B54",
  success: "#4CAF50",
  error: "#FF6B6B",
  gold: "#FFD700",
  background: "#FDF6E3",
  backgroundLight: "#FFF9F0",
  backgroundPeach: "#FFE8D6",
  backgroundOrange: "#FFDCC3",
  backgroundSage: "#F5F0E6",
  textDark: "#2D3436",
  textMedium: "#636E72",
  textLight: "#B2BEC3",
  white: "#FFFFFF",
  lightBackground: "#FAFAFA",
  borderGray: "#E0E0E0",
  divider: "#F0F0F0",
};

export const chartColors = [
  colors.primary,
  colors.success,
  colors.gold,
  "#9C27B0",
  "#2196F3",
  "#FF6B3D",
];

export const chartConfig = {
  backgroundColor: colors.white,
  backgroundGradientFrom: colors.white,
  backgroundGradientTo: colors.white,
  decimalPlaces: 0,
  color: (opacity = 1) => `rgba(255, 123, 84, ${opacity})`,
  labelColor: (opacity = 1) => `rgba(45, 52, 54, ${opacity})`,
  style: {
    borderRadius: 16,
  },
  propsForBackgroundLines: {
    strokeDasharray: "",
    stroke: colors.divider,
  },
};
