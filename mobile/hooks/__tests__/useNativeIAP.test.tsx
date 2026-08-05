import { act, renderHook, waitFor } from "@testing-library/react-native";
import { ErrorCode, type Purchase } from "expo-iap";
import { AppState, Linking } from "react-native";

import {
  getMobileIAPContext,
  restoreMobileIAPPurchases,
  verifyApplePurchase,
} from "@/api/iap";
import { useNativeIAP } from "@/hooks/useNativeIAP";

const mockFetchProducts = jest.fn().mockResolvedValue(undefined);
const mockRequestPurchase = jest.fn().mockResolvedValue(undefined);
const mockFinishTransaction = jest.fn().mockResolvedValue(undefined);
const mockReconnect = jest.fn().mockResolvedValue(true);
const mockGetAvailablePurchases = jest.fn().mockResolvedValue([]);
const mockIsEligibleForIntroOfferIOS = jest.fn().mockResolvedValue(false);
let iapCallbacks: Record<string, (...args: any[]) => void> = {};
let mockSubscriptions: unknown[] = [];
let mockCurrentStore: "apple" | "google" = "apple";
let mockAppStateListener: (state: string) => void;
let mockLinkListener: () => void;
const mockCapture = jest.fn();
const mockPosthog = { capture: mockCapture };

jest.mock("posthog-react-native", () => ({
  usePostHog: () => mockPosthog,
}));

jest.mock("expo-iap", () => ({
  ErrorCode: {
    UserCancelled: "user-cancelled",
    Pending: "pending",
    DeferredPayment: "deferred-payment",
    Unknown: "unknown",
    ItemUnavailable: "item-unavailable",
    SkuNotFound: "sku-not-found",
    FeatureNotSupported: "feature-not-supported",
  },
  getAvailablePurchases: (...args: any[]) => mockGetAvailablePurchases(...args),
  isEligibleForIntroOfferIOS: (...args: any[]) =>
    mockIsEligibleForIntroOfferIOS(...args),
  useIAP: (options: Record<string, (...args: any[]) => void>) => {
    iapCallbacks = options;
    return {
      connected: true,
      subscriptions: mockSubscriptions,
      fetchProducts: mockFetchProducts,
      requestPurchase: mockRequestPurchase,
      finishTransaction: mockFinishTransaction,
      reconnect: mockReconnect,
    };
  },
}));

jest.mock("@/api/iap", () => ({
  getMobileIAPContext: jest.fn(),
  restoreMobileIAPPurchases: jest.fn(),
  verifyApplePurchase: jest.fn(),
  verifyGooglePurchase: jest.fn(),
}));

jest.mock("@/utils/iapProducts", () => {
  const actual = jest.requireActual("@/utils/iapProducts");
  return { ...actual, currentIAPStore: () => mockCurrentStore };
});

const appleProduct = {
  id: "premium.annual",
  type: "subs",
  platform: "ios",
  title: "Premium Annual",
  displayPrice: "$39.99",
  subscriptionPeriodUnitIOS: "year",
  subscriptionPeriodNumberIOS: "1",
  subscriptionOffers: [],
};

const googleProduct = {
  id: "premium.monthly",
  type: "subs",
  platform: "android",
  title: "Premium Monthly",
  displayPrice: "$4.99",
  subscriptionOffers: [
    {
      id: "base-monthly",
      type: "introductory",
      displayPrice: "$4.99",
      price: 4.99,
      offerTokenAndroid: "offer-token",
      period: { unit: "month", value: 1 },
      pricingPhasesAndroid: {
        pricingPhaseList: [
          {
            billingCycleCount: 0,
            billingPeriod: "P1M",
            formattedPrice: "$4.99",
            priceAmountMicros: "4990000",
            priceCurrencyCode: "USD",
            recurrenceMode: 1,
          },
        ],
      },
    },
  ],
};

const baseContext = {
  products: [
    {
      store: "apple" as const,
      productId: "premium.annual",
      entitlement: "premium" as const,
      billingPeriod: "annual" as const,
    },
  ],
  purchaseContext: {
    store: "apple" as const,
    appleAppAccountToken: "00000000-0000-4000-8000-000000000001",
  },
  currentEntitlement: null,
  pending: null,
  error: null,
  restore: null,
  serverTime: "2026-08-05T12:00:00Z",
};

