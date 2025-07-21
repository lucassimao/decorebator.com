import { getAuthorization } from "./users";
import { DEFAULT_ERROR } from "./constants";
import { Platform, Linking } from "react-native";

const API_URL = process.env.EXPO_PUBLIC_API_URL;

export type SubscriptionStatus = {
  plan: "free" | "monthly" | "annual";
  status?: "active" | "cancelled" | "past_due" | "trialing" | "unpaid";
  currentPeriodEnd?: string;
  cancelAtPeriodEnd?: boolean;
  trialEnd?: string;
  hasOptimisticSubscription?: boolean;

  // New computed fields from backend
  isActive?: boolean;
  isCancelledButActive?: boolean;
  isInGracePeriod?: boolean;
  daysRemaining?: number;
};

export type CheckoutSessionResponse = {
  checkoutUrl: string;
  sessionId: string;
};

export async function createCheckoutSession(
  plan: "monthly" | "annual",
  expoUri: string,
): Promise<CheckoutSessionResponse> {
  const endpoint = `${process.env.EXPO_PUBLIC_API_URL}/subscription/checkout-session`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: authorization,
    },
    body: JSON.stringify({ plan, expoUri }),
  });

  if (!response.ok) {
    const body = await response.json();
    throw new Error(body?.error || DEFAULT_ERROR);
  }

  return response.json();
}

export async function getSubscriptionStatus(): Promise<SubscriptionStatus> {
  const endpoint = `${API_URL}/subscription/status`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
    method: "GET",
    headers: {
      Authorization: authorization,
    },
  });

  if (!response.ok) {
    const body = await response.json();
    throw new Error(body?.error || DEFAULT_ERROR);
  }

  return response.json();
}

export async function openNativeSubscriptionManagement(): Promise<void> {
  // Universal subscription management using platform-specific URL deeplinks
  // Works across ALL iOS versions (not just 13.4+) and Android devices

  if (Platform.OS === "web") {
    // For web, we could redirect to Stripe customer portal or show instructions
    throw new Error("Subscription management not available on web");
  }

  if (Platform.OS === "ios") {
    // iOS: Try multiple URLs in order of preference for maximum compatibility
    const iosUrls = [
      "itms-apps://apps.apple.com/account/subscriptions", // Modern iOS (10.3+)
      "itms://buy.itunes.apple.com/WebObjects/MZFinance.woa/wa/manageSubscriptions", // Legacy iOS
      "https://apps.apple.com/account/subscriptions", // Web fallback
    ];

    for (const url of iosUrls) {
      try {
        const canOpen = await Linking.canOpenURL(url);
        if (canOpen) {
          await Linking.openURL(url);
          return;
        }
      } catch (error) {
        // Continue to next URL if this one fails
        console.warn(`Failed to open iOS subscription URL: ${url}`, error);
      }
    }

    // If all URLs fail, throw error
    throw new Error(
      "Could not open subscription management on this iOS device",
    );
  }

  if (Platform.OS === "android") {
    // Android: Open Google Play Store subscription management
    const androidUrls = [
      "https://play.google.com/store/account/subscriptions", // Primary
      "market://details?id=com.yourapp.package", // Fallback (would need actual package name)
    ];

    for (const url of androidUrls) {
      try {
        const canOpen = await Linking.canOpenURL(url);
        if (canOpen) {
          await Linking.openURL(url);
          return;
        }
      } catch (error) {
        console.warn(`Failed to open Android subscription URL: ${url}`, error);
      }
    }

    // If all URLs fail, throw error
    throw new Error(
      "Could not open subscription management on this Android device",
    );
  }

  throw new Error(
    `Subscription management not supported on platform: ${Platform.OS}`,
  );
}
