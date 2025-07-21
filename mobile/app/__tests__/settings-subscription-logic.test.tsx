import React from "react";
import { render } from "@testing-library/react-native";
import SettingsScreen from "../settings";
import { useQuery } from "@tanstack/react-query";
import {
  createSubscriptionData,
  SUBSCRIPTION_SCENARIOS,
} from "./helpers/settings-helpers";

// Mock API modules that depend on expo-secure-store
jest.mock("@/api/subscriptions", () => ({
  getSubscriptionStatus: jest.fn(),
  createCheckoutSession: jest.fn(),
  openNativeSubscriptionManagement: jest.fn(),
}));

jest.mock("@/api/users", () => ({
  sigout: jest.fn(),
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

      const { getByText } = render(<SettingsScreen />);
      // Should show free plan
      expect(getByText("settings.subscription.freePlan")).toBeTruthy();
      // Should show upgrade section
      expect(getByText("upgrade.title")).toBeTruthy();
    });

    it("should be true for active monthly premium users", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.ACTIVE_MONTHLY,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByText, queryByText } = render(<SettingsScreen />);
      // Should show monthly premium plan
      expect(getByText("settings.subscription.monthlyPremium")).toBeTruthy();
      // Should NOT show upgrade section for premium users
      expect(queryByText("upgrade.title")).toBeNull();
    });

    it("should be true for active annual premium users", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.ACTIVE_ANNUAL,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByText, queryByText } = render(<SettingsScreen />);
      // Should show annual premium plan
      expect(getByText("settings.subscription.yearlyPremium")).toBeTruthy();
      // Should NOT show upgrade section for premium users
      expect(queryByText("upgrade.title")).toBeNull();
    });

    it("should be true for cancelled but still active premium users", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.CANCELLED_BUT_ACTIVE,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByText, queryByText } = render(<SettingsScreen />);
      // Should show monthly premium plan
      expect(getByText("settings.subscription.monthlyPremium")).toBeTruthy();
      // Should NOT show upgrade section for still-active cancelled users
      expect(queryByText("upgrade.title")).toBeNull();
    });

    it("should be false for expired cancelled premium users", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.EXPIRED_CANCELLED,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByText } = render(<SettingsScreen />);
      // Should show upgrade section for expired users
      expect(getByText("upgrade.title")).toBeTruthy();
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

      const { getByText, queryByText } = render(<SettingsScreen />);
      // Should show annual premium plan
      expect(getByText("settings.subscription.yearlyPremium")).toBeTruthy();
      // Should NOT show upgrade section for users in grace period
      expect(queryByText("upgrade.title")).toBeNull();
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

      const { getByText } = render(<SettingsScreen />);
      // Should fallback to free plan display
      expect(getByText("settings.subscription.freePlan")).toBeTruthy();
      // Should show upgrade section for unknown plan types
      expect(getByText("upgrade.title")).toBeTruthy();
    });
  });
});