const purchase = {
  id: "purchase-1",
  productId: "premium.annual",
  transactionId: "transaction-1",
  purchaseState: "purchased",
  purchaseToken: "apple-jws",
} as Purchase;

describe("useNativeIAP", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    iapCallbacks = {};
    mockCurrentStore = "apple";
    mockSubscriptions = [appleProduct];
    (AppState.addEventListener as jest.Mock).mockImplementation(
      (_event: string, listener: (next: string) => void) => {
        mockAppStateListener = listener;
        return { remove: jest.fn() };
      },
    );
    (Linking.addEventListener as jest.Mock).mockImplementation(
      (_event: string, listener: () => void) => {
        mockLinkListener = listener;
        return { remove: jest.fn() };
      },
    );
    (getMobileIAPContext as jest.Mock).mockResolvedValue(baseContext);
    (verifyApplePurchase as jest.Mock).mockResolvedValue({
      ...baseContext,
      currentEntitlement: {
        store: "apple",
        productId: "premium.annual",
        entitlement: "premium",
        status: "active",
        periodEnd: "2027-08-05T12:00:00Z",
        autoRenewEnabled: true,
      },
    });
  });

  it("loads the allowlisted store intersection without default selection", async () => {
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    expect(result.current.state.plans).toHaveLength(1);
    expect(result.current.state.selectedProductId).toBeNull();
    expect(mockFetchProducts).toHaveBeenCalledWith({
      skus: ["premium.annual"],
      type: "subs",
    });
  });

  it("binds the Apple account token and finishes only after backend access", async () => {
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    act(() => result.current.selectPlan("premium.annual"));
    await act(async () => result.current.purchase());
    expect(mockRequestPurchase).toHaveBeenCalledWith(
      expect.objectContaining({
        request: {
          apple: expect.objectContaining({
            sku: "premium.annual",
            appAccountToken: "00000000-0000-4000-8000-000000000001",
            andDangerouslyFinishTransactionAutomatically: false,
          }),
        },
      }),
    );

    await act(async () => iapCallbacks.onPurchaseSuccess(purchase));
    await waitFor(() => expect(result.current.state.status).toBe("entitled"));
    expect(verifyApplePurchase).toHaveBeenCalledWith("transaction-1");
    expect(mockFinishTransaction).toHaveBeenCalledWith({
      purchase,
      isConsumable: false,
    });
  });

  it("keeps pending purchases unfinished and locks the product", async () => {
    (verifyApplePurchase as jest.Mock).mockResolvedValueOnce({
      ...baseContext,
      pending: {
        store: "apple",
        productId: "premium.annual",
        startedAt: "2026-08-05T12:00:00Z",
      },
    });
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    await act(async () => iapCallbacks.onPurchaseSuccess(purchase));
    await waitFor(() => expect(result.current.state.status).toBe("pending"));
    expect(result.current.state.selectedProductId).toBe("premium.annual");
    expect(mockFinishTransaction).not.toHaveBeenCalled();
  });

  it("finishes and records a local pending purchase after foreground resolution", async () => {
    (verifyApplePurchase as jest.Mock).mockResolvedValueOnce({
      ...baseContext,
      pending: {
        store: "apple",
        productId: "premium.annual",
        startedAt: "2026-08-05T12:00:00Z",
      },
    });
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    await act(async () => iapCallbacks.onPurchaseSuccess(purchase));
    await waitFor(() => expect(result.current.state.status).toBe("pending"));
    (getMobileIAPContext as jest.Mock).mockResolvedValueOnce({
      ...baseContext,
      currentEntitlement: {
        store: "apple",
        productId: "premium.annual",
        entitlement: "premium",
        status: "active",
        periodEnd: "2027-08-05T12:00:00Z",
        autoRenewEnabled: true,
      },
    });
    act(() => mockAppStateListener("active"));
    await waitFor(() => expect(result.current.state.status).toBe("entitled"));
    expect(mockFinishTransaction).toHaveBeenCalledWith({
      purchase,
      isConsumable: false,
    });
    expect(mockCapture).toHaveBeenCalledWith(
      "purchase_succeeded",
      expect.objectContaining({
        store: "apple",
        productId: "premium.annual",
        billingPeriod: "annual",
      }),
    );
  });

  it("does not let a native pending flag mask backend rejection", async () => {
    (verifyApplePurchase as jest.Mock).mockResolvedValueOnce({
      ...baseContext,
      error: {
        code: "invalid_purchase",
        messageKey: "subscription.iap.invalidPurchase",
        retryable: false,
      },
    });
    const pendingPurchase = { ...purchase, purchaseState: "pending" };
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    await act(async () => iapCallbacks.onPurchaseSuccess(pendingPurchase));
    await waitFor(() => expect(result.current.state.status).toBe("failed"));
    expect(result.current.state.errorCode).toBe("invalid_purchase");
    expect(mockFinishTransaction).not.toHaveBeenCalled();
  });

  it("keeps granted access when finishing the store transaction is deferred", async () => {
    mockFinishTransaction.mockRejectedValueOnce(new Error("store offline"));
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    await act(async () => iapCallbacks.onPurchaseSuccess(purchase));
    await waitFor(() => expect(result.current.state.status).toBe("entitled"));
    expect(result.current.state.errorCode).toBeNull();
  });

  it("coalesces duplicate purchase callbacks while verification is active", async () => {
    let resolveVerification!: (value: typeof baseContext) => void;
    (verifyApplePurchase as jest.Mock).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveVerification = resolve;
        }),
    );
    renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(mockFetchProducts).toHaveBeenCalled());
    act(() => {
      iapCallbacks.onPurchaseSuccess(purchase);
      iapCallbacks.onPurchaseSuccess(purchase);
    });
    expect(verifyApplePurchase).toHaveBeenCalledTimes(1);
    await act(async () => resolveVerification(baseContext));
  });

  it("queues an unfinished transaction replay until backend context loads", async () => {
    let resolveContext!: (value: typeof baseContext) => void;
    (getMobileIAPContext as jest.Mock).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveContext = resolve;
        }),
    );
    const { result } = renderHook(() => useNativeIAP(true));
    act(() => iapCallbacks.onPurchaseSuccess(purchase));
    expect(verifyApplePurchase).not.toHaveBeenCalled();
    await act(async () => resolveContext(baseContext));
    await waitFor(() => expect(verifyApplePurchase).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(result.current.state.status).toBe("entitled"));
    expect(mockFinishTransaction).toHaveBeenCalledWith({
      purchase,
      isConsumable: false,
    });
  });

  it("returns to ready after native cancellation without failure analytics", async () => {
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    act(() => result.current.selectPlan("premium.annual"));
    await act(async () => result.current.purchase());
    act(() =>
      iapCallbacks.onPurchaseError({
        code: ErrorCode.UserCancelled,
        productId: "premium.annual",
      }),
    );
    expect(result.current.state.status).toBe("ready");
    expect(mockCapture).not.toHaveBeenCalledWith(
      "purchase_failed",
      expect.anything(),
    );
  });

  it("treats a rejected native cancellation as cancellation exactly once", async () => {
    mockRequestPurchase.mockRejectedValueOnce({
      code: ErrorCode.UserCancelled,
      message: "User cancelled",
    });
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    act(() => result.current.selectPlan("premium.annual"));
    await act(async () => result.current.purchase());
    expect(result.current.state.status).toBe("ready");
    expect(mockCapture).not.toHaveBeenCalledWith(
      "purchase_failed",
      expect.anything(),
    );
    act(() =>
      iapCallbacks.onPurchaseError({
        code: ErrorCode.UserCancelled,
        productId: "premium.annual",
      }),
    );
    expect(result.current.state.status).toBe("ready");
  });

  it("records deferred payment as pending without failure analytics", async () => {
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    act(() => result.current.selectPlan("premium.annual"));
    await act(async () => result.current.purchase());
    act(() =>
      iapCallbacks.onPurchaseError({
        code: ErrorCode.DeferredPayment,
        productId: "premium.annual",
      }),
    );
    expect(result.current.state.status).toBe("pending");
    expect(mockCapture).toHaveBeenCalledWith(
      "purchase_pending",
      expect.objectContaining({
        store: "apple",
        productId: "premium.annual",
        billingPeriod: "annual",
      }),
    );
    expect(mockCapture).not.toHaveBeenCalledWith(
      "purchase_failed",
      expect.anything(),
    );
  });

  it("binds Google purchases to the server account and selected offer", async () => {
    mockCurrentStore = "google";
    mockSubscriptions = [googleProduct];
    (getMobileIAPContext as jest.Mock).mockResolvedValueOnce({
      ...baseContext,
      products: [
        {
          store: "google",
          productId: "premium.monthly",
          entitlement: "premium",
          billingPeriod: "monthly",
        },
      ],
      purchaseContext: {
        store: "google",
        googleObfuscatedAccountId: "server-account-hash",
      },
    });
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    act(() => result.current.selectPlan("premium.monthly"));
    await act(async () => result.current.purchase());
    expect(mockRequestPurchase).toHaveBeenCalledWith({
      type: "subs",
      request: {
        google: {
          skus: ["premium.monthly"],
          obfuscatedAccountId: "server-account-hash",
          subscriptionOffers: [
            { sku: "premium.monthly", offerToken: "offer-token" },
          ],
        },
      },
    });
  });

  it("refreshes authoritative state on foreground and deep-link entry", async () => {
    renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(getMobileIAPContext).toHaveBeenCalledTimes(1));
    act(() => mockAppStateListener("background"));
    expect(getMobileIAPContext).toHaveBeenCalledTimes(1);
    act(() => mockAppStateListener("active"));
    await waitFor(() => expect(getMobileIAPContext).toHaveBeenCalledTimes(2));
    act(() => mockLinkListener());
    await waitFor(() => expect(getMobileIAPContext).toHaveBeenCalledTimes(3));
    await act(async () => {});
  });

  it("suppresses foreground refresh while the native purchase is in flight", async () => {
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    act(() => result.current.selectPlan("premium.annual"));
    await act(async () => result.current.purchase());
    expect(result.current.state.status).toBe("purchasing");
    act(() => mockAppStateListener("active"));
    expect(getMobileIAPContext).toHaveBeenCalledTimes(1);
  });

  it("shows an Apple introductory offer only after StoreKit eligibility", async () => {
    mockSubscriptions = [
      {
        ...appleProduct,
        subscriptionGroupIdIOS: "premium-group",
        subscriptionOffers: [
          {
            id: "intro",
            type: "introductory",
            displayPrice: "$0.00",
            price: 0,
            period: { unit: "week", value: 1 },
          },
        ],
      },
    ];
    mockIsEligibleForIntroOfferIOS.mockResolvedValueOnce(true);
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() =>
      expect(mockIsEligibleForIntroOfferIOS).toHaveBeenCalledWith(
        "premium-group",
      ),
    );
    await waitFor(() =>
      expect(result.current.state.plans[0]?.offerDisplay?.displayPrice).toBe(
        "$0.00",
      ),
    );
    expect(result.current.state.plans[0].period).toEqual({
      unit: "year",
      value: 1,
    });
  });

  it("finishes only backend-restored evidence indexes", async () => {
    mockGetAvailablePurchases.mockResolvedValueOnce([purchase]);
    (restoreMobileIAPPurchases as jest.Mock).mockResolvedValueOnce({
      ...baseContext,
      restore: {
        status: "restored",
        restoredPurchases: 1,
        unchangedPurchases: 0,
        rejectedPurchases: 0,
        retryablePurchases: 0,
        outcomes: [{ index: 0, status: "restored" }],
      },
      currentEntitlement: {
        store: "apple",
        productId: "premium.annual",
        entitlement: "premium",
        status: "active",
        periodEnd: "2027-08-05T12:00:00Z",
        autoRenewEnabled: true,
      },
    });
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    await act(async () => result.current.restore());
    expect(restoreMobileIAPPurchases).toHaveBeenCalledWith("apple", [
      "transaction-1",
    ]);
    expect(mockFinishTransaction).toHaveBeenCalledWith({
      purchase,
      isConsumable: false,
    });
    expect(result.current.state.status).toBe("entitled");
  });

  it("does not finish rejected restore evidence", async () => {
    mockGetAvailablePurchases.mockResolvedValueOnce([purchase]);
    (restoreMobileIAPPurchases as jest.Mock).mockResolvedValueOnce({
      ...baseContext,
      restore: {
        status: "partial",
        restoredPurchases: 0,
        unchangedPurchases: 0,
        rejectedPurchases: 1,
        retryablePurchases: 0,
        outcomes: [{ index: 0, status: "rejected" }],
      },
    });
    const { result } = renderHook(() => useNativeIAP(true));
    await waitFor(() => expect(result.current.state.status).toBe("ready"));
    await act(async () => result.current.restore());
    expect(mockFinishTransaction).not.toHaveBeenCalled();
    expect(result.current.state.status).toBe("ready");
  });
});
