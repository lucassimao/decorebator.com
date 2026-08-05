import { useCallback, useEffect, useRef, useState } from "react";
import { AppState, Linking } from "react-native";
import {
  ErrorCode,
  getAvailablePurchases,
  isEligibleForIntroOfferIOS,
  type Purchase,
  useIAP,
} from "expo-iap";
import { usePostHog } from "posthog-react-native";

import {
  getMobileIAPContext,
  restoreMobileIAPPurchases,
  type IAPStore,
  type MobileIAPResponse,
  verifyApplePurchase,
  verifyGooglePurchase,
} from "@/api/iap";
import {
  ACTIVATION_EVENT_NAMES,
  captureActivationEvent,
} from "@/utils/activationEvents";
import {
  currentIAPStore,
  intersectStorePlans,
  type NativeIAPPlan,
} from "@/utils/iapProducts";

export type NativeIAPStatus =
  | "loading"
  | "ready"
  | "unavailable"
  | "empty"
  | "purchasing"
  | "pending"
  | "entitled"
  | "failed"
  | "restoring";

export interface NativeIAPState {
  status: NativeIAPStatus;
  plans: NativeIAPPlan[];
  selectedProductId: string | null;
  errorCode: string | null;
  retryable: boolean;
  restoreStatus: string | null;
  entitlementOrigin: "purchase" | "restore" | null;
}

const ACCESS_STATUSES = new Set(["active", "grace_period"]);

