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
import { useTranslation } from "react-i18next";
import i18n from "@/i18n";
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
const { width: SCREEN_WIDTH } = Dimensions.get("window");

const getPricingPlans = (t: any): PricingPlan[] => [
  {
    id: "monthly",
    name: t("settings.subscription.monthly"),
    price: 6.99,
    currency: "USD",
    interval: "month",
    features: [
      t("settings.subscription.features.unlimitedWordlists"),
      t("settings.subscription.features.aiEnrichment"),
      t("settings.subscription.features.allQuizModes"),
      t("settings.subscription.features.progressTracking"),
      t("settings.subscription.features.prioritySupport"),
    ],
  },
  {
    id: "annual",
    name: t("settings.subscription.yearly"),
    price: 69.99,
    currency: "USD",
    interval: "year",
    popular: true,
    savings: t("settings.subscription.savePercent", { percent: 17 }),
    features: [
      t("settings.subscription.features.unlimitedWordlists"),
      t("settings.subscription.features.aiEnrichment"),
      t("settings.subscription.features.allQuizModes"),
      t("settings.subscription.features.progressTracking"),
      t("settings.subscription.features.prioritySupport"),
    ],
  },
];

const SettingsScreen: React.FC = () => {
  const navigation = useNavigation();
  const queryClient = useQueryClient();
  const [selectedPlan, setSelectedPlan] = useState<PlanRecurrence | null>(null);
  const { t } = useTranslation();

  const PRICING_PLANS = React.useMemo(() => getPricingPlans(t), [t]);

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
        scheme: "decorebator",
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

      console.log(result);

      // Handle all possible result types for better Android compatibility
      if (result.type === "success") {
        // Refresh subscription data
        refetchSubscription();
        Alert.alert(
          t("common.success"),
          t("settings.subscription.activatedSuccess"),
        );
      } else if (result.type === "cancel") {
        // User cancelled checkout - no action needed
        console.log("Checkout cancelled by user");
      } else if (result.type === "dismiss") {
        // Browser was dismissed - could be user closing or redirect completion
        // Silently refresh subscription to check if payment succeeded
        console.log("Browser dismissed, checking subscription status");

        // Refresh subscription without showing success alert
        // If payment succeeded, the subscription query will reflect the change
        refetchSubscription();
      }
    },
    onError: () => {
      Alert.alert(t("common.error"), t("settings.subscription.checkoutError"));
    },
  });

  // Cancel subscription mutation
  const cancelMutation = useMutation({
    mutationFn: subscriptionsApi.cancelSubscription,
    onSuccess: () => {
      refetchSubscription();
      Alert.alert(
        t("common.success"),
        t("settings.subscription.cancelSuccess"),
      );
    },
    onError: () => {
      Alert.alert(t("common.error"), t("settings.subscription.cancelError"));
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
      Alert.alert(t("settings.noEmailClient"));
      return;
    }

    MailComposer.composeAsync({
      recipients: ["support@decorerbator.com"],
      subject: t("settings.supportEmailSubject"),
      body: "",
    });
  };

  const profileSettings = () => router.push("/profileSettings");

  const handleCancelSubscription = () => {
    Alert.alert(
      t("settings.subscription.cancelSubscription"),
      t("settings.subscription.cancelConfirmMessage"),
      [
        { text: t("settings.subscription.keepSubscription"), style: "cancel" },
        {
          text: t("settings.subscription.cancelSubscription"),
          style: "destructive",
          onPress: () => cancelMutation.mutate(),
        },
      ],
    );
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString(
      i18n.language.startsWith("en") ? "en-US" : i18n.language,
      {
        year: "numeric",
        month: "long",
        day: "numeric",
      },
    );
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
            <Text style={styles.headerTitle}>{t("settings.title")}</Text>
            <View style={{ width: 40 }} />
          </View>

          {/* Current Subscription */}
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>
              {t("settings.subscription.currentPlan")}
            </Text>

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
                        ? t("settings.subscription.freePlan")
                        : subscription?.plan === "monthly"
                          ? t("settings.subscription.monthlyPremium")
                          : t("settings.subscription.yearlyPremium")}
                    </Text>
                    <Text style={styles.planStatus}>
                      {subscription?.status === "active"
                        ? t("settings.subscription.statusActive")
                        : subscription?.status === "cancelled"
                          ? t("settings.subscription.statusCanceling")
                          : subscription?.status}
                    </Text>
                  </View>
                </View>

                {isPremium && subscription?.currentPeriodEnd && (
                  <View style={styles.subscriptionDetails}>
                    <View style={styles.detailRow}>
                      <Text style={styles.detailLabel}>
                        {subscription.cancelAtPeriodEnd
                          ? t("settings.subscription.expiresOn")
                          : t("settings.subscription.renewsOn")}
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
                              {t("settings.subscription.cancelSubscription")}
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
                      {t("settings.subscription.freePlanLimit")}
                    </Text>
                  </View>
                )}
              </View>
            )}
          </View>

          {/* Upgrade Section */}
          {!isPremium && (
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>{t("upgrade.title")}</Text>
              <Text style={styles.sectionSubtitle}>
                {t("upgrade.subtitle")}
              </Text>

              {/* Premium Features */}
              <View style={styles.featuresCard}>
                <Text style={styles.featuresTitle}>
                  {t("settings.subscription.premiumFeatures")}
                </Text>
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
                        <Text style={styles.popularText}>
                          {t("settings.subscription.bestValue")}
                        </Text>
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
                      {selectedPlan
                        ? t("settings.subscription.continueToPayment")
                        : t("settings.subscription.selectPlan")}
                    </Text>
                    <Ionicons name="arrow-forward" size={20} color="#FFFFFF" />
                  </>
                )}
              </TouchableOpacity>
            </View>
          )}

          {/* Other Settings */}
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>
              {t("settings.otherSettings")}
            </Text>

            <TouchableOpacity
              style={styles.settingItem}
              onPress={profileSettings}
            >
              <MaterialIcons name="person-outline" size={24} color="#636E72" />
              <Text style={styles.settingText}>
                {t("settings.account.title")}
              </Text>
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
              <Text style={styles.settingText}>
                {t("settings.helpAndSupport")}
              </Text>
              <Ionicons name="chevron-forward" size={20} color="#636E72" />
            </TouchableOpacity>

            <TouchableOpacity style={styles.settingItem} onPress={signOut}>
              <MaterialIcons name="logout" size={24} color="#FF6B6B" />
              <Text style={[styles.settingText, { color: "#FF6B6B" }]}>
                {t("settings.account.logOut")}
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
