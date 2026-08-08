import {
  getMobileIAPContext,
  restoreMobileIAPPurchases,
  verifyApplePurchase,
  verifyGooglePurchase,
} from "@/api/iap";
import { authenticationHeaders, hasAuthenticationSession } from "@/api/users";

jest.mock("@/api/baseUrl", () => ({
  getApiBaseUrl: () => "https://api.test",
}));
jest.mock("@/api/users", () => ({
  getAuthorization: jest.fn(),
  hasAuthenticationSession: jest.fn(() => true),
  authenticationHeaders: jest.fn(() => ({ Authorization: "test-token" })),
  authenticatedFetch: (...args: Parameters<typeof fetch>) => fetch(...args),
}));

const envelope = {
  products: [],
  purchaseContext: null,
  currentEntitlement: null,
  pending: null,
  error: null,
  restore: null,
  serverTime: "2026-08-05T12:00:00Z",
};

describe("native IAP API", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (hasAuthenticationSession as jest.Mock).mockReturnValue(true);
    (authenticationHeaders as jest.Mock).mockReturnValue({
      Authorization: "Bearer test",
    });
    global.fetch = jest.fn().mockResolvedValue({
      status: 200,
      json: async () => envelope,
    }) as jest.Mock;
  });

  it("loads an explicit-null store context without leaking identity", async () => {
    await expect(getMobileIAPContext("apple")).resolves.toEqual(envelope);
    expect(fetch).toHaveBeenCalledWith(
      "https://api.test/subscription/iap/context?store=apple",
      expect.objectContaining({
        method: "GET",
        headers: { Authorization: "Bearer test" },
      }),
    );
    expect(JSON.stringify((fetch as jest.Mock).mock.calls)).not.toContain(
      "userId",
    );
  });

  it("sends only the provider evidence required by each verification route", async () => {
    await verifyApplePurchase("transaction-1");
    await verifyGooglePurchase("purchase-token-1");
    expect((fetch as jest.Mock).mock.calls[0][1].body).toBe(
      JSON.stringify({ transactionId: "transaction-1" }),
    );
    expect((fetch as jest.Mock).mock.calls[1][1].body).toBe(
      JSON.stringify({ purchaseToken: "purchase-token-1" }),
    );
  });

  it("preserves restore evidence order for backend outcome indexes", async () => {
    await restoreMobileIAPPurchases("google", ["first", "second"]);
    expect((fetch as jest.Mock).mock.calls[0][1].body).toBe(
      JSON.stringify({ store: "google", evidence: ["first", "second"] }),
    );
  });

  it("rejects malformed envelopes and missing authorization", async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      status: 503,
      json: async () => ({ error: "raw provider error" }),
    });
    await expect(getMobileIAPContext("google")).rejects.toThrow(
      "Invalid native IAP response",
    );

    (hasAuthenticationSession as jest.Mock).mockReturnValue(false);
    await expect(getMobileIAPContext("apple")).rejects.toThrow(
      "Authentication required",
    );
  });
});
