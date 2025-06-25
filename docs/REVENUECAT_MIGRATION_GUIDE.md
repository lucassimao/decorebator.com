# RevenueCat Integration Migration Guide

This guide walks through adding RevenueCat support to the existing Decorebator subscription system.

## Prerequisites

1. RevenueCat account with iOS and Android apps configured
2. App Store Connect and Google Play Console access
3. In-app products created in both stores
4. Development environment set up

## Step 1: Database Migration

Run the migration to add RevenueCat support:

```bash
cd api
make migrate-up
```

This adds:
- `subscription_provider` enum type
- `platform_type` enum type
- `revenuecat_events` table for webhook tracking
- New columns to `users` and `subscriptions` tables

## Step 2: Configure RevenueCat Dashboard

### 2.1 Create Apps
1. Go to RevenueCat dashboard
2. Create new project
3. Add iOS app with bundle ID: `com.decorebator.app`
4. Add Android app with package name: `com.decorebator.app`

### 2.2 Configure Products

#### iOS Products
1. In App Store Connect, create:
   - Monthly: `com.decorebator.premium.monthly` ($6.99)
   - Annual: `com.decorebator.premium.annual` ($69.90)

2. In RevenueCat, import these products

#### Android Products
1. In Google Play Console, create:
   - Monthly: `premium_monthly` ($6.99)
   - Annual: `premium_annual` ($69.90)

2. In RevenueCat, import these products

### 2.3 Create Entitlement
1. Go to Entitlements in RevenueCat
2. Create entitlement named `premium`
3. Attach all 4 products to this entitlement

### 2.4 Configure Webhook
1. Go to Integrations → Webhooks
2. Add webhook URL: `https://api.decorebator.com/webhook/revenuecat`
3. Add Authorization header with your secret
4. Select all event types

### 2.5 Get API Keys
1. Go to API Keys section
2. Copy iOS SDK key (starts with `appl_`)
3. Copy Android SDK key (starts with `goog_`)

## Step 3: Backend Configuration

### 3.1 Environment Variables

Add to `.env`:
```bash
REVENUECAT_API_KEY_IOS=appl_your_ios_key
REVENUECAT_API_KEY_ANDROID=goog_your_android_key
REVENUECAT_WEBHOOK_SECRET=your_webhook_secret
```

### 3.2 Verify Services

The following services are already implemented:
- `RevenueCatService` for API interactions
- Webhook handler at `/webhook/revenuecat`
- Purchase restoration at `/subscription/revenuecat/restore`
- Provider detection is handled locally in mobile app (no API endpoint needed)

## Step 4: Mobile App Configuration

### 4.1 Environment Variables

Add to `mobile/.env`:
```bash
EXPO_PUBLIC_REVENUECAT_API_KEY_IOS=appl_your_ios_key
EXPO_PUBLIC_REVENUECAT_API_KEY_ANDROID=goog_your_android_key
```

### 4.2 Install Dependencies

Dependencies are already in `package.json`:
```bash
cd mobile
npm install
```

### 4.3 iOS Configuration

1. Run `npx pod-install` to install native dependencies
2. Add to `Info.plist` if needed:
```xml
<key>SKAdNetworkItems</key>
<array>
  <dict>
    <key>SKAdNetworkIdentifier</key>
    <string>v79kvwwj4g.skadnetwork</string>
  </dict>
</array>
```

### 4.4 Android Configuration

No additional configuration needed - RevenueCat SDK handles everything.

## Step 5: Testing

### 5.1 Test Provider Detection

Provider detection is now handled locally in the mobile app:
- Android users automatically use RevenueCat
- US iOS users use Stripe
- Non-US iOS users use RevenueCat

No API endpoint is needed for provider detection.

### 5.2 Test Purchase Flow

1. **iOS Testing**:
   - Use sandbox test account
   - Install app via TestFlight
   - Complete purchase flow
   - Verify webhook received

2. **Android Testing**:
   - Add test account in Play Console
   - Install app via internal testing track
   - Complete purchase flow
   - Verify webhook received

### 5.3 Test Webhook

Send test webhook:
```bash
curl -X POST https://api.decorebator.com/webhook/revenuecat \
  -H "Authorization: your_webhook_secret" \
  -H "Content-Type: application/json" \
  -d '{
    "api_version": "1.0",
    "event": {
      "type": "TEST",
      "id": "test_123",
      "app_user_id": "123"
    }
  }'
```

### 5.4 Verify Database

```sql
-- Check webhook was recorded
SELECT * FROM revenuecat_events ORDER BY created_at DESC LIMIT 1;

-- Check user subscription
SELECT * FROM subscriptions WHERE user_id = 123;
```

## Step 6: Monitoring

### 6.1 Set Up Alerts

Monitor for:
- Webhook failures
- Purchase failures
- Provider detection errors

### 6.2 Analytics Queries

```sql
-- Subscriptions by provider
SELECT provider, COUNT(*) 
FROM subscriptions 
WHERE status = 'active' 
GROUP BY provider;

-- Recent RevenueCat events
SELECT event_type, COUNT(*) 
FROM revenuecat_events 
WHERE created_at > NOW() - INTERVAL '24 hours' 
GROUP BY event_type;
```

## Step 7: Go Live

### 7.1 Pre-Launch Checklist
- [ ] Products approved in App Store Connect
- [ ] Products approved in Google Play Console
- [ ] RevenueCat webhook tested
- [ ] Environment variables set in production
- [ ] Database migrations run in production

### 7.2 Gradual Rollout
1. Enable for internal testers first
2. Monitor for 24 hours
3. Enable for 10% of users
4. Monitor for 1 week
5. Enable for all users

### 7.3 Rollback Plan
If issues arise:
1. Force all users to Stripe by modifying provider detection
2. Fix issues
3. Re-enable RevenueCat gradually

## Common Issues & Solutions

### Issue: Webhook not receiving events
**Solution**: Check Authorization header matches exactly

### Issue: User not getting premium after purchase
**Solution**: Verify user ID linkage between RevenueCat and database

### Issue: Wrong provider selected
**Solution**: Check user's country and platform in database

### Issue: Purchase restoration failing
**Solution**: Ensure RevenueCat customer ID is set for user

## Support Resources

- RevenueCat Documentation: https://docs.revenuecat.com
- RevenueCat Support: support@revenuecat.com
- Internal Docs: [Subscription System Documentation](SUBSCRIPTION_SYSTEM.md)