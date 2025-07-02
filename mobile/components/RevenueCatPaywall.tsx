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
import { PurchasesPackage } from "react-native-purchases";
import { useResponsive, useResponsiveSpacing } from "@/hooks/useResponsive";

interface RevenueCatPaywallProps {
  onClose: () => void;
  onSuccess: () => void;
}

export default function RevenueCatPaywall({
  onClose,
  onSuccess,
}: RevenueCatPaywallProps) {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { isTablet, contentWidth } = useResponsive();
  const spacing = useResponsiveSpacing();
  const styles = createStyles(theme, isTablet, contentWidth, spacing);

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
      await purchasePackage(pkg);
      Alert.alert(
        t("common.success"),
        t("settings.subscription.purchaseSuccess"),
      );
      onSuccess();
    } catch (error: any) {
      if (error.message !== "Purchase cancelled") {
        Alert.alert(
          t("common.error"),
          error.message || t("settings.subscription.purchaseError"),
        );
      }
    }
  };

  const handleRestore = async () => {
    try {
      await restorePurchases();
      Alert.alert(
        t("common.success"),
        t("settings.subscription.restoreSuccess"),
      );
      onSuccess();
    } catch (error: any) {
      Alert.alert(
        t("common.error"),
        error.message || t("settings.subscription.restoreError"),
      );
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
        <View
          style={[
            styles.contentContainer,
            isTablet && styles.tabletContentContainer,
          ]}
        >
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
          <View
            style={[
              styles.packagesSection,
              isTablet && styles.tabletPackagesSection,
            ]}
          >
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
                    isTablet && styles.tabletPackageCard,
                  ]}
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
          <Text style={styles.termsText}>
            {t("settings.subscription.terms")}
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
  isTablet: boolean,
  contentWidth: number,
  spacing: ReturnType<typeof useResponsiveSpacing>,
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
      marginTop: spacing.md,
      fontSize: 16,
      color: theme.colors.text.secondary,
    },
    header: {
      flexDirection: "row",
      justifyContent: "space-between",
      alignItems: "center",
      padding: spacing.lg,
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
      padding: spacing.sm,
    },
    content: {
      flex: 1,
    },
    contentContainer: {
      width: "100%",
    },
    tabletContentContainer: {
      maxWidth: Math.min(contentWidth * 0.8, 600),
      alignSelf: "center",
      paddingHorizontal: spacing.lg,
    },
    featuresSection: {
      padding: spacing.lg,
    },
    sectionTitle: {
      fontSize: 18,
      fontWeight: "600",
      color: theme.colors.text.primary,
      marginBottom: spacing.md,
    },
    featureItem: {
      flexDirection: "row",
      alignItems: "center",
      marginBottom: spacing.sm,
    },
    featureText: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      marginLeft: spacing.sm,
      flex: 1,
    },
    packagesSection: {
      padding: spacing.lg,
      paddingTop: 0,
    },
    tabletPackagesSection: {
      flexDirection: isTablet ? "row" : "column",
      gap: isTablet ? spacing.md : 0,
    },
    packageCard: {
      backgroundColor: theme.colors.background.surface,
      borderRadius: theme.borderRadius.lg,
      padding: spacing.lg,
      marginBottom: isTablet ? 0 : spacing.md,
      borderWidth: 2,
      borderColor: theme.colors.ui.border,
      flex: isTablet ? 1 : undefined,
      ...theme.shadows.md,
    },
    tabletPackageCard: {
      marginBottom: 0,
    },
    popularPackage: {
      borderColor: theme.colors.primary,
      position: "relative",
    },
    popularBadge: {
      position: "absolute",
      top: -12,
      right: spacing.lg,
      backgroundColor: theme.colors.primary,
      paddingHorizontal: spacing.sm,
      paddingVertical: spacing.xs,
      borderRadius: theme.borderRadius.sm,
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
      marginBottom: spacing.sm,
    },
    packagePrice: {
      fontSize: 32,
      fontWeight: "700",
      color: theme.colors.primary,
      marginBottom: spacing.xs,
    },
    savingsText: {
      fontSize: 14,
      color: theme.colors.success,
      fontWeight: "600",
      marginBottom: spacing.sm,
    },
    packageDescription: {
      fontSize: 14,
      color: theme.colors.text.secondary,
    },
    restoreButton: {
      alignItems: "center",
      padding: spacing.md,
      marginHorizontal: spacing.lg,
      marginBottom: spacing.md,
    },
    restoreButtonText: {
      fontSize: 16,
      color: theme.colors.primary,
      fontWeight: "600",
    },
    termsText: {
      fontSize: 12,
      color: theme.colors.text.tertiary,
      textAlign: "center",
      paddingHorizontal: spacing.lg,
      paddingBottom: spacing.lg,
      lineHeight: 18,
    },
    errorContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      padding: spacing.xl,
    },
    errorText: {
      fontSize: 16,
      color: theme.colors.text.secondary,
      textAlign: "center",
      marginTop: spacing.md,
    },
    purchasingOverlay: {
      ...StyleSheet.absoluteFillObject,
      backgroundColor: "rgba(0, 0, 0, 0.5)",
      justifyContent: "center",
      alignItems: "center",
    },
    purchasingText: {
      marginTop: spacing.md,
      fontSize: 16,
      color: "#FFFFFF",
    },
  });
  return styles;
};
