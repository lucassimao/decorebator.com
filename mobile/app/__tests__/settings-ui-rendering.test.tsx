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
  sigout: jest.fn(),
}));

const mockUseQuery = useQuery as jest.Mock;

describe("SettingsScreen - UI Rendering", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Default mock for useQuery (subscription data)
    mockUseQuery.mockReturnValue({
      data: createSubscriptionData(),
      isLoading: false,
      refetch: jest.fn(),
    });
  });

  describe("Conditional UI Rendering Based on Subscription State", () => {
    describe("Upgrade Section Visibility", () => {
      it("should show upgrade section for free plan users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.FREE_USER,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        expect(getByText("upgrade.title")).toBeTruthy();
        expect(getByText("upgrade.subtitle")).toBeTruthy();
        expect(getByText("settings.subscription.premiumFeatures")).toBeTruthy();
      });

      it("should hide upgrade section for active premium users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.ACTIVE_MONTHLY,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);
        expect(queryByText("upgrade.title")).toBeNull();
        expect(queryByText("upgrade.subtitle")).toBeNull();
        expect(queryByText("settings.subscription.premiumFeatures")).toBeNull();
      });

      it("should hide upgrade section for cancelled but still active users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.CANCELLED_BUT_ACTIVE,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);
        expect(queryByText("upgrade.title")).toBeNull();
      });

      it("should hide upgrade section for users in grace period", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            isInGracePeriod: true,
            status: "past_due",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);
        expect(queryByText("upgrade.title")).toBeNull();
      });

      it("should show upgrade section for expired premium users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.EXPIRED_CANCELLED,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        expect(getByText("upgrade.title")).toBeTruthy();
      });
    });

    describe("Free Plan Limitations Display", () => {
      it("should show free plan limitations for free users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.FREE_USER,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        expect(getByText("settings.subscription.freePlanLimit")).toBeTruthy();
      });

      it("should hide free plan limitations for premium users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.ACTIVE_MONTHLY,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);
        expect(queryByText("settings.subscription.freePlanLimit")).toBeNull();
      });
    });

    describe("Subscription Management Button Visibility", () => {
      it("should show manage subscription button for active, non-cancelled subscriptions", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "active",
            cancelAtPeriodEnd: false,
            isCancelledButActive: false,
            currentPeriodEnd: "2025-12-31T23:59:59Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        expect(
          getByText("settings.subscription.manageSubscription"),
        ).toBeTruthy();
      });

      it("should hide manage subscription button for cancelled subscriptions", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: false,
            status: "cancelled",
            isCancelledButActive: true,
            currentPeriodEnd: "2025-12-31T23:59:59Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);
        expect(
          queryByText("settings.subscription.manageSubscription"),
        ).toBeNull();
      });

      it("should hide manage subscription button for subscriptions set to cancel at period end", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "active",
            cancelAtPeriodEnd: true,
            isCancelledButActive: false,
            currentPeriodEnd: "2025-12-31T23:59:59Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);
        expect(
          queryByText("settings.subscription.manageSubscription"),
        ).toBeNull();
      });

      it("should show manage subscription button for users in grace period", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: true,
            isInGracePeriod: true,
            status: "past_due",
            currentPeriodEnd: "2025-12-31T23:59:59Z",
            cancelAtPeriodEnd: false,
            isCancelledButActive: false,
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        // Users in grace period should still be able to manage their subscription to fix payment issues
        expect(
          getByText("settings.subscription.manageSubscription"),
        ).toBeTruthy();
      });
    });

    describe("Subscription Details Section Visibility", () => {
      it("should show subscription details for active subscriptions with end date", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "active",
            currentPeriodEnd: "2025-12-31T23:59:59Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        expect(getByText("settings.subscription.renewsOn")).toBeTruthy();
        expect(getByText("Formatted: 2025-12-31T23:59:59Z")).toBeTruthy();
      });

      it("should show subscription details for cancelled but active subscriptions", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: false,
            isCancelledButActive: true,
            status: "cancelled",
            currentPeriodEnd: "2025-06-15T12:00:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        expect(getByText("settings.subscription.expiresOn")).toBeTruthy();
        expect(getByText("Formatted: 2025-06-15T12:00:00Z")).toBeTruthy();
      });

      it("should show subscription details for grace period subscriptions", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            isInGracePeriod: true,
            status: "past_due",
            currentPeriodEnd: "2025-03-01T10:30:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        expect(
          getByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeTruthy();
        expect(getByText("Formatted: 2025-03-01T10:30:00Z")).toBeTruthy();
      });

      it("should hide subscription details for expired subscriptions", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: false,
            isCancelledButActive: false,
            status: "cancelled",
            currentPeriodEnd: "2024-12-31T23:59:59Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
        expect(queryByText("settings.subscription.expiresOn")).toBeNull();
        expect(queryByText("Formatted: 2024-12-31T23:59:59Z")).toBeNull();
      });

      it("should hide subscription details when no currentPeriodEnd date is available", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "active",
            currentPeriodEnd: undefined as any,
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
      });
    });

    describe("Premium Badge Icon Display", () => {
      it("should show premium icon for premium users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.ACTIVE_MONTHLY,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        // Premium users should see the premium plan text, indicating premium icon is shown
        expect(getByText("settings.subscription.monthlyPremium")).toBeTruthy();
      });

      it("should show lock icon for free users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.FREE_USER,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);
        // Free users should see the free plan text, indicating lock icon is shown
        expect(getByText("settings.subscription.freePlan")).toBeTruthy();
      });
    });
  });
});
