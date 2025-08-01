import { useTheme } from "@/contexts/ThemeContext";
import { useRevenueCat } from "@/hooks/useRevenueCat";
import { MaterialIcons } from "@expo/vector-icons";
import { LinearGradient } from "expo-linear-gradient";
import * as WebBrowser from "expo-web-browser";
import React, { useEffect } from "react";
import { useTranslation } from "react-i18next";
import {
  ActivityIndicator,
  Alert,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  Image,
} from "react-native";
import { PACKAGE_TYPE, PurchasesPackage } from "react-native-purchases";

interface RevenueCatPaywallProps {
  onClose: () => void;
  onSuccess: (packageType: PACKAGE_TYPE) => void;
}

export default function RevenueCatPaywall({
  onClose,
  onSuccess,
}: RevenueCatPaywallProps) {
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

  const {
    isInitialized,
    isLoading,
    error,
    purchasePackage,
    isPurchasing,
    restorePurchases,
    isRestoring,
    getCurrentOffering,
  } = useRevenueCat();

  const currentOffering = getCurrentOffering();

  // Link handlers for Terms of Service and Privacy Policy
  const openTerms = () => {
    WebBrowser.openBrowserAsync("https://decorebator.com/en/terms");
  };

  const openPrivacy = () => {
    WebBrowser.openBrowserAsync("https://decorebator.com/en/privacy");
  };

  useEffect(() => {
    if (error) {
      Alert.alert(
        t("common.error"),
        error.message || t("settings.subscription.loadError"),
      );
    }
  }, [error, t]);

  const handlePurchase = async (pkg: PurchasesPackage) => {
    try {
      // Purchase the package and get the updated CustomerInfo
      const customerInfo = await purchasePackage(pkg);

      // If customerInfo is null, user cancelled - just return
      if (!customerInfo) {
        return;
      }

      // Check if user now has active premium entitlements from the fresh CustomerInfo
      const premiumEntitlement =
        customerInfo?.entitlements?.active?.["Premium"];

      if (premiumEntitlement) {
        // Pass package type to settings screen for optimistic update
        onSuccess(pkg.packageType);
      } else {
        // Purchase succeeded but no premium entitlements - show error
        Alert.alert(
          t("common.error"),
          t("settings.subscription.purchaseSuccessButNoEntitlements", {
            defaultValue:
              "Purchase completed but premium access not activated. Please contact support if the issue persists.",
          }),
        );
      }
    } catch (error: any) {
      // Handle RevenueCat specific errors
      console.log("RevenueCat purchase error:", JSON.stringify(error));

      // Check if user cancelled the purchase
      if (error.userCancelled || error.code === "1") {
        // User cancelled the purchase - just log it, don't show error
        console.log("User cancelled purchase");
        return;
      }

      // Handle specific error codes (RevenueCat uses string codes)
      switch (error.code) {
        case "2": // Store problem error
          Alert.alert(
            t("common.error"),
            t("settings.subscription.storeProblem", {
              defaultValue:
                "There was a problem with the store. Please try again later.",
            }),
          );
          break;
        case "3": // Purchase not allowed error
          Alert.alert(
            t("common.error"),
            t("settings.subscription.paymentNotAllowed", {
              defaultValue:
                "Payment failed. Please try a different payment method.",
            }),
          );
          break;
        case "4": // Payment pending error
          Alert.alert(
            t("common.info"),
            t("settings.subscription.paymentPending", {
              defaultValue:
                "Payment is pending. You'll receive access once payment is confirmed.",
            }),
          );
          break;
        case "5": // Invalid credentials error
          Alert.alert(
            t("common.error"),
            t("settings.subscription.paymentMethodInvalid", {
              defaultValue:
                "Payment method is invalid. Please check your payment information and try again.",
            }),
          );
          break;
        case "6": // Receipt already in use error
          Alert.alert(
            t("common.error"),
            t("settings.subscription.receiptValidationFailed", {
              defaultValue:
                "Receipt validation failed. Please try again or contact support.",
            }),
          );
          break;
        case "22": // Configuration error
        case "23": // Configuration error (iOS specific)
          Alert.alert(
            t("common.error"),
            t("settings.subscription.paymentSystemError", {
              defaultValue:
                "Payment system configuration error. Please try again later.",
            }),
          );
          break;
        default:
          Alert.alert(
            t("common.error"),
            error.message || t("settings.subscription.purchaseError"),
          );
          break;
      }
    }
  };

  const handleRestore = async () => {
    try {
      // Restore purchases and get the updated CustomerInfo
      const customerInfo = await restorePurchases();

      // Check if user now has active premium entitlements from the fresh CustomerInfo
      const premiumEntitlement =
        customerInfo?.entitlements?.active?.["Premium"];

      if (premiumEntitlement) {
        // For restore, we don't know the exact package type, use UNKNOWN
        // Backend webhook will sync the correct plan type
        onSuccess(PACKAGE_TYPE.UNKNOWN);
      } else {
        // Restore succeeded but no premium entitlements found
        Alert.alert(
          t("common.error"),
          t("settings.subscription.restoreNoEntitlements", {
            defaultValue: "No active premium subscriptions found to restore.",
          }),
        );
      }
    } catch (error: any) {
      // Handle RevenueCat specific errors for restore
      console.log("RevenueCat restore error:", JSON.stringify(error));

      switch (error.code) {
        case "2": // Store problem error
          Alert.alert(
            t("common.error"),
            t("settings.subscription.storeProblem", {
              defaultValue:
                "There was a problem with the store. Please try again later.",
            }),
          );
          break;
        case "7": // Network error
          Alert.alert(
            t("common.error"),
            t("settings.subscription.networkError", {
              defaultValue:
                "Network error. Please check your connection and try again.",
            }),
          );
          break;
        default:
          Alert.alert(
            t("common.error"),
            error.message || t("settings.subscription.restoreError"),
          );
          break;
      }
    }
  };

  if (!isInitialized || isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={theme.colors.primary} />
        <Text style={styles.loadingText}>{t("common.loading")}</Text>
      </View>
    );
  }

  if (!currentOffering || !currentOffering.availablePackages.length) {
    return (
      <View style={styles.container}>
        <View style={styles.header}>
          <TouchableOpacity onPress={onClose} style={styles.closeButton}>
            <MaterialIcons
              name="close"
              size={24}
              color={theme.colors.text.primary}
            />
          </TouchableOpacity>
        </View>
        <View style={styles.errorContainer}>
          <MaterialIcons
            name="error-outline"
            size={48}
            color={theme.colors.error}
          />
          <Text style={styles.errorText}>
            {t("settings.subscription.noOfferings")}
          </Text>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <LinearGradient
        colors={
          theme.mode === "light"
            ? ["#1A237E", "#3949AB", "#64B5F6"]
            : ["#0D47A1", "#1565C0", "#42A5F5"]
        }
        style={styles.headerGradient}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
        locations={[0, 0.5, 1]}
      >
        <View style={styles.header}>
          <TouchableOpacity onPress={onClose} style={styles.closeButton}>
            <MaterialIcons
              name="close"
              size={24}
              color={theme.colors.text.inverse}
            />
          </TouchableOpacity>
          <View style={styles.headerContent}>
            <View style={styles.premiumIconWrapper}>
              <Image
                source={require("@/assets/images/icon.png")}
                style={styles.decorebatorIcon}
                resizeMode="contain"
              />
            </View>
            <Text style={styles.title}>
              {t("settings.subscription.upgradeToPremium")}
            </Text>
            <Text style={styles.subtitle}>{t("upgrade.subtitle")}</Text>
          </View>
        </View>
      </LinearGradient>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* Packages */}
        <View style={styles.packagesSection}>
          <View style={styles.packagesGrid}>
            {currentOffering.availablePackages.map((pkg) => {
              const isMonthly = pkg.packageType === "MONTHLY";
              const isAnnual = pkg.packageType === "ANNUAL";
              const isPopular = isAnnual; // Mark annual as popular

              return (
                <TouchableOpacity
                  key={pkg.identifier}
                  style={[
                    styles.packageCard,
                    isPopular && styles.popularPackage,
                  ]}
                  onPress={() => handlePurchase(pkg)}
                  disabled={isPurchasing}
                  activeOpacity={0.8}
                >
                  {isPopular && (
                    <LinearGradient
                      colors={[
                        theme.colors.primary,
                        theme.mode === "light" ? "#FF9A76" : "#FF8A65",
                      ]}
                      style={styles.popularGradient}
                      start={{ x: 0, y: 0 }}
                      end={{ x: 1, y: 0 }}
                    >
                      <Text style={styles.popularBadgeText}>
                        {t("settings.subscription.bestValue")}
                      </Text>
                    </LinearGradient>
                  )}

                  <View style={styles.packageContent}>
                    <Text style={styles.packageTitle}>
                      {isMonthly
                        ? t("settings.subscription.monthly")
                        : t("settings.subscription.yearly")}
                    </Text>

                    <View style={styles.priceContainer}>
                      <Text style={styles.packagePrice}>
                        {pkg.product.priceString}
                      </Text>
                      <Text style={styles.packagePeriod}>
                        {isMonthly
                          ? `/${t("settings.subscription.perMonth").replace("per ", "")}`
                          : `/${t("settings.subscription.perYear").replace("per ", "")}`}
                      </Text>
                    </View>

                    {isAnnual && (
                      <View style={styles.savingsContainer}>
                        <MaterialIcons
                          name="local-offer"
                          size={16}
                          color={theme.colors.success}
                        />
                        <Text style={styles.savingsText}>
                          {t("settings.subscription.savePercent", {
                            percent: 17,
                          })}
                        </Text>
                      </View>
                    )}

                    <Text style={styles.packageDescription}>
                      {isMonthly
                        ? t("settings.subscription.billedMonthly")
                        : t("settings.subscription.billedYearly")}
                    </Text>
                  </View>
                </TouchableOpacity>
              );
            })}
          </View>
        </View>

        {/* Divider */}
        <View style={styles.sectionDivider} />

        {/* Terms */}
        <View style={styles.termsContainer}>
          <View style={styles.termsRow}>
            <Text style={styles.termsText}>
              {t("settings.subscription.termsPrefix")}{" "}
            </Text>
            <TouchableOpacity
              onPress={openTerms}
              hitSlop={{ top: 8, bottom: 8, left: 4, right: 4 }}
            >
              <Text style={styles.linkText}>
                {t("settings.termsOfService")}
              </Text>
            </TouchableOpacity>
            <Text style={styles.termsText}>
              {" "}
              {t("settings.subscription.termsMiddle")}{" "}
            </Text>
            <TouchableOpacity
              onPress={openPrivacy}
              hitSlop={{ top: 8, bottom: 8, left: 4, right: 4 }}
            >
              <Text style={styles.linkText}>{t("settings.privacyPolicy")}</Text>
            </TouchableOpacity>
            <Text style={styles.termsText}>. </Text>
          </View>
          <Text style={styles.termsText}>
            {t("settings.subscription.termsSuffix")}
          </Text>
        </View>

        {/* Premium Features List */}
        <View style={styles.featuresContainer}>
          <View style={styles.featuresList}>
            {[
              t("settings.subscription.features.unlimitedWordlists"),
              t("settings.subscription.features.aiEnrichment"),
              t("settings.subscription.features.allQuizModes"),
              t("settings.subscription.features.progressTracking"),
              t("settings.subscription.features.prioritySupport"),
            ].map((feature, index) => (
              <View key={index} style={styles.featureItem}>
                <LinearGradient
                  colors={[
                    theme.colors.primary,
                    theme.mode === "light" ? "#FF9A76" : "#FF8A65",
                  ]}
                  style={styles.checkmarkContainer}
                  start={{ x: 0, y: 0 }}
                  end={{ x: 1, y: 1 }}
                >
                  <MaterialIcons
                    name="check"
                    size={responsive.getValueForSize(10, 12, 14, 16)}
                    color={theme.colors.text.inverse}
                  />
                </LinearGradient>
                <Text style={styles.featureText}>{feature}</Text>
              </View>
            ))}
          </View>
        </View>

        {/* Restore button */}
        <TouchableOpacity
          style={styles.restoreButton}
          onPress={handleRestore}
          disabled={isRestoring}
        >
          {isRestoring ? (
            <ActivityIndicator size="small" color={theme.colors.primary} />
          ) : (
            <Text style={styles.restoreButtonText}>
              {t("settings.subscription.restorePurchases")}
            </Text>
          )}
        </TouchableOpacity>
      </ScrollView>

      {isPurchasing && (
        <View style={styles.purchasingOverlay}>
          <ActivityIndicator size="large" color={theme.colors.primary} />
          <Text style={styles.purchasingText}>
            {t("settings.subscription.processing")}
          </Text>
        </View>
      )}
    </View>
  );
}

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) => {
  const styles = StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    loadingContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      backgroundColor: theme.colors.background.default,
    },
    loadingText: {
      marginTop: 16,
      fontSize: 16,
      color: theme.colors.text.secondary,
    },
    headerGradient: {
      paddingTop:
        responsive.spacing.vertical +
        responsive.getValueForSize(10, 12, 16, 20),
      paddingBottom: responsive.spacing.vertical,
    },
    header: {
      alignItems: "center",
    },
    headerContent: {
      alignItems: "center",
      marginTop: responsive.spacing.elementSpacing / 2,
      paddingHorizontal: responsive.spacing.horizontal,
    },
    premiumIconWrapper: {
      width: responsive.getValueForSize(60, 68, 76, 84),
      height: responsive.getValueForSize(60, 68, 76, 84),
      borderRadius: responsive.getValueForSize(30, 34, 38, 42),
      backgroundColor: "rgba(255, 255, 255, 0.95)",
      justifyContent: "center",
      alignItems: "center",
      marginBottom: responsive.spacing.elementSpacing / 3,
      overflow: "hidden",
      ...theme.shadows.sm,
    },
    decorebatorIcon: {
      width: responsive.getValueForSize(40, 48, 54, 60),
      height: responsive.getValueForSize(40, 48, 54, 60),
      maxWidth: "100%",
      maxHeight: "100%",
    },
    title: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.headline,
        responsive.fontSizes.title,
        responsive.fontSizes.title,
        responsive.fontSizes.title,
      ),
      fontWeight: "700",
      color: theme.colors.text.inverse,
      marginTop: responsive.getValueForSize(8, 10, 12, 14),
      textAlign: "center",
    },
    subtitle: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.label,
        responsive.fontSizes.body,
        responsive.fontSizes.body,
        responsive.fontSizes.body,
      ),
      color: theme.colors.text.inverse,
      marginTop: responsive.getValueForSize(4, 6, 8, 10),
      opacity: 0.95,
      textAlign: "center",
      paddingHorizontal: responsive.spacing.horizontal,
      lineHeight: responsive.fontSizes.body * 1.3,
    },
    closeButton: {
      position: "absolute",
      top: responsive.getValueForSize(8, 10, 12, 14),
      right: responsive.spacing.horizontal,
      padding: responsive.getValueForSize(6, 8, 10, 12),
      zIndex: 10,
      backgroundColor: "rgba(255, 255, 255, 0.15)",
      borderRadius: theme.borderRadius.full,
    },
    content: {
      flex: 1,
    },
    featuresContainer: {
      marginTop: responsive.getValueForSize(12, 16, 20, 24),
      marginBottom: responsive.getValueForSize(12, 16, 20, 24),
      marginHorizontal: responsive.spacing.horizontal,
      backgroundColor:
        theme.mode === "light"
          ? `${theme.colors.primary}05`
          : theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      padding: responsive.getValueForSize(12, 16, 20, 24),
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? `${theme.colors.primary}10`
          : theme.colors.ui.border,
    },
    featuresTitle: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.label,
        responsive.fontSizes.body,
        responsive.fontSizes.body,
        responsive.fontSizes.body,
      ),
      fontWeight: "600",
      color: theme.colors.text.secondary,
      marginBottom: responsive.getValueForSize(8, 10, 12, 14),
      textAlign: "center",
    },
    featuresList: {
      gap: responsive.getValueForSize(6, 8, 10, 12),
    },
    featureItem: {
      flexDirection: "row",
      alignItems: "center",
      gap: responsive.getValueForSize(10, 12, 14, 16),
    },
    checkmarkContainer: {
      width: responsive.getValueForSize(20, 22, 24, 26),
      height: responsive.getValueForSize(20, 22, 24, 26),
      borderRadius: theme.borderRadius.full,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    featureText: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.body,
        responsive.fontSizes.body,
        responsive.fontSizes.headline,
        responsive.fontSizes.headline,
      ),
      color: theme.colors.text.primary,
      fontWeight: "400",
      flex: 1,
    },
    sectionDivider: {
      height: 1,
      backgroundColor:
        theme.mode === "light"
          ? theme.colors.ui.divider
          : theme.colors.border.light,
      marginHorizontal: responsive.spacing.horizontal * 2,
      marginVertical: responsive.getValueForSize(8, 10, 12, 14),
      opacity: 0.5,
    },
    packagesSection: {
      paddingHorizontal: responsive.spacing.horizontal,
      paddingTop: responsive.spacing.vertical / 2,
      paddingBottom: responsive.spacing.vertical / 3,
    },
    packagesGrid: {
      flexDirection: "row",
      flexWrap: "wrap",
      justifyContent: "space-between",
      gap: responsive.getValueForSize(8, 10, 12, 14),
    },
    packageCard: {
      backgroundColor:
        theme.mode === "light"
          ? theme.colors.background.surface
          : theme.colors.background.elevated,
      borderRadius: theme.borderRadius.xl,
      padding: responsive.getValueForSize(14, 16, 18, 20),
      borderWidth: 2,
      borderColor:
        theme.mode === "light"
          ? theme.colors.border.light
          : theme.colors.ui.border,
      ...theme.shadows.md,
      overflow: "hidden",
      transform: [{ scale: 1 }],
      width: `${48}%`,
      minHeight: responsive.getValueForSize(140, 160, 180, 200),
      justifyContent: "space-between",
    },
    popularPackage: {
      borderColor: theme.colors.primary,
      borderWidth: 3,
      ...theme.shadows.lg,
      shadowColor: theme.colors.primary,
      shadowOpacity: 0.2,
      elevation: 12,
    },
    popularGradient: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
      paddingVertical: responsive.getValueForSize(3, 4, 6, 8),
      alignItems: "center",
      borderTopLeftRadius: theme.borderRadius.lg - 2,
      borderTopRightRadius: theme.borderRadius.lg - 2,
    },
    popularBadgeText: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.micro,
        responsive.fontSizes.micro,
        responsive.fontSizes.label,
        responsive.fontSizes.label,
      ),
      fontWeight: "700",
      color: theme.colors.text.inverse,
      textTransform: "uppercase",
      letterSpacing: 0.5,
    },
    packageContent: {
      marginTop: responsive.getValueForSize(16, 20, 24, 28),
    },
    packageTitle: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.body,
        responsive.fontSizes.body,
        responsive.fontSizes.headline,
        responsive.fontSizes.headline,
      ),
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: responsive.getValueForSize(8, 10, 12, 14),
    },
    priceContainer: {
      flexDirection: "row",
      alignItems: "baseline",
      marginBottom: responsive.spacing.elementSpacing / 2,
    },
    packagePrice: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.headline,
        responsive.fontSizes.title,
        responsive.fontSizes.title,
        responsive.fontSizes.display,
      ),
      fontWeight: "700",
      color: theme.colors.primary,
    },
    packagePeriod: {
      fontSize: responsive.fontSizes.body,
      color: theme.colors.text.secondary,
    },
    savingsContainer: {
      flexDirection: "row",
      alignItems: "center",
      backgroundColor:
        theme.mode === "light"
          ? `${theme.colors.success}15`
          : `${theme.colors.success}20`,
      paddingHorizontal: responsive.getValueForSize(8, 10, 12, 14),
      paddingVertical: responsive.getValueForSize(3, 4, 5, 6),
      borderRadius: theme.borderRadius.md,
      alignSelf: "flex-start",
      marginBottom: responsive.getValueForSize(6, 8, 10, 12),
      borderWidth: 1,
      borderColor: `${theme.colors.success}30`,
    },
    savingsText: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.micro,
        responsive.fontSizes.label,
        responsive.fontSizes.label,
        responsive.fontSizes.label,
      ),
      color: theme.colors.success,
      fontWeight: "600",
      marginLeft: 4,
    },
    packageDescription: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.micro,
        responsive.fontSizes.label,
        responsive.fontSizes.label,
        responsive.fontSizes.label,
      ),
      color: theme.colors.text.secondary,
    },
    termsContainer: {
      marginTop: responsive.spacing.vertical / 3,
      marginBottom: responsive.getValueForSize(8, 10, 12, 14),
      paddingHorizontal: responsive.spacing.horizontal * 1.5,
      alignItems: "center",
    },
    termsRow: {
      flexDirection: "row",
      flexWrap: "wrap",
      justifyContent: "center",
      alignItems: "baseline",
    },
    termsText: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.body,
        responsive.fontSizes.body,
        responsive.fontSizes.body,
        responsive.fontSizes.body,
      ),
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: responsive.fontSizes.body * 1.4,
    },
    linkText: {
      color: theme.colors.primary,
      fontWeight: "500",
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.body,
        responsive.fontSizes.body,
        responsive.fontSizes.body,
        responsive.fontSizes.body,
      ),
      lineHeight: responsive.fontSizes.body * 1.4,
      textDecorationLine: "underline",
    },
    restoreButton: {
      alignItems: "center",
      paddingVertical: responsive.getValueForSize(8, 10, 12, 14),
      paddingHorizontal: responsive.spacing.horizontal,
      marginHorizontal: responsive.spacing.horizontal * 2,
      marginBottom: responsive.spacing.vertical,
    },
    restoreButtonText: {
      fontSize: responsive.getValueForSize(
        responsive.fontSizes.label,
        responsive.fontSizes.label,
        responsive.fontSizes.body,
        responsive.fontSizes.body,
      ),
      color: theme.colors.text.secondary,
      fontWeight: "500",
      textDecorationLine: "underline",
      textDecorationColor: theme.colors.text.secondary,
    },
    errorContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      padding: 40,
    },
    errorText: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginTop: 16,
    },
    purchasingOverlay: {
      ...StyleSheet.absoluteFillObject,
      backgroundColor: "rgba(0, 0, 0, 0.5)",
      justifyContent: "center",
      alignItems: "center",
    },
    purchasingText: {
      marginTop: 16,
      fontSize: 16,
      color: "#FFFFFF",
    },
  });
  return styles;
};
