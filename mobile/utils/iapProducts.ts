import { Platform } from "react-native";
import type { ProductSubscription, SubscriptionOffer } from "expo-iap";

import type { IAPBillingPeriod, IAPStore, MobileIAPProduct } from "@/api/iap";

export interface NativeIAPPlan {
  billingPeriod: IAPBillingPeriod;
  productId: string;
  title: string;
  displayPrice: string;
  period: { unit: "day" | "week" | "month" | "year"; value: number };
  offer: SubscriptionOffer | null;
  offerDisplay: {
    displayPrice: string;
    period: NativeIAPPlan["period"];
  } | null;
  product: ProductSubscription;
}

export function currentIAPStore(): IAPStore | null {
  if (Platform.OS === "ios") return "apple";
  if (Platform.OS === "android") return "google";
  return null;
}

export function intersectStorePlans(
  catalog: MobileIAPProduct[],
  subscriptions: ProductSubscription[],
  eligibleAppleIntroGroups: ReadonlySet<string> = new Set(),
): NativeIAPPlan[] {
  const storeProducts = new Map(subscriptions.map((item) => [item.id, item]));
  return catalog.flatMap((entry) => {
    const product = storeProducts.get(entry.productId);
    if (!product || product.type !== "subs") return [];
    const offer = selectStoreOffer(product, eligibleAppleIntroGroups);
    const renewal = renewalTerms(product, offer);
    if (!renewal) return [];
    return [
      {
        billingPeriod: entry.billingPeriod,
        productId: entry.productId,
        title: product.title,
        displayPrice: renewal.displayPrice,
        period: renewal.period,
        offer,
        offerDisplay: introductoryTerms(product, offer),
        product,
      },
    ];
  });
}

export function selectStoreOffer(
  product: ProductSubscription,
  eligibleAppleIntroGroups: ReadonlySet<string> = new Set(),
): SubscriptionOffer | null {
  const offers = product.subscriptionOffers ?? [];
  if (product.platform === "android") {
    return (
      offers.find((offer) => Boolean(offer.offerTokenAndroid)) ??
      offers[0] ??
      null
    );
  }
  if (
    !product.subscriptionGroupIdIOS ||
    !eligibleAppleIntroGroups.has(product.subscriptionGroupIdIOS)
  ) {
    return null;
  }
  return offers.find((offer) => offer.type === "introductory") ?? null;
}

function renewalTerms(
  product: ProductSubscription,
  offer: SubscriptionOffer | null,
): Pick<NativeIAPPlan, "displayPrice" | "period"> | null {
  if (
    product.platform === "ios" &&
    product.subscriptionPeriodUnitIOS &&
    product.subscriptionPeriodUnitIOS !== "empty"
  ) {
    const value = Number(product.subscriptionPeriodNumberIOS ?? "1");
    return {
      displayPrice: product.displayPrice,
      period: {
        unit: product.subscriptionPeriodUnitIOS,
        value: Number.isFinite(value) && value > 0 ? value : 1,
      },
    };
  }
  if (product.platform === "android" && offer) {
    const phases = offer.pricingPhasesAndroid?.pricingPhaseList ?? [];
    const recurring = phases.find((phase) => phase.recurrenceMode === 1);
    const period = recurring
      ? parseGoogleBillingPeriod(recurring.billingPeriod)
      : null;
    if (recurring && period) {
      return { displayPrice: recurring.formattedPrice, period };
    }
  }
  return null;
}

function introductoryTerms(
  product: ProductSubscription,
  offer: SubscriptionOffer | null,
): NativeIAPPlan["offerDisplay"] {
  if (!offer) return null;
  if (product.platform === "ios") {
    return offer.period && offer.period.unit !== "unknown"
      ? {
          displayPrice: offer.displayPrice,
          period: { unit: offer.period.unit, value: offer.period.value },
        }
      : null;
  }
  const introductory = offer.pricingPhasesAndroid?.pricingPhaseList.find(
    (phase) => phase.recurrenceMode !== 1,
  );
  const period = introductory
    ? parseGoogleBillingPeriod(introductory.billingPeriod)
    : null;
  return introductory && period
    ? { displayPrice: introductory.formattedPrice, period }
    : null;
}

function parseGoogleBillingPeriod(
  value: string,
): NativeIAPPlan["period"] | null {
  const match = /^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?$/.exec(value);
  if (!match) return null;
  const entries = [
    ["year", match[1]],
    ["month", match[2]],
    ["week", match[3]],
    ["day", match[4]],
  ] as const;
  for (const [unit, raw] of entries) {
    const count = Number(raw);
    if (Number.isFinite(count) && count > 0) return { unit, value: count };
  }
  return null;
}
