
# Subscription System Documentation

## Overview

Decorebator implements a dual-provider subscription system that intelligently routes users to the appropriate payment provider based on their platform and location. This architecture ensures compliance with app store policies while maximizing revenue opportunities.

## Table of Contents

1. [Architecture](#architecture)
2. [Payment Providers](#payment-providers)
3. [Database Schema](#database-schema)
4. [User Flow & Lifecycle](#user-flow--lifecycle)
5. [API Endpoints](#api-endpoints)
6. [Mobile Integration](#mobile-integration)
7. [Webhook Processing](#webhook-processing)
8. [Testing](#testing)
9. [Configuration](#configuration)
10. [Monitoring & Analytics](#monitoring--analytics)
11. [Migration Guide](#migration-guide)

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
    external_event_id TEXT UNIQUE,  -- Provider-agnostic event ID (Stripe/RevenueCat)
    provider subscription_provider NOT NULL,  -- 'stripe' or 'revenuecat'
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

## User Flow & Lifecycle

### Complete Subscription Purchase Flow

#### For RevenueCat Users (Android + Non-US iOS)

1. **User taps "View Plans"** → Opens RevenueCat native paywall modal
2. **RevenueCat Paywall displays**:
   - Native App Store/Google Play pricing
   - Available packages (Monthly $6.99/Annual $69.90)
   - Premium features list with translations
   - "Restore Purchases" option
3. **User selects plan** → Native platform purchase flow begins
4. **Purchase completes** → RevenueCat SDK immediately updates customer info
5. **App syncs with backend** → JWT token refreshed with premium status
6. **Premium features activate** → User sees premium status immediately

#### For Stripe Users (US iOS + Web)

1. **User selects plan** → Pricing cards show monthly/annual options
2. **User taps "Continue to Payment"** → Redirects to Stripe checkout
3. **User completes payment** → Stripe webhook processes subscription
4. **User returns to app** → Automatic session refresh
5. **Premium status activates** → New JWT issued with subscription plan

### When Users See Premium Status

Users see their account as premium in these scenarios:

1. **Immediately after purchase** - RevenueCat SDK provides instant feedback
2. **On app focus/resume** - Automatic subscription status check
3. **JWT refresh** - New token includes updated subscription plan  
4. **Manual refresh** - Pull-to-refresh or settings screen reload

**Key Difference**: RevenueCat users see premium status **instantly** due to SDK integration, while Stripe users see it **after webhook processing** (usually within seconds).

### Payment Failure Handling

#### RevenueCat (App Store/Google Play)
- **Platform handles failures automatically**
- `BILLING_ISSUE` webhook sets subscription to `past_due` status  
- **3-day grace period implemented in backend** for consistent experience
- **User retains access during grace period**
- App stores retry failed payments automatically
- Users receive platform notifications about payment issues

#### Stripe
- Failed payment webhooks trigger **email notifications via SendGrid**
- Customer portal allows users to update payment methods
- **3-day grace period implemented in backend**
- Multiple retry attempts over 23 days before cancellation
- Email notifications at each stage of dunning sequence

### Grace Period Implementation

Both providers now benefit from a **unified 3-day grace period** implemented at the backend level:

**Technical Implementation:**
```go
// Grace period logic in subscription.IsActive()
func (s *Subscription) IsActive() bool {
    if s.Status == StatusActive {
        return true
    }
    
    // Allow access during grace period for past_due subscriptions
    if s.Status == StatusPastDue {
        gracePeriodEnd := s.CurrentPeriodEnd.Add(GracePeriodDays * 24 * time.Hour)
        return time.Now().Before(gracePeriodEnd)
    }
    
    return false
}
```

**Key Features:**
- **Duration**: 3 days from subscription's `current_period_end` date
- **Access**: Users maintain full premium features during grace period
- **Automatic Downgrade**: Backend automatically downgrades expired grace period users to free plan
- **JWT Refresh**: Tokens updated with new subscription status after grace period expires
- **Mobile Cache**: Automatic cache invalidation ensures seamless UX during status changes

**Benefits:**
- **Consistent Experience**: Same grace period across all payment providers
- **Revenue Protection**: Reduces immediate churn from temporary payment issues
- **User Satisfaction**: Fair handling of billing problems
- **Simple Implementation**: Mathematical calculation, no additional database fields required

### Subscription Renewal Process

#### Automatic Renewals

**RevenueCat:**
- App Store/Google Play handles renewals automatically
- `RENEWAL` webhook updates local subscription period
- No user action required for successful renewals
- Users get renewal notifications from app stores

**Stripe:**
- Automatic renewal handled by Stripe
- `customer.subscription.updated` webhook processes renewal
- Invoice generated and sent automatically
- SendGrid sends renewal confirmation emails

#### Near Renewal Notifications

The system includes **subscription reminder emails** via background workers:
- Reminder emails sent 3-7 days before renewal date
- Uses SendGrid for delivery via `subscription_reminder` worker queue
- Includes renewal amount and next billing date
- Configurable timing in worker configuration

### Purchase Restoration Flow

For users who reinstall the app or switch devices:

**RevenueCat Users:**
1. **"Restore Purchases" button** in paywall component
2. **Calls RevenueCat SDK** `Purchases.restorePurchases()`
3. **Syncs with backend** via `POST /subscription/revenuecat/restore` endpoint
4. **Links RevenueCat customer** to user account automatically
5. **Premium status restored** immediately in app

**Stripe Users:**
- Subscription linked to email address in database
- Login automatically restores premium status
- No additional restoration process needed

### Subscription Cancellation Scenarios

#### User-Initiated Cancellation

**RevenueCat:**
- Users cancel through App Store/Google Play settings (external to app)
- `CANCELLATION` webhook sets `cancel_at_period_end = true`
- User retains premium access until period ends
- No immediate loss of premium features

**Stripe:**
- Users cancel through customer portal or in-app cancellation
- Immediate webhook processing updates status
- Access continues until current period end
- Cancellation confirmation email sent via SendGrid

#### Failed Payment Cancellation

**RevenueCat:**
- App stores handle failed payment grace periods
- `EXPIRATION` webhook marks subscription as canceled after grace period
- Multiple platform retry attempts before final cancellation

**Stripe:**
- Failed payments trigger comprehensive dunning sequence
- Multiple retry attempts over 23-day period
- Email notifications at each dunning stage
- Final cancellation if payment not recovered

### Error Handling & Edge Cases

**Network Issues:**
- Mobile app caches purchase state locally until sync possible
- Retry mechanisms for failed API calls with exponential backoff
- Graceful degradation when offline (cached subscription status)

**Purchase Cancellation During Flow:**
- Handles user cancelling during purchase process
- No charges occur for cancelled purchases
- App handles `PURCHASE_CANCELLED_ERROR` gracefully with user-friendly messages

**Account Linking Security:**
- RevenueCat customer ID automatically linked to prevent subscription sharing
- Robust validation prevents subscription transfer between accounts
- Handles device switching and app reinstallation seamlessly

**Webhook Processing Reliability:**
- **All webhooks (Stripe + RevenueCat) processed asynchronously** via River background jobs
- **Immediate 200 OK response** prevents provider retries and webhook timeouts
- **Provider-agnostic event tracking** in unified `subscription_events` table
- Failed webhook jobs can be retried manually via River queue management
- Comprehensive audit trail with provider identification

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

### Stripe Webhook Processing (Asynchronous)

Stripe webhooks are now processed asynchronously using River workers to ensure reliability and prevent webhook timeouts.

#### HTTP Handler
```go
// internal/http/subscription.go
func HandleStripeWebhook(subService *SubscriptionService, riverClient *river.Client[pgx.Tx]) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Verify webhook signature and construct event
        event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook signature verification failed"})
            return
        }

        // 2. Enqueue the event for async processing
        _, err = riverClient.Insert(c.Request.Context(), service.StripeWebhookArgs{
            EventID:   event.ID,
            EventType: string(event.Type),
            EventData: event.Data.Raw,
        }, &river.InsertOpts{Queue: "stripe-webhook"})

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
            return
        }

        // 3. Return 200 immediately to prevent retries
        c.JSON(http.StatusOK, gin.H{"status": "success"})
    }
}
```

#### Background Worker
```go
// internal/service/stripe_webhook_worker.go
func (w *StripeWebhookWorker) Work(ctx context.Context, job *river.Job[StripeWebhookArgs]) error {
    // Reconstruct Stripe event from job args
    event := stripe.Event{
        ID:   job.Args.EventID,
        Type: stripe.EventType(job.Args.EventType),
        Data: &stripe.EventData{Raw: job.Args.EventData},
    }

    // Process based on event type with full error handling
    switch event.Type {
    case "customer.subscription.created":
        return w.subscriptionService.handleSubscriptionCreated(ctx, event)
    case "customer.subscription.updated":
        return w.subscriptionService.handleSubscriptionUpdated(ctx, event)
    case "customer.subscription.deleted":
        return w.subscriptionService.handleSubscriptionDeleted(ctx, event)
    case "invoice.payment_failed":
        return w.subscriptionService.handleInvoicePaymentFailed(ctx, event)
    }
}
```

### RevenueCat Webhook Handler

RevenueCat webhooks are handled asynchronously using a background worker to ensure reliability and prevent webhook timeouts. The HTTP handler immediately enqueues a job and returns a `200 OK` to RevenueCat to prevent retries.

#### Asynchronous Processing Flow

1. **HTTP Handler** receives webhook → Validates authorization → Enqueues River job → Returns 200 OK
2. **Background Worker** processes job → Updates subscription in database → Logs event for audit
3. **Error Handling** → Failed jobs can be manually retried → Comprehensive logging for debugging

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

#### Background Worker Implementation

```go
// internal/service/revenuecat_worker.go
type RevenueCatWebhookWorker struct {
	service RevenueCatService
}

func (w *RevenueCatWebhookWorker) Work(ctx context.Context, job *river.Job[RevenueCatWebhookArgs]) error {
	// Process the webhook payload
	return w.service.HandleWebhook(ctx, job.Args.Payload)
}
```

**Worker Queue Configuration:**
- **Stripe Queue**: `stripe-webhook` (5 max workers)
- **RevenueCat Queue**: `revenuecat-webhook` (5 max workers)
- **Retry Policy**: Exponential backoff with maximum 5 retries
- **Job Timeout**: 30 seconds
- **Duplicate Prevention**: Both workers check for already-processed events

**Supported Event Types:**

**Stripe Events:**
- `customer.subscription.created` → Creates new subscription record
- `customer.subscription.updated` → Updates subscription period and status
- `customer.subscription.deleted` → Marks subscription as canceled
- `invoice.payment_failed` → Sets status to `past_due` and sends email notifications

**RevenueCat Events:**
- `INITIAL_PURCHASE` → Creates new subscription record
- `RENEWAL` → Updates subscription period and status
- `CANCELLATION` → Sets `cancel_at_period_end = true`
- `EXPIRATION` → Marks subscription as canceled
- `BILLING_ISSUE` → Sets status to `past_due`

**Unified Event Tracking:**
- All webhook events (both Stripe and RevenueCat) are stored in the `subscription_events` table
- Provider field distinguishes between event sources
- Complete audit trail for debugging and compliance

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
- [ ] **Stripe webhooks process asynchronously** and update subscription status
- [ ] **RevenueCat webhooks process asynchronously** and update subscription status
- [ ] **Both webhook types create unified subscription events** for audit trail
- [ ] Purchase restoration works correctly
- [ ] Subscription cancellation works
- [ ] Premium features unlock after purchase
- [ ] **Webhook failures are properly retried** via River queue system

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
