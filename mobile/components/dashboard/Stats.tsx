import { useTheme } from "@/contexts/ThemeContext";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import React, { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import {
  Animated,
  Image,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";

import { useWordlistProgress } from "@/hooks/useWordlistProgress";
import { StatsSkeleton } from "@/components/ui/StatsSkeleton";

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
type DashboardStatsProps = {
  onAddFirstWords?: () => void;
  wordlists?: { id: number; wordsCount?: number; totalWords?: number }[];
};

const PROGRESS_OVERVIEW_ENABLED = true;

const DashboardStats: React.FC<DashboardStatsProps> = ({
  onAddFirstWords,
  wordlists,
}) => {
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

  // Use shared wordlist progress hook to avoid duplicate API calls
  const {
    data: progressSummary,
    isLoading,
    isError,
    refetch,
  } = useWordlistProgress();

  const totalWordsFromWordlists = wordlists
    ? wordlists.reduce(
        (sum, wl) => sum + (wl.wordsCount ?? wl.totalWords ?? 0),
        0,
      )
    : null;
  const wordlistCountFromWordlists = wordlists ? wordlists.length : null;

  // Transform progress summary to user stats format
  const stats = progressSummary
    ? {
        totalWords:
          (totalWordsFromWordlists ??
            progressSummary.wordlists?.reduce(
              (sum, wl) => sum + wl.totalWords,
              0,
            )) ||
          0,
        wordlists:
          (wordlistCountFromWordlists ?? progressSummary.wordlists?.length) ||
          0,
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
    return <StatsSkeleton />;
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

  const singleWordlistFromStats =
    progressSummary?.wordlists && progressSummary.wordlists.length === 1
      ? progressSummary.wordlists[0]
      : null;
  const singleWordlistFromData =
    wordlists && wordlists.length === 1 ? wordlists[0] : null;
  const isSingleEmptyWordlist = singleWordlistFromData
    ? (singleWordlistFromData.wordsCount ?? 0) === 0
    : singleWordlistFromStats !== null &&
      singleWordlistFromStats.totalWords === 0;

  if (isSingleEmptyWordlist) {
    return (
      <View style={[styles.statsContainer, styles.readyStateContainer]}>
        <View style={styles.readyHeaderRow}>
          <View style={styles.readyImageWrap}>
            <Image
              source={require("../../assets/images/wordlist-ready.png")}
              style={styles.readyImage}
              resizeMode="contain"
            />
          </View>
          <View style={styles.readyTextBlock}>
            <Text style={styles.readyTitle}>
              {t("dashboard.stats.readyState.title")}
            </Text>
            <Text style={styles.readySubtitle} numberOfLines={1}>
              {t("dashboard.stats.readyState.subtitle")}
            </Text>
            <TouchableOpacity
              style={styles.readyButton}
              onPress={onAddFirstWords}
              activeOpacity={0.85}
              disabled={!onAddFirstWords}
            >
              <Ionicons
                name="add"
                size={18}
                color={theme.colors.text.inverse}
              />
              <Text style={styles.readyButtonText} numberOfLines={1}>
                {t("dashboard.stats.readyState.cta")}
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.statsContainer}>
      {/* Progress Overview */}
      {PROGRESS_OVERVIEW_ENABLED && (
        <View style={styles.progressOverview}>
          <View style={styles.progressHeader}>
            <Text style={styles.progressLabel}>
              {t("dashboard.stats.learningProgress")}
            </Text>
            {stats && stats.currentStreak > 0 ? (
              <View style={styles.streakContainer}>
                <MaterialIcons
                  name="local-fire-department"
                  size={18}
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

          <View style={styles.progressBottom}>
            <Text style={styles.motivationalText}>
              {getMotivationalMessage()}
            </Text>
          </View>
        </View>
      )}

      {/* Stats Grid */}
      <View style={styles.statsGrid}>
        <View style={[styles.statItem, styles.statItemFirst]}>
          <View style={styles.iconContainer}>
            <Ionicons name="list" size={22} color={theme.colors.success} />
          </View>
          <View style={styles.labelContainer}>
            <Text style={styles.statLabel}>
              {t("dashboard.stats.wordlists")}
            </Text>
          </View>
          <AnimatedCounter
            value={stats?.wordlists || 0}
            style={styles.statValue}
            delay={0}
          />
        </View>

        <View style={styles.statDivider} />

        <View style={styles.statItem}>
          <View style={styles.iconContainer}>
            <MaterialIcons
              name="library-books"
              size={22}
              color={theme.colors.primary}
            />
          </View>
          <View style={styles.labelContainer}>
            <Text style={styles.statLabel}>{t("dashboard.stats.words")}</Text>
          </View>
          <AnimatedCounter
            value={stats?.totalWords || 0}
            style={styles.statValue}
            delay={200}
          />
        </View>

        <View style={styles.statDivider} />

        <View style={[styles.statItem, styles.statItemLast]}>
          <View style={styles.iconContainer}>
            <MaterialIcons
              name="school"
              size={22}
              color={theme.colors.semantic.info}
            />
          </View>
          <View style={styles.labelContainer}>
            <Text style={styles.statLabel}>{t("dashboard.stats.learned")}</Text>
          </View>
          <AnimatedCounter
            value={stats?.wordsLearned || 0}
            style={[styles.statValue, { color: theme.colors.success }]}
            delay={400}
          />
        </View>
      </View>
    </View>
  );
};

export default DashboardStats;

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) =>
  StyleSheet.create({
    statsContainer: {
      borderRadius: theme.borderRadius.xl,
      paddingTop: theme.spacing.md,
      paddingBottom: theme.spacing.sm,
      paddingHorizontal: theme.spacing.md,
      marginHorizontal: 20,
      marginBottom: 16,
      backgroundColor:
        theme.mode === "light"
          ? theme.colors.background.surface
          : theme.colors.background.elevated,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(253, 246, 227, 0.5)" // Subtle web beige border
          : theme.colors.ui.divider,
      shadowColor:
        theme.mode === "light"
          ? theme.colors.primary
          : theme.colors.text.primary,
      shadowOffset: { width: 0, height: 8 },
      shadowOpacity: theme.mode === "light" ? 0.15 : 0.12,
      shadowRadius: 20,
      elevation: theme.mode === "light" ? 10 : 14,
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
      marginBottom: 8,
      paddingBottom: 8,
      borderBottomWidth: 1,
      borderBottomColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.05)"
          : "rgba(255, 255, 255, 0.05)",
    },
    progressHeader: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      marginBottom: 6,
    },
    progressLabel: {
      fontSize: 13.5,
      color: theme.colors.text.secondary,
      fontWeight: "600",
      flex: 1,
      letterSpacing: 0.3,
    },
    progressBarContainer: {
      flexDirection: "row",
      alignItems: "center",
      gap: 12,
    },
    progressBarBackground: {
      flex: 1,
      height: 7,
      backgroundColor: theme.colors.ui.divider,
      borderRadius: 6,
      overflow: "hidden",
    },
    progressBarFill: {
      height: "100%",
      backgroundColor: theme.colors.success,
      borderRadius: 6,
    },
    progressPercentage: {
      fontSize: 15,
      fontWeight: "700",
      color: theme.colors.success,
      minWidth: 45,
      letterSpacing: -0.3,
    },
    progressBottom: {
      paddingTop: 4,
    },
    motivationalText: {
      fontSize: 12.5,
      color: theme.colors.text.secondary,
      fontStyle: "italic",
      opacity: 0.85,
    },
    statsGrid: {
      flexDirection: "row",
      alignItems: "stretch",
      minHeight: 95,
      paddingTop: 4,
    },
    statItem: {
      alignItems: "center",
      flex: 1,
      paddingVertical: 4,
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
      alignSelf: "center",
      height: "65%",
      backgroundColor: theme.colors.ui.divider,
      marginHorizontal: 16,
      borderRadius: 0.5,
      opacity: 0.6,
    },
    readyStateContainer: {
      paddingVertical: theme.spacing.md,
      paddingLeft: theme.spacing.md,
      paddingRight: theme.spacing.lg,
    },
    readyHeaderRow: {
      flexDirection: "row",
      alignItems: "center",
      gap: 0,
    },
    readyImageWrap: {
      width: 96,
      height: 68,
      alignItems: "center",
      justifyContent: "center",
    },
    readyImage: {
      width: 170,
      height: 112,
      marginLeft: -20,
    },
    readyTextBlock: {
      flex: 1,
      alignItems: "flex-start",
    },
    readyTitle: {
      fontSize: 17,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    readySubtitle: {
      fontSize: 12.5,
      color: theme.colors.text.secondary,
      marginTop: 4,
    },
    readyButton: {
      alignSelf: "flex-start",
      flexDirection: "row",
      alignItems: "center",
      gap: 8,
      marginTop: 10,
      paddingHorizontal: 16,
      paddingVertical: 8,
      borderRadius: 999,
      backgroundColor: theme.colors.primary,
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 8 },
      shadowOpacity: 0.14,
      shadowRadius: 14,
      elevation: 10,
    },
    readyButtonText: {
      fontSize: 13,
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    iconContainer: {
      width: 44,
      height: 44,
      borderRadius: 22,
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
      marginBottom: 8,
      ...theme.shadows.sm,
    },
    labelContainer: {
      minHeight: 22,
      justifyContent: "center",
      alignItems: "center",
      marginBottom: 4,
      paddingHorizontal: 4,
    },
    statLabel: {
      fontSize: 12.5,
      color: theme.colors.text.secondary,
      fontWeight: "600",
      textAlign: "center",
      lineHeight: 16,
      letterSpacing: 0.2,
    },
    statValue: {
      fontSize: 26,
      fontWeight: "700",
      color: theme.colors.text.primary,
      letterSpacing: -0.5,
    },
    streakContainer: {
      flexDirection: "row",
      alignItems: "center",
      gap: 4,
    },
    streakText: {
      fontSize: 12.5,
      fontWeight: "700",
      color: theme.colors.semantic.warning,
      letterSpacing: 0.2,
    },
  });
