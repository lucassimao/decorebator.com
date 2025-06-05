import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { LinearGradient } from "expo-linear-gradient";
import React, { useEffect, useRef } from "react";
import {
  Animated,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { useTranslation } from "react-i18next";

import * as wordlistsApi from "@/api/wordlists";
import { useUserInfo } from "@/hooks/users";
import { useRouter } from "expo-router";

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
type DashboardStatsProps = {
  onUpgradePress?: () => void;
};

const PROGRESS_OVERVIEW_ENABLED = false;

const DashboardStats: React.FC<DashboardStatsProps> = ({ onUpgradePress }) => {
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const scaleAnim = useRef(new Animated.Value(0.9)).current;
  const { t } = useTranslation();
  const { isPremium } = useUserInfo();
  const router = useRouter();

  const {
    data: stats,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ["dashboardStats"],
    queryFn: wordlistsApi.getUserStats,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Animation on load
  useEffect(() => {
    if (!isLoading && stats) {
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
  }, [isLoading, stats]);

  const getProgressPercentage = () => {
    if (!stats || stats.totalWords === 0) return 0;
    return Math.round((stats.wordsLearned / stats.totalWords) * 100);
  };

  const onStatsPress = () => {
    // router.push(`/analytics?wordlistId=${item.id}`);
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
        <MaterialIcons name="error-outline" size={24} color="#FF6B6B" />
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
      <TouchableOpacity
        onPress={isPremium ? onStatsPress : undefined}
        activeOpacity={isPremium ? 0.9 : 1}
        disabled={!isPremium}
      >
        <LinearGradient
          colors={["rgba(255, 255, 255, 0.9)", "rgba(255, 255, 255, 0.7)"]}
          style={styles.statsContainer}
          start={{ x: 0, y: 0 }}
          end={{ x: 1, y: 1 }}
        >
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
                      color="#FF6B3D"
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
                <MaterialIcons name="library-books" size={24} color="#FF7B54" />
              </View>
              <Text style={styles.statLabel}>
                {t("dashboard.stats.totalWords")}
              </Text>
              <AnimatedCounter
                value={stats?.totalWords || 0}
                style={styles.statValue}
                delay={0}
              />
            </View>

            <View style={styles.statDivider} />

            <View style={styles.statItem}>
              <View style={styles.iconContainer}>
                <Ionicons name="list" size={24} color="#4CAF50" />
              </View>
              <Text style={styles.statLabel}>
                {t("dashboard.stats.wordlists")}
              </Text>
              <AnimatedCounter
                value={stats?.wordlists || 0}
                style={styles.statValue}
                delay={200}
              />
            </View>

            <View style={styles.statDivider} />

            <View style={[styles.statItem, styles.statItemLast]}>
              <View style={styles.iconContainer}>
                <MaterialIcons name="school" size={24} color="#2196F3" />
              </View>
              <Text style={styles.statLabel}>
                {t("dashboard.stats.learned")}
              </Text>
              <AnimatedCounter
                value={stats?.wordsLearned || 0}
                style={[styles.statValue, { color: "#4CAF50" }]}
                delay={400}
              />
            </View>
          </View>

          {/* Premium Analytics Upsell */}
          {!isPremium && (
            <TouchableOpacity
              style={styles.premiumUpsellContainer}
              onPress={onUpgradePress}
              activeOpacity={0.8}
            >
              <LinearGradient
                colors={["#FFD700", "#FFA500"]}
                style={styles.premiumGradient}
                start={{ x: 0, y: 0 }}
                end={{ x: 1, y: 0 }}
              >
                <View style={styles.premiumContent}>
                  <View style={styles.premiumLeft}>
                    <View style={styles.premiumIconContainer}>
                      <MaterialIcons name="analytics" size={20} color="#FFF" />
                    </View>
                    <View style={styles.premiumTextContainer}>
                      <Text style={styles.premiumTitle}>
                        {t("dashboard.stats.premium.title")}
                      </Text>
                      <Text style={styles.premiumSubtitle}>
                        {t("dashboard.stats.premium.subtitle")}
                      </Text>
                    </View>
                  </View>
                  <MaterialIcons name="chevron-right" size={24} color="#FFF" />
                </View>
              </LinearGradient>
            </TouchableOpacity>
          )}

          {/* Premium Analytics Access Indicator */}
          {isPremium && (
            <TouchableOpacity
              style={styles.analyticsAccessContainer}
              onPress={onStatsPress}
              activeOpacity={0.8}
            >
              <View style={styles.analyticsAccessContent}>
                <MaterialIcons name="analytics" size={20} color="#FFD700" />
                <Text style={styles.analyticsAccessText}>
                  {t("dashboard.stats.premium.viewAnalytics")}
                </Text>
                <MaterialIcons name="chevron-right" size={20} color="#636E72" />
              </View>
            </TouchableOpacity>
          )}
        </LinearGradient>
      </TouchableOpacity>
    </Animated.View>
  );
};

export default DashboardStats;

const styles = StyleSheet.create({
  statsContainer: {
    borderRadius: 20,
    paddingVertical: 20,
    paddingHorizontal: 16,
    marginHorizontal: 20,
    marginBottom: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.08,
    shadowRadius: 12,
    elevation: 5,
  },
  loadingContainer: {
    marginHorizontal: 20,
    marginBottom: 24,
  },
  skeletonContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    backgroundColor: "rgba(255, 255, 255, 0.7)",
    borderRadius: 20,
    paddingVertical: 20,
    paddingHorizontal: 16,
  },
  skeletonItem: {
    alignItems: "center",
    flex: 1,
  },
  skeletonCircle: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: "#F0F0F0",
    marginBottom: 8,
  },
  skeletonText: {
    width: 60,
    height: 12,
    borderRadius: 6,
    backgroundColor: "#F0F0F0",
    marginBottom: 8,
  },
  skeletonNumber: {
    width: 40,
    height: 24,
    borderRadius: 12,
    backgroundColor: "#F0F0F0",
  },
  errorContainer: {
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "rgba(255, 255, 255, 0.7)",
    borderRadius: 20,
    paddingVertical: 30,
    marginHorizontal: 20,
    marginBottom: 24,
  },
  errorText: {
    fontSize: 16,
    color: "#636E72",
    marginTop: 8,
  },
  retryText: {
    fontSize: 14,
    color: "#FF7B54",
    marginTop: 4,
    fontWeight: "500",
  },
  progressOverview: {
    marginBottom: 5,
  },
  progressLabel: {
    fontSize: 14,
    color: "#636E72",
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
    backgroundColor: "#F0F0F0",
    borderRadius: 4,
    overflow: "hidden",
  },
  progressBarFill: {
    height: "100%",
    backgroundColor: "#4CAF50",
    borderRadius: 4,
  },
  progressPercentage: {
    fontSize: 16,
    fontWeight: "600",
    color: "#4CAF50",
    minWidth: 45,
  },
  motivationalText: {
    fontSize: 13,
    color: "#636E72",
    fontStyle: "italic",
  },
  statsGrid: {
    flexDirection: "row",
    alignItems: "center",
  },
  statItem: {
    alignItems: "center",
    flex: 1,
    paddingVertical: 8,
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
    backgroundColor: "#E0E0E0",
    marginHorizontal: 16,
  },
  iconContainer: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: "rgba(255, 123, 84, 0.1)",
    justifyContent: "center",
    alignItems: "center",
    marginBottom: 8,
  },
  statLabel: {
    fontSize: 13,
    color: "#636E72",
    marginBottom: 4,
    fontWeight: "500",
  },
  statValue: {
    fontSize: 28,
    fontWeight: "700",
    color: "#2D3436",
  },
  streakContainer: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
  },
  streakText: {
    fontSize: 14,
    fontWeight: "600",
    color: "#FF6B3D",
  },
  // Premium Upsell Styles
  premiumUpsellContainer: {
    marginTop: 16,
    borderRadius: 12,
    overflow: "hidden",
  },
  premiumGradient: {
    padding: 1,
    borderRadius: 12,
  },
  premiumContent: {
    backgroundColor: "rgba(255, 215, 0, 0.1)",
    borderRadius: 11,
    padding: 12,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
  },
  premiumLeft: {
    flexDirection: "row",
    alignItems: "center",
    flex: 1,
  },
  premiumIconContainer: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: "#FFD700",
    justifyContent: "center",
    alignItems: "center",
    marginRight: 12,
  },
  premiumTextContainer: {
    flex: 1,
  },
  premiumTitle: {
    fontSize: 15,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 2,
  },
  premiumSubtitle: {
    fontSize: 13,
    color: "#636E72",
  },
  // Analytics Access Styles (for premium users)
  analyticsAccessContainer: {
    marginTop: 16,
    backgroundColor: "rgba(255, 215, 0, 0.1)",
    borderRadius: 12,
    padding: 12,
  },
  analyticsAccessContent: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
  },
  analyticsAccessText: {
    fontSize: 14,
    fontWeight: "500",
    color: "#636E72",
    flex: 1,
    textAlign: "center",
  },
});
