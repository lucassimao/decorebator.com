import React from "react";
import { Animated, StyleSheet, View } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";

interface ShimmerElementProps {
  style: any;
  skeletonStyles: {
    elementColor: string;
    shimmerHighlight: string;
  };
  shimmerAnim: Animated.Value;
}

const ShimmerElement: React.FC<ShimmerElementProps> = ({
  style,
  skeletonStyles,
  shimmerAnim,
}) => {
  const shimmerOpacity = shimmerAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [0.4, 1.0],
  });

  const backgroundColor = shimmerAnim.interpolate({
    inputRange: [0, 0.5, 1],
    outputRange: [
      skeletonStyles.elementColor,
      skeletonStyles.shimmerHighlight,
      skeletonStyles.elementColor,
    ],
  });

  return (
    <Animated.View
      style={[
        style,
        {
          backgroundColor,
          opacity: shimmerOpacity,
        },
      ]}
    />
  );
};

export const StatsSkeleton: React.FC = React.memo(() => {
  const { theme } = useTheme();

  // Internal shimmer animation
  const shimmerAnim = React.useRef(new Animated.Value(0)).current;

  // Shimmer animation loop
  React.useEffect(() => {
    const shimmerLoop = Animated.loop(
      Animated.sequence([
        Animated.timing(shimmerAnim, {
          toValue: 1,
          duration: 1000,
          useNativeDriver: false, // Using backgroundColor interpolation
        }),
        Animated.timing(shimmerAnim, {
          toValue: 0,
          duration: 1000,
          useNativeDriver: false,
        }),
      ]),
    );
    shimmerLoop.start();

    return () => shimmerLoop.stop();
  }, [shimmerAnim]);

  // Get theme-aware skeleton styles
  const getSkeletonStyles = () => {
    const isDark = theme.mode === "dark";
    return {
      elementColor: isDark
        ? "rgba(255, 255, 255, 0.08)"
        : "rgba(0, 0, 0, 0.08)",
      shimmerHighlight: isDark
        ? "rgba(255, 255, 255, 0.15)"
        : "rgba(0, 0, 0, 0.15)",
      progressFillColor: isDark
        ? "rgba(34, 197, 94, 0.3)" // theme.colors.success with opacity
        : "rgba(34, 197, 94, 0.2)",
    };
  };

  const skeletonStyles = getSkeletonStyles();
  const styles = createStyles(theme, skeletonStyles);

  return (
    <View style={styles.statsContainer}>
      {/* Progress Overview Skeleton */}
      <View style={styles.progressOverview}>
        <ShimmerElement
          style={styles.progressLabel}
          skeletonStyles={skeletonStyles}
          shimmerAnim={shimmerAnim}
        />
        <View style={styles.progressBarContainer}>
          <View style={styles.progressBarBackground}>
            <Animated.View style={styles.progressBarFill} />
          </View>
          <ShimmerElement
            style={styles.progressPercentage}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>

        <View style={styles.progressFooter}>
          <ShimmerElement
            style={styles.motivationalText}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <View style={styles.streakContainer}>
            <ShimmerElement
              style={styles.streakIcon}
              skeletonStyles={skeletonStyles}
              shimmerAnim={shimmerAnim}
            />
            <ShimmerElement
              style={styles.streakText}
              skeletonStyles={skeletonStyles}
              shimmerAnim={shimmerAnim}
            />
          </View>
        </View>
      </View>

      {/* Stats Grid Skeleton */}
      <View style={styles.statsGrid}>
        <View style={[styles.statItem, styles.statItemFirst]}>
          <ShimmerElement
            style={styles.iconContainer}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <View style={styles.labelContainer}>
            <ShimmerElement
              style={styles.statLabel}
              skeletonStyles={skeletonStyles}
              shimmerAnim={shimmerAnim}
            />
          </View>
          <ShimmerElement
            style={styles.statValue}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>

        <View style={styles.statDivider} />

        <View style={styles.statItem}>
          <ShimmerElement
            style={styles.iconContainer}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <View style={styles.labelContainer}>
            <ShimmerElement
              style={styles.statLabel}
              skeletonStyles={skeletonStyles}
              shimmerAnim={shimmerAnim}
            />
          </View>
          <ShimmerElement
            style={styles.statValue}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>

        <View style={styles.statDivider} />

        <View style={[styles.statItem, styles.statItemLast]}>
          <ShimmerElement
            style={styles.iconContainer}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <View style={styles.labelContainer}>
            <ShimmerElement
              style={styles.statLabel}
              skeletonStyles={skeletonStyles}
              shimmerAnim={shimmerAnim}
            />
          </View>
          <ShimmerElement
            style={styles.statValue}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>
      </View>
    </View>
  );
});