export function useNativeIAP(visible: boolean, source = "settings") {
  const posthog = usePostHog();
  const store = currentIAPStore();
  const [context, setContext] = useState<MobileIAPResponse | null>(null);
  const [state, setState] = useState<NativeIAPState>({
    status: "loading",
    plans: [],
    selectedProductId: null,
    errorCode: null,
    retryable: true,
    restoreStatus: null,
    entitlementOrigin: null,
  });
  const mountedRef = useRef(true);
  const contextRef = useRef(context);
  const plansRef = useRef(state.plans);
  const stateRef = useRef(state);
  const processingRef = useRef(new Set<string>());
  const completedPurchasesRef = useRef(new Set<string>());
  const queuedPurchasesRef = useRef(new Map<string, Purchase>());
  const pendingPurchasesRef = useRef(new Map<string, Purchase>());
  const purchaseAttemptRef = useRef<{
    productId: string;
    errorHandled: boolean;
  } | null>(null);
  const finishTransactionRef = useRef<
    (input: { purchase: Purchase; isConsumable: boolean }) => Promise<void>
  >(async () => undefined);
  const [productsFetched, setProductsFetched] = useState(false);
  const [eligibleAppleIntroGroups, setEligibleAppleIntroGroups] = useState<
    ReadonlySet<string>
  >(new Set());
  const eligibilityGenerationRef = useRef(0);

  useEffect(() => {
    contextRef.current = context;
  }, [context]);
  useEffect(() => {
    plansRef.current = state.plans;
  }, [state.plans]);
  useEffect(() => {
    stateRef.current = state;
  }, [state]);

  const fail = useCallback((errorCode: string, retryable = true) => {
    if (!mountedRef.current) return;
    stateRef.current = {
      ...stateRef.current,
      status: "failed",
      errorCode,
      retryable,
    };
    setState((current) => ({
      ...current,
      status: "failed",
      errorCode,
      retryable,
    }));
  }, []);

  const processPurchase = useCallback(
    async (purchase: Purchase) => {
      const activeStore = store;
      if (
        !activeStore ||
        processingRef.current.has(purchase.id) ||
        completedPurchasesRef.current.has(purchase.id)
      )
        return;
      if (purchaseAttemptRef.current?.productId === purchase.productId) {
        purchaseAttemptRef.current.errorHandled = true;
      }
      if (!contextRef.current) {
        queuedPurchasesRef.current.set(purchase.id, purchase);
        return;
      }
      const plan = plansRef.current.find(
        (candidate) => candidate.productId === purchase.productId,
      );
      const catalogProduct = contextRef.current.products.find(
        (candidate) =>
          candidate.store === activeStore &&
          candidate.productId === purchase.productId,
      );
      const analytics = plan
        ? analyticsPlan(activeStore, plan)
        : {
            store: activeStore,
            productId: purchase.productId,
            billingPeriod: catalogProduct?.billingPeriod,
          };

      const evidence = purchaseEvidence(activeStore, purchase);
      if (!evidence) {
        fail("missing_purchase_evidence", false);
        return;
      }

      processingRef.current.add(purchase.id);
      try {
        const response =
          activeStore === "apple"
            ? await verifyApplePurchase(evidence)
            : await verifyGooglePurchase(evidence);
        if (!mountedRef.current) return;
        contextRef.current = response;
        setContext(response);

        if (response.error) {
          fail(response.error.code, response.error.retryable);
          capturePurchaseFailure(posthog, analytics, response.error.code);
          return;
        }

        if (response.pending || purchase.purchaseState === "pending") {
          setState((current) => ({
            ...current,
            status: "pending",
            selectedProductId: purchase.productId,
            errorCode: null,
            retryable: true,
            entitlementOrigin: "purchase",
          }));
          pendingPurchasesRef.current.set(purchase.productId, purchase);
          captureActivationEvent(
            posthog,
            ACTIVATION_EVENT_NAMES.PURCHASE_PENDING,
            analytics,
          );
          return;
        }

        if (
          response.currentEntitlement &&
          ACCESS_STATUSES.has(response.currentEntitlement.status)
        ) {
          completedPurchasesRef.current.add(purchase.id);
          pendingPurchasesRef.current.delete(purchase.productId);
          setState((current) => ({
            ...current,
            status: "entitled",
            selectedProductId: purchase.productId,
            errorCode: null,
            retryable: false,
            entitlementOrigin: "purchase",
          }));
          captureActivationEvent(
            posthog,
            ACTIVATION_EVENT_NAMES.PURCHASE_SUCCEEDED,
            analytics,
          );
          try {
            await finishTransactionRef.current({
              purchase,
              isConsumable: false,
            });
          } catch {
            // Backend access is authoritative. The unfinished store transaction
            // will be replayed by the SDK, but must not revoke granted access.
          }
          return;
        }

        fail("entitlement_not_granted", true);
        capturePurchaseFailure(posthog, analytics, "entitlement_not_granted");
      } catch {
        fail("verification_unavailable", true);
        capturePurchaseFailure(posthog, analytics, "verification_unavailable");
      } finally {
        processingRef.current.delete(purchase.id);
      }
    },
    [fail, posthog, store],
  );

  const handlePurchaseError = useCallback(
    (error: { code?: ErrorCode; productId?: string | null }) => {
      const code = error.code ?? ErrorCode.Unknown;
      const attempt = purchaseAttemptRef.current;
      if (
        attempt &&
        (!error.productId || attempt.productId === error.productId)
      ) {
        if (attempt.errorHandled) return;
        attempt.errorHandled = true;
      }
      const plan = plansRef.current.find(
        (candidate) => candidate.productId === error.productId,
      );
      if (code === ErrorCode.UserCancelled) {
        stateRef.current = {
          ...stateRef.current,
          status: "ready",
          errorCode: null,
          entitlementOrigin: null,
        };
        setState((current) => ({
          ...current,
          status: "ready",
          errorCode: null,
          entitlementOrigin: null,
        }));
        return;
      }
      if (code === ErrorCode.Pending || code === ErrorCode.DeferredPayment) {
        stateRef.current = {
          ...stateRef.current,
          status: "pending",
          selectedProductId:
            error.productId ?? stateRef.current.selectedProductId,
          entitlementOrigin: "purchase",
        };
        setState((current) => ({
          ...current,
          status: "pending",
          selectedProductId: error.productId ?? current.selectedProductId,
          entitlementOrigin: "purchase",
        }));
        if (store && plan) {
          captureActivationEvent(
            posthog,
            ACTIVATION_EVENT_NAMES.PURCHASE_PENDING,
            analyticsPlan(store, plan),
          );
        }
        return;
      }
      fail(code, isRetryableStoreError(code));
      if (store && plan)
        capturePurchaseFailure(posthog, analyticsPlan(store, plan), code);
    },
    [fail, posthog, store],
  );

  const {
    connected,
    subscriptions,
    fetchProducts,
    requestPurchase,
    finishTransaction,
    reconnect,
  } = useIAP({
    onPurchaseSuccess: (purchase) => void processPurchase(purchase),
    onPurchaseError: handlePurchaseError,
    onError: () => fail("store_unavailable", true),
  });
  finishTransactionRef.current = finishTransaction;

  const loadContext = useCallback(async () => {
    if (!visible || !store) {
      if (visible) fail("unsupported_platform", false);
      return;
    }
    if (
      stateRef.current.status === "purchasing" ||
      stateRef.current.status === "restoring" ||
      processingRef.current.size > 0
    ) {
      return;
    }
    setState((current) => ({
      ...current,
      status: ["pending", "entitled"].includes(current.status)
        ? current.status
        : "loading",
      errorCode: null,
    }));
    contextRef.current = null;
    setContext(null);
    setProductsFetched(false);
    eligibilityGenerationRef.current += 1;
    setEligibleAppleIntroGroups(new Set());
    try {
      const response = await getMobileIAPContext(store);
      if (!mountedRef.current) return;
      contextRef.current = response;
      setContext(response);
      if (response.error) {
        fail(response.error.code, response.error.retryable);
      }
    } catch {
      fail("context_unavailable", true);
    }
  }, [fail, store, visible]);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    if (!visible) return;
    captureActivationEvent(posthog, ACTIVATION_EVENT_NAMES.PAYWALL_IMPRESSION, {
      source,
      store: store ?? undefined,
    });
    void loadContext();
  }, [loadContext, posthog, source, store, visible]);

  useEffect(() => {
    if (!visible || !context || context.error || !store) return;
    for (const [id, purchase] of queuedPurchasesRef.current) {
      queuedPurchasesRef.current.delete(id);
      void processPurchase(purchase);
    }
  }, [context, processPurchase, store, visible]);

  useEffect(() => {
    if (!visible || !context || context.error || !store) return;
    if (
      context.currentEntitlement &&
      ACCESS_STATUSES.has(context.currentEntitlement.status)
    ) {
      const previous = stateRef.current;
      const resolvedLocalPurchase =
        previous.status === "pending" &&
        previous.entitlementOrigin === "purchase";
      if (resolvedLocalPurchase) {
        const productId = context.currentEntitlement.productId;
        const purchase = pendingPurchasesRef.current.get(productId);
        const product = context.products.find(
          (candidate) => candidate.productId === productId,
        );
        if (product) {
          captureActivationEvent(
            posthog,
            ACTIVATION_EVENT_NAMES.PURCHASE_SUCCEEDED,
            {
              store,
              productId,
              billingPeriod: product.billingPeriod,
            },
          );
        }
        if (purchase) {
          completedPurchasesRef.current.add(purchase.id);
          pendingPurchasesRef.current.delete(productId);
          void finishTransactionRef
            .current({ purchase, isConsumable: false })
            .catch(() => undefined);
        }
      }
      setState((current) => ({
        ...current,
        status: "entitled",
        selectedProductId: context.currentEntitlement?.productId ?? null,
        entitlementOrigin: resolvedLocalPurchase
          ? "purchase"
          : current.status === "entitled"
            ? current.entitlementOrigin
            : null,
      }));
    } else if (context.pending) {
      setState((current) => ({
        ...current,
        status: "pending",
        selectedProductId: context.pending?.productId ?? null,
        entitlementOrigin:
          current.entitlementOrigin === "purchase" ? "purchase" : null,
      }));
    }
    if (!connected) return;
    if (context.products.length === 0) {
      setProductsFetched(true);
      return;
    }
    const requestContext = context;
    void fetchProducts({
      skus: context.products.map((product) => product.productId),
      type: "subs",
    })
      .then(() => {
        if (contextRef.current === requestContext) setProductsFetched(true);
      })
      .catch(() => {
        if (contextRef.current === requestContext)
          fail("store_unavailable", true);
      });
  }, [connected, context, fail, fetchProducts, posthog, store, visible]);

  useEffect(() => {
    if (
      !visible ||
      store !== "apple" ||
      !productsFetched ||
      !context ||
      context.error
    )
      return;
    const groups = Array.from(
      new Set(
        subscriptions.flatMap((product) =>
          product.platform === "ios" &&
          product.subscriptionGroupIdIOS &&
          product.subscriptionOffers?.some(
            (offer) => offer.type === "introductory",
          )
            ? [product.subscriptionGroupIdIOS]
            : [],
        ),
      ),
    );
    if (groups.length === 0) return;
    const generation = ++eligibilityGenerationRef.current;
    void Promise.all(
      groups.map(async (group) => ({
        group,
        eligible: await isEligibleForIntroOfferIOS(group).catch(() => false),
      })),
    ).then((results) => {
      if (
        !mountedRef.current ||
        generation !== eligibilityGenerationRef.current
      )
        return;
      setEligibleAppleIntroGroups(
        new Set(
          results
            .filter((result) => result.eligible)
            .map((result) => result.group),
        ),
      );
    });
  }, [context, productsFetched, store, subscriptions, visible]);

  useEffect(() => {
    if (!visible || !context || context.error || !productsFetched) return;
    const plans = intersectStorePlans(
      context.products,
      subscriptions,
      eligibleAppleIntroGroups,
    );
    if (plans.length === 0 && context.products.length > 0 && !connected) return;
    setState((current) => {
      if (
        processingRef.current.size > 0 ||
        ["purchasing", "pending", "entitled", "restoring"].includes(
          current.status,
        )
      ) {
        return { ...current, plans };
      }
      return {
        ...current,
        plans,
        status: plans.length > 0 ? "ready" : "empty",
        errorCode: null,
      };
    });
  }, [
    connected,
    context,
    eligibleAppleIntroGroups,
    productsFetched,
    subscriptions,
    visible,
  ]);

  useEffect(() => {
    if (!visible) return;
    const refresh = () => void loadContext();
    const appState = AppState.addEventListener("change", (next) => {
      if (next === "active") refresh();
    });
    const link = Linking.addEventListener("url", refresh);
    return () => {
      appState.remove();
      link.remove();
    };
  }, [loadContext, visible]);

  const selectPlan = useCallback(
    (productId: string) => {
      if (state.status !== "ready") return;
      const plan = state.plans.find(
        (candidate) => candidate.productId === productId,
      );
      if (!plan || !store) return;
      setState((current) => ({ ...current, selectedProductId: productId }));
      captureActivationEvent(
        posthog,
        ACTIVATION_EVENT_NAMES.PAYWALL_PLAN_SELECTED,
        { source, store, billingPeriod: plan.billingPeriod },
      );
    },
    [posthog, source, state.plans, state.status, store],
  );

  const purchase = useCallback(async () => {
    if (state.status !== "ready" || !context?.purchaseContext || !store) return;
    const plan = state.plans.find(
      (candidate) => candidate.productId === state.selectedProductId,
    );
    if (!plan) return;
    stateRef.current = { ...state, status: "purchasing" };
    purchaseAttemptRef.current = {
      productId: plan.productId,
      errorHandled: false,
    };
    setState((current) => ({ ...current, status: "purchasing" }));
    try {
      if (store === "apple" && context.purchaseContext.appleAppAccountToken) {
        await requestPurchase({
          type: "subs",
          request: {
            apple: {
              sku: plan.productId,
              appAccountToken: context.purchaseContext.appleAppAccountToken,
              andDangerouslyFinishTransactionAutomatically: false,
            },
          },
        });
      } else if (
        store === "google" &&
        context.purchaseContext.googleObfuscatedAccountId
      ) {
        await requestPurchase({
          type: "subs",
          request: {
            google: {
              skus: [plan.productId],
              obfuscatedAccountId:
                context.purchaseContext.googleObfuscatedAccountId,
              subscriptionOffers: plan.offer?.offerTokenAndroid
                ? [
                    {
                      sku: plan.productId,
                      offerToken: plan.offer.offerTokenAndroid,
                    },
                  ]
                : null,
            },
          },
        });
      } else {
        fail("invalid_purchase_context", false);
      }
    } catch (error) {
      handlePurchaseError({
        code: canonicalPurchaseErrorCode(error),
        productId: plan.productId,
      });
    }
  }, [context, fail, handlePurchaseError, requestPurchase, state, store]);

  const restore = useCallback(async () => {
    if (!store || stateRef.current.status === "restoring") return;
    stateRef.current = { ...stateRef.current, status: "restoring" };
    setState((current) => ({ ...current, status: "restoring" }));
    try {
      const purchases = await getAvailablePurchases();
      const eligible = purchases
        .map((purchase) => ({
          purchase,
          evidence: purchaseEvidence(store, purchase),
        }))
        .filter((entry): entry is { purchase: Purchase; evidence: string } =>
          Boolean(entry.evidence),
        );
      const response = await restoreMobileIAPPurchases(
        store,
        eligible.map((entry) => entry.evidence),
      );
      contextRef.current = response;
      setContext(response);
      for (const outcome of response.restore?.outcomes ?? []) {
        if (outcome.status !== "restored" && outcome.status !== "unchanged")
          continue;
        const purchaseToFinish = eligible[outcome.index]?.purchase;
        if (purchaseToFinish) {
          try {
            await finishTransaction({
              purchase: purchaseToFinish,
              isConsumable: false,
            });
          } catch {
            // Restore authority already came from the backend. The SDK can
            // replay unfinished transactions without hiding granted access.
          }
        }
      }
      captureActivationEvent(
        posthog,
        ACTIVATION_EVENT_NAMES.RESTORE_COMPLETED,
        {
          store,
          restoreStatus: response.restore?.status ?? "failed",
        },
      );
      if (
        response.currentEntitlement &&
        ACCESS_STATUSES.has(response.currentEntitlement.status)
      ) {
        setState((current) => ({
          ...current,
          status: "entitled",
          selectedProductId: response.currentEntitlement?.productId ?? null,
          restoreStatus: response.restore?.status ?? null,
          entitlementOrigin: "restore",
        }));
      } else if (response.pending || response.restore?.status === "pending") {
        setState((current) => ({
          ...current,
          status: "pending",
          selectedProductId: response.pending?.productId ?? null,
          restoreStatus: response.restore?.status ?? null,
          entitlementOrigin: null,
        }));
      } else if (response.error || response.restore?.status === "failed") {
        fail(
          response.error?.code ??
            response.restore?.error?.code ??
            "restore_failed",
          true,
        );
      } else {
        setState((current) => ({
          ...current,
          status: current.plans.length > 0 ? "ready" : "empty",
          restoreStatus: response.restore?.status ?? null,
          entitlementOrigin: null,
        }));
      }
    } catch {
      fail("restore_unavailable", true);
      captureActivationEvent(
        posthog,
        ACTIVATION_EVENT_NAMES.RESTORE_COMPLETED,
        {
          store,
          restoreStatus: "failed",
        },
      );
    }
  }, [fail, finishTransaction, posthog, store]);

  const retry = useCallback(async () => {
    if (!connected) await reconnect();
    await loadContext();
  }, [connected, loadContext, reconnect]);

  return {
    state,
    store,
    selectPlan,
    purchase,
    restore,
    retry,
    refresh: loadContext,
  };
}

