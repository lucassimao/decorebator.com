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

describe("SettingsScreen - Dates and Labels", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Default mock for useQuery (subscription data)
    mockUseQuery.mockReturnValue({
      data: createSubscriptionData(),
      isLoading: false,
      refetch: jest.fn(),
    });
  });

  describe("Label and Date Logic (Renewal vs Expiration)", () => {
    describe("Date Label Selection Logic", () => {
      it("should show 'renewsOn' label for active, non-cancelled subscriptions", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "active",
            isInGracePeriod: false,
            cancelAtPeriodEnd: false,
            isCancelledButActive: false,
            currentPeriodEnd: "2025-12-31T23:59:59Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText, queryByText } = render(<SettingsScreen />);

        // Should show renewsOn label
        expect(getByText("settings.subscription.renewsOn")).toBeTruthy();

        // Should NOT show other date labels
        expect(queryByText("settings.subscription.expiresOn")).toBeNull();
        expect(
          queryByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeNull();
      });

      it("should show 'expiresOn' label for subscriptions set to cancel at period end", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: true,
            status: "active",
            isInGracePeriod: false,
            cancelAtPeriodEnd: true,
            isCancelledButActive: false,
            currentPeriodEnd: "2025-06-30T23:59:59Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText, queryByText } = render(<SettingsScreen />);

        // Should show expiresOn label
        expect(getByText("settings.subscription.expiresOn")).toBeTruthy();

        // Should NOT show other date labels
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
        expect(
          queryByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeNull();
      });

      it("should show 'expiresOn' label for cancelled but still active subscriptions", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: false,
            status: "cancelled",
            isInGracePeriod: false,
            cancelAtPeriodEnd: false,
            isCancelledButActive: true,
            daysRemaining: 20,
            currentPeriodEnd: "2025-03-15T10:30:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText, queryByText } = render(<SettingsScreen />);

        // Should show expiresOn label
        expect(getByText("settings.subscription.expiresOn")).toBeTruthy();

        // Should NOT show other date labels
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
        expect(
          queryByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeNull();
      });

      it("should show 'gracePeriodEndsOn' label for subscriptions in grace period", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: true,
            status: "past_due",
            isInGracePeriod: true,
            cancelAtPeriodEnd: false,
            isCancelledButActive: false,
            currentPeriodEnd: "2025-02-28T15:45:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText, queryByText } = render(<SettingsScreen />);

        // Should show gracePeriodEndsOn label
        expect(
          getByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeTruthy();

        // Should NOT show other date labels
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
        expect(queryByText("settings.subscription.expiresOn")).toBeNull();
      });

      it("should prioritize grace period label over cancelled/expiry labels", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "past_due",
            isInGracePeriod: true,
            cancelAtPeriodEnd: true, // This would normally show "expiresOn"
            isCancelledButActive: false,
            currentPeriodEnd: "2025-01-31T12:00:00Z",
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText, queryByText } = render(<SettingsScreen />);

        // Grace period should take precedence
        expect(
          getByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeTruthy();
        expect(queryByText("settings.subscription.expiresOn")).toBeNull();
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
      });
    });

    describe("Date Display with Correct Labels", () => {
      it("should display correct date with renewsOn label for regular active subscriptions", () => {
        const testDate = "2025-08-15T14:30:00Z";
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
            currentPeriodEnd: testDate,
            cancelAtPeriodEnd: false,
            isCancelledButActive: false,
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should show renewsOn label with correctly formatted date
        expect(getByText("settings.subscription.renewsOn")).toBeTruthy();
        expect(mockFormatDate).toHaveBeenCalledWith(testDate);
        expect(getByText(`Formatted: ${testDate}`)).toBeTruthy();
      });

      it("should display correct date with expiresOn label for cancelled subscriptions", () => {
        const testDate = "2025-04-10T09:15:30Z";
        const { formatDate: mockFormatDate } = jest.requireMock(
          "@/utils/dateUtils",
        ) as {
          formatDate: jest.Mock;
        };

        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: false,
            status: "cancelled",
            currentPeriodEnd: testDate,
            isCancelledButActive: true,
            daysRemaining: 5,
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should show expiresOn label with correctly formatted date
        expect(getByText("settings.subscription.expiresOn")).toBeTruthy();
        expect(mockFormatDate).toHaveBeenCalledWith(testDate);
        expect(getByText(`Formatted: ${testDate}`)).toBeTruthy();
      });

      it("should display correct date with gracePeriodEndsOn label for grace period subscriptions", () => {
        const testDate = "2025-05-22T18:00:00Z";
        const { formatDate: mockFormatDate } = jest.requireMock(
          "@/utils/dateUtils",
        ) as {
          formatDate: jest.Mock;
        };

        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "past_due",
            currentPeriodEnd: testDate,
            isInGracePeriod: true,
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should show gracePeriodEndsOn label with correctly formatted date
        expect(
          getByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeTruthy();
        expect(mockFormatDate).toHaveBeenCalledWith(testDate);
        expect(getByText(`Formatted: ${testDate}`)).toBeTruthy();
      });
    });

    describe("Date Display Conditions", () => {
      it("should not show any date labels for free plan users", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "free",
            isActive: true,
            status: "active",
            currentPeriodEnd: undefined, // Free users never have currentPeriodEnd in production
            cancelAtPeriodEnd: false,
            isCancelledButActive: false,
            isInGracePeriod: false,
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);

        // Should not show any date-related labels for free users
        // because they don't have currentPeriodEnd in the backend
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
        expect(queryByText("settings.subscription.expiresOn")).toBeNull();
        expect(
          queryByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeNull();
      });

      it("should not show date labels for expired subscriptions without remaining access", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: false,
            status: "cancelled",
            currentPeriodEnd: "2024-12-31T23:59:59Z", // Past date
            isInGracePeriod: false,
            isCancelledButActive: false,
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);

        // Should not show any date labels for fully expired subscriptions
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
        expect(queryByText("settings.subscription.expiresOn")).toBeNull();
        expect(
          queryByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeNull();
      });

      it("should not show date labels when currentPeriodEnd is missing", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "annual",
            isActive: true,
            status: "active",
            currentPeriodEnd: undefined as any, // Missing date
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByText } = render(<SettingsScreen />);

        // Should not show any date labels when date is missing
        expect(queryByText("settings.subscription.renewsOn")).toBeNull();
        expect(queryByText("settings.subscription.expiresOn")).toBeNull();
        expect(
          queryByText("settings.subscription.gracePeriodEndsOn"),
        ).toBeNull();
      });

      it("should show date labels only when subscription meets all display criteria", () => {
        mockUseQuery.mockReturnValue({
          data: createSubscriptionData({
            plan: "monthly",
            isActive: true,
            status: "active",
            currentPeriodEnd: "2025-11-30T20:00:00Z",
            cancelAtPeriodEnd: false,
            isCancelledButActive: false,
            isInGracePeriod: false,
          }),
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByText } = render(<SettingsScreen />);

        // Should meet all criteria: has currentPeriodEnd AND (isActive OR isCancelledButActive OR isInGracePeriod)
        expect(getByText("settings.subscription.renewsOn")).toBeTruthy();
        expect(getByText("Formatted: 2025-11-30T20:00:00Z")).toBeTruthy();
      });
    });
  });
});
