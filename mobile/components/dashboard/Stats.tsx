import { useTheme } from "@/contexts/ThemeContext";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import React, { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import {
  Animated,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";

import { useWordlistProgress } from "@/hooks/useWordlistProgress";
import { useResponsive } from "@/hooks/useResponsive";

// Animated Counter Component
interface AnimatedCounterProps {
  value: number;
  duration?: number;
  delay?: number;
  style?: any;
}

const AnimatedCounter: React.FC<AnimatedCounterProps> = ({
  value,
  duration = 1000,
  delay = 0,
  style,
}) => {
  const animatedValue = useRef(new Animated.Value(0)).current;
  const [displayValue, setDisplayValue] = React.useState(0);

  useEffect(() => {
    const listener = animatedValue.addListener(({ value }) => {
      setDisplayValue(Math.floor(value));
    });

    Animated.timing(animatedValue, {
      toValue: value,
      duration,
      delay,
      useNativeDriver: false,
    }).start();

    return () => {
      animatedValue.removeListener(listener);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  return <Text style={style}>{displayValue}</Text>;
};

// Main Component
// eslint-disable-next-line @typescript-eslint/no-empty-object-type
type DashboardStatsProps = {};

const PROGRESS_OVERVIEW_ENABLED = true;

const DashboardStats: React.FC<DashboardStatsProps> = () => {
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const scaleAnim = useRef(new Animated.Value(0.9)).current;
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, type: deviceType } = useResponsive();
  const styles = createStyles(theme, isTablet, deviceType);

  // Use shared wordlist progress hook to avoid duplicate API calls
  const {
    data: progressSummary,
    isLoading,
    isError,
    refetch,
  } = useWordlistProgress();

  // Transform progress summary to user stats format
  const stats = progressSummary
    ? {
        totalWords:
          progressSummary.wordlists?.reduce(
            (sum, wl) => sum + wl.totalWords,
            0,
          ) || 0,
        wordlists: progressSummary.wordlists?.length || 0,
        wordsLearned:
          progressSummary.wordlists?.reduce(
            (sum, wl) => sum + wl.wordsMastered,
            0,
          ) || 0,
        currentStreak:
          progressSummary.wordlists?.length > 0
            ? Math.max(
                0,
                ...progressSummary.wordlists.map((wl) => wl.currentStreak || 0),
              ) || 0
            : 0,
      }
    : null;

  // Animation on load
  useEffect(() => {
    if (!isLoading && progressSummary) {
      Animated.parallel([
        Animated.timing(fadeAnim, {
          toValue: 1,
          duration: 500,
          useNativeDriver: true,
        }),
        Animated.spring(scaleAnim, {
          toValue: 1,
          friction: 8,
          tension: 40,
          useNativeDriver: true,
        }),
      ]).start();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, progressSummary]);

  const getProgressPercentage = () => {
    if (!stats || stats.totalWords === 0) return 0;
    return Math.round((stats.wordsLearned / stats.totalWords) * 100);
  };

  const getMotivationalMessage = () => {
    const progress = getProgressPercentage();
    if (progress === 0) return t("dashboard.stats.progress.ready");
    if (progress < 25) return t("dashboard.stats.progress.greatStart");
    if (progress < 50) return t("dashboard.stats.progress.makingProgress");
    if (progress < 75) return t("dashboard.stats.progress.halfwayThere");
    if (progress < 100) return t("dashboard.stats.progress.almostThere");
    return t("dashboard.stats.progress.congratulations");
  };

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <View style={styles.skeletonContainer}>
          {[1, 2, 3].map((index) => (
            <View key={index} style={styles.skeletonItem}>
              <View style={styles.skeletonCircle} />
              <View style={styles.skeletonText} />
              <View style={styles.skeletonNumber} />
            </View>
          ))}
        </View>
      </View>
    );
  }

  if (isError) {
    return (
      <TouchableOpacity onPress={() => refetch()} style={styles.errorContainer}>
        <MaterialIcons
          name="error-outline"
          size={24}
          color={theme.colors.error}
        />
        <Text style={styles.errorText}>
          {t("dashboard.stats.failedToLoad")}
        </Text>
        <Text style={styles.retryText}>{t("dashboard.stats.tapToRetry")}</Text>
      </TouchableOpacity>
    );
  }

  return (
    <Animated.View
      style={{ opacity: fadeAnim, transform: [{ scale: scaleAnim }] }}
    >
      <View style={styles.statsContainer}>
        {/* Progress Overview */}
        {PROGRESS_OVERVIEW_ENABLED && (
          <View style={styles.progressOverview}>
            <Text style={styles.progressLabel}>
              {t("dashboard.stats.learningProgress")}
            </Text>
            <View style={styles.progressBarContainer}>
              <View style={styles.progressBarBackground}>
                <Animated.View
                  style={[
                    styles.progressBarFill,
                    {
                      width: `${getProgressPercentage()}%`,
                    },
                  ]}
                />
              </View>
              <Text style={styles.progressPercentage}>
                {getProgressPercentage()}%
              </Text>
            </View>

            <View
              style={{
                flexDirection: "row",
                justifyContent: "space-between",
                alignItems: "center",
                paddingTop: 6,
              }}
            >
              <Text style={styles.motivationalText}>
                {getMotivationalMessage()}
              </Text>
              {stats && stats.currentStreak > 0 ? (
                <View style={styles.streakContainer}>
                  <MaterialIcons
                    name="local-fire-department"
                    size={20}
                    color={theme.colors.semantic.warning}
                  />
                  <Text style={styles.streakText}>
                    {t("dashboard.stats.dayStreak", {
                      count: stats?.currentStreak || 0,
                    })}
                  </Text>
                </View>
              ) : null}
            </View>
          </View>
        )}

        {/* Stats Grid */}
        <View style={styles.statsGrid}>
          <View style={[styles.statItem, styles.statItemFirst]}>
            <View style={styles.iconContainer}>
              <MaterialIcons
                name="library-books"
                size={24}
                color={theme.colors.primary}
              />
            </View>
            <View style={styles.labelContainer}>
              <Text style={styles.statLabel}>
                {t("dashboard.stats.totalWords")}
              </Text>
            </View>
            <AnimatedCounter
              value={stats?.totalWords || 0}
              style={styles.statValue}
              delay={0}
            />
          </View>

          <View style={styles.statDivider} />

          <View style={styles.statItem}>
            <View style={styles.iconContainer}>
              <Ionicons name="list" size={24} color={theme.colors.success} />
            </View>
            <View style={styles.labelContainer}>
              <Text style={styles.statLabel}>
                {t("dashboard.stats.wordlists")}
              </Text>
            </View>
            <AnimatedCounter
              value={stats?.wordlists || 0}
              style={styles.statValue}
              delay={200}
            />
          </View>

          <View style={styles.statDivider} />

          <View style={[styles.statItem, styles.statItemLast]}>
            <View style={styles.iconContainer}>
              <MaterialIcons
                name="school"
                size={24}
                color={theme.colors.semantic.info}
              />
            </View>
            <View style={styles.labelContainer}>
              <Text style={styles.statLabel}>
                {t("dashboard.stats.learned")}
              </Text>
            </View>
            <AnimatedCounter
              value={stats?.wordsLearned || 0}
              style={[styles.statValue, { color: theme.colors.success }]}
              delay={400}
            />
          </View>
        </View>
      </View>
    </Animated.View>
  );
};

export default DashboardStats;

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  isTablet: boolean,
  deviceType: string,
) =>
  StyleSheet.create({
    statsContainer: {
      borderRadius: theme.borderRadius.xl,
      paddingVertical: theme.spacing.lg,
      paddingHorizontal: theme.spacing.lg,
      marginHorizontal: 0, // Remove for responsive container
      marginBottom: theme.spacing.lg,
      backgroundColor:
        theme.mode === "light"
          ? theme.colors.background.surface
          : theme.colors.background.elevated,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.5)" // Subtle web beige border
          : theme.colors.ui.divider,
      ...theme.shadows.md,
      // Extra elevation for contrast with subtle orange shadow
      shadowColor:
        theme.mode === "light"
          ? theme.colors.primary
          : theme.colors.text.primary,
      shadowOpacity: theme.mode === "light" ? 0.15 : 0.1,
      elevation: theme.mode === "light" ? 8 : 12,
    },
    loadingContainer: {
      marginHorizontal: 0, // Remove for responsive container
      marginBottom: theme.spacing.lg,
    },
    skeletonContainer: {
      flexDirection: "row",
      justifyContent: "space-between",
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.xl,
      paddingVertical: theme.spacing.lg,
      paddingHorizontal: theme.spacing.lg,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      ...theme.shadows.md,
      shadowColor:
        theme.mode === "light"
          ? theme.colors.primary
          : theme.colors.text.primary,
      shadowOpacity: theme.mode === "light" ? 0.15 : 0.1,
    },
    skeletonItem: {
      alignItems: "center",
      flex: 1,
    },
    skeletonCircle: {
      width: isTablet ? 48 : 40,
      height: isTablet ? 48 : 40,
      borderRadius: isTablet ? 24 : 20,
      backgroundColor: theme.colors.ui.divider,
      marginBottom: theme.spacing.xs,
    },
    skeletonText: {
      width: isTablet ? 80 : 60,
      height: theme.spacing.sm,
      borderRadius: theme.borderRadius.xs,
      backgroundColor: theme.colors.ui.divider,
      marginBottom: theme.spacing.xs,
    },
    skeletonNumber: {
      width: isTablet ? 50 : 40,
      height: isTablet ? 30 : 24,
      borderRadius: theme.borderRadius.sm,
      backgroundColor: theme.colors.ui.divider,
    },
    errorContainer: {
      flexDirection: "column",
      alignItems: "center",
      justifyContent: "center",
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.xl,
      paddingVertical: theme.spacing.xl,
      marginHorizontal: 0, // Remove for responsive container
      marginBottom: theme.spacing.lg,
      borderWidth: 1,
      borderColor: theme.colors.state.incorrectBackground,
      ...theme.shadows.md,
    },
    errorText: {
      fontSize: theme.typography.sizes.body,
      color: theme.colors.text.secondary,
      marginTop: theme.spacing.xs,
    },
    retryText: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.primary,
      marginTop: theme.spacing.xs,
      fontWeight: theme.typography.weights.medium,
    },
    progressOverview: {
      marginBottom: theme.spacing.xs,
    },
    progressLabel: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
      marginBottom: theme.spacing.xs,
      fontWeight: theme.typography.weights.medium,
    },
    progressBarContainer: {
      flexDirection: "row",
      alignItems: "center",
      gap: theme.spacing.sm,
    },
    progressBarBackground: {
      flex: 1,
      height: isTablet ? 10 : 8,
      backgroundColor: theme.colors.ui.divider,
      borderRadius: theme.borderRadius.xs,
      overflow: "hidden",
    },
    progressBarFill: {
      height: "100%",
      backgroundColor: theme.colors.success,
      borderRadius: theme.borderRadius.xs,
    },
    progressPercentage: {
      fontSize: theme.typography.sizes.body,
      fontWeight: theme.typography.weights.semibold,
      color: theme.colors.success,
      minWidth: isTablet ? 55 : 45,
    },
    motivationalText: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
      fontStyle: "italic",
    },
    statsGrid: {
      flexDirection: "row",
      alignItems: "stretch",
      minHeight: isTablet ? 140 : 120,
    },
    statItem: {
      alignItems: "center",
      flex: 1,
      paddingVertical: theme.spacing.xs,
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
      height: isTablet ? 80 : 60,
      backgroundColor: theme.colors.ui.divider,
      marginHorizontal: theme.spacing.md,
      borderRadius: theme.borderRadius.xs,
    },
    iconContainer: {
      width: isTablet ? 56 : 48,
      height: isTablet ? 56 : 48,
      borderRadius: isTablet ? 28 : 24,
      backgroundColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.4)" // Web beige background
          : theme.colors.background.subtle,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(255, 123, 84, 0.15)"
          : theme.colors.ui.border,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: theme.spacing.sm,
      ...theme.shadows.sm,
    },
    labelContainer: {
      minHeight: isTablet ? 40 : 32,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: theme.spacing.xs,
      paddingHorizontal: theme.spacing.xs,
    },
    statLabel: {
      fontSize: theme.typography.sizes.caption,
      color: theme.colors.text.secondary,
      fontWeight: theme.typography.weights.medium,
      textAlign: "center",
      lineHeight: theme.typography.lineHeights.caption,
    },
    statValue: {
      fontSize: isTablet
        ? theme.typography.sizes.display
        : theme.typography.sizes.title,
      fontWeight: theme.typography.weights.bold,
      color: theme.colors.text.primary,
    },
    streakContainer: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
    },
    streakText: {
      fontSize: theme.typography.sizes.caption,
      fontWeight: theme.typography.weights.semibold,
      color: theme.colors.semantic.warning,
    },
  });
