import React from "react";
import { act, fireEvent, render } from "@testing-library/react-native";
import { AccessibilityInfo } from "react-native";

import NativeIAPPaywall from "@/components/NativeIAPPaywall";
import type { NativeIAPState } from "@/hooks/useNativeIAP";

const mockSelectPlan = jest.fn();
const mockPurchase = jest.fn();
const mockRestore = jest.fn();
const mockRetry = jest.fn();
let mockReducedMotion = false;
const mockWithDelay = jest.fn((_delay: number, value: number) => value);

const plan = {
  billingPeriod: "annual",
  productId: "premium.annual",
  title: "Premium Annual",
  displayPrice: "$39.99",
  period: { unit: "year", value: 1 },
  offer: null,
  offerDisplay: null,
  product: {},
} as any;

let mockState: NativeIAPState;

jest.mock("react-i18next", () => ({
  useTranslation: () => ({
    t: (key: string, values?: Record<string, unknown>) => {
      if (key.endsWith("period.year")) return `${values?.count} year`;
      if (key.endsWith("renewal"))
        return `${values?.price} per ${values?.period}`;
      if (key.endsWith("failure")) return `Failed: ${values?.code}`;
      return key;
    },
  }),
}));

jest.mock("react-native-reanimated", () => ({
  __esModule: true,
  default: { View: "Animated.View" },
  useAnimatedStyle: (fn: () => unknown) => fn(),
  useReducedMotion: () => mockReducedMotion,
  useSharedValue: (value: number) => ({ value }),
  withDelay: (delay: number, value: number) => mockWithDelay(delay, value),
  withSpring: (value: number) => value,
}));

jest.mock("@/hooks/useNativeIAP", () => ({
  useNativeIAP: () => ({
    state: mockState,
    store: "apple",
    selectPlan: mockSelectPlan,
    purchase: mockPurchase,
    restore: mockRestore,
    retry: mockRetry,
  }),
}));

function readyState(overrides: Partial<NativeIAPState> = {}): NativeIAPState {
  return {
    status: "ready",
    plans: [plan],
    selectedProductId: null,
    errorCode: null,
    retryable: true,
    restoreStatus: null,
    entitlementOrigin: null,
    ...overrides,
  };
}

describe("NativeIAPPaywall", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockReducedMotion = false;
    mockState = readyState();
  });

  it("requires explicit plan selection before purchase", () => {
    const view = render(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={jest.fn()} />,
    );

    const purchaseButton = view.getByTestId("native-iap-purchase");
    expect(purchaseButton.props.accessibilityState.disabled).toBe(true);
    fireEvent.press(view.getByLabelText("Premium Annual, $39.99, 1 year"));
    expect(mockSelectPlan).toHaveBeenCalledWith("premium.annual");
    expect(mockPurchase).not.toHaveBeenCalled();
  });

  it("pairs the selected store price with disclosure and one purchase action", () => {
    mockState = readyState({ selectedProductId: "premium.annual" });
    const view = render(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={jest.fn()} />,
    );

    expect(view.getAllByText("$39.99")).toHaveLength(1);
    expect(view.getByText("$39.99 per 1 year")).toBeTruthy();
    fireEvent.press(view.getByTestId("native-iap-purchase"));
    expect(mockPurchase).toHaveBeenCalledTimes(1);
  });

  it("locks plan and restore actions while a purchase is pending", () => {
    mockState = readyState({
      status: "pending",
      selectedProductId: "premium.annual",
    });
    const view = render(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={jest.fn()} />,
    );

    expect(
      view.getByLabelText("Premium Annual, $39.99, 1 year").props
        .accessibilityState.disabled,
    ).toBe(true);
    expect(
      view.getByText("settings.subscription.nativeIap.pendingMessage"),
    ).toBeTruthy();
    expect(
      view.getByTestId("native-iap-restore").props.accessibilityState.disabled,
    ).toBe(true);
  });

  it("never exposes stale prices when catalog loading or recovery fails", () => {
    mockState = readyState({ status: "loading", plans: [] });
    const view = render(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={jest.fn()} />,
    );
    expect(view.queryByText("$39.99")).toBeNull();
    expect(view.getByTestId("native-iap-loading")).toBeTruthy();

    mockState = readyState({
      status: "failed",
      plans: [],
      errorCode: "store_unavailable",
    });
    view.rerender(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={jest.fn()} />,
    );
    fireEvent.press(view.getByText("settings.subscription.nativeIap.retry"));
    fireEvent.press(view.getByText("settings.subscription.restorePurchases"));
    expect(mockRetry).toHaveBeenCalledTimes(1);
    expect(mockRestore).toHaveBeenCalledTimes(1);
    expect(view.queryByText("$39.99")).toBeNull();
  });

  it("announces backend-granted entitlement once", async () => {
    mockState = readyState({
      status: "entitled",
      selectedProductId: "premium.annual",
      entitlementOrigin: "purchase",
    });
    const onEntitled = jest.fn();
    const view = render(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={onEntitled} />,
    );
    await act(async () => {});
    view.rerender(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={onEntitled} />,
    );
    expect(onEntitled).toHaveBeenCalledTimes(1);
    expect(AccessibilityInfo.announceForAccessibility).toHaveBeenCalledWith(
      "settings.subscription.nativeIap.entitled",
    );
  });

  it("keeps an existing entitlement visible without a false activation callback", async () => {
    mockState = readyState({
      status: "entitled",
      selectedProductId: "premium.annual",
      entitlementOrigin: null,
    });
    const onEntitled = jest.fn();
    const view = render(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={onEntitled} />,
    );
    await act(async () => {});
    expect(onEntitled).not.toHaveBeenCalled();
    expect(
      view.getByText("settings.subscription.nativeIap.entitledMessage"),
    ).toBeTruthy();
  });

  it("renders the shelf in its static state when reduced motion is enabled", () => {
    mockReducedMotion = true;
    render(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={jest.fn()} />,
    );
    expect(mockWithDelay).not.toHaveBeenCalled();
  });

  it("keeps the title, price, and legal block scalable inside one scroll surface", () => {
    mockState = readyState({ selectedProductId: "premium.annual" });
    const view = render(
      <NativeIAPPaywall visible onClose={jest.fn()} onEntitled={jest.fn()} />,
    );
    expect(view.getByTestId("native-iap-scroll")).toBeTruthy();
    for (const text of [
      "Premium Annual",
      "$39.99",
      "settings.subscription.nativeIap.legal",
    ]) {
      expect(view.getByText(text).props.allowFontScaling).toBe(true);
      expect(view.getByText(text).props.numberOfLines).toBeUndefined();
    }
  });
});
