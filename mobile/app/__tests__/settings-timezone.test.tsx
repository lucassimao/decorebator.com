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

describe("SettingsScreen - Timezone", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Default mock for useQuery (subscription data)
    mockUseQuery.mockReturnValue({
      data: createSubscriptionData(),
      isLoading: false,
      refetch: jest.fn(),
    });
  });

  describe("Timezone and Date Formatting", () => {
    it("should display formatted subscription dates with timezone conversion", () => {
      const { formatDate: mockFormatDate } = jest.requireMock(
        "@/utils/dateUtils",
      ) as {
        formatDate: jest.Mock;
      };

      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "monthly",
          isActive: true,
          currentPeriodEnd: "2025-01-15T23:59:59Z", // UTC date
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);

      // Verify formatDate was called with the UTC date
      expect(mockFormatDate).toHaveBeenCalledWith("2025-01-15T23:59:59Z");

      // Should show the formatted date (mocked to return "Formatted: ...")
      expect(getByTestId("subscription-date")).toBeTruthy();
    });

    it("should handle subscription dates for cancelled but active subscriptions", () => {
      const { formatDate: mockFormatDate } = jest.requireMock(
        "@/utils/dateUtils",
      ) as {
        formatDate: jest.Mock;
      };

      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "annual",
          isActive: false,
          isCancelledButActive: true,
          status: "cancelled",
          currentPeriodEnd: "2025-03-20T12:30:00Z",
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);

      // Should show "expires on" label for cancelled subscriptions
      expect(getByTestId("expires-on-label")).toBeTruthy();

      // Should format the expiration date
      expect(mockFormatDate).toHaveBeenCalledWith("2025-03-20T12:30:00Z");
    });

    it("should handle subscription dates for grace period users", () => {
      const { formatDate: mockFormatDate } = jest.requireMock(
        "@/utils/dateUtils",
      ) as {
        formatDate: jest.Mock;
      };

      mockUseQuery.mockReturnValue({
        data: createSubscriptionData({
          plan: "monthly",
          isActive: true,
          isInGracePeriod: true,
          status: "past_due",
          currentPeriodEnd: "2025-02-01T08:15:30Z",
        }),
        isLoading: false,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);

      // Should show grace period warning (includes emoji prefix)
      expect(getByTestId("grace-period-warning")).toBeTruthy();

      // Should format the grace period end date
      expect(mockFormatDate).toHaveBeenCalledWith("2025-02-01T08:15:30Z");
    });
  });
});
