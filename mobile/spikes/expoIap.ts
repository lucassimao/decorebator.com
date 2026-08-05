import {
  endConnection,
  fetchProducts,
  finishTransaction,
  getAvailablePurchases,
  initConnection,
  requestPurchase,
  type ProductOrSubscription,
  type Purchase,
} from "expo-iap";

export type IAPSpikePurchaseContext =
  | { store: "apple"; appAccountToken: string }
  | { store: "google"; obfuscatedAccountId: string };

export async function connectIAPSpike(): Promise<() => Promise<void>> {
  await initConnection();
  return async () => {
    await endConnection();
  };
}

export async function fetchSubscriptionProductsSpike(
  productIds: string[],
): Promise<ProductOrSubscription[]> {
  return (await fetchProducts({ skus: productIds, type: "subs" })) ?? [];
}

export async function requestSubscriptionSpike(
  productId: string,
  context: IAPSpikePurchaseContext,
  googleOfferToken?: string,
): Promise<void> {
  if (context.store === "apple") {
    await requestPurchase({
      type: "subs",
      request: {
        apple: {
          sku: productId,
          appAccountToken: context.appAccountToken,
          andDangerouslyFinishTransactionAutomatically: false,
        },
      },
    });
    return;
  }

  await requestPurchase({
    type: "subs",
    request: {
      google: {
        skus: [productId],
        obfuscatedAccountId: context.obfuscatedAccountId,
        subscriptionOffers: googleOfferToken
          ? [{ sku: productId, offerToken: googleOfferToken }]
          : null,
      },
    },
  });
}

export async function restorePurchasesSpike(): Promise<Purchase[]> {
  return getAvailablePurchases();
}

// Call only after the backend has verified the purchase and returned its
// authoritative entitlement envelope. Never finish from the native callback.
export async function finishBackendVerifiedPurchaseSpike(
  purchase: Purchase,
): Promise<void> {
  await finishTransaction({ purchase, isConsumable: false });
}
