import React from "react";
import { render } from "@testing-library/react-native";
import SettingsScreen from "../settings";
import { useQuery } from "@tanstack/react-query";
import {
  createSubscriptionData,
  SUBSCRIPTION_SCENARIOS,
} from "./utils/test-helpers";

// Mock API modules that depend on expo-secure-store
jest.mock("@/api/subscriptions", () => ({
  getSubscriptionStatus: jest.fn(),
  createCheckoutSession: jest.fn(),
  openNativeSubscriptionManagement: jest.fn(),
}));

jest.mock("@/api/users", () => ({
  getAuthorization: jest.fn(() => null),
  hasAuthenticationSession: jest.fn(() => false),
  getAuthenticationSessionEpoch: jest.fn(() => null),
  isAuthenticationSessionReady: jest.fn(() => true),
  subscribeAuthenticationSessionChanges: jest.fn(() => jest.fn()),
  clearSessionScopedCaches: jest.fn(() => Promise.resolve()),
  persistSessionSnapshot: jest.fn(() => Promise.resolve()),
  getProfile: jest.fn(),
  sigout: jest.fn(),
  update: jest.fn(),
}));

const mockUseQuery = useQuery as jest.Mock;

describe("SettingsScreen - Subscription Logic", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Default mock for useQuery (subscription data)
    mockUseQuery.mockReturnValue({
      data: createSubscriptionData(),
      isLoading: false,
      refetch: jest.fn(),
    });
  });

  describe("isPremium Logic", () => {
    it("should be false for free plan users", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.FREE_USER,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      // Should show free plan
      expect(getByTestId("free-plan-name")).toBeTruthy();
      // Should show upgrade section
      expect(getByTestId("upgrade-title")).toBeTruthy();
    });

    it("should be true for active monthly premium users", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.ACTIVE_MONTHLY,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId, queryByTestId } = render(<SettingsScreen />);
      // Should show monthly premium plan
      expect(getByTestId("monthly-premium-name")).toBeTruthy();
      // Should NOT show upgrade section for premium users
      expect(queryByTestId("upgrade-title")).toBeNull();
    });

    it("should be true for active annual premium users", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.ACTIVE_ANNUAL,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId, queryByTestId } = render(<SettingsScreen />);
      // Should show annual premium plan
      expect(getByTestId("yearly-premium-name")).toBeTruthy();
      // Should NOT show upgrade section for premium users
      expect(queryByTestId("upgrade-title")).toBeNull();
    });

    it("should be true for cancelled but still active premium users", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.CANCELLED_BUT_ACTIVE,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId, queryByTestId } = render(<SettingsScreen />);
      // Should show monthly premium plan
      expect(getByTestId("monthly-premium-name")).toBeTruthy();
      // Should NOT show upgrade section for still-active cancelled users
      expect(queryByTestId("upgrade-title")).toBeNull();
    });

    it("should be false for expired cancelled premium users", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.EXPIRED_CANCELLED,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      // Should show upgrade section for expired users
      expect(getByTestId("upgrade-title")).toBeTruthy();
    });

    it("should be true for users in grace period", () => {
      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "annual",
          isActive: true,
          isInGracePeriod: true,
          status: "past_due",
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId, queryByTestId } = render(<SettingsScreen />);
      // Should show annual premium plan
      expect(getByTestId("yearly-premium-name")).toBeTruthy();
      // Should NOT show upgrade section for users in grace period
      expect(queryByTestId("upgrade-title")).toBeNull();
    });

    it("should be false for unknown/invalid plan types", () => {
      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "unknown" as any,
          isActive: true,
          status: "active",
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      // Should fallback to free plan display
      expect(getByTestId("free-plan-name")).toBeTruthy();
      // Should show upgrade section for unknown plan types
      expect(getByTestId("upgrade-title")).toBeTruthy();
    });

    it("should be true for optimistic monthly subscriptions", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.OPTIMISTIC_MONTHLY,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId, queryByTestId } = render(<SettingsScreen />);
      // Should show monthly premium plan
      expect(getByTestId("monthly-premium-name")).toBeTruthy();
      // Should show active status via plan-status test ID
      expect(getByTestId("plan-status")).toBeTruthy();
      // Should NOT show upgrade section for optimistic premium users
      expect(queryByTestId("upgrade-title")).toBeNull();
    });

    it("should be true for optimistic annual subscriptions", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.OPTIMISTIC_ANNUAL,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId, queryByTestId } = render(<SettingsScreen />);
      // Should show annual premium plan
      expect(getByTestId("yearly-premium-name")).toBeTruthy();
      // Should show active status via plan-status test ID
      expect(getByTestId("plan-status")).toBeTruthy();
      // Should NOT show upgrade section for optimistic premium users
      expect(queryByTestId("upgrade-title")).toBeNull();
    });

    it("should treat optimistic subscriptions as premium with backend-computed fields", () => {
      // This test ensures optimistic subscriptions work with proper backend-computed fields for immediate UI updates
      const optimisticData = SUBSCRIPTION_SCENARIOS.OPTIMISTIC_MONTHLY;

      // Verify the optimistic data has proper backend-computed fields for immediate UI functionality
      expect(optimisticData.isActive).toBe(true);
      expect(optimisticData.isCancelledButActive).toBe(false);
      expect(optimisticData.isInGracePeriod).toBe(false);

      mockUseQuery.mockReturnValue({
        data: optimisticData,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId, queryByTestId } = render(<SettingsScreen />);

      // Should still be treated as premium despite missing backend-computed fields
      expect(getByTestId("monthly-premium-name")).toBeTruthy();
      expect(queryByTestId("upgrade-title")).toBeNull();
      expect(queryByTestId("free-plan-limit")).toBeNull();
    });
  });
});
