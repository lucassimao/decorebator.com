import * as subscriptionsApi from "@/api/subscriptions";
import { Wordlist } from "@/api/wordlists";
import { CreateWordlistModal } from "@/components/dashboard/CreateWordlistModal";
import { CongratulationsModal } from "@/components/dashboard/CongratulationsModal";
import { Header } from "@/components/dashboard/Header";
import DashboardStats from "@/components/dashboard/Stats";
import { WordlistDetailModal } from "@/components/dashboard/WordlistDetailModal";
import Wordlistitem from "@/components/dashboard/WordlistItem";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { WordlistItemSkeleton } from "@/components/ui/WordlistItemSkeleton";
import { useTheme } from "@/contexts/ThemeContext";
import { useUpgradePromptDialog } from "@/hooks/useUpgradePromptDialog";
import { useWordlistProgress } from "@/hooks/useWordlistProgress";
import { useWordlists } from "@/hooks/useWordlists";
import { createCommonStyles } from "@/styles/common";
import { Ionicons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "expo-router";
import React, { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
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
  const [showCongratulationsModal, setShowCongratulationsModal] =
    useState(false);
  const [congratulationsWordlist, setCongratulationsWordlist] =
    useState<Wordlist | null>(null);
  const upgradeDialog = useUpgradePromptDialog();
  const router = useRouter();
  const { t } = useTranslation();
  const { theme } = useTheme();
  const commonStyles = createCommonStyles(theme);
  const styles = createStyles(theme);

  // Animation refs
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const slideAnim = useRef(new Animated.Value(50)).current;
  const welcomeOpacity = useRef(new Animated.Value(0)).current;

  // Fetch subscription
  const { data: subscription } = useQuery({
    queryKey: ["subscription"],
    queryFn: subscriptionsApi.getSubscriptionStatus,
    staleTime: 0, // data is stale as soon as it arrives
    // ---- always refetch on mount or when window regains focus ----
    refetchOnMount: "always",
    refetchOnWindowFocus: "always",
  });

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

  // Check if user is first-time user on mount
  useEffect(() => {
    const checkFirstTimeUser = async () => {
      try {
        const hasSeenDashboard = await AsyncStorage.getItem("hasSeenDashboard");
        const isJustSignedUp = await AsyncStorage.getItem("justSignedUp");

        if (__DEV__) {
          console.log("Dashboard welcome check:", {
            hasSeenDashboard,
            isJustSignedUp,
          });
        }

        // Only show welcome if user literally just signed up
        if (isJustSignedUp) {
          if (__DEV__) {
            console.log("Showing welcome overlay for newly signed up user");
          }
          setIsNewUser(true);
          setShowWelcomeOverlay(true);

          // Clear the signup flag and mark dashboard as seen
          await AsyncStorage.removeItem("justSignedUp");
          await AsyncStorage.setItem("hasSeenDashboard", "true");
        }
      } catch (error) {
        console.warn("Error checking first-time user status:", error);
      }
    };

    // Check immediately when component mounts, don't wait for wordlists
    if (!isLoading) {
      // Small delay to ensure component is fully mounted
      const timer = setTimeout(() => {
        checkFirstTimeUser();
      }, 300);

      return () => clearTimeout(timer);
    }
  }, [isLoading]);

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
    const isFreePlan = !subscription || subscription.plan === "free";

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
            <TouchableOpacity
              style={styles.addButton}
              onPress={handleAddNewWordlist}
            >
              <Ionicons
                name="add-circle"
                size={34}
                color={theme.colors.primary}
              />
            </TouchableOpacity>
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

  // Welcome overlay for first-time users
  const renderWelcomeOverlay = () => {
    if (!showWelcomeOverlay) return null;

    return (
      <View style={styles.welcomeOverlay}>
        <Animated.View
          style={[styles.welcomeModal, { opacity: welcomeOpacity }]}
        >
          <View style={styles.welcomeHeader}>
            <Text style={styles.welcomeTitle}>
              {t("welcome.title", "Welcome to Decorebator! 🎉")}
            </Text>
            <Text style={styles.welcomeSubtitle}>
              {t(
                "welcome.subtitle",
                "Your AI-powered language learning journey starts here",
              )}
            </Text>
          </View>

          <View style={styles.welcomeFeatures}>
            <View style={styles.welcomeFeature}>
              <View style={styles.welcomeFeatureIcon}>
                <Ionicons name="book" size={24} color={theme.colors.primary} />
              </View>
              <Text style={styles.welcomeFeatureText}>
                {t("welcome.feature1", "Create custom wordlists")}
              </Text>
            </View>

            <View style={styles.welcomeFeature}>
              <View style={styles.welcomeFeatureIcon}>
                <Ionicons name="bulb" size={24} color={theme.colors.primary} />
              </View>
              <Text style={styles.welcomeFeatureText}>
                {t("welcome.feature2", "AI-powered definitions & images")}
              </Text>
            </View>

            <View style={styles.welcomeFeature}>
              <View style={styles.welcomeFeatureIcon}>
                <Ionicons
                  name="repeat"
                  size={24}
                  color={theme.colors.primary}
                />
              </View>
              <Text style={styles.welcomeFeatureText}>
                {t("welcome.feature3", "Smart spaced repetition")}
              </Text>
            </View>
          </View>

          <TouchableOpacity
            style={styles.welcomeButton}
            onPress={() => {
              setShowWelcomeOverlay(false);
              setShowCreateModal(true);
            }}
            activeOpacity={0.8}
          >
            <Text style={styles.welcomeButtonText}>
              {t("welcome.getStarted", "Create Your First Wordlist")}
            </Text>
            <Ionicons name="arrow-forward" size={20} color="#FFFFFF" />
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.welcomeSkip}
            onPress={() => setShowWelcomeOverlay(false)}
            activeOpacity={0.7}
          >
            <Text style={styles.welcomeSkipText}>
              {t("welcome.skip", "Skip for now")}
            </Text>
          </TouchableOpacity>
        </Animated.View>
      </View>
    );
  };

  // Animate welcome overlay when shown
  useEffect(() => {
    if (showWelcomeOverlay) {
      welcomeOpacity.setValue(0);
      Animated.timing(welcomeOpacity, {
        toValue: 1,
        duration: 300,
        useNativeDriver: true,
      }).start();
    }
  }, [showWelcomeOverlay, welcomeOpacity]);

  // Fallback: Show welcome for empty wordlists if no welcome was shown yet
  // Note: Don't set isNewUser flag here as this could be an existing user with no wordlists
  useEffect(() => {
    if (!isLoading && hasNoWordlist && !showWelcomeOverlay && !isNewUser) {
      const timer = setTimeout(() => {
        if (__DEV__) {
          console.log(
            "Fallback: Showing welcome for empty wordlist (not marking as new user)",
          );
        }
        // Only show welcome overlay, don't mark as new user
        setShowWelcomeOverlay(true);
      }, 1000);

      return () => clearTimeout(timer);
    }
  }, [isLoading, hasNoWordlist, showWelcomeOverlay, isNewUser]);

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
      {renderWelcomeOverlay()}
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
      padding: 4,
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
    // Welcome overlay styles
    welcomeOverlay: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: "rgba(0, 0, 0, 0.6)",
      justifyContent: "center",
      alignItems: "center",
      zIndex: 1000,
    },
    welcomeModal: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: 20,
      padding: 24,
      marginHorizontal: 20,
      maxWidth: 400,
      width: "100%",
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.3,
      shadowRadius: 12,
      elevation: 8,
    },
    welcomeHeader: {
      alignItems: "center",
      marginBottom: 24,
    },
    welcomeTitle: {
      fontSize: 24,
      fontWeight: "700",
      color: theme.colors.text.primary,
      textAlign: "center",
      marginBottom: 8,
    },
    welcomeSubtitle: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: 22,
    },
    welcomeFeatures: {
      marginBottom: 24,
    },
    welcomeFeature: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: 16,
    },
    welcomeFeatureIcon: {
      width: 40,
      height: 40,
      borderRadius: 20,
      backgroundColor: `${theme.colors.primary}15`,
      justifyContent: "center",
      alignItems: "center",
      marginRight: 16,
    },
    welcomeFeatureText: {
      flex: 1,
      fontSize: 16,
      color: theme.colors.text.primary,
      fontWeight: "500",
    },
    welcomeButton: {
      backgroundColor: theme.colors.primary,
      borderRadius: 12,
      paddingVertical: 16,
      paddingHorizontal: 24,
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      marginBottom: 12,
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.3,
      shadowRadius: 8,
      elevation: 5,
    },
    welcomeButtonText: {
      color: "#FFFFFF",
      fontSize: 18,
      fontWeight: "600",
      marginRight: 8,
    },
    welcomeSkip: {
      paddingVertical: 12,
      alignItems: "center",
    },
    welcomeSkipText: {
      color: theme.colors.text.secondary,
      fontSize: 16,
      fontWeight: "500",
    },
  });
