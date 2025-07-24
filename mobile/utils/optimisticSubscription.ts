import type { SubscriptionStatus } from "@/api/subscriptions";

/**
 * Creates optimistic subscription data for immediate UI updates after subscription purchase
 * This matches the exact logic used in useRevenueCat hook for consistency
 *
 * @param plan - The subscription plan type
 * @param expiresDate - The actual expiry date from RevenueCat (optional, will calculate if not provided)
 * @returns Optimistic subscription data structure
 */
export function createOptimisticSubscriptionData(
  plan: "annual" | "monthly",
  expiresDate?: string | null,
): SubscriptionStatus {
  let currentPeriodEnd: string;

  if (expiresDate) {
    // Use the actual expiry date from RevenueCat
    currentPeriodEnd = expiresDate;
  } else {
    // Fallback: Calculate realistic currentPeriodEnd dates
    const now = new Date();

    if (plan === "annual") {
      // Annual: 1 year from now
      const yearlyEnd = new Date(now);
      yearlyEnd.setFullYear(now.getFullYear() + 1);
      currentPeriodEnd = yearlyEnd.toISOString();
    } else {
      // Monthly: 30 days from now
      const monthlyEnd = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
      currentPeriodEnd = monthlyEnd.toISOString();
    }
  }

  // Return optimistic subscription data (matches backend subscription status endpoint exactly)
  return {
    plan,
    status: "active" as const,
    currentPeriodEnd,
    cancelAtPeriodEnd: false,
    trialEnd: undefined,
    hasOptimisticSubscription: true,
    // Backend-computed fields for optimistic state
    isActive: true,
    isCancelledButActive: false,
    isInGracePeriod: false,
    daysRemaining: undefined, // Only set for cancelled subscriptions
  };
}
