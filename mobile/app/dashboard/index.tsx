import { Wordlist } from "@/api/wordlists";
import { useUserSession } from "@/hooks/useUserSession";
import { CreateWordlistModal } from "@/components/dashboard/CreateWordlistModal";
import { CongratulationsModal } from "@/components/dashboard/CongratulationsModal";
import { Header } from "@/components/dashboard/Header";
import DashboardStats from "@/components/dashboard/Stats";
import { WordlistDetailModal } from "@/components/dashboard/WordlistDetailModal";
import Wordlistitem from "@/components/dashboard/WordlistItem";
import { WelcomeOverlay } from "@/components/dashboard/WelcomeOverlay";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { WordlistItemSkeleton } from "@/components/ui/WordlistItemSkeleton";
import { useTheme } from "@/contexts/ThemeContext";
import { useUpgradePromptDialog } from "@/hooks/useUpgradePromptDialog";
import { useWordlistProgress } from "@/hooks/useWordlistProgress";
import { useWordlists } from "@/hooks/useWordlists";
import { createCommonStyles } from "@/styles/common";
import { Ionicons } from "@expo/vector-icons";
import { useRouter, useFocusEffect } from "expo-router";
import React, { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { usePostHog } from "posthog-react-native";
import {
  Alert,
  Animated,
  Dimensions,
  FlatList,
  Image,
  ImageBackground,
  RefreshControl,
  SafeAreaView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  Easing,
} from "react-native";
import AsyncStorage from "@react-native-async-storage/async-storage";
// eslint-disable-next-line @typescript-eslint/no-empty-object-type
interface DashboardProps {}

const Dashboard: React.FC<DashboardProps> = () => {
  const [refreshing, setRefreshing] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedWordlist, setSelectedWordlist] =
    React.useState<Wordlist | null>(null);
  const [showWelcomeOverlay, setShowWelcomeOverlay] = useState(false);
  const [isNewUser, setIsNewUser] = useState(false);
  const [hasSeenDashboardFlag, setHasSeenDashboardFlag] = useState(false);
  const [hasCheckedWelcomeState, setHasCheckedWelcomeState] = useState(false);
  const [showCongratulationsModal, setShowCongratulationsModal] =
    useState(false);
  const [congratulationsWordlist, setCongratulationsWordlist] =
    useState<Wordlist | null>(null);
  const upgradeDialog = useUpgradePromptDialog();
  const router = useRouter();
  const { t } = useTranslation();
  const posthog = usePostHog();
  const { theme, responsive } = useTheme();
  const commonStyles = createCommonStyles(theme, responsive);
  const styles = createStyles(theme);
  const AnimatedTouchable = Animated.createAnimatedComponent(TouchableOpacity);

  // Animation refs
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const slideAnim = useRef(new Animated.Value(50)).current;
  const pulseAnim = useRef(new Animated.Value(1)).current;
  const pulseLoopRef = useRef<Animated.CompositeAnimation | null>(null);

  // Get premium status from centralized session
  const { isPremium } = useUserSession();

  // Fetch wordlists using the centralized hook
  const { data: wordlists, isLoading, refetch } = useWordlists();

  // Fetch batch progress data
  const { data: progressData } = useWordlistProgress();

  // Create progress map for O(1) lookup
  const progressMap = useMemo(() => {
    if (!progressData?.wordlists) return new Map();
    return new Map(progressData.wordlists.map((p) => [p.wordlistId, p]));
  }, [progressData]);

  const hasNoWordlist = wordlists && wordlists.length === 0;

  const stopPulse = React.useCallback(() => {
    if (pulseLoopRef.current) {
      pulseLoopRef.current.stop();
      pulseLoopRef.current = null;
    }
    pulseAnim.stopAnimation();
    pulseAnim.setValue(1);
  }, [pulseAnim]);

  const startPulse = React.useCallback(() => {
    if (pulseLoopRef.current) return; // avoid duplicates
    const maxScale = hasNoWordlist ? 1.1 : 1.05;
    const loop = Animated.loop(
      Animated.sequence([
        Animated.timing(pulseAnim, {
          toValue: maxScale,
          duration: 1200,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
        Animated.timing(pulseAnim, {
          toValue: 1,
          duration: 1200,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
      ]),
    );
    pulseLoopRef.current = loop;
    loop.start();
  }, [hasNoWordlist, pulseAnim]);

  // Check if user is first-time user immediately on mount
  useEffect(() => {
    let cancelled = false;
    const checkFirstTimeUser = async () => {
      try {
        const entries = await AsyncStorage.multiGet([
          "justSignedUp",
          "hasSeenDashboard",
        ]);
        const isJustSignedUp = entries.find(
          ([key]) => key === "justSignedUp",
        )?.[1];
        const hasSeenDashboard = entries.find(
          ([key]) => key === "hasSeenDashboard",
        )?.[1];

        if (!cancelled) {
          setHasSeenDashboardFlag(hasSeenDashboard === "true");
        }

        if (__DEV__) {
          console.log("Dashboard welcome check (immediate):", {
            isJustSignedUp,
            hasSeenDashboard,
          });
        }

        if (isJustSignedUp && !cancelled) {
          setIsNewUser(true);
          setShowWelcomeOverlay(true);

          await AsyncStorage.removeItem("justSignedUp");
          await AsyncStorage.setItem("hasSeenDashboard", "true");

          if (!cancelled) {
            setHasSeenDashboardFlag(true);
          }
        }
      } catch (error) {
        console.warn("Error checking first-time user status:", error);
      } finally {
        if (!cancelled) {
          setHasCheckedWelcomeState(true);
        }
      }
    };

    checkFirstTimeUser();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    posthog.capture("dashboard_viewed");
  }, [posthog]);

  // Animate when wordlist state changes
  useEffect(() => {
    if (!isLoading) {
      fadeAnim.setValue(0);
      slideAnim.setValue(50);

      Animated.parallel([
        Animated.timing(fadeAnim, {
          toValue: 1,
          duration: 400,
          useNativeDriver: true,
        }),
        Animated.timing(slideAnim, {
          toValue: 0,
          duration: 400,
          useNativeDriver: true,
        }),
      ]).start();
    }
  }, [hasNoWordlist, isLoading, fadeAnim, slideAnim]);

  // Start pulse when focused; stop on blur
  useFocusEffect(
    React.useCallback(() => {
      if (!isLoading) startPulse();
      return () => stopPulse();
    }, [isLoading, startPulse, stopPulse]),
  );

  // Restart pulse when empty state changes to update amplitude
  useEffect(() => {
    if (pulseLoopRef.current) {
      stopPulse();
      startPulse();
    }
  }, [hasNoWordlist, startPulse, stopPulse]);

  // Pause while the create modal is open
  useEffect(() => {
    if (showCreateModal) stopPulse();
    else startPulse();
  }, [showCreateModal, startPulse, stopPulse]);

  const markDashboardSeen = React.useCallback(() => {
    setHasSeenDashboardFlag(true);
    AsyncStorage.setItem("hasSeenDashboard", "true").catch((error) => {
      console.warn("Failed to persist dashboard seen flag:", error);
    });
  }, []);

  const handleWelcomeDismiss = React.useCallback(() => {
    setShowWelcomeOverlay(false);
    markDashboardSeen();
  }, [markDashboardSeen]);

  const handleWelcomeGetStarted = React.useCallback(() => {
    handleWelcomeDismiss();
    setShowCreateModal(true);
  }, [handleWelcomeDismiss]);

  const handleRefresh = async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  };

  const onUpgradePress = () => router.push(`/settings`);

  const renderWordlistItem = ({ item }: { item: Wordlist }) => (
    <Wordlistitem
      item={item}
      progress={progressMap.get(item.id)}
      onPressed={() => setSelectedWordlist(item)}
      onUpgradePress={onUpgradePress}
    />
  );
  const hideWordlistDetailModal = () => setSelectedWordlist(null);

  // Render skeleton items while loading
  const renderSkeletonItems = () => (
    <View>
      {[1, 2, 3].map((item) => (
        <WordlistItemSkeleton key={item} />
      ))}
    </View>
  );

  const handleAddNewWordlist = () => {
    const wordlistCount = wordlists?.length || 0;
    const isFreePlan = !isPremium;

    // Check if user has reached free plan limit
    if (isFreePlan && wordlistCount >= 1) {
      upgradeDialog.show();
    } else {
      setShowCreateModal(true);
    }
  };

  const renderStatsAndSection = () => (
    <>
      {/* Only show stats if user has wordlists */}
      {!hasNoWordlist && (
        <>
          {/* Stats */}
          <DashboardStats />

          {/* Section Header */}
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>
              {t("dashboard.wordlists.myWordlists")}
            </Text>
            <AnimatedTouchable
              style={[styles.addButton, { transform: [{ scale: pulseAnim }] }]}
              onPress={handleAddNewWordlist}
              onPressIn={stopPulse}
              onPressOut={startPulse}
              activeOpacity={0.85}
              accessibilityLabel={t(
                "dashboard.wordlists.addNewWordlist",
                "Add New Wordlist",
              )}
            >
              <Ionicons
                name="add-circle"
                size={34}
                color={theme.colors.primary}
              />
            </AnimatedTouchable>
          </View>
        </>
      )}
    </>
  );

  // Render empty state illustration when no wordlists
  const renderEmptyState = () => (
    <Animated.View
      style={[
        styles.emptyStateContainer,
        {
          opacity: fadeAnim,
          transform: [{ translateY: slideAnim }],
        },
      ]}
    >
      {/* Illustration Image */}
      <View style={styles.illustrationContainer}>
        <Image
          source={require("../../assets/images/empty-dashboard-bg.png")}
          style={styles.illustrationImage}
          resizeMode="contain"
        />
      </View>

      {/* Bottom content */}
      <View style={styles.emptyStateContent}>
        {/* No wordlists message */}
        <Text
          style={
            theme.mode === "dark"
              ? styles.noWordlistsTextDark
              : styles.noWordlistsText
          }
        >
          {t("dashboard.wordlists.noWordlistsYet")}
        </Text>

        {/* CTA Button */}
        <TouchableOpacity
          style={styles.ctaButton}
          onPress={() => setShowCreateModal(true)}
          activeOpacity={0.8}
        >
          <Text style={styles.ctaButtonText}>
            {t("dashboard.wordlists.createFirstWordlist")}
          </Text>
          <Ionicons
            name="add-circle"
            size={24}
            color="#FFFFFF"
            style={styles.ctaIcon}
          />
        </TouchableOpacity>
      </View>
    </Animated.View>
  );

  // Fallback: Show welcome for empty wordlists if no welcome was shown yet
  // Note: Don't set isNewUser flag here as this could be an existing user with no wordlists
  useEffect(() => {
    if (
      !hasCheckedWelcomeState ||
      isLoading ||
      !hasNoWordlist ||
      showWelcomeOverlay ||
      isNewUser ||
      hasSeenDashboardFlag
    ) {
      return;
    }

    const timer = setTimeout(() => {
      if (__DEV__) {
        console.log(
          "Fallback: Showing welcome for empty wordlist (persisting dismissal after first display)",
        );
      }
      setShowWelcomeOverlay(true);
      markDashboardSeen();
    }, 1000);

    return () => clearTimeout(timer);
  }, [
    hasCheckedWelcomeState,
    isLoading,
    hasNoWordlist,
    showWelcomeOverlay,
    isNewUser,
    hasSeenDashboardFlag,
    markDashboardSeen,
  ]);

  return (
    <>
      {theme.mode === "dark" ? (
        <SafeAreaView style={[commonStyles.safeArea, styles.containerDark]}>
          <OfflineIndicator />
          <Header />
          <Animated.View
            style={[
              { flex: 1 },
              !hasNoWordlist &&
                !isLoading && {
                  opacity: fadeAnim,
                  transform: [{ translateY: slideAnim }],
                },
            ]}
          >
            <FlatList
              data={isLoading ? [] : hasNoWordlist ? [] : wordlists}
              renderItem={renderWordlistItem}
              keyExtractor={(item) => String(item.id)}
              ListHeaderComponent={renderStatsAndSection}
              contentContainerStyle={styles.listContent}
              showsVerticalScrollIndicator={false}
              onScrollBeginDrag={stopPulse}
              onMomentumScrollEnd={startPulse}
              refreshControl={
                <RefreshControl
                  refreshing={refreshing}
                  onRefresh={handleRefresh}
                  colors={[theme.colors.primary]}
                  tintColor={theme.colors.primary}
                />
              }
              ListEmptyComponent={
                isLoading
                  ? renderSkeletonItems()
                  : hasNoWordlist
                    ? renderEmptyState()
                    : null
              }
            />
          </Animated.View>
        </SafeAreaView>
      ) : (
        <ImageBackground
          source={require("../../assets/images/signup-bg3.png")}
          style={styles.backgroundImage}
          imageStyle={styles.backgroundImageStyle}
          resizeMode="cover"
        >
          <SafeAreaView style={[commonStyles.safeArea, styles.container]}>
            <OfflineIndicator />
            <Header />
            <Animated.View
              style={[
                { flex: 1 },
                !hasNoWordlist &&
                  !isLoading && {
                    opacity: fadeAnim,
                    transform: [{ translateY: slideAnim }],
                  },
              ]}
            >
              <FlatList
                data={isLoading ? [] : hasNoWordlist ? [] : wordlists}
                renderItem={renderWordlistItem}
                keyExtractor={(item) => String(item.id)}
                ListHeaderComponent={renderStatsAndSection}
                contentContainerStyle={styles.listContent}
                showsVerticalScrollIndicator={false}
                onScrollBeginDrag={stopPulse}
                onMomentumScrollEnd={startPulse}
                refreshControl={
                  <RefreshControl
                    refreshing={refreshing}
                    onRefresh={handleRefresh}
                    colors={[theme.colors.primary]}
                    tintColor={theme.colors.primary}
                  />
                }
                ListEmptyComponent={
                  isLoading
                    ? renderSkeletonItems()
                    : hasNoWordlist
                      ? renderEmptyState()
                      : null
                }
              />
            </Animated.View>
          </SafeAreaView>
        </ImageBackground>
      )}

      {/* Create Wordlist Modal */}
      <CreateWordlistModal
        visible={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={(wordlist) => {
          setShowCreateModal(false);
          // Refresh the wordlists to show the new one
          refetch();

          // Special handling for first-time users
          if (isNewUser) {
            setIsNewUser(false);
            setCongratulationsWordlist(wordlist);
            setShowCongratulationsModal(true);
          } else {
            Alert.alert(
              t("common.success"),
              t("createWordlist.successMessage", { name: wordlist.name }),
            );
          }
        }}
      />

      {selectedWordlist && (
        <WordlistDetailModal
          visible
          onClose={hideWordlistDetailModal}
          wordlist={selectedWordlist}
        />
      )}

      {/* Congratulations Modal for First Wordlist */}
      <CongratulationsModal
        visible={showCongratulationsModal}
        onClose={() => setShowCongratulationsModal(false)}
        onAddWords={() => {
          if (congratulationsWordlist) {
            setSelectedWordlist(congratulationsWordlist);
          }
        }}
        wordlist={congratulationsWordlist}
      />

      {/* Welcome Overlay for First-Time Users */}
      <WelcomeOverlay
        visible={showWelcomeOverlay}
        onGetStarted={handleWelcomeGetStarted}
        onSkip={handleWelcomeDismiss}
      />
    </>
  );
};

export default Dashboard;

const { width } = Dimensions.get("window");

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    backgroundImage: {
      flex: 1,
      width: "100%",
      height: "100%",
      backgroundColor: "#FFF9F0", // Fallback warm background color
    },
    backgroundImageStyle: {
      opacity: 0.7, // Prominent yet readable background for dashboard
      // Force full coverage regardless of aspect ratio
      width: "100%",
      height: "100%",
    },
    container: {
      flex: 1,
      backgroundColor: "transparent", // Let background image show through
    },
    containerDark: {
      flex: 1,
      backgroundColor: theme.colors.background.default, // Use theme default for dark mode
    },
    listContent: {
      paddingBottom: 30,
    },
    header: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      paddingTop: 10,
      paddingBottom: 20,
      paddingHorizontal: 20,
    },
    settingsButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: theme.colors.background.surface,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    profileButton: {
      width: 40,
      height: 40,
      borderRadius: 20,
      overflow: "hidden",
      backgroundColor: theme.colors.background.surface,
      ...theme.shadows.sm,
    },
    profileImage: {
      width: "100%",
      height: "100%",
    },
    statsContainer: {
      flexDirection: "row",
      justifyContent: "space-between",
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      paddingVertical: theme.spacing.lg,
      paddingHorizontal: 10,
      marginHorizontal: 20,
      marginBottom: 24,
      ...theme.shadows.sm,
    },
    statItem: {
      alignItems: "center",
      flex: 1,
    },
    statLabel: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginBottom: 8,
    },
    statValue: {
      fontSize: 36,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    sectionHeader: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      paddingHorizontal: 20,
      marginBottom: 16,
    },
    sectionTitle: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    addButton: {
      width: 44,
      height: 44,
      borderRadius: 22,
      alignItems: "center",
      justifyContent: "center",
      backgroundColor:
        theme.mode === "light" ? "#FFFFFF" : theme.colors.background.surface,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? theme.colors.border.light
          : theme.colors.ui.border,
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: theme.mode === "light" ? 0.18 : 0.3,
      shadowRadius: 12,
      elevation: 10,
    },

    emptyContainer: {
      paddingHorizontal: 20,
      paddingTop: 40,
      alignItems: "center",
    },
    emptyText: {
      fontSize: 18,
      color: theme.colors.text.secondary,
      marginBottom: 20,
    },
    ctaButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: theme.borderRadius.md,
      paddingVertical: theme.spacing.md,
      paddingHorizontal: theme.spacing.lg,
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      gap: 8,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
    },
    ctaButtonText: {
      color: theme.colors.text.inverse,
      fontSize: 18,
      fontWeight: "600",
    },
    ctaIcon: {
      marginLeft: 4,
    },
    // Empty state styles
    emptyStateContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: 20,
    },
    illustrationContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
    },
    illustrationImage: {
      width: width,
      height: width * 0.7, // Adjust ratio based on illustration
      maxHeight: 300,
    },
    emptyStateContent: {
      paddingHorizontal: 20,
      paddingBottom: 30,
      alignItems: "center",
    },
    noWordlistsText: {
      fontSize: 20,
      fontWeight: "500",
      color: "#2D3436",
      textAlign: "center",
      marginBottom: 30,
    },
    noWordlistsTextDark: {
      fontSize: 20,
      fontWeight: "500",
      color: "#FFFFFF",
      textAlign: "center",
      marginBottom: 30,
    },
  });
