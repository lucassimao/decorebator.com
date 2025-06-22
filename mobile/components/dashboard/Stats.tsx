import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import React, { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import {
  Animated,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";

import { useUserInfo } from "@/hooks/users";
import { useWordlistProgress } from "@/hooks/useWordlistProgress";

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
  }, [value]);

  return <Text style={style}>{displayValue}</Text>;
};

// Main Component
type DashboardStatsProps = {};

const PROGRESS_OVERVIEW_ENABLED = true;

const DashboardStats: React.FC<DashboardStatsProps> = () => {
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const scaleAnim = useRef(new Animated.Value(0.9)).current;
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = createStyles(theme);

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
        <MaterialIcons name="error-outline" size={24} color={theme.colors.error} />
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
              {stats?.currentStreak && stats.currentStreak > 0 && (
                <View style={styles.streakContainer}>
                  <MaterialIcons
                    name="local-fire-department"
                    size={20}
                    color={theme.colors.semantic.warning}
                  />
                  <Text style={styles.streakText}>
                    {t("dashboard.stats.dayStreak", {
                      count: stats.currentStreak,
                    })}
                  </Text>
                </View>
              )}
            </View>
          </View>
        )}

        {/* Stats Grid */}
        <View style={styles.statsGrid}>
          <View style={[styles.statItem, styles.statItemFirst]}>
            <View style={styles.iconContainer}>
              <MaterialIcons name="library-books" size={24} color={theme.colors.primary} />
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
              <MaterialIcons name="school" size={24} color={theme.colors.semantic.info} />
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

const createStyles = (theme: ReturnType<typeof useTheme>['theme']) => StyleSheet.create({
  statsContainer: {
    borderRadius: theme.borderRadius.xl,
    paddingVertical: theme.spacing.lg,
    paddingHorizontal: theme.spacing.lg,
    marginHorizontal: 20,
    marginBottom: 24,
    backgroundColor: theme.mode === 'light' ? theme.colors.background.surface : theme.colors.background.elevated,
    borderWidth: 1,
    borderColor: theme.mode === 'light' ? theme.colors.ui.border : theme.colors.ui.divider,
    ...theme.shadows.md,
    // Extra elevation for contrast
    elevation: theme.mode === 'light' ? 8 : 12,
  },
  loadingContainer: {
    marginHorizontal: 20,
    marginBottom: 24,
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
  },
  skeletonItem: {
    alignItems: "center",
    flex: 1,
  },
  skeletonCircle: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: theme.colors.ui.divider,
    marginBottom: 8,
  },
  skeletonText: {
    width: 60,
    height: 12,
    borderRadius: 6,
    backgroundColor: theme.colors.ui.divider,
    marginBottom: 8,
  },
  skeletonNumber: {
    width: 40,
    height: 24,
    borderRadius: 12,
    backgroundColor: theme.colors.ui.divider,
  },
  errorContainer: {
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: theme.colors.background.surface,
    borderRadius: theme.borderRadius.xl,
    paddingVertical: 30,
    marginHorizontal: 20,
    marginBottom: 24,
    borderWidth: 1,
    borderColor: theme.colors.state.incorrectBackground,
    ...theme.shadows.md,
  },
  errorText: {
    fontSize: 16,
    color: theme.colors.text.secondary,
    marginTop: 8,
  },
  retryText: {
    fontSize: 14,
    color: theme.colors.primary,
    marginTop: 4,
    fontWeight: "500",
  },
  progressOverview: {
    marginBottom: 5,
  },
  progressLabel: {
    fontSize: 14,
    color: theme.colors.text.secondary,
    marginBottom: 8,
    fontWeight: "500",
  },
  progressBarContainer: {
    flexDirection: "row",
    alignItems: "center",
    gap: 12,
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
    backgroundColor: theme.colors.success,
    borderRadius: 4,
  },
  progressPercentage: {
    fontSize: 16,
    fontWeight: "600",
    color: theme.colors.success,
    minWidth: 45,
  },
  motivationalText: {
    fontSize: 13,
    color: theme.colors.text.secondary,
    fontStyle: "italic",
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
    backgroundColor: theme.mode === 'light' ? 'rgba(255, 123, 84, 0.08)' : theme.colors.background.subtle,
    borderWidth: 1,
    borderColor: theme.mode === 'light' ? 'rgba(255, 123, 84, 0.15)' : theme.colors.ui.border,
    justifyContent: "center",
    alignItems: "center",
    marginBottom: 12,
    ...theme.shadows.sm,
  },
  labelContainer: {
    minHeight: 32,
    justifyContent: "center",
    alignItems: "center",
    marginBottom: 4,
    paddingHorizontal: 4,
  },
  statLabel: {
    fontSize: 13,
    color: theme.colors.text.secondary,
    fontWeight: "500",
    textAlign: "center",
    lineHeight: 16,
  },
  statValue: {
    fontSize: 28,
    fontWeight: "700",
    color: theme.colors.text.primary,
  },
  streakContainer: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
  },
  streakText: {
    fontSize: 14,
    fontWeight: "600",
    color: theme.colors.semantic.warning,
  },
});
