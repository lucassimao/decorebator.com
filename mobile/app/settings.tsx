import * as subscriptionsApi from "@/api/subscriptions";
import * as usersApi from "@/api/users";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as AuthSession from "expo-auth-session";
import * as MailComposer from "expo-mail-composer";
import { router } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import React, { useState, useRef } from "react";
import { useTranslation } from "react-i18next";
import i18n from "@/i18n";
import { useFocusEffect } from "@react-navigation/native";
import { useTheme } from "@/contexts/ThemeContext";
import { createCommonStyles } from "@/styles/common";
import { usePaymentProvider } from "@/hooks/useRevenueCat";
import RevenueCatPaywall from "@/components/RevenueCatPaywall";
import {
  ActivityIndicator,
  Alert,
  Modal,
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
// const { width: SCREEN_WIDTH } = Dimensions.get("window"); // Removed unused variable

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
  const [showRevenueCatPaywall, setShowRevenueCatPaywall] = useState(false);
  const { t } = useTranslation();
  const previousSubscriptionRef = useRef<string | null>(null);
  const { theme, themeMode, setThemeMode } = useTheme();
  const commonStyles = createCommonStyles(theme);

  // Get payment provider
  const { data: providerInfo } = usePaymentProvider();

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

  // Detect subscription changes when returning to the screen (e.g., after Stripe checkout)
  useFocusEffect(
    React.useCallback(() => {
      if (subscription) {
        const currentPlan = subscription.plan;

        // Check if subscription upgraded from free to premium
        if (
          previousSubscriptionRef.current === "free" &&
          currentPlan !== "free" &&
          currentPlan !== null
        ) {
          // Show success message for completed checkout
          Alert.alert(
            t("common.success"),
            t("settings.subscription.activatedSuccess"),
          );
        }

        // Update the reference for next comparison
        previousSubscriptionRef.current = currentPlan;
      }
    }, [subscription, t]),
  );

  const checkoutMutation = useMutation({
    mutationFn: async (plan: PlanRecurrence) => {
      const redirectUri = AuthSession.makeRedirectUri({
        scheme: "decorebator",
        path: "settings", // Use settings path instead of stripe
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

      console.log("Stripe checkout result:", result);

      // Handle all possible result types for better Android compatibility
      if (result.type === "success") {
        // Direct success - refresh subscription and show success message
        console.log("Checkout completed successfully");
        await refetchSubscription();
        Alert.alert(
          t("common.success"),
          t("settings.subscription.activatedSuccess"),
        );
      } else if (result.type === "cancel") {
        // User cancelled checkout - no action needed
        console.log("Checkout cancelled by user");
      } else if (result.type === "dismiss") {
        // Browser was dismissed - could be user closing or redirect completion
        // This is common on Android when Stripe redirects back to the app
        console.log("Browser dismissed, checking subscription status");

        // Wait a moment for any background processing, then refresh
        setTimeout(async () => {
          await refetchSubscription();

          // Check if subscription changed from free to premium
          // Note: We'll let the subscription query refetch handle the UI update
          console.log("Subscription status checked after browser dismiss");
        }, 1000);
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

  const styles = createStyles(theme);

  return (
    <SafeAreaView style={[commonStyles.safeArea, styles.container]}>
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
            <Ionicons
              name="arrow-back"
              size={24}
              color={theme.colors.text.primary}
            />
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
              <ActivityIndicator size="small" color={theme.colors.primary} />
            </View>
          ) : (
            <View style={styles.subscriptionCard}>
              <View style={styles.subscriptionHeader}>
                <View style={styles.planBadge}>
                  <MaterialIcons
                    name={isPremium ? "workspace-premium" : "lock-outline"}
                    size={24}
                    color={
                      isPremium
                        ? theme.colors.premium
                        : theme.colors.text.secondary
                    }
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
                        <ActivityIndicator
                          size="small"
                          color={theme.colors.error}
                        />
                      ) : (
                        <>
                          <MaterialIcons
                            name="cancel"
                            size={20}
                            color={theme.colors.error}
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
                      color={theme.colors.text.secondary}
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
            <Text style={styles.sectionSubtitle}>{t("upgrade.subtitle")}</Text>

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
                    color={theme.colors.success}
                  />
                  <Text style={styles.featureText}>{feature}</Text>
                </View>
              ))}
            </View>

            {/* Pricing Plans - Only show for Stripe */}
            {providerInfo?.provider === "stripe" && (
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
            )}

            {/* Subscribe Button */}
            <TouchableOpacity
              style={[
                styles.subscribeButton,
                !selectedPlan &&
                  providerInfo?.provider === "stripe" &&
                  styles.subscribeButtonDisabled,
              ]}
              onPress={() => {
                if (providerInfo?.provider === "revenuecat") {
                  // Show RevenueCat paywall
                  setShowRevenueCatPaywall(true);
                } else {
                  // Use Stripe checkout
                  selectedPlan && checkoutMutation.mutateAsync(selectedPlan);
                }
              }}
              disabled={
                providerInfo?.provider === "stripe" &&
                (!selectedPlan || checkoutMutation.isPending)
              }
              activeOpacity={0.8}
            >
              {checkoutMutation.isPending ? (
                <ActivityIndicator
                  size="small"
                  color={theme.colors.text.inverse}
                />
              ) : (
                <>
                  <Text style={styles.subscribeButtonText}>
                    {providerInfo?.provider === "revenuecat"
                      ? t("settings.subscription.viewPlans")
                      : selectedPlan
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
          <Text style={styles.sectionTitle}>{t("settings.otherSettings")}</Text>

          <TouchableOpacity
            style={styles.settingItem}
            onPress={profileSettings}
          >
            <MaterialIcons
              name="person-outline"
              size={24}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.settingText}>
              {t("settings.account.title")}
            </Text>
            <Ionicons
              name="chevron-forward"
              size={20}
              color={theme.colors.text.secondary}
            />
          </TouchableOpacity>

          {/* Theme Toggle */}
          <View style={styles.settingItem}>
            <MaterialIcons
              name={theme.mode === "dark" ? "dark-mode" : "light-mode"}
              size={24}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.settingText}>
              {t("settings.theme", { defaultValue: "Theme" })}
            </Text>
            <View style={styles.themeToggleContainer}>
              <TouchableOpacity
                style={[
                  styles.themeOption,
                  themeMode === "light" && styles.themeOptionActive,
                ]}
                onPress={() => setThemeMode("light")}
              >
                <MaterialIcons
                  name="light-mode"
                  size={16}
                  color={
                    themeMode === "light"
                      ? theme.colors.primary
                      : theme.colors.text.secondary
                  }
                />
              </TouchableOpacity>
              <TouchableOpacity
                style={[
                  styles.themeOption,
                  themeMode === "system" && styles.themeOptionActive,
                ]}
                onPress={() => setThemeMode("system")}
              >
                <MaterialIcons
                  name="phone-iphone"
                  size={16}
                  color={
                    themeMode === "system"
                      ? theme.colors.primary
                      : theme.colors.text.secondary
                  }
                />
              </TouchableOpacity>
              <TouchableOpacity
                style={[
                  styles.themeOption,
                  themeMode === "dark" && styles.themeOptionActive,
                ]}
                onPress={() => setThemeMode("dark")}
              >
                <MaterialIcons
                  name="dark-mode"
                  size={16}
                  color={
                    themeMode === "dark"
                      ? theme.colors.primary
                      : theme.colors.text.secondary
                  }
                />
              </TouchableOpacity>
            </View>
          </View>

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
            <MaterialIcons
              name="help-outline"
              size={24}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.settingText}>
              {t("settings.helpAndSupport")}
            </Text>
            <Ionicons
              name="chevron-forward"
              size={20}
              color={theme.colors.text.secondary}
            />
          </TouchableOpacity>
        </View>

        {/* Legal & Privacy Section */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>
            {t("settings.legalAndPrivacy")}
          </Text>

          <TouchableOpacity
            style={styles.settingItem}
            onPress={async () => {
              await WebBrowser.openBrowserAsync(
                `https://decorebator.com/${i18n.language}/privacy`,
              );
            }}
          >
            <MaterialIcons
              name="privacy-tip"
              size={24}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.settingText}>
              {t("settings.privacyPolicy")}
            </Text>
            <Ionicons
              name="chevron-forward"
              size={20}
              color={theme.colors.text.secondary}
            />
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.settingItem}
            onPress={async () => {
              await WebBrowser.openBrowserAsync(
                `https://decorebator.com/${i18n.language}/terms`,
              );
            }}
          >
            <MaterialIcons
              name="gavel"
              size={24}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.settingText}>
              {t("settings.termsOfService")}
            </Text>
            <Ionicons
              name="chevron-forward"
              size={20}
              color={theme.colors.text.secondary}
            />
          </TouchableOpacity>
        </View>

        {/* Sign Out */}
        <View style={styles.section}>
          <TouchableOpacity style={styles.settingItem} onPress={signOut}>
            <MaterialIcons name="logout" size={24} color={theme.colors.error} />
            <Text style={[styles.settingText, { color: theme.colors.error }]}>
              {t("settings.account.logOut")}
            </Text>
            <Ionicons
              name="chevron-forward"
              size={20}
              color={theme.colors.error}
            />
          </TouchableOpacity>
        </View>
      </ScrollView>

      {/* RevenueCat Paywall Modal */}
      <Modal
        visible={showRevenueCatPaywall}
        animationType="slide"
        presentationStyle="pageSheet"
        onRequestClose={() => setShowRevenueCatPaywall(false)}
      >
        <RevenueCatPaywall
          onClose={() => setShowRevenueCatPaywall(false)}
          onSuccess={() => {
            setShowRevenueCatPaywall(false);
            refetchSubscription();
            Alert.alert(
              t("common.success"),
              t("settings.subscription.activatedSuccess"),
            );
          }}
        />
      </Modal>
    </SafeAreaView>
  );
};

export default SettingsScreen;

const createStyles = (theme: ReturnType<typeof useTheme>["theme"]) =>
  StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
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
      backgroundColor: theme.colors.background.elevated,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    headerTitle: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    section: {
      marginBottom: 24,
    },
    sectionTitle: {
      fontSize: 18,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 8,
      paddingHorizontal: 20,
    },
    sectionSubtitle: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      marginBottom: 16,
      paddingHorizontal: 20,
    },
    loadingContainer: {
      padding: 40,
      alignItems: "center",
    },
    subscriptionCard: {
      backgroundColor: theme.colors.background.surface,
      marginHorizontal: 20,
      borderRadius: theme.borderRadius.lg,
      padding: theme.spacing.lg,
      ...theme.shadows.md,
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
      backgroundColor:
        theme.mode === "light" ? "#FFF5F0" : theme.colors.background.elevated,
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
      color: theme.colors.text.primary,
      marginBottom: 4,
    },
    planStatus: {
      fontSize: 14,
      color: theme.colors.success,
      textTransform: "capitalize",
    },
    subscriptionDetails: {
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
      paddingTop: theme.spacing.md,
    },
    detailRow: {
      flexDirection: "row",
      justifyContent: "space-between",
      marginBottom: 12,
    },
    detailLabel: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
    detailValue: {
      fontSize: 14,
      fontWeight: "500",
      color: theme.colors.text.primary,
    },
    cancelButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      paddingVertical: 12,
      borderRadius: theme.borderRadius.sm,
      backgroundColor: theme.colors.state.incorrectBackground,
      gap: 8,
    },
    cancelButtonText: {
      fontSize: 14,
      fontWeight: "500",
      color: theme.colors.error,
    },
    freeplanLimits: {
      flexDirection: "row",
      alignItems: "center",
      paddingTop: 12,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
    },
    limitText: {
      fontSize: 14,
      color: theme.colors.text.secondary,
      flex: 1,
    },
    featuresCard: {
      backgroundColor: theme.colors.background.surface,
      marginHorizontal: 20,
      borderRadius: theme.borderRadius.lg,
      padding: theme.spacing.lg,
      marginBottom: 20,
      ...theme.shadows.sm,
    },
    featuresTitle: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.primary,
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
      color: theme.colors.text.primary,
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
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      padding: theme.spacing.md,
      borderWidth: 2,
      borderColor: "transparent",
      position: "relative",
      ...theme.shadows.sm,
    },
    pricingCardSelected: {
      borderColor: theme.colors.primary,
    },
    pricingCardPopular: {
      borderColor: theme.colors.premium,
    },
    popularBadge: {
      position: "absolute",
      top: -12,
      alignSelf: "center",
      backgroundColor: theme.colors.premium,
      paddingHorizontal: 12,
      paddingVertical: 4,
      borderRadius: 12,
    },
    popularText: {
      fontSize: 10,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    planInterval: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.primary,
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
      color: theme.colors.text.primary,
      fontWeight: "500",
    },
    priceAmount: {
      fontSize: 28,
      fontWeight: "700",
      color: theme.colors.text.primary,
    },
    priceInterval: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
    savingsBadge: {
      backgroundColor: theme.colors.state.correctBackground,
      paddingHorizontal: 8,
      paddingVertical: 4,
      borderRadius: theme.borderRadius.sm,
      alignSelf: "center",
      marginBottom: 12,
    },
    savingsText: {
      fontSize: 12,
      fontWeight: "600",
      color: theme.colors.success,
    },
    radioButton: {
      width: 20,
      height: 20,
      borderRadius: 10,
      borderWidth: 2,
      borderColor: theme.colors.ui.disabled,
      alignSelf: "center",
      justifyContent: "center",
      alignItems: "center",
    },
    radioButtonSelected: {
      borderColor: theme.colors.primary,
    },
    radioButtonInner: {
      width: 10,
      height: 10,
      borderRadius: 5,
      backgroundColor: theme.colors.primary,
    },
    subscribeButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      backgroundColor: theme.colors.primary,
      marginHorizontal: 20,
      paddingVertical: theme.spacing.md,
      borderRadius: theme.borderRadius.md,
      gap: 8,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
    },
    subscribeButtonDisabled: {
      backgroundColor: theme.colors.ui.disabled,
      shadowOpacity: 0,
    },
    subscribeButtonText: {
      fontSize: 16,
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    settingItem: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor: theme.colors.background.surface,
      marginHorizontal: 20,
      padding: theme.spacing.md,
      borderRadius: theme.borderRadius.md,
      marginBottom: 8,
      gap: 12,
      ...theme.shadows.sm,
    },
    settingText: {
      flex: 1,
      fontSize: 16,
      color: theme.colors.text.primary,
    },
    themeToggleContainer: {
      flexDirection: "row",
      gap: 8,
    },
    themeOption: {
      width: 32,
      height: 32,
      borderRadius: theme.borderRadius.sm,
      backgroundColor: theme.colors.background.elevated,
      justifyContent: "center",
      alignItems: "center",
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
    },
    themeOptionActive: {
      backgroundColor: theme.colors.background.subtle,
      borderColor: theme.colors.primary,
    },
  });
