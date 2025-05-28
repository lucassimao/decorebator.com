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

import * as wordlistsApi from "@/api/wordlists";

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
  onStatsPress?: () => void;
};

const DashboardStats: React.FC<DashboardStatsProps> = ({ onStatsPress }) => {
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const scaleAnim = useRef(new Animated.Value(0.9)).current;

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

  const getMotivationalMessage = () => {
    const progress = getProgressPercentage();
    if (progress === 0) return "Ready to start learning! 🚀";
    if (progress < 25) return "Great start! Keep going! 💪";
    if (progress < 50) return "Making progress! 📈";
    if (progress < 75) return "More than halfway there! 🎯";
    if (progress < 100) return "Almost there! You got this! 🔥";
    return "Congratulations! 🎉";
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
        <Text style={styles.errorText}>Failed to load stats</Text>
        <Text style={styles.retryText}>Tap to retry</Text>
      </TouchableOpacity>
    );
  }

  return (
    <Animated.View
      style={{ opacity: fadeAnim, transform: [{ scale: scaleAnim }] }}
    >
      <TouchableOpacity
        onPress={onStatsPress}
        activeOpacity={0.9}
        disabled={!onStatsPress}
      >
        <LinearGradient
          colors={["rgba(255, 255, 255, 0.9)", "rgba(255, 255, 255, 0.7)"]}
          style={styles.statsContainer}
          start={{ x: 0, y: 0 }}
          end={{ x: 1, y: 1 }}
        >
          {/* Progress Overview */}
          <View style={styles.progressOverview}>
            <Text style={styles.progressLabel}>Learning Progress</Text>
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

            <View style={{ flexDirection: "row",justifyContent:"space-between",alignItems:"center" ,paddingTop:6}}>
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
                    {stats.currentStreak} day streak!
                  </Text>
                </View>
              )}
            </View>
          </View>

          {/* Stats Grid */}
          <View style={styles.statsGrid}>
            <View style={[styles.statItem, styles.statItemFirst]}>
              <View style={styles.iconContainer}>
                <MaterialIcons name="library-books" size={24} color="#FF7B54" />
              </View>
              <Text style={styles.statLabel}>Total Words</Text>
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
              <Text style={styles.statLabel}>Wordlists</Text>
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
              <Text style={styles.statLabel}>Learned</Text>
              <AnimatedCounter
                value={stats?.wordsLearned || 0}
                style={[styles.statValue, { color: "#4CAF50" }]}
                delay={400}
              />
            </View>
          </View>
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
    // marginTop: 16,
    // paddingTop: 16,
    // borderTopWidth: 1,
    // borderTopColor: "#F0F0F0",
    // gap: 6,
  },
  streakText: {
    fontSize: 14,
    fontWeight: "600",
    color: "#FF6B3D",
  },
});
