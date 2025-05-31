type PlanRecurrence = "annual" | "monthly";

export interface PricingPlan {
  id: PlanRecurrence;
  name: string;
  price: number;
  currency: string;
  interval: "month" | "year";
  features: string[];
  popular?: boolean;
  savings?: string;
}

import * as subscriptionsApi from "@/api/subscriptions";
import * as usersApi from "@/api/users";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as AuthSession from "expo-auth-session";
import * as MailComposer from "expo-mail-composer";
import { router } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import React, { useState } from "react";
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  ImageBackground,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
const { width: SCREEN_WIDTH } = Dimensions.get("window");

const PRICING_PLANS: PricingPlan[] = [
  {
    id: "monthly",
    name: "Monthly",
    price: 6.99,
    currency: "USD",
    interval: "month",
    features: [
      "Unlimited wordlists",
      "AI-powered word enrichment",
      "All quiz modes unlocked",
      "Progress tracking",
      "Priority support",
    ],
  },
  {
    id: "annual",
    name: "Yearly",
    price: 69.99,
    currency: "USD",
    interval: "year",
    popular: true,
    savings: "Save 17%",
    features: [
      "Unlimited wordlists",
      "AI-powered word enrichment",
      "All quiz modes unlocked",
      "Progress tracking",
      "Priority support",
    ],
  },
];

