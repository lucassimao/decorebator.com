import React from "react";
import { render } from "@testing-library/react-native";
import SettingsScreen from "../settings";
import { useQuery } from "@tanstack/react-query";
import { createSubscriptionData } from "./utils/test-helpers";

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

describe("SettingsScreen - Edge Cases", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Default mock for useQuery (subscription data)
    mockUseQuery.mockReturnValue({
      data: createSubscriptionData(),
      isLoading: false,
      refetch: jest.fn(),
    });
  });

  describe("Edge Cases and Error Scenarios", () => {
    describe("Subscription Data Loading States", () => {
      it("should handle loading state gracefully", () => {
        mockUseQuery.mockReturnValue({
          data: null,
          isLoading: true,
          refetch: jest.fn(),
        });

        const { getByText, queryByText } = render(<SettingsScreen />);

        // Should show the section header even when loading
        expect(getByText("settings.subscription.currentPlan")).toBeTruthy();

        // Should show loading indicator and not crash
        // (Component should show basic structure and not crash)
        expect(queryByText("settings.subscription.freePlan")).toBeNull();
        expect(queryByText("settings.subscription.monthlyPremium")).toBeNull();
        // During loading, upgrade section is not shown (loading takes precedence)
        expect(queryByText("upgrade.title")).toBeNull();
      });

      it("should handle missing subscription data gracefully", () => {
        mockUseQuery.mockReturnValue({
          data: null,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should show basic structure
        expect(getByText("settings.subscription.currentPlan")).toBeTruthy();

        // Should show fallback free plan when subscription data is missing
        expect(getByText("settings.subscription.freePlan")).toBeTruthy();
        expect(getByText("settings.subscription.statusActive")).toBeTruthy();

        // Should show upgrade section for fallback free plan state
        expect(getByText("upgrade.title")).toBeTruthy();
      });

      it("should handle undefined subscription properties", () => {
        mockUseQuery.mockReturnValue({
          data: {
            // Minimal data with most properties undefined
            plan: "monthly",
            // All other properties undefined
          } as any,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should not crash with undefined properties
        expect(getByText("settings.subscription.currentPlan")).toBeTruthy();
        // Should show some form of plan information even with incomplete data
        expect(getByText("settings.subscription.monthlyPremium")).toBeTruthy();
      });
    });

    describe("Invalid Subscription Data", () => {
      it("should handle invalid date strings gracefully", () => {
        const { formatDate: mockFormatDate } = jest.requireMock(
          "@/utils/dateUtils",
        ) as {
          formatDate: jest.Mock;
        };

        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "active",
            currentPeriodEnd: "invalid-date-string", // Invalid date format
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        // Mock formatDate to simulate error handling
        mockFormatDate.mockReturnValue("Invalid Date");

        const { getByText } = render(<SettingsScreen />);

        // Should still show the renewsOn label
        expect(getByText("settings.subscription.renewsOn")).toBeTruthy();
        // Should call formatDate with the invalid date
        expect(mockFormatDate).toHaveBeenCalledWith("invalid-date-string");
        // Should display whatever formatDate returns (error handling is in dateUtils)
        expect(getByText("Invalid Date")).toBeTruthy();
      });

      it("should handle null/undefined currentPeriodEnd", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: true,
            status: "active",
            currentPeriodEnd: null as any, // Null date
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);

        // Should not show date labels when currentPeriodEnd is null
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
        expect(queryByText("settings.subscription.expiresOn")).toBeNull();
      });

      it("should handle unknown subscription status values", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: false,
            status: "unknown_status" as any, // Unknown status
            isInGracePeriod: false,
            isCancelledButActive: false,
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should display the raw status value when status is unknown
        expect(getByText("unknown_status")).toBeTruthy();
      });

      it("should handle unknown plan types gracefully", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "enterprise" as any, // Unknown plan type
            isActive: true,
            status: "active",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should not crash with unknown plan type
        expect(getByText("settings.subscription.currentPlan")).toBeTruthy();

        // Should fallback to free plan display for unknown plan types
        expect(getByText("settings.subscription.freePlan")).toBeTruthy();

        // Should show upgrade section since unknown plans are treated as free
        expect(getByText("upgrade.title")).toBeTruthy();
      });
    });

    describe("Boundary Conditions", () => {
      it("should handle zero days remaining", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: false,
            status: "cancelled",
            isCancelledButActive: true,
            daysRemaining: 0, // Zero days left
            currentPeriodEnd: "2025-01-01T00:00:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should still show days remaining section even with 0 days
        expect(getByText("settings.subscription.daysRemaining")).toBeTruthy();
        expect(getByText("settings.subscription.daysCount")).toBeTruthy();
      });

      it("should handle negative days remaining", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: false,
            status: "cancelled",
            isCancelledButActive: true,
            daysRemaining: -5, // Negative days (should not happen but test edge case)
            currentPeriodEnd: "2024-12-01T00:00:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should handle negative days gracefully
        expect(getByText("settings.subscription.daysRemaining")).toBeTruthy();
        expect(getByText("settings.subscription.daysCount")).toBeTruthy();
      });

      it("should handle very large days remaining", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: false,
            status: "cancelled",
            isCancelledButActive: true,
            daysRemaining: 999999, // Very large number
            currentPeriodEnd: "2999-12-31T23:59:59Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should handle large numbers gracefully
        expect(getByText("settings.subscription.daysRemaining")).toBeTruthy();
        expect(getByText("settings.subscription.daysCount")).toBeTruthy();
      });
    });

    describe("Conflicting State Combinations", () => {
      it("should handle subscription marked as both active and cancelled", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true, // Active
            status: "cancelled", // But cancelled status
            isCancelledButActive: false, // But not cancelled but active
            isInGracePeriod: false,
            currentPeriodEnd: "2025-06-15T12:00:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should prioritize isActive flag for isPremium logic
        expect(getByText("settings.subscription.monthlyPremium")).toBeTruthy();
        // Should show status as cancelled (raw status display)
        expect(getByText("settings.subscription.statusExpired")).toBeTruthy();
      });

      it("should handle grace period with conflicting active states", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: false, // Not active
            status: "past_due",
            isInGracePeriod: true, // But in grace period
            isCancelledButActive: false,
            currentPeriodEnd: "2025-03-01T10:00:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText, queryByText } = render(<SettingsScreen />);

        // Grace period users should be treated as premium (even if isActive is false)
        // Our isPremium logic needs to check isInGracePeriod
        expect(getByText("settings.subscription.yearlyPremium")).toBeTruthy();

        // Should show grace period status and warning
        expect(
          getByText("settings.subscription.statusGracePeriod"),
        ).toBeTruthy();
        expect(
          getByText("⚠️ settings.subscription.gracePeriodWarning"),
        ).toBeTruthy();

        // Grace period should be treated as premium (no upgrade section)
        expect(queryByText("upgrade.title")).toBeNull();
      });

      it("should handle subscription with multiple end-of-period flags", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "active",
            cancelAtPeriodEnd: true, // Set to cancel
            isCancelledButActive: true, // Also marked as cancelled but active
            isInGracePeriod: false,
            currentPeriodEnd: "2025-04-20T18:30:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText, queryByText } = render(<SettingsScreen />);

        // Should still be treated as premium
        expect(getByText("settings.subscription.monthlyPremium")).toBeTruthy();

        // Should show expiresOn label (either flag should trigger this)
        expect(getByText("settings.subscription.expiresOn")).toBeTruthy();

        // Should not show manage button (cancelled subscriptions can't be managed)
        expect(
          queryByText("settings.subscription.manageSubscription"),
        ).toBeNull();
      });
    });

    describe("Component Error Recovery", () => {
      it("should not crash when subscription object has circular references", () => {
        const circularData = createSubscriptionData({
          plan: "monthly",
          isActive: true,
          status: "active",
        });

        // Create circular reference (this shouldn't happen but test resilience)
        (circularData as any).self = circularData;

        mockUseQuery.mockReturnValue({
          data: circularData,
          isLoading: false,
          refetch: jest.fn(),
        });

        // Should not throw an error during rendering
        expect(() => render(<SettingsScreen />)).not.toThrow();
      });

      it("should handle extremely long strings in subscription data", () => {
        const veryLongString = "x".repeat(10000); // 10k character string

        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: veryLongString as any, // Very long status string
            currentPeriodEnd: "2025-12-31T23:59:59Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should handle long strings without crashing
        expect(getByText("settings.subscription.currentPlan")).toBeTruthy();
        // Should display the long string (React Native will handle text wrapping)
        expect(getByText(veryLongString)).toBeTruthy();
      });
    });
  });
});
