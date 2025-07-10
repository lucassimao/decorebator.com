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

export const WordlistItemSkeleton: React.FC = React.memo(() => {
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
      containerColor: isDark
        ? theme.colors.background.elevated
        : theme.colors.background.surface,
      progressFillColor: isDark
        ? "rgba(34, 197, 94, 0.3)" // theme.colors.success with opacity
        : "rgba(34, 197, 94, 0.2)",
    };
  };

  const skeletonStyles = getSkeletonStyles();
  const styles = createStyles(theme, skeletonStyles);

  return (
    <View style={styles.wordlistCard}>
      {/* Card Header */}
      <View style={styles.cardHeader}>
        <ShimmerElement
          style={styles.languageFlag}
          skeletonStyles={skeletonStyles}
          shimmerAnim={shimmerAnim}
        />
        <View style={styles.cardTitleContainer}>
          <ShimmerElement
            style={styles.wordlistTitle}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <ShimmerElement
            style={styles.wordlistDescription}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>
      </View>

      {/* Card Stats */}
      <View style={styles.cardStats}>
        <View style={styles.cardStat}>
          <ShimmerElement
            style={styles.statIcon}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <ShimmerElement
            style={styles.statText}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>
        <View style={styles.cardStat}>
          <ShimmerElement
            style={styles.languageName}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>
        <View style={styles.cardStat}>
          <ShimmerElement
            style={styles.statIcon}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
          <ShimmerElement
            style={styles.statText}
            skeletonStyles={skeletonStyles}
            shimmerAnim={shimmerAnim}
          />
        </View>
      </View>

      {/* Progress Bar */}
      <View style={styles.progressBar}>
        <Animated.View style={styles.progressFill} />
      </View>

      {/* Action Buttons Row */}
      <View style={styles.actionButtonsRow}>
        {[1, 2, 3, 4].map((item) => (
          <View key={item} style={styles.actionButton}>
            <ShimmerElement
              style={styles.actionButtonIcon}
              skeletonStyles={skeletonStyles}
              shimmerAnim={shimmerAnim}
            />
            <ShimmerElement
              style={styles.actionButtonText}
              skeletonStyles={skeletonStyles}
              shimmerAnim={shimmerAnim}
            />
          </View>
        ))}
      </View>
    </View>
  );
});

WordlistItemSkeleton.displayName = "WordlistItemSkeleton";

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  skeletonStyles: {
    elementColor: string;
    containerColor: string;
    progressFillColor: string;
  },
) =>
  StyleSheet.create({
    wordlistCard: {
      backgroundColor: skeletonStyles.containerColor,
      borderRadius: theme.borderRadius.lg,
      padding: theme.spacing.md,
      marginHorizontal: 20,
      marginBottom: 12,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.3)"
          : theme.colors.ui.divider,
      ...theme.shadows.md,
      shadowColor:
        theme.mode === "light"
          ? theme.colors.primary
          : theme.colors.text.primary,
      shadowOpacity: theme.mode === "light" ? 0.15 : 0.1,
      elevation: theme.mode === "light" ? 6 : 10,
    },
    cardHeader: {
      flexDirection: "row",
      alignItems: "flex-start",
      marginBottom: 12,
    },
    languageFlag: {
      width: 32,
      height: 32,
      borderRadius: 16,
      marginRight: 12,
    },
    cardTitleContainer: {
      flex: 1,
    },
    wordlistTitle: {
      width: "75%",
      height: 18,
      borderRadius: 9,
      marginBottom: 8,
    },
    wordlistDescription: {
      width: "60%",
      height: 14,
      borderRadius: 7,
    },
    cardStats: {
      flexDirection: "row",
      flexWrap: "wrap",
      gap: 16,
      marginBottom: 12,
    },
    cardStat: {
      flexDirection: "row",
      alignItems: "center",
      gap: 4,
    },
    statIcon: {
      width: 16,
      height: 16,
      borderRadius: 8,
    },
    statText: {
      width: 50,
      height: 14,
      borderRadius: 7,
    },
    languageName: {
      width: 60,
      height: 14,
      borderRadius: 7,
    },
    progressBar: {
      height: 4,
      backgroundColor: theme.colors.ui.divider,
      borderRadius: 2,
      overflow: "hidden",
    },
    progressFill: {
      height: "100%",
      width: "45%", // Show partial progress
      backgroundColor: skeletonStyles.progressFillColor,
      borderRadius: 2,
    },
    actionButtonsRow: {
      flexDirection: "row",
      marginTop: 12,
      gap: 8,
      paddingTop: 8,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
    },
    actionButton: {
      flex: 1,
      flexDirection: "column",
      alignItems: "center",
      paddingVertical: 12,
      paddingHorizontal: 8,
      borderRadius: theme.borderRadius.sm,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.2)"
          : theme.colors.background.subtle,
      gap: 4,
    },
    actionButtonIcon: {
      width: 20,
      height: 20,
      borderRadius: 10,
    },
    actionButtonText: {
      width: 40,
      height: 12,
      borderRadius: 6,
    },
  });
