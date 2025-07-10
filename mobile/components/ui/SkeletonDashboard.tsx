import React from "react";
import { Animated, StyleSheet, View } from "react-native";
import { useTheme } from "@/contexts/ThemeContext";

// No props needed - component is fully self-contained

interface ShimmerElementProps {
  style: any;
  skeletonStyles: {
    elementColor: string;
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
    outputRange: [0.3, 0.8],
  });

  return (
    <Animated.View
      style={[
        style,
        {
          backgroundColor: skeletonStyles.elementColor,
          opacity: shimmerOpacity,
        },
      ]}
    />
  );
};

export const SkeletonDashboard: React.FC = React.memo(() => {
  const { theme } = useTheme();

  // Internal shimmer animation - fully self-contained
  const shimmerAnim = React.useRef(new Animated.Value(0)).current;

  // Shimmer animation loop - starts when component mounts
  React.useEffect(() => {
    const shimmerLoop = Animated.loop(
      Animated.sequence([
        Animated.timing(shimmerAnim, {
          toValue: 1,
          duration: 1200,
          useNativeDriver: true,
        }),
        Animated.timing(shimmerAnim, {
          toValue: 0,
          duration: 1200,
          useNativeDriver: true,
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
      elementColor: isDark ? "rgba(255, 255, 255, 0.1)" : "rgba(0, 0, 0, 0.1)",
      containerColor: isDark
        ? "rgba(255, 255, 255, 0.05)"
        : "rgba(255, 255, 255, 0.9)",
      shadowColor: isDark ? "#FFFFFF" : "#000000",
    };
  };

  const skeletonStyles = getSkeletonStyles();

  return (
    <View style={styles.skeletonContainer}>
      {/* Header skeleton */}
      <View style={styles.skeletonHeader}>
        <ShimmerElement
          style={styles.skeletonCircle}
          skeletonStyles={skeletonStyles}
          shimmerAnim={shimmerAnim}
        />
        <ShimmerElement
          style={styles.skeletonTitle}
          skeletonStyles={skeletonStyles}
          shimmerAnim={shimmerAnim}
        />
        <ShimmerElement
          style={styles.skeletonCircle}
          skeletonStyles={skeletonStyles}
          shimmerAnim={shimmerAnim}
        />
      </View>

      {/* Stats skeleton */}
      <View
        style={[
          styles.skeletonStats,
          { backgroundColor: skeletonStyles.containerColor },
        ]}
      >
        <View style={styles.skeletonStatItem}>
          <ShimmerElement
            style={styles.skeletonStatValue}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <ShimmerElement
            style={styles.skeletonStatLabel}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>
        <View style={styles.skeletonStatItem}>
          <ShimmerElement
            style={styles.skeletonStatValue}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <ShimmerElement
            style={styles.skeletonStatLabel}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>
        <View style={styles.skeletonStatItem}>
          <ShimmerElement
            style={styles.skeletonStatValue}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <ShimmerElement
            style={styles.skeletonStatLabel}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>
      </View>

      {/* Section header skeleton */}
      <View style={styles.skeletonSectionHeader}>
        <ShimmerElement
          style={styles.skeletonSectionTitle}
          skeletonStyles={skeletonStyles}
          shimmerAnim={shimmerAnim}
        />
        <ShimmerElement
          style={styles.skeletonAddButton}
          skeletonStyles={skeletonStyles}
          shimmerAnim={shimmerAnim}
        />
      </View>

      {/* Wordlist items skeleton */}
      <View style={styles.skeletonWordlists}>
        {[1, 2, 3].map((item) => (
          <View
            key={item}
            style={[
              styles.skeletonWordlistItem,
              { backgroundColor: skeletonStyles.containerColor },
            ]}
          >
            <View style={styles.skeletonWordlistInfo}>
              <ShimmerElement
                style={styles.skeletonWordlistTitle}
                skeletonStyles={skeletonStyles}
                shimmerAnim={shimmerAnim}
              />
              <ShimmerElement
                style={styles.skeletonWordlistSubtitle}
                skeletonStyles={skeletonStyles}
                shimmerAnim={shimmerAnim}
              />
            </View>
            <ShimmerElement
              style={styles.skeletonWordlistProgress}
              skeletonStyles={skeletonStyles}
              shimmerAnim={shimmerAnim}
            />
          </View>
        ))}
      </View>
    </View>
  );
});

SkeletonDashboard.displayName = "SkeletonDashboard";

const styles = StyleSheet.create({
  // Skeleton styles - theme-aware with better contrast
  skeletonContainer: {
    flex: 1,
    paddingHorizontal: 20,
  },
  skeletonHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingVertical: 20,
  },
  skeletonCircle: {
    width: 40,
    height: 40,
    borderRadius: 20,
  },
  skeletonTitle: {
    width: 120,
    height: 24,
    borderRadius: 12,
  },
  skeletonStats: {
    flexDirection: "row",
    justifyContent: "space-between",
    borderRadius: 16,
    paddingVertical: 20,
    paddingHorizontal: 10,
    marginBottom: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  skeletonStatItem: {
    alignItems: "center",
    flex: 1,
  },
  skeletonStatValue: {
    width: 40,
    height: 36,
    borderRadius: 18,
    marginBottom: 8,
  },
  skeletonStatLabel: {
    width: 60,
    height: 14,
    borderRadius: 7,
  },
  skeletonSectionHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 16,
  },
  skeletonSectionTitle: {
    width: 150,
    height: 20,
    borderRadius: 10,
  },
  skeletonAddButton: {
    width: 34,
    height: 34,
    borderRadius: 17,
  },
  skeletonWordlists: {
    gap: 16,
  },
  skeletonWordlistItem: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    borderRadius: 12,
    padding: 16,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  skeletonWordlistInfo: {
    flex: 1,
  },
  skeletonWordlistTitle: {
    width: "80%",
    height: 18,
    borderRadius: 9,
    marginBottom: 8,
  },
  skeletonWordlistSubtitle: {
    width: "60%",
    height: 14,
    borderRadius: 7,
  },
  skeletonWordlistProgress: {
    width: 60,
    height: 60,
    borderRadius: 30,
  },
});
