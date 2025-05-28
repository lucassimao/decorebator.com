
import * as usersApi from "@/api/users";
import * as wordlistsApi from "@/api/wordlists";
import { Wordlist } from "@/api/wordlists";
import { CreateWordlistModal } from "@/components/dashboard/CreateWordlistModal";
import DashboardStats from "@/components/dashboard/Stats";
import Wordlistitem from "@/components/dashboard/WordlistItem";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useFocusEffect } from "@react-navigation/native";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "expo-router";
import React, { useEffect, useState } from "react";
import {
  ActivityIndicator,
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

const { width, height } = Dimensions.get("window");



interface DashboardProps {
}

const Dashboard: React.FC<DashboardProps> = () => {
  const [refreshing, setRefreshing] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [user, setUser] = React.useState(usersApi.getUserInfo());
  const router = useRouter()

  const {
    data: wordlists,
    isLoading,
    refetch,
  } = useQuery<wordlistsApi.Wordlist[], Error>({
    queryFn: () => wordlistsApi.getUserWordlists(),
    queryKey: ["wordlists"],
  });

  useEffect(() => {
    if (isLoading) return;

    const isEmpty = !wordlists || wordlists.length == 0;

    if (isEmpty) {
      router.push("/dashboard/welcome");
    }
  }, [wordlists, isLoading, router]);

  const handleRefresh = async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  };



  const handleSettingsPress = () => {
    // navigation.navigate('Settings');
  };

  const handleProfilePress = () => {
    // navigation.navigate('Profile');
  };

  // Refresh user session when screen comes into focus
  useFocusEffect(
    React.useCallback(() => {
      const refreshUserSession = async () => {
        try {
          const updatedUser = await usersApi.refreshToken();
          setUser(updatedUser);
        } catch (error) {
          // If refresh fails, just use the cached user info
          console.error("Failed to refresh user session:", error);
        }
      };

      refreshUserSession();
    }, []),
  );

  const renderWordlistItem = ({ item }: { item: any }) => <Wordlistitem item={item}/>

  const renderHeader = () => (
    <>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity
          style={styles.settingsButton}
          onPress={handleSettingsPress}
        >
          <Ionicons name="settings-outline" size={24} color="#2D3436" />
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.profileButton}
          onPress={handleProfilePress}
        >
          <Image
            source={{ uri: "https://via.placeholder.com/40" }}
            style={styles.profileImage}
          />
        </TouchableOpacity>
      </View>

      {/* Greeting */}
      <View style={styles.greetingContainer}>
        <Text style={styles.greeting}>Good Morning,</Text>
        <Text style={styles.userName}>{user?.firstName}!</Text>
        <Text style={styles.subtitle}>Ready to learn something new today?</Text>
      </View>

      {/* Stats */}
     <DashboardStats />

      {/* Section Header */}
      <View style={styles.sectionHeader}>
        <Text style={styles.sectionTitle}>My Wordlists</Text>
        <TouchableOpacity
          style={styles.addButton}
          onPress={() => setShowCreateModal(true)}
        >
          <Ionicons name="add-circle" size={24} color="#FF7B54" />
        </TouchableOpacity>
      </View>
    </>
  );

  if (isLoading && !wordlists) {
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
                <Text style={styles.emptyText}>No wordlists yet</Text>
                <TouchableOpacity
                  style={styles.ctaButton}
                  onPress={() => setShowCreateModal(true)}
                >
                  <Text style={styles.ctaButtonText}>
                    Create your first wordlist
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
  greetingContainer: {
    marginBottom: 20,
    paddingHorizontal: 20,
  },
  greeting: {
    fontSize: 28,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 4,
  },
  userName: {
    fontSize: 28,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: "#636E72",
    marginTop: 4,
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