const SettingsScreen: React.FC = () => {
  const navigation = useNavigation();
  const queryClient = useQueryClient();
  const [selectedPlan, setSelectedPlan] = useState<PlanRecurrence | null>(null);

  // Fetch subscription
  const {
    data: subscription,
    isLoading,
    refetch: refetchSubscription,
  } = useQuery({
    queryKey: ["subscription"],
    queryFn: subscriptionsApi.getSubscriptionStatus,

    // ---- make everything immediately stale & un-cached ----
    staleTime: 0, // data is stale as soon as it arrives

    // ---- always refetch on mount or when window regains focus ----
    refetchOnMount: "always",
    refetchOnWindowFocus: "always",
  });

  const checkoutMutation = useMutation({
    mutationFn: async (plan: PlanRecurrence) => {
      const redirectUri = AuthSession.makeRedirectUri({
        scheme: "decorerbator",
        path: "stripe",
      });

      const res = await subscriptionsApi.createCheckoutSession(
        plan,
        redirectUri,
      );

      return { checkoutUrl: res.checkoutUrl, redirectUri };
    },
    onSuccess: async (data) => {
      const result = await WebBrowser.openAuthSessionAsync(
        data.checkoutUrl,
        data.redirectUri,
      );

      if (result.type === "success") {
        // Refresh subscription data
        refetchSubscription();
        Alert.alert("Success", "Subscription activated successfully!");
      }
    },
    onError: () => {
      Alert.alert("Error", "Failed to open checkout. Please try again.");
    },
  });

  // Cancel subscription mutation
  const cancelMutation = useMutation({
    mutationFn: subscriptionsApi.cancelSubscription,
    onSuccess: () => {
      refetchSubscription();
      Alert.alert(
        "Success",
        "Your subscription will be canceled at the end of the current period.",
      );
    },
    onError: () => {
      Alert.alert("Error", "Failed to cancel subscription. Please try again.");
    },
  });

  const signOut = () => {
    usersApi.sigout();
    queryClient.cancelQueries();
    queryClient.clear();
    queryClient.removeQueries();
    router.replace("/signin");
  };

  const support = async () => {
    const isAvailable = await MailComposer.isAvailableAsync();
    if (!isAvailable) {
      Alert.alert(
        "No email client available. Could you please email us at support@decorerbator.com ?",
      );
      return;
    }

    MailComposer.composeAsync({
      recipients: ["support@decorerbator.com"],
      subject: "Support request",
      body: "",
    });
  };

  const profileSettings = () => router.push("/profileSettings");

  const handleCancelSubscription = () => {
    Alert.alert(
      "Cancel Subscription",
      "Are you sure you want to cancel your subscription? You will lose access to premium features at the end of your current billing period.",
      [
        { text: "Keep Subscription", style: "cancel" },
        {
          text: "Cancel Subscription",
          style: "destructive",
          onPress: () => cancelMutation.mutate(),
        },
      ],
    );
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const isPremium = subscription?.plan !== "free";

  return (
    <ImageBackground
      source={require("@/assets/images/dashboard-bg.png")}
      style={styles.backgroundImage}
      resizeMode="cover"
    >
      <SafeAreaView style={styles.container}>
        <ScrollView
          contentContainerStyle={styles.scrollContent}
          showsVerticalScrollIndicator={false}
        >
          {/* Header */}
          <View style={styles.header}>
            <TouchableOpacity
              style={styles.backButton}
              onPress={() => navigation.goBack()}
            >
              <Ionicons name="arrow-back" size={24} color="#2D3436" />
            </TouchableOpacity>
            <Text style={styles.headerTitle}>Settings</Text>
            <View style={{ width: 40 }} />
          </View>

          {/* Current Subscription */}
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Current Subscription</Text>

            {isLoading ? (
              <View style={styles.loadingContainer}>
                <ActivityIndicator size="small" color="#FF7B54" />
              </View>
            ) : (
              <View style={styles.subscriptionCard}>
                <View style={styles.subscriptionHeader}>
                  <View style={styles.planBadge}>
                    <MaterialIcons
                      name={isPremium ? "workspace-premium" : "lock-outline"}
                      size={24}
                      color={isPremium ? "#FFD700" : "#636E72"}
                    />
                  </View>
                  <View style={styles.subscriptionInfo}>
                    <Text style={styles.planName}>
                      {subscription?.plan === "free"
                        ? "Free Plan"
                        : subscription?.plan === "monthly"
                          ? "Monthly Premium"
                          : "Yearly Premium"}
                    </Text>
                    <Text style={styles.planStatus}>
                      {subscription?.status === "active"
                        ? "Active"
                        : subscription?.status === "cancelled"
                          ? "Canceling"
                          : subscription?.status}
                    </Text>
                  </View>
                </View>

                {isPremium && subscription?.currentPeriodEnd && (
                  <View style={styles.subscriptionDetails}>
                    <View style={styles.detailRow}>
                      <Text style={styles.detailLabel}>
                        {subscription.cancelAtPeriodEnd
                          ? "Expires on"
                          : "Renews on"}
                      </Text>
                      <Text style={styles.detailValue}>
                        {formatDate(subscription.currentPeriodEnd)}
                      </Text>
                    </View>

                    {!subscription.cancelAtPeriodEnd && (
                      <TouchableOpacity
                        style={styles.cancelButton}
                        onPress={handleCancelSubscription}
                        disabled={cancelMutation.isPending}
                      >
                        {cancelMutation.isPending ? (
                          <ActivityIndicator size="small" color="#FF6B6B" />
                        ) : (
                          <>
                            <MaterialIcons
                              name="cancel"
                              size={20}
                              color="#FF6B6B"
                            />
                            <Text style={styles.cancelButtonText}>
                              Cancel Subscription
                            </Text>
                          </>
                        )}
                      </TouchableOpacity>
                    )}
                  </View>
                )}

                {!isPremium && (
                  <View style={styles.freeplanLimits}>
                    <Text style={styles.limitText}>
                      <MaterialIcons
                        name="info-outline"
                        size={16}
                        color="#636E72"
                      />{" "}
                      Limited to 1 wordlist with up to 10 words
                    </Text>
                  </View>
                )}
              </View>
            )}
          </View>

          {/* Upgrade Section */}
          {!isPremium && (
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>Upgrade to Premium</Text>
              <Text style={styles.sectionSubtitle}>
                Unlock unlimited learning potential
              </Text>

              {/* Premium Features */}
              <View style={styles.featuresCard}>
                <Text style={styles.featuresTitle}>Premium Features</Text>
                {PRICING_PLANS[0].features.map((feature, index) => (
                  <View key={index} style={styles.featureRow}>
                    <MaterialIcons
                      name="check-circle"
                      size={20}
                      color="#4CAF50"
                    />
                    <Text style={styles.featureText}>{feature}</Text>
                  </View>
                ))}
              </View>

              {/* Pricing Plans */}
              <View style={styles.pricingContainer}>
                {PRICING_PLANS.map((plan) => (
                  <TouchableOpacity
                    key={plan.id}
                    style={[
                      styles.pricingCard,
                      selectedPlan === plan.id && styles.pricingCardSelected,
                      plan.popular && styles.pricingCardPopular,
                    ]}
                    onPress={() => setSelectedPlan(plan.id)}
                    activeOpacity={0.8}
                  >
                    {plan.popular && (
                      <View style={styles.popularBadge}>
                        <Text style={styles.popularText}>BEST VALUE</Text>
                      </View>
                    )}

                    <Text style={styles.planInterval}>{plan.name}</Text>
                    <View style={styles.priceRow}>
                      <Text style={styles.priceSymbol}>$</Text>
                      <Text style={styles.priceAmount}>{plan.price}</Text>
                      <Text style={styles.priceInterval}>/{plan.interval}</Text>
                    </View>

                    {plan.savings && (
                      <View style={styles.savingsBadge}>
                        <Text style={styles.savingsText}>{plan.savings}</Text>
                      </View>
                    )}

                    <View
                      style={[
                        styles.radioButton,
                        selectedPlan === plan.id && styles.radioButtonSelected,
                      ]}
                    >
                      {selectedPlan === plan.id && (
                        <View style={styles.radioButtonInner} />
                      )}
                    </View>
                  </TouchableOpacity>
                ))}
              </View>

              {/* Subscribe Button */}
              <TouchableOpacity
                style={[
                  styles.subscribeButton,
                  !selectedPlan && styles.subscribeButtonDisabled,
                ]}
                onPress={() =>
                  selectedPlan && checkoutMutation.mutateAsync(selectedPlan)
                }
                disabled={!selectedPlan || checkoutMutation.isPending}
                activeOpacity={0.8}
              >
                {checkoutMutation.isPending ? (
                  <ActivityIndicator size="small" color="#FFFFFF" />
                ) : (
                  <>
                    <Text style={styles.subscribeButtonText}>
                      {selectedPlan ? "Continue to Payment" : "Select a Plan"}
                    </Text>
                    <Ionicons name="arrow-forward" size={20} color="#FFFFFF" />
                  </>
                )}
              </TouchableOpacity>
            </View>
          )}

          {/* Other Settings */}
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Other Settings</Text>

            <TouchableOpacity
              style={styles.settingItem}
              onPress={profileSettings}
            >
              <MaterialIcons name="person-outline" size={24} color="#636E72" />
              <Text style={styles.settingText}>Account</Text>
              <Ionicons name="chevron-forward" size={20} color="#636E72" />
            </TouchableOpacity>

            {/* <TouchableOpacity style={styles.settingItem}>
              <MaterialIcons
                name="notifications-none"
                size={24}
                color="#636E72"
              />
              <Text style={styles.settingText}>Notifications</Text>
              <Ionicons name="chevron-forward" size={20} color="#636E72" />
            </TouchableOpacity> */}

            <TouchableOpacity style={styles.settingItem} onPress={support}>
              <MaterialIcons name="help-outline" size={24} color="#636E72" />
              <Text style={styles.settingText}>Help & Support</Text>
              <Ionicons name="chevron-forward" size={20} color="#636E72" />
            </TouchableOpacity>

            <TouchableOpacity style={styles.settingItem} onPress={signOut}>
              <MaterialIcons name="logout" size={24} color="#FF6B6B" />
              <Text style={[styles.settingText, { color: "#FF6B6B" }]}>
                Sign Out
              </Text>
              <Ionicons name="chevron-forward" size={20} color="#FF6B6B" />
            </TouchableOpacity>
          </View>
        </ScrollView>
      </SafeAreaView>
    </ImageBackground>
  );
};

export default SettingsScreen;

const styles = StyleSheet.create({
  backgroundImage: {
    flex: 1,
    width: SCREEN_WIDTH,
  },
  container: {
    flex: 1,
  },
  scrollContent: {
    paddingBottom: 30,
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 20,
    paddingVertical: 16,
  },
  backButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    justifyContent: "center",
    alignItems: "center",
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: "600",
    color: "#2D3436",
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 8,
    paddingHorizontal: 20,
  },
  sectionSubtitle: {
    fontSize: 14,
    color: "#636E72",
    marginBottom: 16,
    paddingHorizontal: 20,
  },
  loadingContainer: {
    padding: 40,
    alignItems: "center",
  },
  subscriptionCard: {
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    marginHorizontal: 20,
    borderRadius: 16,
    padding: 20,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
    elevation: 3,
  },
  subscriptionHeader: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 16,
  },
  planBadge: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: "#FFF5F0",
    justifyContent: "center",
    alignItems: "center",
    marginRight: 16,
  },
  subscriptionInfo: {
    flex: 1,
  },
  planName: {
    fontSize: 18,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 4,
  },
  planStatus: {
    fontSize: 14,
    color: "#4CAF50",
    textTransform: "capitalize",
  },
  subscriptionDetails: {
    borderTopWidth: 1,
    borderTopColor: "#F0F0F0",
    paddingTop: 16,
  },
  detailRow: {
    flexDirection: "row",
    justifyContent: "space-between",
    marginBottom: 12,
  },
  detailLabel: {
    fontSize: 14,
    color: "#636E72",
  },
  detailValue: {
    fontSize: 14,
    fontWeight: "500",
    color: "#2D3436",
  },
  cancelButton: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    paddingVertical: 12,
    borderRadius: 8,
    backgroundColor: "#FFF5F5",
    gap: 8,
  },
  cancelButtonText: {
    fontSize: 14,
    fontWeight: "500",
    color: "#FF6B6B",
  },
  freeplanLimits: {
    flexDirection: "row",
    alignItems: "center",
    paddingTop: 12,
    borderTopWidth: 1,
    borderTopColor: "#F0F0F0",
  },
  limitText: {
    fontSize: 14,
    color: "#636E72",
    flex: 1,
  },
  featuresCard: {
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    marginHorizontal: 20,
    borderRadius: 16,
    padding: 20,
    marginBottom: 20,
  },
  featuresTitle: {
    fontSize: 16,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 16,
  },
  featureRow: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 12,
    gap: 12,
  },
  featureText: {
    fontSize: 14,
    color: "#2D3436",
    flex: 1,
  },
  pricingContainer: {
    flexDirection: "row",
    paddingHorizontal: 20,
    gap: 12,
    marginBottom: 20,
  },
  pricingCard: {
    flex: 1,
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    borderRadius: 16,
    padding: 16,
    borderWidth: 2,
    borderColor: "transparent",
    position: "relative",
  },
  pricingCardSelected: {
    borderColor: "#FF7B54",
  },
  pricingCardPopular: {
    borderColor: "#FFD700",
  },
  popularBadge: {
    position: "absolute",
    top: -12,
    alignSelf: "center",
    backgroundColor: "#FFD700",
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 12,
  },
  popularText: {
    fontSize: 10,
    fontWeight: "700",
    color: "#2D3436",
  },
  planInterval: {
    fontSize: 16,
    fontWeight: "600",
    color: "#2D3436",
    marginBottom: 8,
    textAlign: "center",
  },
  priceRow: {
    flexDirection: "row",
    alignItems: "baseline",
    justifyContent: "center",
    marginBottom: 8,
  },
  priceSymbol: {
    fontSize: 16,
    color: "#2D3436",
    fontWeight: "500",
  },
  priceAmount: {
    fontSize: 28,
    fontWeight: "700",
    color: "#2D3436",
  },
  priceInterval: {
    fontSize: 14,
    color: "#636E72",
  },
  savingsBadge: {
    backgroundColor: "#E8F5E9",
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 8,
    alignSelf: "center",
    marginBottom: 12,
  },
  savingsText: {
    fontSize: 12,
    fontWeight: "600",
    color: "#4CAF50",
  },
  radioButton: {
    width: 20,
    height: 20,
    borderRadius: 10,
    borderWidth: 2,
    borderColor: "#DFE6E9",
    alignSelf: "center",
    justifyContent: "center",
    alignItems: "center",
  },
  radioButtonSelected: {
    borderColor: "#FF7B54",
  },
  radioButtonInner: {
    width: 10,
    height: 10,
    borderRadius: 5,
    backgroundColor: "#FF7B54",
  },
  subscribeButton: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "#FF7B54",
    marginHorizontal: 20,
    paddingVertical: 16,
    borderRadius: 12,
    gap: 8,
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 5,
  },
  subscribeButtonDisabled: {
    backgroundColor: "#DFE6E9",
    shadowOpacity: 0,
  },
  subscribeButtonText: {
    fontSize: 16,
    fontWeight: "600",
    color: "#FFFFFF",
  },
  settingItem: {
    flexDirection: "row",
    alignItems: "center",
    backgroundColor: "rgba(255, 255, 255, 0.9)",
    marginHorizontal: 20,
    padding: 16,
    borderRadius: 12,
    marginBottom: 8,
    gap: 12,
  },
  settingText: {
    flex: 1,
    fontSize: 16,
    color: "#2D3436",
  },
});
