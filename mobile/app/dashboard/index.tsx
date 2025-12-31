import { Wordlist } from "@/api/wordlists";
import { useUserSession } from "@/hooks/useUserSession";
import { CreateWordlistModal } from "@/components/dashboard/CreateWordlistModal";
import { CongratulationsModal } from "@/components/dashboard/CongratulationsModal";
import { Header } from "@/components/dashboard/Header";
import DashboardStats from "@/components/dashboard/Stats";
import { WordlistDetailModal } from "@/components/dashboard/WordlistDetailModal";
import Wordlistitem from "@/components/dashboard/WordlistItem";
import { WelcomeOverlay } from "@/components/dashboard/WelcomeOverlay";
import { EmptyState } from "@/components/dashboard/EmptyState";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { WordlistItemSkeleton } from "@/components/ui/WordlistItemSkeleton";
import { useTheme } from "@/contexts/ThemeContext";
import { useUpgradePromptDialog } from "@/hooks/useUpgradePromptDialog";
import { useWordlistProgress } from "@/hooks/useWordlistProgress";
import { useWordlists } from "@/hooks/useWordlists";
import { useDashboardAnimations } from "@/hooks/useDashboardAnimations";
import { useWelcomeState } from "@/hooks/useWelcomeState";
import { createCommonStyles } from "@/styles/common";
import { Ionicons } from "@expo/vector-icons";
import { useRouter } from "expo-router";
import React, { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { usePostHog } from "posthog-react-native";
import {
  Alert,
  Animated,
  FlatList,
  ImageBackground,
  RefreshControl,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
// eslint-disable-next-line @typescript-eslint/no-empty-object-type
interface DashboardProps {}

const Dashboard: React.FC<DashboardProps> = () => {
  const [refreshing, setRefreshing] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedWordlist, setSelectedWordlist] =
    React.useState<Wordlist | null>(null);
  const [openAddWordOnSelect, setOpenAddWordOnSelect] = useState(false);
  const [showCongratulationsModal, setShowCongratulationsModal] =
    useState(false);
  const [congratulationsWordlist, setCongratulationsWordlist] =
    useState<Wordlist | null>(null);
  const [lockTemporarilyUnlocked, setLockTemporarilyUnlocked] = useState(false);

  const upgradeDialog = useUpgradePromptDialog();
  const router = useRouter();
  const { t } = useTranslation();
  const posthog = usePostHog();
  const { theme, responsive } = useTheme();
  const commonStyles = createCommonStyles(theme, responsive);
  const styles = createStyles(theme);
  const AnimatedTouchable = Animated.createAnimatedComponent(TouchableOpacity);

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

  const wordlistsData = Array.isArray(wordlists) ? wordlists : [];
  const hasNoWordlist = !isLoading && wordlistsData.length === 0;
  const wordlistCount = wordlistsData.length;

  // Use custom animation hook
  const {
    fadeAnim,
    slideAnim,
    pulseAnim,
    lockNudgeAnim,
    startPulse,
    stopPulse,
    triggerLockNudge,
  } = useDashboardAnimations({
    hasNoWordlist,
    isLoading,
    showCreateModal,
  });

  // Use custom welcome state hook
  const {
    showWelcomeOverlay,
    isNewUser,
    setIsNewUser,
    handleWelcomeDismiss,
    handleWelcomeGetStarted,
  } = useWelcomeState({
    hasNoWordlist,
    isLoading,
  });

  useEffect(() => {
    posthog.capture("dashboard_viewed");
  }, [posthog]);

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
      onAddWords={() => {
        setOpenAddWordOnSelect(true);
        setSelectedWordlist(item);
      }}
      onUpgradePress={onUpgradePress}
    />
  );
  const hideWordlistDetailModal = () => {
    setSelectedWordlist(null);
    setOpenAddWordOnSelect(false);
  };

  // Render skeleton items while loading
  const renderSkeletonItems = () => (
    <View>
      {[1, 2, 3].map((item) => (
        <WordlistItemSkeleton key={item} />
      ))}
    </View>
  );

  const handleAddNewWordlist = () => {
    const wordlistCount = wordlistsData.length;
    const isFreePlan = !isPremium;

    // Check if user has reached free plan limit
    if (isFreePlan && wordlistCount >= 1) {
      upgradeDialog.show();
    } else {
      setShowCreateModal(true);
    }
  };

  const triggerLockUnlock = React.useCallback(() => {
    setLockTemporarilyUnlocked(true);
    const timeout = setTimeout(() => {
      setLockTemporarilyUnlocked(false);
    }, 600);
    return () => clearTimeout(timeout);
  }, []);

  const renderStatsAndSection = () => (
    <>
      {/* Only show stats if user has wordlists */}
      {!hasNoWordlist && (
        <>
          {/* Stats */}
          <DashboardStats
            onAddFirstWords={() => {
              if (wordlistsData.length === 1) {
                setOpenAddWordOnSelect(true);
                setSelectedWordlist(wordlistsData[0]);
              }
            }}
            wordlists={wordlistsData}
          />

          {/* Section Header */}
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>
              {t("dashboard.wordlists.myWordlists")}
            </Text>
          </View>
        </>
      )}
    </>
  );

  // Render empty state illustration when no wordlists
  const renderEmptyState = () => (
    <EmptyState
      fadeAnim={fadeAnim}
      slideAnim={slideAnim}
      lockNudgeAnim={lockNudgeAnim}
      lockTemporarilyUnlocked={lockTemporarilyUnlocked}
      onCreateWordlist={() => setShowCreateModal(true)}
      onTriggerLockNudge={triggerLockNudge}
      onTriggerLockUnlock={triggerLockUnlock}
    />
  );

  // Shared dashboard content
  const renderDashboardContent = () => (
    <>
      <OfflineIndicator />
      <Header />
      <Animated.View
        key={`wordlist-content-${wordlistCount}`}
        style={[
          { flex: 1 },
          !hasNoWordlist &&
            !isLoading && {
              opacity: fadeAnim,
              transform: [{ translateY: slideAnim }],
            },
        ]}
      >
        {hasNoWordlist && !isLoading ? (
          renderEmptyState()
        ) : (
          <FlatList
            data={isLoading ? [] : wordlistsData}
            renderItem={renderWordlistItem}
            keyExtractor={(item) => String(item.id)}
            ListHeaderComponent={renderStatsAndSection}
            contentContainerStyle={[
              styles.listContent,
              hasNoWordlist && styles.listContentEmpty,
            ]}
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
            ListEmptyComponent={isLoading ? renderSkeletonItems() : null}
          />
        )}
      </Animated.View>

      {/* Floating Action Button */}
      {!hasNoWordlist && (
        <AnimatedTouchable
          style={[styles.fab, { transform: [{ scale: pulseAnim }] }]}
          onPress={handleAddNewWordlist}
          onPressIn={stopPulse}
          onPressOut={startPulse}
          activeOpacity={0.85}
          accessibilityLabel={t(
            "dashboard.wordlists.addNewWordlist",
            "Add New Wordlist",
          )}
        >
          <Ionicons name="add" size={28} color="#FFFFFF" />
        </AnimatedTouchable>
      )}
    </>
  );

  return (
    <>
      {theme.mode === "dark" ? (
        <SafeAreaView style={[commonStyles.safeArea, styles.containerDark]}>
          {renderDashboardContent()}
        </SafeAreaView>
      ) : (
        <ImageBackground
          source={require("../../assets/images/signup-bg3.png")}
          style={styles.backgroundImage}
          imageStyle={styles.backgroundImageStyle}
          resizeMode="cover"
        >
          <SafeAreaView style={[commonStyles.safeArea, styles.container]}>
            {renderDashboardContent()}
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
          startInAddMode={openAddWordOnSelect}
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
        onGetStarted={() =>
          handleWelcomeGetStarted(() => setShowCreateModal(true))
        }
        onSkip={handleWelcomeDismiss}
      />
    </>
  );
};

export default Dashboard;

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
    listContentEmpty: {
      flexGrow: 1,
      justifyContent: "center",
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
    fab: {
      position: "absolute",
      right: 20,
      bottom: 20,
      width: 56,
      height: 56,
      borderRadius: 28,
      backgroundColor: theme.colors.primary,
      alignItems: "center",
      justifyContent: "center",
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.3,
      shadowRadius: 8,
      elevation: 8,
    },
  });
