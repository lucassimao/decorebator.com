import { getAuthorization } from "./users";
import {  DEFAULT_ERROR } from "./constants";

const API_URL =process.env.EXPO_PUBLIC_API_URL

export type SubscriptionStatus = {
  plan: 'free' | 'monthly' | 'annual';
  status?: 'active' | 'cancelled' | 'past_due' | 'trialing' | 'unpaid';
  currentPeriodEnd?: string;
  cancelAtPeriodEnd?: boolean;
  trialEnd?: string;
};

export type CheckoutSessionResponse = {
  checkoutUrl: string;
  sessionId: string;
};

export async function createCheckoutSession(plan: 'monthly' | 'annual',expoUri:string): Promise<CheckoutSessionResponse> {
  const endpoint = `${process.env.EXPO_PUBLIC_API_URL}/subscription/checkout-session`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": authorization,
    },
    body: JSON.stringify({ plan, expoUri }),
  });

  if (!response.ok) {
    const body = await response.json();
    throw new Error(body?.error || DEFAULT_ERROR);
  }

  return response.json();
}

export async function getSubscriptionStatus(): Promise<SubscriptionStatus> {
  const endpoint = `${API_URL}/subscription/status`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
    method: "GET",
    headers: {
      "Authorization": authorization,
    },
  });

  if (!response.ok) {
    const body = await response.json();
    throw new Error(body?.error || DEFAULT_ERROR);
  }

  return response.json();
}

export async function cancelSubscription(): Promise<void> {
  const endpoint = `${API_URL}/subscription/cancel`;
  const authorization = getAuthorization();

  if (!authorization) {
    throw new Error("Authentication required");
  }

  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      "Authorization": authorization,
    },
  });

  if (!response.ok) {
    const body = await response.json();
    throw new Error(body?.error || DEFAULT_ERROR);
  }
}