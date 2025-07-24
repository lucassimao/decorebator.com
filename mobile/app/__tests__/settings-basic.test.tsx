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

describe("SettingsScreen - Basic Rendering", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Default mock for useQuery (subscription data)
    mockUseQuery.mockReturnValue({
      data: createSubscriptionData(),
      isLoading: false,
      refetch: jest.fn(),
    });
  });

  describe("Basic Rendering", () => {
    it("renders without crashing", () => {
      const { getByTestId } = render(<SettingsScreen />);
      expect(getByTestId("settings-title")).toBeTruthy();
    });

    it("shows loading state when subscription is loading", () => {
      mockUseQuery.mockReturnValue({
        data: null,
        isLoading: true,
        refetch: jest.fn(),
      });

      const { getByTestId } = render(<SettingsScreen />);
      // Should show current plan section header
      expect(getByTestId("current-plan-section")).toBeTruthy();
    });
  });
});
