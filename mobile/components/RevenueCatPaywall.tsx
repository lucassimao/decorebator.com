import React, { useEffect } from "react";
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from "react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useRevenueCat } from "@/hooks/useRevenueCat";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { PurchasesPackage, PACKAGE_TYPE } from "react-native-purchases";
import * as WebBrowser from "expo-web-browser";
import i18n from "@/i18n";

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
    const language = i18n.language || "en";
    WebBrowser.openBrowserAsync(`https://decorebator.com/${language}/terms`);
  };

  const openPrivacy = () => {
    const language = i18n.language || "en";
    WebBrowser.openBrowserAsync(`https://decorebator.com/${language}/privacy`);
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
        // User cancelled the purchase - don't show error
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
      <View style={styles.header}>
        <Text style={styles.title}>
          {t("settings.subscription.upgradeToPremium")}
        </Text>
        <TouchableOpacity onPress={onClose} style={styles.closeButton}>
          <MaterialIcons
            name="close"
            size={24}
            color={theme.colors.text.primary}
          />
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* Features */}
        <View style={styles.featuresSection}>
          <Text style={styles.sectionTitle}>
            {t("settings.subscription.features.title")}
          </Text>
          {[
            t("settings.subscription.features.unlimitedWordlists"),
            t("settings.subscription.features.aiEnrichment"),
            t("settings.subscription.features.allQuizModes"),
            t("settings.subscription.features.progressTracking"),
            t("settings.subscription.features.prioritySupport"),
          ].map((feature, index) => (
            <View key={index} style={styles.featureItem}>
              <MaterialIcons
                name="check-circle"
                size={20}
                color={theme.colors.success}
              />
              <Text style={styles.featureText}>{feature}</Text>
            </View>
          ))}
        </View>

        {/* Packages */}
        <View style={styles.packagesSection}>
          {currentOffering.availablePackages.map((pkg) => {
            const isMonthly = pkg.packageType === "MONTHLY";
            const isAnnual = pkg.packageType === "ANNUAL";
            const isPopular = isAnnual; // Mark annual as popular

            return (
              <TouchableOpacity
                key={pkg.identifier}
                style={[styles.packageCard, isPopular && styles.popularPackage]}
                onPress={() => handlePurchase(pkg)}
                disabled={isPurchasing}
              >
                {isPopular && (
                  <View style={styles.popularBadge}>
                    <Text style={styles.popularBadgeText}>
                      {t("settings.subscription.mostPopular")}
                    </Text>
                  </View>
                )}

                <Text style={styles.packageTitle}>
                  {isMonthly
                    ? t("settings.subscription.monthly")
                    : t("settings.subscription.yearly")}
                </Text>

                <Text style={styles.packagePrice}>
                  {pkg.product.priceString}
                </Text>

                {isAnnual && (
                  <Text style={styles.savingsText}>
                    {t("settings.subscription.savePercent", { percent: 17 })}
                  </Text>
                )}

                <Text style={styles.packageDescription}>
                  {isMonthly
                    ? t("settings.subscription.billedMonthly")
                    : t("settings.subscription.billedYearly")}
                </Text>
              </TouchableOpacity>
            );
          })}
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

        {/* Terms */}
        <View style={styles.termsContainer}>
          <Text style={styles.termsText}>
            {t("settings.subscription.termsPrefix")}{" "}
            <TouchableOpacity
              onPress={openTerms}
              style={styles.inlineLinkContainer}
              hitSlop={{ top: 8, bottom: 8, left: 4, right: 4 }}
            >
              <Text style={styles.linkText}>{t("termsOfService")}</Text>
            </TouchableOpacity>{" "}
            {t("settings.subscription.termsMiddle")}{" "}
            <TouchableOpacity
              onPress={openPrivacy}
              style={styles.inlineLinkContainer}
              hitSlop={{ top: 8, bottom: 8, left: 4, right: 4 }}
            >
              <Text style={styles.linkText}>{t("privacyPolicy")}</Text>
            </TouchableOpacity>
            . {t("settings.subscription.termsSuffix")}
          </Text>
        </View>
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
    header: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      padding: 20,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.ui.border,
    },
    title: {
      fontSize: 24,
      fontWeight: "700",
      color: theme.colors.text.primary,
      flex: 1,
    },
    closeButton: {
      padding: 8,
    },
    content: {
      flex: 1,
    },
    featuresSection: {
      padding: 20,
    },
    sectionTitle: {
      fontSize: 18,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 16,
    },
    featureItem: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: 12,
    },
    featureText: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      marginLeft: 12,
      flex: 1,
    },
    packagesSection: {
      padding: 20,
      paddingTop: 0,
    },
    packageCard: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: 16,
      padding: 20,
      marginBottom: 16,
      borderWidth: 2,
      borderColor: theme.colors.ui.border,
    },
    popularPackage: {
      borderColor: theme.colors.primary,
      position: "relative",
    },
    popularBadge: {
      position: "absolute",
      top: -12,
      right: 20,
      backgroundColor: theme.colors.primary,
      paddingHorizontal: 12,
      paddingVertical: 4,
      borderRadius: 12,
    },
    popularBadgeText: {
      fontSize: 12,
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    packageTitle: {
      fontSize: 20,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: 8,
    },
    packagePrice: {
      fontSize: 32,
      fontWeight: "700",
      color: theme.colors.primary,
      marginBottom: 4,
    },
    savingsText: {
      fontSize: 14,
      color: theme.colors.success,
      fontWeight: "600",
      marginBottom: 8,
    },
    packageDescription: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
    restoreButton: {
      alignItems: "center",
      padding: 16,
      marginHorizontal: 20,
      marginBottom: 16,
    },
    restoreButtonText: {
      fontSize: 16,
      color: theme.colors.primary,
      fontWeight: "600",
    },
    termsContainer: {
      marginTop: responsive.spacing.elementSpacing,
      paddingHorizontal: responsive.spacing.horizontal / 2,
      paddingBottom: responsive.spacing.vertical,
    },
    termsText: {
      fontSize: responsive.fontSizes.micro,
      color: theme.colors.text.secondary,
      textAlign: "center",
      lineHeight: responsive.fontSizes.micro * 1.4,
    },
    inlineLinkContainer: {
      minHeight: responsive.spacing.minTouchTarget,
      justifyContent: "center",
    },
    linkText: {
      color: theme.colors.primary,
      fontWeight: "600",
      fontSize: responsive.fontSizes.micro,
      textDecorationLine: "underline",
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
