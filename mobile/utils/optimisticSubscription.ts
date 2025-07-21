import type { SubscriptionStatus } from "@/api/subscriptions";

/**
 * Creates optimistic subscription data for immediate UI updates after subscription purchase
 * This matches the exact logic used in useRevenueCat hook for consistency
 *
 * @param plan - The subscription plan ("monthly" or "annual")
 * @returns Optimistic subscription data structure
 */
export function createOptimisticSubscriptionData(
  plan: "monthly" | "annual",
): SubscriptionStatus {
  // Calculate realistic currentPeriodEnd dates (matches useRevenueCat logic)
  const now = new Date();
  let currentPeriodEnd: string;

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

  // Return optimistic subscription data (matches useRevenueCat structure exactly)
  return {
    plan,
    status: "active" as const,
    currentPeriodEnd,
    cancelAtPeriodEnd: false,
    trialEnd: null,
    hasOptimisticSubscription: true,
    // Note: isActive, isCancelledButActive, isInGracePeriod are NOT set
    // These are backend-computed fields that will be undefined in optimistic state
  };
}
