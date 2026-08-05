import React from "react";
import { fireEvent, render } from "@testing-library/react-native";
import SettingsScreen from "../settings";
import { useQuery } from "@tanstack/react-query";
import {
  createSubscriptionData,
  SUBSCRIPTION_SCENARIOS,
} from "./utils/test-helpers";

jest.mock("@/components/NativeIAPPaywall", () => {
  const { Text } = jest.requireMock("react-native");
  function MockNativeIAPPaywall() {
    return <Text>settings.subscription.nativeIap.title</Text>;
  }
  return MockNativeIAPPaywall;
});

// Mock API modules that depend on expo-secure-store
jest.mock("@/api/subscriptions", () => ({
  getSubscriptionStatus: jest.fn(),
  createCheckoutSession: jest.fn(),
  openNativeSubscriptionManagement: jest.fn(),
}));

jest.mock("@/api/users", () => ({
  getAuthorization: jest.fn(() => null),
  getProfile: jest.fn(),
  sigout: jest.fn(),
  update: jest.fn(),
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

        const { getByTestId } = render(<SettingsScreen />);
        expect(getByTestId("upgrade-title")).toBeTruthy();
        expect(getByTestId("upgrade-subtitle")).toBeTruthy();
        expect(getByTestId("premium-features-title")).toBeTruthy();
      });

      it("opens the single native-store paywall without provider routing", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.FREE_USER,
          isLoading: false,
          refetch: jest.fn(),
        });

        const view = render(<SettingsScreen />);
        fireEvent.press(view.getByTestId("open-native-iap-paywall"));
        expect(
          view.getByText("settings.subscription.nativeIap.title"),
        ).toBeTruthy();
      });

      it("should hide upgrade section for active premium users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.ACTIVE_MONTHLY,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByTestId } = render(<SettingsScreen />);
        expect(queryByTestId("upgrade-title")).toBeNull();
        expect(queryByTestId("upgrade-subtitle")).toBeNull();
        expect(queryByTestId("premium-features-title")).toBeNull();
      });

      it("should hide upgrade section for cancelled but still active users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.CANCELLED_BUT_ACTIVE,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByTestId } = render(<SettingsScreen />);
        expect(queryByTestId("upgrade-title")).toBeNull();
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

        const { queryByTestId } = render(<SettingsScreen />);
        expect(queryByTestId("upgrade-title")).toBeNull();
      });

      it("should show upgrade section for expired premium users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.EXPIRED_CANCELLED,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByTestId } = render(<SettingsScreen />);
        expect(getByTestId("upgrade-title")).toBeTruthy();
      });
    });

    describe("Free Plan Limitations Display", () => {
      it("should show free plan limitations for free users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.FREE_USER,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByTestId } = render(<SettingsScreen />);
        expect(getByTestId("free-plan-limit")).toBeTruthy();
      });

      it("should hide free plan limitations for premium users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.ACTIVE_MONTHLY,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { queryByTestId } = render(<SettingsScreen />);
        expect(queryByTestId("free-plan-limit")).toBeNull();
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

        const { getByTestId } = render(<SettingsScreen />);
        expect(getByTestId("manage-subscription-button")).toBeTruthy();
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

        const { queryByTestId } = render(<SettingsScreen />);
        expect(queryByTestId("manage-subscription-button")).toBeNull();
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

        const { queryByTestId } = render(<SettingsScreen />);
        expect(queryByTestId("manage-subscription-button")).toBeNull();
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

        const { getByTestId } = render(<SettingsScreen />);
        // Users in grace period should still be able to manage their subscription to fix payment issues
        expect(getByTestId("manage-subscription-button")).toBeTruthy();
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

        const { getByTestId } = render(<SettingsScreen />);
        expect(getByTestId("renews-on-label")).toBeTruthy();
        expect(getByTestId("subscription-date")).toBeTruthy();
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

        const { getByTestId } = render(<SettingsScreen />);
        expect(getByTestId("expires-on-label")).toBeTruthy();
        expect(getByTestId("subscription-date")).toBeTruthy();
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

        const { getByTestId } = render(<SettingsScreen />);
        expect(getByTestId("grace-period-ends-on")).toBeTruthy();
        expect(getByTestId("subscription-date")).toBeTruthy();
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

        const { queryByTestId } = render(<SettingsScreen />);
        expect(queryByTestId("renews-on-label")).toBeNull();
        expect(queryByTestId("expires-on-label")).toBeNull();
        expect(queryByTestId("subscription-date")).toBeNull();
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

        const { queryByTestId } = render(<SettingsScreen />);
        expect(queryByTestId("renews-on-label")).toBeNull();
      });
    });

    describe("Premium Badge Icon Display", () => {
      it("should show premium icon for premium users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.ACTIVE_MONTHLY,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByTestId } = render(<SettingsScreen />);
        // Premium users should see the premium plan text, indicating premium icon is shown
        expect(getByTestId("monthly-premium-name")).toBeTruthy();
      });

      it("should show lock icon for free users", () => {
        mockUseQuery.mockReturnValue({
          data: SUBSCRIPTION_SCENARIOS.FREE_USER,
          isLoading: false,
          refetch: jest.fn(),
        });

        const { getByTestId } = render(<SettingsScreen />);
        // Free users should see the free plan text, indicating lock icon is shown
        expect(getByTestId("free-plan-name")).toBeTruthy();
      });
    });
  });
});