export function purchaseEvidence(
  store: IAPStore,
  purchase: Purchase,
): string | null {
  if (store === "apple") {
    return purchase.transactionId?.trim() || null;
  }
  return purchase.purchaseToken?.trim() || null;
}

function analyticsPlan(store: IAPStore, plan: NativeIAPPlan) {
  return {
    store,
    productId: plan.productId,
    billingPeriod: plan.billingPeriod,
  };
}

function capturePurchaseFailure(
  posthog: Parameters<typeof captureActivationEvent>[0],
  analytics: {
    store: IAPStore;
    productId: string;
    billingPeriod?: "monthly" | "annual";
  },
  errorCode: string,
) {
  captureActivationEvent(posthog, ACTIVATION_EVENT_NAMES.PURCHASE_FAILED, {
    ...analytics,
    errorCode,
  });
}

function isRetryableStoreError(code: string): boolean {
  return ![
    ErrorCode.UserCancelled,
    ErrorCode.ItemUnavailable,
    ErrorCode.SkuNotFound,
    ErrorCode.FeatureNotSupported,
  ].includes(code as ErrorCode);
}

function canonicalPurchaseErrorCode(error: unknown): ErrorCode {
  if (!error || typeof error !== "object" || !("code" in error)) {
    return ErrorCode.Unknown;
  }
  const code = (error as { code?: unknown }).code;
  return typeof code === "string" &&
    Object.values(ErrorCode).includes(code as ErrorCode)
    ? (code as ErrorCode)
    : ErrorCode.Unknown;
}
