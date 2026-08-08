import { getApiBaseUrl } from "@/api/baseUrl";
import {
  authenticatedFetch,
  authenticationHeaders,
  hasAuthenticationSession,
} from "@/api/users";

export type IAPStore = "apple" | "google";
export type IAPBillingPeriod = "monthly" | "annual";

export interface MobileIAPProduct {
  store: IAPStore;
  productId: string;
  entitlement: "premium";
  billingPeriod: IAPBillingPeriod;
}

export interface MobileIAPPurchaseContext {
  store: IAPStore;
  appleAppAccountToken?: string;
  googleObfuscatedAccountId?: string;
}

export interface MobileIAPEntitlement {
  store: IAPStore;
  productId: string;
  entitlement: "premium";
  status: string;
  periodEnd?: string;
  autoRenewEnabled: boolean;
  canceledAt?: string;
  revokedAt?: string;
}

export interface MobileIAPPending {
  store: IAPStore;
  productId: string;
  startedAt: string;
}

export interface MobileIAPError {
  code: string;
  messageKey: string;
  retryable: boolean;
}

export type IAPRestoreStatus =
  | "restored"
  | "no_purchases"
  | "partial"
  | "pending"
  | "failed";

export interface MobileIAPRestoreResult {
  status: IAPRestoreStatus;
  restoredPurchases: number;
  unchangedPurchases: number;
  rejectedPurchases: number;
  retryablePurchases: number;
  outcomes: {
    index: number;
    status: "restored" | "unchanged" | "rejected" | "retryable";
    error?: MobileIAPError;
  }[];
  error?: MobileIAPError;
}

export interface MobileIAPResponse {
  products: MobileIAPProduct[];
  purchaseContext: MobileIAPPurchaseContext | null;
  currentEntitlement: MobileIAPEntitlement | null;
  pending: MobileIAPPending | null;
  error: MobileIAPError | null;
  restore: MobileIAPRestoreResult | null;
  serverTime: string;
}

const API_URL = getApiBaseUrl();
const REQUEST_TIMEOUT_MS = 25_000;

async function requestIAP(
  path: string,
  method: "GET" | "POST",
  body?: unknown,
): Promise<MobileIAPResponse> {
  if (!hasAuthenticationSession()) throw new Error("Authentication required");

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
  try {
    const response = await authenticatedFetch(`${API_URL}${path}`, {
      method,
      headers: {
        ...authenticationHeaders(),
        ...(body === undefined ? {} : { "Content-Type": "application/json" }),
      },
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: controller.signal,
    });
    const payload: unknown = await response.json();
    if (!isMobileIAPResponse(payload)) {
      throw new Error(`Invalid native IAP response (${response.status})`);
    }
    return payload;
  } finally {
    clearTimeout(timeout);
  }
}

export function getMobileIAPContext(store: IAPStore) {
  return requestIAP(`/subscription/iap/context?store=${store}`, "GET");
}

export function verifyApplePurchase(transactionId: string) {
  return requestIAP("/subscription/iap/apple/verify", "POST", {
    transactionId,
  });
}

export function verifyGooglePurchase(purchaseToken: string) {
  return requestIAP("/subscription/iap/google/verify", "POST", {
    purchaseToken,
  });
}

export function restoreMobileIAPPurchases(store: IAPStore, evidence: string[]) {
  return requestIAP("/subscription/iap/restore", "POST", {
    store,
    evidence,
  });
}

function isMobileIAPResponse(value: unknown): value is MobileIAPResponse {
  if (!value || typeof value !== "object") return false;
  const response = value as Partial<MobileIAPResponse>;
  return (
    Array.isArray(response.products) &&
    typeof response.serverTime === "string" &&
    "purchaseContext" in response &&
    "currentEntitlement" in response &&
    "pending" in response &&
    "error" in response &&
    "restore" in response
  );
}
