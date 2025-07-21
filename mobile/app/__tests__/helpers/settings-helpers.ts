import { SubscriptionStatus } from "@/api/subscriptions";

/**
 * Helper function to create subscription data for testing
 * Provides default values and allows overriding specific properties
 */
export const createSubscriptionData = (
  overrides: Partial<SubscriptionStatus> = {},
): SubscriptionStatus => ({
  plan: "free",
  status: "active",
  isActive: true,
  isCancelledButActive: false,
  isInGracePeriod: false,
  currentPeriodEnd: "2024-12-31T23:59:59Z",
  cancelAtPeriodEnd: false,
  daysRemaining: undefined,
  ...overrides,
});

/**
 * Common mock setup for useQuery hook
 * Returns a mock implementation that can be customized per test
 */
export const createMockUseQuery = (mockImplementation?: any) => {
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const mockUseQuery = require("@tanstack/react-query").useQuery as jest.Mock;

  if (mockImplementation) {
    mockUseQuery.mockReturnValue(mockImplementation);
  } else {
    // Default implementation
    mockUseQuery.mockReturnValue({
      data: createSubscriptionData(),
      isLoading: false,
      refetch: jest.fn(),
    });
  }

  return mockUseQuery;
};

/**
 * Test constants for consistent values across test suites
 */
export const TEST_CONSTANTS = {
  UTC_DATE: "2024-12-31T23:59:59Z",
  FORMATTED_DATE: "December 31st, 2024",
  GRACE_PERIOD_DAYS: 15,
  DEFAULT_USER_ID: 123,
} as const;

/**
 * Common test scenarios for subscription states
 */
export const SUBSCRIPTION_SCENARIOS = {
  FREE_USER: createSubscriptionData({
    plan: "free",
    isActive: true,
    status: "active",
  }),

  ACTIVE_MONTHLY: createSubscriptionData({
    plan: "monthly",
    isActive: true,
    status: "active",
  }),

  ACTIVE_ANNUAL: createSubscriptionData({
    plan: "annual",
    isActive: true,
    status: "active",
  }),

  CANCELLED_BUT_ACTIVE: createSubscriptionData({
    plan: "monthly",
    isActive: false,
    isCancelledButActive: true,
    status: "cancelled",
    daysRemaining: 15,
  }),

  EXPIRED_CANCELLED: createSubscriptionData({
    plan: "monthly",
    isActive: false,
    isCancelledButActive: false,
    status: "cancelled",
    daysRemaining: 0,
  }),

  GRACE_PERIOD: createSubscriptionData({
    plan: "monthly",
    isActive: false,
    isCancelledButActive: false,
    isInGracePeriod: true,
    status: "past_due",
    daysRemaining: 2,
  }),

  PAST_DUE: createSubscriptionData({
    plan: "monthly",
    isActive: false,
    isCancelledButActive: false,
    isInGracePeriod: false,
    status: "past_due",
    daysRemaining: 0,
  }),
} as const;
