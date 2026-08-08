import * as subscriptionsApi from "@/api/subscriptions";
import * as usersApi from "@/api/users";
import { Platform, Switch } from "react-native";
import { Ionicons, MaterialIcons } from "@expo/vector-icons";
import { useNavigation } from "@react-navigation/native";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import * as MailComposer from "expo-mail-composer";
import { router } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import * as Notifications from "expo-notifications";
import React, { useState, useRef } from "react";
import { useTranslation } from "react-i18next";
import i18n from "@/i18n";
import { useFocusEffect } from "@react-navigation/native";
import { useTheme } from "@/contexts/ThemeContext";
import { createCommonStyles } from "@/styles/common";
import NativeIAPPaywall from "@/components/NativeIAPPaywall";
import { formatDate } from "@/utils/dateUtils";
import { SafeAreaView } from "react-native-safe-area-context";
import {
  ActivityIndicator,
  Alert,
  Modal,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import AsyncStorage from "@react-native-async-storage/async-storage";
import Constants from "expo-constants";
import * as Sentry from "@sentry/react-native";
import { useUserSession } from "@/hooks/useUserSession";
import {
  registerDevicePushToken,
  unregisterDevicePushToken,
} from "@/utils/pushNotifications";

const PREMIUM_FEATURE_KEYS = [
  "unlimitedWordlists",
  "aiEnrichment",
  "allQuizModes",
  "progressTracking",
  "prioritySupport",
] as const;

const SettingsScreen: React.FC = () => {
  const navigation = useNavigation();
  const queryClient = useQueryClient();
  const [showNativeIAPPaywall, setShowNativeIAPPaywall] = useState(false);
  const { t } = useTranslation();
  const previousSubscriptionRef = useRef<string | null>(null);
  const { theme, themeMode, setThemeMode, responsive } = useTheme();
  const commonStyles = createCommonStyles(theme, responsive);
  const { user: profile } = useUserSession();
  const [notificationsEnabled, setNotificationsEnabled] = useState<boolean>(
    profile?.notificationsEnabled ?? false,
  );
  const [notificationsLoading, setNotificationsLoading] = useState(false);

  // Fetch subscription
  const {
    data: subscription,
    isLoading,
    refetch: refetchSubscription,
  } = useQuery({
    queryKey: ["subscription"],
    queryFn: async () => {
      const data = await subscriptionsApi.getSubscriptionStatus();
      return data;
    },

    // ---- Allow optimistic data to persist for 5 minutes ----
    staleTime: 5 * 60 * 1000, // 5 minutes - prevents immediate overwrite of optimistic data
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

  React.useEffect(() => {
    if (profile?.notificationsEnabled !== undefined) {
      setNotificationsEnabled(profile.notificationsEnabled);
    }
  }, [profile?.notificationsEnabled]);

  const handleToggleNotifications = async (nextValue: boolean) => {
    if (!profile) {
      return;
    }

    setNotificationsLoading(true);
    try {
      if (nextValue) {
        if (Platform.OS === "web") {
          Alert.alert(
            t("common.error"),
            t("settings.notificationsUnsupported", {
              defaultValue: "Push notifications are not supported on web.",
            }),
          );
          setNotificationsEnabled(false);
          return;
        }
        const token = await registerDevicePushToken({ prompt: true });
        if (!token) {
          Alert.alert(
            t("common.error"),
            t("settings.notificationsPermissionDenied", {
              defaultValue:
                "Notifications permission was denied. Enable it in Settings to receive reminders.",
            }),
          );
          setNotificationsEnabled(false);
          return;
        }

        await usersApi.update({ notificationsEnabled: true });
        queryClient.invalidateQueries({ queryKey: ["userProfile"] });
        setNotificationsEnabled(true);
      } else {
        await usersApi.update({ notificationsEnabled: false });
        try {
          await unregisterDevicePushToken();
        } catch (error) {
          console.warn("Failed to unregister push token:", error);
        }
        queryClient.invalidateQueries({ queryKey: ["userProfile"] });
        setNotificationsEnabled(false);
      }
    } catch (error) {
      console.error("Failed to update notifications:", error);
      let permissionStatus: string | null = null;
      let canAskAgain: boolean | null = null;
      if (Platform.OS !== "web") {
        try {
          const permission = await Notifications.getPermissionsAsync();
          permissionStatus = permission.status;
          canAskAgain = permission.canAskAgain;
        } catch (permissionError) {
          console.warn(
            "Failed to read notification permission state:",
            permissionError,
          );
        }
      }
      Sentry.addBreadcrumb({
        message: "Notification toggle failure context",
        category: "notifications",
        level: "error",
        data: {
          permissionStatus,
          canAskAgain,
          platform: Platform.OS,
        },
      });
      Sentry.captureException(error, {
        data: {
          action: "toggle_notifications",
          nextValue,
          userId: profile?.id,
          platform: Platform.OS,
        },
      });
      Alert.alert(
        t("common.error"),
        t("settings.notificationsUpdateError", {
          defaultValue:
            "Failed to update notification preferences. Please try again.",
        }),
      );
      setNotificationsEnabled(profile.notificationsEnabled ?? false);
    } finally {
      setNotificationsLoading(false);
    }
  };

  // Open native subscription management (no API call needed)
  const openNativeManagement = async () => {
    try {
      await subscriptionsApi.openNativeSubscriptionManagement();
      // Show info message about what happens next
      Alert.alert(
        t("settings.subscription.managementOpened"),
        t("settings.subscription.managementInstructions"),
      );
    } catch {
      // Use platform-specific error messages for better user guidance
      const errorMessage =
        Platform.OS === "ios"
          ? t("settings.subscription.managementErrorIOS")
          : Platform.OS === "android"
            ? t("settings.subscription.managementErrorAndroid")
            : t("settings.subscription.managementError");

      Alert.alert(t("common.error"), errorMessage);
    }
  };

  const clearUserAsyncStorage = async () => {
    try {
      const allKeys = await AsyncStorage.getAllKeys();

      // Find user-specific keys to remove (but preserve theme preference)
      const userKeys = allKeys.filter(
        (key) =>
          key === "cachedUserProfile" ||
          key === "cachedSubscription" ||
          key === "hasSeenDashboard" ||
          key === "justSignedUp" ||
          key.startsWith("flashcard_position_") ||
          key.startsWith("flashcard_save_position_"),
      );

      if (userKeys.length > 0) {
        await AsyncStorage.multiRemove(userKeys);
      }
    } catch (error) {
      console.error("Error clearing user AsyncStorage:", error);
    }
  };

  const signOut = async () => {
    try {
      // Clear auth token and offline cache
      await usersApi.sigout();

      // Clear user-specific AsyncStorage data
      await clearUserAsyncStorage();

      // Aggressively clear ALL React Query cache
      queryClient.clear();

      // Force garbage collection of removed queries
      queryClient.getQueryCache().clear();
      queryClient.getMutationCache().clear();

      // Navigate to signin
      router.replace("/signin");
    } catch (error) {
      console.error("Logout error:", error);
      Alert.alert(t("common.error"), t("errors.general"));
    }
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
      t("settings.subscription.manageSubscription"),
      t("settings.subscription.manageSubscriptionMessage"),
      [
        { text: t("common.cancel"), style: "cancel" },
        {
          text: t("settings.subscription.openManagement"),
          style: "default",
          onPress: openNativeManagement,
        },
      ],
    );
  };

  // Note: formatDate function has been replaced with formatDate from utils/dateUtils.ts
  // This now handles proper timezone conversion from UTC to user's local timezone

  // More sophisticated logic to determine if user has premium access
  // Handle null/undefined subscription gracefully and include optimistic subscriptions
  const isPremium =
    (subscription &&
      (subscription.plan === "monthly" || subscription.plan === "annual") && // Only known premium plans
      (subscription.isActive ||
        subscription.isCancelledButActive ||
        subscription.isInGracePeriod)) ||
    subscription?.hasOptimisticSubscription; // Treat optimistic subscriptions as premium

  // Helper functions for safe data access
  const getPlanDisplayName = () => {
    if (!subscription) return t("settings.subscription.freePlan");

    switch (subscription.plan) {
      case "free":
        return t("settings.subscription.freePlan");
      case "monthly":
        return t("settings.subscription.monthlyPremium");
      case "annual":
        return t("settings.subscription.yearlyPremium");
      default:
        // Handle unknown plan types gracefully
        return t("settings.subscription.freePlan");
    }
  };

  const getStatusDisplayName = () => {
    if (!subscription) return t("settings.subscription.statusActive");

    if (subscription.isInGracePeriod) {
      return t("settings.subscription.statusGracePeriod");
    }
    if (subscription.isCancelledButActive) {
      return t("settings.subscription.statusCancelledActive");
    }
    if (subscription.status === "active") {
      return t("settings.subscription.statusActive");
    }
    if (subscription.status === "cancelled") {
      return t("settings.subscription.statusExpired");
    }
    // Return raw status for unknown values, with fallback
    return subscription.status || t("settings.subscription.statusActive");
  };

  // Debug subscription data changes
  React.useEffect(() => {}, [subscription, isPremium]);

  const styles = createStyles(theme, responsive);

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
          <Text testID="settings-title" style={styles.headerTitle}>
            {t("settings.title")}
          </Text>
          <View style={{ width: 40 }} />
        </View>

        {/* Current Subscription */}
        <View style={styles.section}>
          <Text testID="current-plan-section" style={styles.sectionTitle}>
            {t("settings.subscription.currentPlan")}
          </Text>

          {isLoading ? (
            <View style={styles.loadingContainer}>
              <ActivityIndicator size="small" color={theme.colors.primary} />
            </View>
          ) : subscription ? (
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
                  <Text
                    testID={
                      subscription?.plan === "monthly"
                        ? "monthly-premium-name"
                        : subscription?.plan === "annual"
                          ? "yearly-premium-name"
                          : "free-plan-name"
                    }
                    style={styles.planName}
                  >
                    {getPlanDisplayName()}
                  </Text>
                  <Text testID="plan-status" style={styles.planStatus}>
                    {getStatusDisplayName()}
                  </Text>
                </View>
              </View>

              {subscription?.currentPeriodEnd &&
                (subscription?.isActive ||
                  subscription?.isCancelledButActive ||
                  subscription?.isInGracePeriod) && (
                  <View style={styles.subscriptionDetails}>
                    <View style={styles.detailRow}>
                      <Text
                        testID={
                          subscription.isInGracePeriod
                            ? "grace-period-ends-on"
                            : subscription.cancelAtPeriodEnd ||
                                subscription.isCancelledButActive
                              ? "expires-on-label"
                              : "renews-on-label"
                        }
                        style={styles.detailLabel}
                      >
                        {subscription.isInGracePeriod
                          ? t("settings.subscription.gracePeriodEndsOn")
                          : subscription.cancelAtPeriodEnd ||
                              subscription.isCancelledButActive
                            ? t("settings.subscription.expiresOn")
                            : t("settings.subscription.renewsOn")}
                      </Text>
                      <Text
                        testID="subscription-date"
                        style={styles.detailValue}
                      >
                        {formatDate(subscription.currentPeriodEnd)}
                      </Text>
                    </View>

                    {/* Show remaining days for cancelled subscriptions */}
                    {subscription &&
                      subscription.isCancelledButActive &&
                      typeof subscription.daysRemaining === "number" && (
                        <View style={styles.detailRow}>
                          <Text
                            testID="days-remaining-label"
                            style={styles.detailLabel}
                          >
                            {t("settings.subscription.daysRemaining")}
                          </Text>
                          <Text
                            testID="days-remaining-count"
                            style={[
                              styles.detailValue,
                              { color: theme.colors.semantic.warning },
                            ]}
                          >
                            {t("settings.subscription.daysCount", {
                              count: subscription.daysRemaining,
                            })}
                          </Text>
                        </View>
                      )}

                    {/* Grace period warning */}
                    {subscription && subscription.isInGracePeriod && (
                      <View
                        style={[
                          styles.detailRow,
                          {
                            backgroundColor:
                              theme.colors.semantic.warning + "20",
                            padding: 12,
                            borderRadius: 8,
                            marginTop: 8,
                          },
                        ]}
                      >
                        <Text
                          testID="grace-period-warning"
                          style={[
                            styles.detailLabel,
                            { color: theme.colors.semantic.warning },
                          ]}
                        >
                          ⚠️ {t("settings.subscription.gracePeriodWarning")}
                        </Text>
                      </View>
                    )}

                    {subscription &&
                      !subscription.cancelAtPeriodEnd &&
                      !subscription.isCancelledButActive &&
                      subscription.isActive && (
                        <TouchableOpacity
                          testID="manage-subscription-button"
                          style={styles.cancelButton}
                          onPress={handleCancelSubscription}
                        >
                          <MaterialIcons
                            name="settings"
                            size={20}
                            color={theme.colors.primary}
                          />
                          <Text style={styles.manageButtonText}>
                            {t("settings.subscription.manageSubscription")}
                          </Text>
                        </TouchableOpacity>
                      )}
                  </View>
                )}

              {!isPremium && (
                <View style={styles.freeplanLimits}>
                  <Text testID="free-plan-limit" style={styles.limitText}>
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
          ) : (
            // Fallback when subscription data is null/undefined
            <View style={styles.subscriptionCard}>
              <View style={styles.subscriptionHeader}>
                <View style={styles.planBadge}>
                  <MaterialIcons
                    name="lock-outline"
                    size={24}
                    color={theme.colors.text.secondary}
                  />
                </View>
                <View style={styles.subscriptionInfo}>
                  <Text testID="free-plan-name" style={styles.planName}>
                    {t("settings.subscription.freePlan")}
                  </Text>
                  <Text testID="plan-status" style={styles.planStatus}>
                    {t("settings.subscription.statusActive")}
                  </Text>
                </View>
              </View>

              <View style={styles.freeplanLimits}>
                <Text testID="free-plan-limit" style={styles.limitText}>
                  <MaterialIcons
                    name="info-outline"
                    size={16}
                    color={theme.colors.text.secondary}
                  />{" "}
                  {t("settings.subscription.freePlanLimit")}
                </Text>
              </View>
            </View>
          )}
        </View>

        {/* Upgrade Section */}
        {!isPremium && !isLoading && (
          <View testID="upgrade-section" style={styles.section}>
            <Text testID="upgrade-title" style={styles.sectionTitle}>
              {t("upgrade.title")}
            </Text>
            <Text testID="upgrade-subtitle" style={styles.sectionSubtitle}>
              {t("upgrade.subtitle")}
            </Text>

            {/* Premium Features */}
            <View style={styles.featuresCard}>
              <Text
                testID="premium-features-title"
                style={styles.featuresTitle}
              >
                {t("settings.subscription.premiumFeatures")}
              </Text>
              {PREMIUM_FEATURE_KEYS.map((featureKey) => (
                <View key={featureKey} style={styles.featureRow}>
                  <MaterialIcons
                    name="check-circle"
                    size={20}
                    color={theme.colors.success}
                  />
                  <Text style={styles.featureText}>
                    {t(`settings.subscription.features.${featureKey}`)}
                  </Text>
                </View>
              ))}
            </View>

            <TouchableOpacity
              testID="open-native-iap-paywall"
              style={styles.subscribeButton}
              onPress={() => setShowNativeIAPPaywall(true)}
              activeOpacity={0.8}
            >
              <Text style={styles.subscribeButtonText}>
                {t("settings.subscription.viewPlans")}
              </Text>
              <Ionicons name="arrow-forward" size={20} color="#FFFFFF" />
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

          {/* Notifications Toggle */}
          <View style={styles.settingItem}>
            <MaterialIcons
              name="notifications-none"
              size={24}
              color={theme.colors.text.secondary}
            />
            <Text style={styles.settingText}>
              {t("settings.preferences.notifications", {
                defaultValue: "Notifications",
              })}
            </Text>
            <Switch
              value={notificationsEnabled}
              onValueChange={handleToggleNotifications}
              disabled={notificationsLoading}
              thumbColor={
                Platform.OS === "android"
                  ? theme.colors.background.surface
                  : undefined
              }
              trackColor={{
                false: theme.colors.border.light,
                true: theme.colors.primary,
              }}
            />
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

        {/* Version Info */}
        <View style={styles.versionSection}>
          <Text style={styles.versionText}>
            Runtime Version: {Constants.expoConfig?.version || "Development"}
          </Text>
        </View>
      </ScrollView>

      <Modal
        visible={showNativeIAPPaywall}
        animationType="slide"
        presentationStyle="pageSheet"
        onRequestClose={() => setShowNativeIAPPaywall(false)}
      >
        {showNativeIAPPaywall ? (
          <NativeIAPPaywall
            visible
            onClose={() => setShowNativeIAPPaywall(false)}
            onEntitled={async () => {
              await Promise.all([
                refetchSubscription(),
                queryClient.invalidateQueries({ queryKey: ["userProfile"] }),
              ]);
              setShowNativeIAPPaywall(false);
              Alert.alert(
                t("common.success"),
                t("settings.subscription.activatedSuccess"),
              );
            }}
          />
        ) : null}
      </Modal>
    </SafeAreaView>
  );
};

export default SettingsScreen;

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) =>
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
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.vertical,
    },
    backButton: {
      width: responsive.spacing.minTouchTarget,
      height: responsive.spacing.minTouchTarget,
      borderRadius: responsive.spacing.minTouchTarget / 2,
      backgroundColor: theme.colors.background.elevated,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    headerTitle: {
      fontSize: responsive.fontSizes.title,
      fontWeight: "600",
      color: theme.colors.text.primary,
    },
    section: {
      marginBottom: responsive.spacing.formPadding,
    },
    sectionTitle: {
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing,
      paddingHorizontal: responsive.spacing.horizontal,
    },
    sectionSubtitle: {
      fontSize: responsive.fontSizes.label,
      color: theme.colors.text.secondary,
      marginBottom: responsive.spacing.horizontal,
      paddingHorizontal: responsive.spacing.horizontal,
    },
    loadingContainer: {
      padding: 40,
      alignItems: "center",
    },
    subscriptionCard: {
      backgroundColor: theme.colors.background.surface,
      marginHorizontal: responsive.spacing.horizontal,
      borderRadius: theme.borderRadius.lg,
      padding: responsive.spacing.formPadding,
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
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.spacing.elementSpacing / 2,
    },
    planStatus: {
      fontSize: responsive.fontSizes.label,
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
      backgroundColor: theme.colors.background.surface,
      borderWidth: 1,
      borderColor: theme.colors.ui.border,
      gap: 8,
    },
    manageButtonText: {
      fontSize: 14,
      fontWeight: "600",
      color: theme.colors.primary,
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
    subscribeButton: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "center",
      backgroundColor: theme.colors.primary,
      marginHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.vertical,
      paddingHorizontal: responsive.spacing.horizontal,
      borderRadius: theme.borderRadius.md,
      gap: responsive.spacing.elementSpacing,
      minHeight: responsive.spacing.minTouchTarget,
      ...theme.shadows.md,
      shadowColor: theme.colors.primary,
    },
    subscribeButtonText: {
      fontSize: responsive.fontSizes.headline,
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    settingItem: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor: theme.colors.background.surface,
      marginHorizontal: responsive.spacing.horizontal,
      padding: responsive.spacing.formPadding,
      borderRadius: theme.borderRadius.md,
      marginBottom: responsive.spacing.elementSpacing,
      gap: responsive.spacing.elementSpacing,
      minHeight: responsive.spacing.minTouchTarget,
      ...theme.shadows.sm,
    },
    settingText: {
      flex: 1,
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.primary,
    },
    themeToggleContainer: {
      flexDirection: "row",
      gap: 8,
    },
    themeOption: {
      width: responsive.spacing.minTouchTarget * 0.75,
      height: responsive.spacing.minTouchTarget * 0.75,
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
    versionSection: {
      marginTop: responsive.spacing.formPadding,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingBottom: responsive.spacing.elementSpacing,
      alignItems: "center",
    },
    versionText: {
      fontSize: responsive.fontSizes.label,
      color: theme.colors.text.secondary,
      opacity: 0.7,
    },
  });
