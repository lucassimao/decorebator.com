import * as subscriptionsApi from "@/api/subscriptions";
import * as wordlistsApi from "@/api/wordlists";
import { Wordlist } from "@/api/wordlists";
import { CreateWordlistModal } from "@/components/dashboard/CreateWordlistModal";
import { Header } from "@/components/dashboard/Header";
import DashboardStats from "@/components/dashboard/Stats";
import { WordlistDetailModal } from "@/components/dashboard/WordlistDetailModal";
import Wordlistitem from "@/components/dashboard/WordlistItem";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { useTheme } from "@/contexts/ThemeContext";
import { useUpgradePromptDialog } from "@/hooks/useUpgradePromptDialog";
import { useWordlistProgress } from "@/hooks/useWordlistProgress";
import { createCommonStyles } from "@/styles/common";
import { Ionicons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "expo-router";
import React, { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  FlatList,
  RefreshControl,
  SafeAreaView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
// eslint-disable-next-line @typescript-eslint/no-empty-object-type
interface DashboardProps {}

const Dashboard: React.FC<DashboardProps> = () => {
  const [refreshing, setRefreshing] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedWordlist, setSelectedWordlist] =
    React.useState<Wordlist | null>(null);
  const upgradeDialog = useUpgradePromptDialog();
  const router = useRouter();
  const { t } = useTranslation();
  const { theme } = useTheme();
  const commonStyles = createCommonStyles(theme);
  const styles = createStyles(theme);

  // Fetch subscription
  const { data: subscription } = useQuery({
    queryKey: ["subscription"],
    queryFn: subscriptionsApi.getSubscriptionStatus,
    staleTime: 0, // data is stale as soon as it arrives
    // ---- always refetch on mount or when window regains focus ----
    refetchOnMount: "always",
    refetchOnWindowFocus: "always",
  });

  // Fetch wordlists
  const {
    data: wordlists,
    isLoading,
    refetch,
  } = useQuery<wordlistsApi.Wordlist[], Error>({
    queryFn: () => wordlistsApi.getUserWordlists(),
    queryKey: ["wordlists"],
    // Ensure fresh data on mount
    staleTime: 0,
    refetchOnMount: "always",
  });

  // Fetch batch progress data
  const { data: progressData } = useWordlistProgress();

  // Create progress map for O(1) lookup
  const progressMap = useMemo(() => {
    if (!progressData?.wordlists) return new Map();
    return new Map(progressData.wordlists.map((p) => [p.wordlistId, p]));
  }, [progressData]);

  const hasNoWordlist = wordlists && wordlists.length === 0;

  useEffect(() => {
    if (isLoading) return;

    if (hasNoWordlist) {
      router.replace("/dashboard/welcome");
    }
  }, [wordlists, isLoading, router, hasNoWordlist]);

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

  const renderHeader = () => (
    <>
      <Header />

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
          <Ionicons name="add-circle" size={34} color={theme.colors.primary} />
        </TouchableOpacity>
      </View>
    </>
  );

  // Redirect to welcome screen if no wordlists
  if (!isLoading && hasNoWordlist) {
    router.replace("/dashboard/welcome");
    return null;
  }

  return (
    <>
      <SafeAreaView style={[commonStyles.safeArea, styles.container]}>
        <OfflineIndicator />
        <FlatList
          data={isLoading ? [] : wordlists}
          renderItem={renderWordlistItem}
          keyExtractor={(item) => String(item.id)}
          ListHeaderComponent={renderHeader}
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
            isLoading ? (
              <View style={styles.loadingContainer}>
                <ActivityIndicator size="large" color={theme.colors.primary} />
                <Text style={styles.loadingText}>{t("common.loading")}</Text>
              </View>
            ) : (
              <View style={styles.emptyContainer}>
                <Text style={styles.emptyText}>
                  {t("dashboard.wordlists.noWordlistsYet")}
                </Text>
                <TouchableOpacity
                  style={styles.ctaButton}
                  onPress={() => setShowCreateModal(true)}
                >
                  <Text style={styles.ctaButtonText}>
                    {t("dashboard.wordlists.createFirstWordlist")}
                  </Text>
                  <Ionicons
                    name="add-circle"
                    size={24}
                    color={theme.colors.text.inverse}
                  />
                </TouchableOpacity>
              </View>
            )
          }
        />
      </SafeAreaView>

      {/* Create Wordlist Modal */}
      <CreateWordlistModal
        visible={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={() => setShowCreateModal(false)}
      />

      {selectedWordlist && (
        <WordlistDetailModal
          visible
          onClose={hideWordlistDetailModal}
          wordlist={selectedWordlist}
        />
      )}
    </>
  );
};

export default Dashboard;

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    loadingContainer: {
      paddingVertical: 60,
      justifyContent: "center",
      alignItems: "center",
    },
    loadingText: {
      marginTop: 12,
      fontSize: 16,
      color: theme.colors.text.secondary,
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
  });
