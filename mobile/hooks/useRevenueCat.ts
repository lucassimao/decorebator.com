import { restorePurchases } from "@/api/revenuecat";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { Platform } from "react-native";
import Purchases, {
  CustomerInfo,
  PURCHASES_ERROR_CODE,
  PurchasesOffering,
  PurchasesPackage,
} from "react-native-purchases";
import { useUserInfo } from "./users";

const REVENUECAT_API_KEY_IOS =
  process.env.EXPO_PUBLIC_REVENUECAT_API_KEY_IOS || "";
const REVENUECAT_API_KEY_ANDROID =
  process.env.EXPO_PUBLIC_REVENUECAT_API_KEY_ANDROID || "";

export interface RevenueCatState {
  isInitialized: boolean;
  customerInfo: CustomerInfo | null;
  offerings: PurchasesOffering[] | null;
  isLoading: boolean;
  error: Error | null;
}

export function useRevenueCat() {
  const [state, setState] = useState<RevenueCatState>({
    isInitialized: false,
    customerInfo: null,
    offerings: null,
    isLoading: true,
    error: null,
  });

  const queryClient = useQueryClient();
  const { userInfo: user } = useUserInfo();

  // Initialize RevenueCat SDK
  useEffect(() => {
    const initializeRevenueCat = async () => {
      if (!user) return;

      try {
        // Configure RevenueCat based on platform
        const apiKey =
          Platform.OS === "ios"
            ? REVENUECAT_API_KEY_IOS
            : REVENUECAT_API_KEY_ANDROID;

        if (!apiKey) {
          throw new Error(
            `RevenueCat API key not configured for ${Platform.OS}`,
          );
        }

        await Purchases.configure({
          apiKey,
          appUserID: user.id.toString(), // Use user ID as app user ID
        });

        // Get initial customer info
        const customerInfo = await Purchases.getCustomerInfo();

        // Get available offerings
        const offerings = await Purchases.getOfferings();

        setState({
          isInitialized: true,
          customerInfo,
          offerings: offerings.all ? Object.values(offerings.all) : null,
          isLoading: false,
          error: null,
        });
      } catch (error) {
        console.error("Failed to initialize RevenueCat:", error);
        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: error as Error,
        }));
      }
    };

    if (Platform.OS !== "web") {
      initializeRevenueCat();
    }
  }, [user]);

  // Purchase a package
  const purchasePackage = useMutation({
    mutationFn: async (purchasePackage: PurchasesPackage) => {
      if (!state.isInitialized) {
        throw new Error("RevenueCat not initialized");
      }

      try {
        const { customerInfo } =
          await Purchases.purchasePackage(purchasePackage);
        return customerInfo;
      } catch (error: any) {
        if (error.code === PURCHASES_ERROR_CODE.PURCHASE_CANCELLED_ERROR) {
          throw new Error("Purchase cancelled");
        }
        throw error;
      }
    },
    onSuccess: (customerInfo) => {
      setState((prev) => ({ ...prev, customerInfo }));
      // Invalidate user query to refresh subscription status
      queryClient.invalidateQueries({ queryKey: ["user"] });
      queryClient.invalidateQueries({ queryKey: ["subscription"] });
    },
  });

  // Restore purchases
  const restorePurchasesMutation = useMutation({
    mutationFn: async () => {
      if (!state.isInitialized || !user) {
        throw new Error("RevenueCat not initialized");
      }

      // First restore in RevenueCat
      const customerInfo = await Purchases.restorePurchases();

      // Then sync with backend
      await restorePurchases(user.id.toString());

      return customerInfo;
    },
    onSuccess: (customerInfo) => {
      setState((prev) => ({ ...prev, customerInfo }));
      // Invalidate user query to refresh subscription status
      queryClient.invalidateQueries({ queryKey: ["user"] });
      queryClient.invalidateQueries({ queryKey: ["subscription"] });
    },
  });

  // Check if user has active entitlements
  const hasActiveSubscription = () => {
    if (!state.customerInfo) return false;

    // Check for "premium" entitlement
    const premiumEntitlement =
      state.customerInfo.entitlements.active["premium"];
    return !!premiumEntitlement;
  };

  // Get current offering for display
  const getCurrentOffering = () => {
    if (!state.offerings || state.offerings.length === 0) return null;

    // Return the first offering (usually "default")
    return state.offerings[0];
  };

  return {
    ...state,
    purchasePackage: purchasePackage.mutateAsync,
    restorePurchases: restorePurchasesMutation.mutateAsync,
    isPurchasing: purchasePackage.isPending,
    isRestoring: restorePurchasesMutation.isPending,
    hasActiveSubscription,
    getCurrentOffering,
  };
}

// Hook to determine which payment provider to use
export function usePaymentProvider() {
  const { userInfo: user } = useUserInfo();

  // Determine provider based on platform and user location
  const getProvider = (): {
    provider: `revenuecat` | "stripe";
    platform: "ios" | "android" | "web";
  } | null => {
    if (!user) return null;

    const platform =
      Platform.OS === "ios"
        ? "ios"
        : Platform.OS === "android"
          ? "android"
          : "web";

    // Android always uses RevenueCat
    if (platform === "android") {
      return { provider: "revenuecat", platform };
    }

    // iOS: US users use Stripe, others use RevenueCat
    if (platform === "ios") {
      const isUS = user.country === "US";
      return { provider: isUS ? "stripe" : "revenuecat", platform };
    }

    // Web always uses Stripe
    return { provider: "stripe", platform };
  };

  const providerInfo = getProvider();

  return {
    data: providerInfo,
    isLoading: false,
    error: null,
  };
}
