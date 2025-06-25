# Subscription System Documentation

## Overview

Decorebator implements a dual-provider subscription system that intelligently routes users to the appropriate payment provider based on their platform and location. This architecture ensures compliance with app store policies while maximizing revenue opportunities.

## Table of Contents

1. [Architecture](#architecture)
2. [Payment Providers](#payment-providers)
3. [Database Schema](#database-schema)
4. [API Endpoints](#api-endpoints)
5. [Mobile Integration](#mobile-integration)
6. [Webhook Processing](#webhook-processing)
7. [Testing](#testing)
8. [Configuration](#configuration)
9. [Monitoring & Analytics](#monitoring--analytics)
10. [Migration Guide](#migration-guide)

## Architecture

### Provider Selection Logic

The system automatically selects the appropriate payment provider based on:

```
┌─────────────────────────────────────────┐
│         Provider Selection Logic         │
├─────────────────────────────────────────┤
│ Platform: Android → RevenueCat          │
│ Platform: iOS + Location: US → Stripe   │
│ Platform: iOS + Location: Non-US → RC   │
│ Platform: Web → Stripe                  │
└─────────────────────────────────────────┘
```

### System Components

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Mobile App  │────▶│   API Server │────▶│   Database   │
└──────────────┘     └──────────────┘     └──────────────┘
       │                     │                     ▲
       │                     ▼                     │
       │            ┌──────────────┐               │
       │            │   Webhooks   │───────────────┘
       │            └──────────────┘
       │                     ▲
       ▼                     │
┌──────────────┐     ┌──────────────┐
│  RevenueCat  │     │    Stripe    │
└──────────────┘     └──────────────┘
```

## Payment Providers

### Stripe Integration

**Used for:**
- US iOS users
- All web users
- Direct payment processing

**Features:**
- Hosted checkout sessions
- Customer portal for subscription management
- Webhook events for real-time updates
- Invoice and receipt generation

### RevenueCat Integration

**Used for:**
- All Android users
- Non-US iOS users
- In-app purchase management

**Features:**
- Native paywall UI
- Cross-platform purchase restoration
- Subscription lifecycle management
- App Store / Google Play integration

## Database Schema

### Core Tables

#### subscriptions
```sql
CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    provider subscription_provider NOT NULL, -- 'stripe' or 'revenuecat'
    stripe_subscription_id TEXT,
    stripe_customer_id TEXT,
    revenuecat_subscription_id TEXT,
    app_store_product_id TEXT,
    platform platform_type,
    plan subscription_plan NOT NULL,
    status subscription_status NOT NULL,
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    cancelled_at TIMESTAMPTZ,
    trial_end TIMESTAMPTZ,
    amount_cents INT,
    currency TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### subscription_events
```sql
CREATE TABLE subscription_events (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT REFERENCES subscriptions(id),
    stripe_event_id TEXT UNIQUE,
    event_type TEXT NOT NULL,
    event_data JSONB,
    processed_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### revenuecat_events
```sql
CREATE TABLE revenuecat_events (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT UNIQUE NOT NULL,
    event_type TEXT NOT NULL,
    app_user_id TEXT,
    event_data JSONB NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### User Table Extensions
```sql
ALTER TABLE users ADD COLUMN revenuecat_customer_id TEXT;
ALTER TABLE users ADD COLUMN platform platform_type;
```

### Enum Types
```sql
CREATE TYPE subscription_provider AS ENUM ('stripe', 'revenuecat');
CREATE TYPE platform_type AS ENUM ('ios', 'android', 'web');
```

## API Endpoints

### Provider Detection

Provider detection is now handled locally in the mobile app. No API endpoint is needed as the mobile client determines the provider based on:
- Platform (iOS/Android)
- User country (from profile)

The logic is implemented in `mobile/hooks/useRevenueCat.ts`:
```typescript
const isUS = user.country === "US";
const provider = platform === "android" ? "revenuecat" : 
                platform === "ios" && !isUS ? "revenuecat" : "stripe";
```

### Stripe Checkout
```http
POST /subscription/checkout
Authorization: Bearer {token}
Content-Type: application/json

{
    "plan": "monthly" | "annual",
    "redirectUri": "decorebator://settings"
}

Response:
{
    "checkoutUrl": "https://checkout.stripe.com/...",
    "sessionId": "cs_test_..."
}
```

### RevenueCat Purchase Restoration
```http
POST /subscription/revenuecat/restore
Authorization: Bearer {token}
Content-Type: application/json

{
    "appUserId": "123",
    "platform": "ios"
}

Response:
{
    "success": true,
    "subscription": { ... }
}
```

### Subscription Status
```http
GET /subscription
Authorization: Bearer {token}

Response:
{
    "plan": "monthly",
    "status": "active",
    "provider": "stripe",
    "currentPeriodEnd": "2024-02-15T10:00:00Z",
    "cancelAtPeriodEnd": false
}
```

### Cancel Subscription
```http
POST /subscription/cancel
Authorization: Bearer {token}

Response:
{
    "success": true,
    "message": "Subscription will be cancelled at end of period"
}
```

### Webhook Endpoints

#### Stripe Webhook
```http
POST /webhook/stripe
Stripe-Signature: {signature}
Content-Type: application/json

{
    "type": "customer.subscription.created",
    "data": { ... }
}
```

#### RevenueCat Webhook
```http
POST /webhook/revenuecat
Authorization: {webhook_secret}
Content-Type: application/json

{
    "api_version": "1.0",
    "event": {
        "type": "INITIAL_PURCHASE",
        ...
    }
}
```

## Mobile Integration

### Provider Detection Hook
```typescript
// hooks/useRevenueCat.ts
export function usePaymentProvider() {
  const { user } = useUserInfo();
  
  return useQuery({
    queryKey: ["paymentProvider", user?.id],
    queryFn: getPaymentProvider,
    enabled: !!user,
    staleTime: 5 * 60 * 1000,
  });
}
```

### Conditional UI Rendering
```typescript
// Settings screen shows different UI based on provider
{providerInfo?.provider === "stripe" ? (
  <StripePlanSelection />
) : (
  <TouchableOpacity onPress={() => setShowRevenueCatPaywall(true)}>
    <Text>View Plans</Text>
  </TouchableOpacity>
)}
```

### RevenueCat Paywall Component
```typescript
// components/RevenueCatPaywall.tsx
export default function RevenueCatPaywall({ onClose, onSuccess }) {
  const { 
    offerings, 
    purchasePackage, 
    restorePurchases 
  } = useRevenueCat();
  
  // Renders native paywall with available packages
  // Handles purchase flow and restoration
}
```

### RevenueCat SDK Initialization
```typescript
// Automatic initialization based on user ID
useEffect(() => {
  if (user && Platform.OS !== 'web') {
    Purchases.configure({
      apiKey: Platform.OS === 'ios' 
        ? REVENUECAT_API_KEY_IOS 
        : REVENUECAT_API_KEY_ANDROID,
      appUserID: user.id.toString(),
    });
  }
}, [user]);
```

## Webhook Processing

### Stripe Webhook Handler
```go
func (s *SubscriptionService) HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error {
    // 1. Verify webhook signature
    event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
    
    // 2. Process based on event type
    switch event.Type {
    case "customer.subscription.created":
        return s.handleSubscriptionCreated(ctx, event)
    case "customer.subscription.updated":
        return s.handleSubscriptionUpdated(ctx, event)
    case "customer.subscription.deleted":
        return s.handleSubscriptionDeleted(ctx, event)
    }
}
```

### RevenueCat Webhook Handler
```go
func (s *RevenueCatService) HandleWebhook(ctx context.Context, payload []byte, authHeader string) error {
    // 1. Verify authorization header
    if authHeader != s.webhookSecret {
        return ErrUnauthorized
    }
    
    // 2. Parse webhook payload
    var webhook RevenueCatWebhook
    json.Unmarshal(payload, &webhook)
    
    // 3. Process based on event type
    switch webhook.Event.Type {
    case "INITIAL_PURCHASE":
        return s.handleInitialPurchase(ctx, webhook.Event)
    case "RENEWAL":
        return s.handleRenewal(ctx, webhook.Event)
    case "CANCELLATION":
        return s.handleCancellation(ctx, webhook.Event)
    }
}
```

## Testing

### Integration Tests
```go
// tests/integration/revenuecat_test.go
func TestRevenueCatIntegration(t *testing.T) {
    t.Run("GetPaymentProvider_ForAndroidUser_ReturnsRevenueCat", func(t *testing.T) {
        // Test Android users get RevenueCat
    })
    
    t.Run("GetPaymentProvider_ForUSiOSUser_ReturnsStripe", func(t *testing.T) {
        // Test US iOS users get Stripe
    })
    
    t.Run("RevenueCatWebhook_ProcessesInitialPurchase", func(t *testing.T) {
        // Test webhook creates subscription
    })
}
```

### Manual Testing Checklist
- [ ] Android user can purchase via RevenueCat
- [ ] US iOS user can purchase via Stripe
- [ ] Non-US iOS user can purchase via RevenueCat
- [ ] Webhooks update subscription status
- [ ] Purchase restoration works correctly
- [ ] Subscription cancellation works
- [ ] Premium features unlock after purchase

## Configuration

### Environment Variables

#### API Backend
```bash
# Stripe Configuration
STRIPE_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_MONTHLY_PRICE_ID=price_...
STRIPE_ANNUAL_PRICE_ID=price_...

# RevenueCat Configuration
REVENUECAT_API_KEY_IOS=appl_...
REVENUECAT_API_KEY_ANDROID=goog_...
REVENUECAT_WEBHOOK_SECRET=rc_webhook_secret
```

#### Mobile App
```bash
# RevenueCat SDK Keys
EXPO_PUBLIC_REVENUECAT_API_KEY_IOS=appl_...
EXPO_PUBLIC_REVENUECAT_API_KEY_ANDROID=goog_...
```

### RevenueCat Dashboard Setup

1. **Create App in RevenueCat**
   - Add iOS app with bundle ID
   - Add Android app with package name

2. **Configure Products**
   - iOS Monthly: `com.decorebator.premium.monthly`
   - iOS Annual: `com.decorebator.premium.annual`
   - Android Monthly: `premium_monthly`
   - Android Annual: `premium_annual`

3. **Create Entitlement**
   - Name: `premium`
   - Include all premium products

4. **Configure Webhook**
   - URL: `https://api.decorebator.com/webhook/revenuecat`
   - Authorization Header: Your webhook secret

### Stripe Dashboard Setup

1. **Create Products**
   - Monthly Premium: $6.99/month
   - Annual Premium: $69.90/year

2. **Configure Webhook**
   - URL: `https://api.decorebator.com/webhook/stripe`
   - Events: All subscription events

3. **Set up Customer Portal**
   - Enable subscription management
   - Configure cancellation flow

## Monitoring & Analytics

### Key Metrics to Track

1. **Conversion Metrics**
   - Provider selection distribution
   - Checkout initiation rate
   - Purchase completion rate
   - Cancellation rate by provider

2. **Revenue Metrics**
   - MRR by provider
   - ARPU by platform
   - Churn rate by provider

3. **Technical Metrics**
   - Webhook processing success rate
   - Provider API response times
   - Failed purchase attempts

### SQL Queries for Analytics

```sql
-- Active subscriptions by provider
SELECT 
    provider,
    COUNT(*) as count,
    SUM(amount_cents) / 100.0 as mrr
FROM subscriptions
WHERE status = 'active'
GROUP BY provider;

-- Conversion funnel by platform
SELECT 
    platform,
    COUNT(DISTINCT user_id) as users,
    COUNT(DISTINCT CASE WHEN plan != 'free' THEN user_id END) as premium_users,
    COUNT(DISTINCT CASE WHEN plan != 'free' THEN user_id END) * 100.0 / 
        COUNT(DISTINCT user_id) as conversion_rate
FROM users
GROUP BY platform;
```

## Migration Guide

### For Existing Stripe Users

No action required. Existing Stripe subscriptions continue to work without modification.

### Adding RevenueCat to Existing App

1. **Database Migration**
   ```bash
   cd api
   make create-migration name=add_revenuecat_support
   # Edit migration file with schema changes
   make migrate-up
   ```

2. **Update API Services**
   - Add RevenueCatService
   - Update subscription endpoints
   - Add webhook handlers

3. **Update Mobile App**
   - Install RevenueCat SDK
   - Add provider detection
   - Implement conditional UI

4. **Configure Services**
   - Set up RevenueCat dashboard
   - Add environment variables
   - Test webhook endpoints

### Rollback Plan

If issues arise:
1. Disable RevenueCat provider detection (force Stripe)
2. Existing subscriptions continue working
3. Fix issues and re-enable gradually

## Best Practices

1. **Always verify webhook signatures** to prevent unauthorized subscription modifications
2. **Store all webhook events** for audit trail and debugging
3. **Implement idempotent webhook processing** to handle duplicate events
4. **Use provider-specific IDs** for tracking (don't mix Stripe and RevenueCat IDs)
5. **Test subscription flows end-to-end** including webhooks in staging
6. **Monitor webhook failures** and implement retry logic
7. **Keep subscription status in sync** between providers and local database

## Troubleshooting

### Common Issues

1. **User sees wrong provider**
   - Check user's country in database
   - Verify platform detection is working
   - Check provider selection endpoint response

2. **Subscription not activating**
   - Check webhook logs for errors
   - Verify webhook secret is correct
   - Ensure user IDs match between systems

3. **Purchase restoration failing**
   - Verify RevenueCat customer ID is set
   - Check RevenueCat API key configuration
   - Ensure user is logged in with correct account

### Debug Queries

```sql
-- Check user's subscription details
SELECT 
    u.email,
    u.subscription_plan,
    u.revenuecat_customer_id,
    s.*
FROM users u
LEFT JOIN subscriptions s ON s.user_id = u.id
WHERE u.email = 'user@example.com';

-- View recent webhook events
SELECT * FROM revenuecat_events 
ORDER BY created_at DESC 
LIMIT 10;

-- Check failed webhook processing
SELECT * FROM revenuecat_events 
WHERE processed = false 
ORDER BY created_at DESC;
```

## Security Considerations

1. **API Keys**: Never expose RevenueCat API keys in client code
2. **Webhook Secrets**: Use strong, unique secrets for each provider
3. **User Association**: Always verify user ownership before modifying subscriptions
4. **PII Handling**: Don't log sensitive customer information
5. **Rate Limiting**: Implement rate limits on subscription endpoints
6. **Audit Trail**: Keep detailed logs of all subscription modifications

## Future Enhancements

1. **Subscription Analytics Dashboard**
   - Real-time metrics
   - Cohort analysis
   - Churn prediction

2. **Promotional Offers**
   - RevenueCat promotional offers
   - Stripe coupon codes
   - Free trial periods

3. **Family Sharing**
   - iOS Family Sharing support
   - Google Play Family Library

4. **B2B Subscriptions**
   - Team/organization accounts
   - Volume discounts
   - Invoice billing