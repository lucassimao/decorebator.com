import * as subscriptionsApi from "@/api/subscriptions";
import * as wordlistsApi from "@/api/wordlists";
import { Wordlist } from "@/api/wordlists";
import { CreateWordlistModal } from "@/components/dashboard/CreateWordlistModal";
import { Header } from "@/components/dashboard/Header";
import DashboardStats from "@/components/dashboard/Stats";
import { WordlistDetailModal } from "@/components/dashboard/WordlistDetailModal";
import Wordlistitem from "@/components/dashboard/WordlistItem";
import { useUpgradePromptDialog } from "@/hooks/useUpgradePromptDialog";
import { useOffline } from "@/hooks/useOffline";
import { OfflineIndicator } from "@/components/OfflineIndicator";
import { Ionicons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "expo-router";
import React, { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Dimensions,
  FlatList,
  ImageBackground,
  RefreshControl,
  SafeAreaView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";

const { width, height } = Dimensions.get("window");

interface DashboardProps {}

const Dashboard: React.FC<DashboardProps> = () => {
  const [refreshing, setRefreshing] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedWordlist, setSelectedWordlist] =
    React.useState<Wordlist | null>(null);
  const upgradeDialog = useUpgradePromptDialog();
  const router = useRouter();
  const { t } = useTranslation();

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
  });

  const hasNoWordlist = wordlists && wordlists.length == 0;

  useEffect(() => {
    if (isLoading) return;

    if (hasNoWordlist) {
      router.push("/dashboard/welcome");
    }
  }, [wordlists, isLoading, router]);

  const handleRefresh = async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  };

  const renderWordlistItem = ({ item }: { item: Wordlist }) => (
    <Wordlistitem item={item} onPressed={() => setSelectedWordlist(item)} />
  );
  const hideWordlistDetailModal = () => setSelectedWordlist(null);

  const handleAddNewWordlist = () => {
    const wordlistCount = wordlists?.length || 0;
    const isFreePlan = !subscription || subscription.plan == "free";

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
          <Ionicons name="add-circle" size={34} color="#FF7B54" />
        </TouchableOpacity>
      </View>
    </>
  );

  // so that we keep the loader till we trigger the redirect to /dashboard/welcome
  if (isLoading || hasNoWordlist) {
    return (
      <ImageBackground
        source={require("@/assets/images/dashboard-bg.png")}
        style={styles.backgroundImage}
        resizeMode="cover"
      >
        <SafeAreaView style={styles.container}>
          <View style={styles.loadingContainer}>
            <ActivityIndicator size="large" color="#FF7B54" />
          </View>
        </SafeAreaView>
      </ImageBackground>
    );
  }

  return (
    <>
      <ImageBackground
        source={require("@/assets/images/dashboard-bg.png")}
        style={styles.backgroundImage}
        resizeMode="cover"
      >
        <SafeAreaView style={styles.container}>
          <OfflineIndicator />
          <FlatList
            data={wordlists}
            renderItem={renderWordlistItem}
            keyExtractor={(item) => String(item.id)}
            ListHeaderComponent={renderHeader}
            contentContainerStyle={styles.listContent}
            showsVerticalScrollIndicator={false}
            refreshControl={
              <RefreshControl
                refreshing={refreshing}
                onRefresh={handleRefresh}
                colors={["#FF7B54"]}
                tintColor="#FF7B54"
              />
            }
            ListEmptyComponent={
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
                  <Ionicons name="add-circle" size={24} color="#FFFFFF" />
                </TouchableOpacity>
              </View>
            }
          />
        </SafeAreaView>
      </ImageBackground>

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

const styles = StyleSheet.create({
  backgroundImage: {
    flex: 1,
    width: width,
    height: height,
  },
  container: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
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
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  profileButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    overflow: "hidden",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
    backgroundColor: "#FFFFFF",
  },
  profileImage: {
    width: "100%",
    height: "100%",
  },
  statsContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    backgroundColor: "rgba(255, 255, 255, 0.7)",
    borderRadius: 16,
    paddingVertical: 20,
    paddingHorizontal: 10,
    marginHorizontal: 20,
    marginBottom: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 8,
    elevation: 2,
  },
  statItem: {
    alignItems: "center",
    flex: 1,
  },
  statLabel: {
    fontSize: 14,
    color: "#636E72",
    marginBottom: 8,
  },
  statValue: {
    fontSize: 36,
    fontWeight: "700",
    color: "#2D3436",
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
    color: "#2D3436",
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
    color: "#636E72",
    marginBottom: 20,
  },
  ctaButton: {
    backgroundColor: "#FF7B54",
    borderRadius: 12,
    paddingVertical: 16,
    paddingHorizontal: 24,
    flexDirection: "row",
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 5,
    gap: 8,
  },
  ctaButtonText: {
    color: "#FFFFFF",
    fontSize: 18,
    fontWeight: "600",
  },
});