StatsSkeleton.displayName = "StatsSkeleton";

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  skeletonStyles: {
    elementColor: string;
    progressFillColor: string;
  },
) =>
  StyleSheet.create({
    statsContainer: {
      borderRadius: theme.borderRadius.xl,
      paddingVertical: theme.spacing.lg,
      paddingHorizontal: theme.spacing.lg,
      marginHorizontal: 20,
      marginBottom: 24,
      backgroundColor:
        theme.mode === "light"
          ? theme.colors.background.surface
          : theme.colors.background.elevated,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.5)"
          : theme.colors.ui.divider,
      ...theme.shadows.md,
      shadowColor:
        theme.mode === "light"
          ? theme.colors.primary
          : theme.colors.text.primary,
      shadowOpacity: theme.mode === "light" ? 0.15 : 0.1,
      elevation: theme.mode === "light" ? 8 : 12,
    },
    progressOverview: {
      marginBottom: 5,
    },
    progressLabel: {
      width: 120,
      height: 14,
      borderRadius: 7,
      marginBottom: 8,
    },
    progressBarContainer: {
      flexDirection: "row",
      alignItems: "center",
      gap: 12,
      marginBottom: 6,
    },
    progressBarBackground: {
      flex: 1,
      height: 8,
      backgroundColor: theme.colors.ui.divider,
      borderRadius: 4,
      overflow: "hidden",
    },
    progressBarFill: {
      height: "100%",
      width: "65%", // Show partial progress
      backgroundColor: skeletonStyles.progressFillColor,
      borderRadius: 4,
    },
    progressPercentage: {
      width: 45,
      height: 16,
      borderRadius: 8,
    },
    progressFooter: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      paddingTop: 6,
    },
    motivationalText: {
      width: 100,
      height: 13,
      borderRadius: 6,
    },
    streakContainer: {
      flexDirection: "row",
      alignItems: "center",
      gap: 4,
    },
    streakIcon: {
      width: 20,
      height: 20,
      borderRadius: 10,
    },
    streakText: {
      width: 60,
      height: 14,
      borderRadius: 7,
    },
    statsGrid: {
      flexDirection: "row",
      alignItems: "stretch",
      minHeight: 120,
    },
    statItem: {
      alignItems: "center",
      flex: 1,
      paddingVertical: 8,
      justifyContent: "space-between",
    },
    statItemFirst: {
      paddingLeft: 0,
    },
    statItemLast: {
      paddingRight: 0,
    },
    statDivider: {
      width: 1,
      height: 60,
      backgroundColor: theme.colors.ui.divider,
      marginHorizontal: 16,
      borderRadius: 0.5,
    },
    iconContainer: {
      width: 48,
      height: 48,
      borderRadius: 24,
      marginBottom: 12,
    },
    labelContainer: {
      minHeight: 32,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: 4,
      paddingHorizontal: 4,
    },
    statLabel: {
      width: 60,
      height: 13,
      borderRadius: 6,
    },
    statValue: {
      width: 40,
      height: 28,
      borderRadius: 14,
    },
  });
