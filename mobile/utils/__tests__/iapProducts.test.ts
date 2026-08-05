import type { ProductSubscription } from "expo-iap";

import type { MobileIAPProduct } from "@/api/iap";
import { intersectStorePlans, selectStoreOffer } from "@/utils/iapProducts";

const catalog: MobileIAPProduct[] = [
  {
    store: "google",
    productId: "premium.annual",
    entitlement: "premium",
    billingPeriod: "annual",
  },
  {
    store: "google",
    productId: "premium.monthly",
    entitlement: "premium",
    billingPeriod: "monthly",
  },
];

const annual = {
  id: "premium.annual",
  type: "subs",
  platform: "android",
  title: "Premium Annual",
  displayPrice: "R$ 199,90",
  subscriptionOffers: [
    {
      id: "base",
      type: "introductory",
      displayPrice: "R$ 99,90",
      price: 99.9,
      period: { unit: "year", value: 1 },
      offerTokenAndroid: "offer-token",
      pricingPhasesAndroid: {
        pricingPhaseList: [
          {
            billingCycleCount: 1,
            billingPeriod: "P1M",
            formattedPrice: "R$ 0,00",
            priceAmountMicros: "0",
            priceCurrencyCode: "BRL",
            recurrenceMode: 2,
          },
          {
            billingCycleCount: 0,
            billingPeriod: "P1Y",
            formattedPrice: "R$ 199,90",
            priceAmountMicros: "199900000",
            priceCurrencyCode: "BRL",
            recurrenceMode: 1,
          },
        ],
      },
    },
  ],
} as ProductSubscription;

describe("native IAP store product adapter", () => {
  it("renders only the backend/store intersection with store-authored metadata", () => {
    const plans = intersectStorePlans(catalog, [annual]);
    expect(plans).toHaveLength(1);
    expect(plans[0]).toMatchObject({
      productId: "premium.annual",
      billingPeriod: "annual",
      title: "Premium Annual",
      displayPrice: "R$ 199,90",
      period: { unit: "year", value: 1 },
      offerDisplay: {
        displayPrice: "R$ 0,00",
        period: { unit: "month", value: 1 },
      },
    });
    expect(JSON.stringify(plans[0])).not.toContain("premium.monthly");
  });

  it("rejects products that have no store billing period", () => {
    const incomplete = {
      ...annual,
      subscriptionOffers: [],
    } as ProductSubscription;
    expect(intersectStorePlans(catalog, [incomplete])).toEqual([]);
  });

  it("selects the same eligible Android offer token that purchase will use", () => {
    expect(selectStoreOffer(annual)?.offerTokenAndroid).toBe("offer-token");
  });

  it("uses StoreKit renewal terms and only eligible introductory metadata", () => {
    const apple = {
      id: "premium.annual",
      type: "subs",
      platform: "ios",
      title: "Annual",
      displayPrice: "$39.99",
      subscriptionPeriodUnitIOS: "year",
      subscriptionPeriodNumberIOS: "1",
      subscriptionGroupIdIOS: "group-1",
      subscriptionOffers: [
        {
          id: "intro",
          type: "introductory",
          displayPrice: "$0.00",
          price: 0,
          period: { unit: "week", value: 1 },
        },
      ],
    } as ProductSubscription;
    const ineligible = intersectStorePlans(
      [{ ...catalog[0], store: "apple" }],
      [apple],
    )[0];
    expect(ineligible.offer).toBeNull();
    expect(ineligible.period).toEqual({ unit: "year", value: 1 });

    const plan = intersectStorePlans(
      [{ ...catalog[0], store: "apple" }],
      [apple],
      new Set(["group-1"]),
    )[0];
    expect(plan.offer?.id).toBe("intro");
    expect(plan.offerDisplay?.period).toEqual({ unit: "week", value: 1 });
    expect(plan.period).toEqual({ unit: "year", value: 1 });
    expect(plan.displayPrice).toBe("$39.99");
  });
});
