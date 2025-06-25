
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
  const { userInfo: user } = useUserInfo();

  // Determine provider based on platform and user location
  const getProvider = (): {
    provider: PaymentProvider;
    platform: "ios" | "android" | "web";
  } | null => {
    if (!user) return null;

    const platform =
      Platform.OS === "ios"
        ? "ios"
        : Platform.OS === "android"
          ? "android"
          : "web";

    // Android always uses RevenueCat
    if (platform === "android") {
      return { provider: "revenuecat", platform };
    }

    // iOS: US users use Stripe, others use RevenueCat
    if (platform === "ios") {
      const isUS = user.country === "US";
      return { provider: isUS ? "stripe" : "revenuecat", platform };
    }

    // Web always uses Stripe
    return { provider: "stripe", platform };
  };

  const providerInfo = getProvider();

  return {
    data: providerInfo,
    isLoading: false,
    error: null,
  };
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

RevenueCat webhooks are handled asynchronously using a background worker to ensure reliability. The HTTP handler immediately enqueues a job and returns a `200 OK` to RevenueCat.

```go
// internal/http/revenuecat.go
func HandleRevenueCatWebhook(riverClient *river.Client[pgx.Tx]) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Verify Authorization header
		authHeader := c.GetHeader("Authorization")
		expectedToken := os.Getenv("REVENUECAT_WEBHOOK_AUTHORIZATION")

		if expectedToken == "" {
			common.Logger.Error("REVENUECAT_WEBHOOK_AUTHORIZATION is not set, skipping validation")
		} else if authHeader != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			return
		}

		// 2. Read the request body
		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Enqueue a job to process the webhook
		_, err = riverClient.Insert(c.Request.Context(), service.RevenueCatWebhookArgs{
			Payload:    payload,
		}, &river.InsertOpts{
			Queue: "revenuecat-webhook",
		})

		if err != nil {
			common.Logger.Error("failed to enqueue revenuecat webhook job", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue webhook job"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}
```

## Testing

### Integration Tests

The RevenueCat integration is tested using a mock RevenueCat service that implements the `RevenueCatService` interface. This allows for testing the webhook processing logic without making actual calls to the RevenueCat API.

```go
// tests/integration/revenuecat_test.go
func TestRevenueCatIntegration(t *testing.T) {
    // ... test setup ...

	// Create a mock RevenueCat service.
	rcServiceMock := &mocks.RevenueCatServiceMock{}

	// Set the mock functions.
	rcServiceMock.HandleWebhookFunc = func(ctx context.Context, payload []byte) error {
		// Simulate the real service's behavior by creating the subscription directly.
		// ... implementation ...
	}

	// Process the job that was just enqueued by the webhook.
	ts.ProcessNextRiverJob(t, "revenuecat-webhook", rcServiceMock)

    // ... assertions ...
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
REVENUECAT_API_KEY=your_revenuecat_api_key
REVENUECAT_WEBHOOK_AUTHORIZATION=your_webhook_authorization_token
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
   - Authorization Header: Your webhook authorization token

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
2. **Webhook Authorization**: Use a strong, unique token for webhook authorization
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
