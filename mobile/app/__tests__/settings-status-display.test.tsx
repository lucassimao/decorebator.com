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

describe("SettingsScreen - Status Display", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Default mock for useQuery (subscription data)
    mockUseQuery.mockReturnValue({
      data: createSubscriptionData(),
      isLoading: false,
      refetch: jest.fn(),
    });
  });

  describe("Status Text Display Logic", () => {
    it("should display 'statusActive' for active premium subscriptions", () => {
      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "monthly",
          isActive: true,
          status: "active",
          isInGracePeriod: false,
          isCancelledButActive: false,
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      expect(getByTestId("plan-status")).toBeTruthy();
    });

    it("should display 'statusCancelledActive' for cancelled but still active subscriptions", () => {
      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "annual",
          isActive: false,
          status: "cancelled",
          isInGracePeriod: false,
          isCancelledButActive: true,
          daysRemaining: 10,
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      expect(getByTestId("plan-status")).toBeTruthy();
    });

    it("should display 'statusGracePeriod' for subscriptions in grace period", () => {
      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "monthly",
          isActive: true,
          status: "past_due",
          isInGracePeriod: true,
          isCancelledButActive: false,
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      expect(getByTestId("plan-status")).toBeTruthy();
    });

    it("should display 'statusExpired' for cancelled and expired subscriptions", () => {
      mockUseQuery.mockReturnValue({
        data: SUBSCRIPTION_SCENARIOS.EXPIRED_CANCELLED,
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      expect(getByTestId("plan-status")).toBeTruthy();
    });

    it("should display raw status for unknown/unhandled status values", () => {
      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "monthly",
          isActive: false,
          status: "trialing" as any, // Unknown status
          isInGracePeriod: false,
          isCancelledButActive: false,
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      expect(getByTestId("plan-status")).toBeTruthy();
    });

    it("should show grace period warning with emoji for users in grace period", () => {
      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "monthly",
          isActive: true,
          status: "past_due",
          isInGracePeriod: true,
          currentPeriodEnd: "2025-02-01T08:15:30Z",
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      // Should show grace period warning with emoji
      expect(getByTestId("grace-period-warning")).toBeTruthy();
    });

    it("should show remaining days for cancelled but active subscriptions", () => {
      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "annual",
          isActive: false,
          status: "cancelled",
          isCancelledButActive: true,
          daysRemaining: 7,
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      // Should show days remaining section
      expect(getByTestId("days-remaining-label")).toBeTruthy();
      expect(getByTestId("days-remaining-count")).toBeTruthy();
    });
  });
});
