# Subscription System Quick Reference

## Provider Selection (TL;DR)
- **Android** → RevenueCat
- **US iOS** → Stripe  
- **Non-US iOS** → RevenueCat
- **Web** → Stripe

## Key Files

### API Backend
- `api/internal/service/revenuecat.go` - RevenueCat service
- `api/internal/service/subscription.go` - Core subscription logic
- `api/internal/http/revenuecat.go` - RevenueCat endpoints
- `api/internal/http/subscription.go` - Subscription endpoints
- `api/internal/repository/subscription.go` - Database queries
- `api/cmd/migrate/migrations/000049_add_revenuecat_support.up.sql` - Schema

### Mobile App
- `mobile/hooks/useRevenueCat.ts` - RevenueCat SDK hook
- `mobile/components/RevenueCatPaywall.tsx` - Paywall UI
- `mobile/app/settings.tsx` - Subscription management UI
- `mobile/api/revenuecat.ts` - API client functions

## Environment Variables

### API (.env)
```bash
REVENUECAT_API_KEY_IOS=appl_xxxxx
REVENUECAT_API_KEY_ANDROID=goog_xxxxx  
REVENUECAT_WEBHOOK_SECRET=your_secret_here
```

### Mobile (.env)
```bash
EXPO_PUBLIC_REVENUECAT_API_KEY_IOS=appl_xxxxx
EXPO_PUBLIC_REVENUECAT_API_KEY_ANDROID=goog_xxxxx
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/subscription/revenuecat/restore` | POST | Restore RevenueCat purchases |
| `/webhook/revenuecat` | POST | RevenueCat webhook handler |

## Database Changes

### New Tables
- `revenuecat_events` - Webhook audit trail

### Modified Tables
- `users` - Added `revenuecat_customer_id`, `platform`
- `subscriptions` - Added `provider`, `revenuecat_subscription_id`, `app_store_product_id`, `platform`

### New Enums
- `subscription_provider` - 'stripe', 'revenuecat'
- `platform_type` - 'ios', 'android', 'web'

## Testing

### Run Tests
```bash
# API tests
cd api
make test

# Specific RevenueCat tests
go test -v ./tests/integration/revenuecat_test.go
```

### Manual Testing
1. Set platform in request: `{"platform": "android"}`
2. Check provider response
3. Complete purchase flow
4. Verify webhook updates subscription

## Common Tasks

### Check User's Provider
```typescript
// Provider is now determined locally in mobile app
const { data: providerInfo } = usePaymentProvider();
// Android → RevenueCat
// US iOS → Stripe
// Non-US iOS → RevenueCat
if (providerInfo?.provider === 'revenuecat') {
  // Show RevenueCat UI
} else {
  // Show Stripe UI
}
```

### Handle RevenueCat Purchase
```typescript
const { purchasePackage } = useRevenueCat();
await purchasePackage(selectedPackage);
```

### Debug Subscription Issues
```sql
-- Check user's subscription
SELECT u.email, u.subscription_plan, s.* 
FROM users u 
LEFT JOIN subscriptions s ON s.user_id = u.id 
WHERE u.email = 'user@example.com';

-- View recent webhooks
SELECT * FROM revenuecat_events 
ORDER BY created_at DESC LIMIT 10;
```

## RevenueCat Product IDs

| Platform | Product | ID |
|----------|---------|-----|
| iOS | Monthly | `com.decorebator.premium.monthly` |
| iOS | Annual | `com.decorebator.premium.annual` |
| Android | Monthly | `premium_monthly` |
| Android | Annual | `premium_annual` |

## Webhook Event Types

### RevenueCat Events
- `INITIAL_PURCHASE` - New subscription
- `RENEWAL` - Subscription renewed
- `CANCELLATION` - User cancelled
- `UNCANCELLATION` - User reactivated
- `EXPIRATION` - Subscription expired

### Stripe Events (existing)
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`

## Troubleshooting Checklist

- [ ] Environment variables set correctly?
- [ ] RevenueCat webhook URL configured?
- [ ] Webhook secret matches?
- [ ] User has `revenuecat_customer_id`?
- [ ] Platform detected correctly?
- [ ] Products configured in RevenueCat dashboard?
- [ ] Entitlements set up properly?
- [ ] App Store / Play Store products approved?