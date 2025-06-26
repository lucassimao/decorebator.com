# RevenueCat Integration Documentation

This document describes the RevenueCat integration added to Decorebator to support non-US and Android subscriptions alongside the existing Stripe integration.

## Overview

The integration implements a dual-provider subscription system:
- **US iOS users**: Continue using Stripe for subscriptions
- **Non-US iOS users**: Use RevenueCat (App Store)
- **Android users**: Use RevenueCat (Google Play Store)

## Architecture Changes

### Database Schema

Added new fields and tables to support RevenueCat:

#### New Columns
- `users.revenuecat_customer_id` - Links user to RevenueCat customer
- `users.platform` - Tracks user's platform (ios/android/web)
- `subscriptions.provider` - Identifies payment provider (stripe/revenuecat)
- `subscriptions.revenuecat_subscription_id` - RevenueCat subscription identifier
- `subscriptions.app_store_product_id` - App Store/Google Play product ID
- `subscriptions.platform` - Platform where subscription was created

#### New Tables
- `revenuecat_events` - Webhook event audit trail for RevenueCat

#### New Types
- `subscription_provider` - ENUM ('stripe', 'revenuecat')
- `platform_type` - ENUM ('ios', 'android', 'web')

### API Backend

#### New Services
- **RevenueCatService** (`internal/service/revenuecat.go`)
  - Handles RevenueCat API interactions
  - Processes webhooks
  - Manages customer info sync
  - Supports purchase restoration

#### New Endpoints
- `POST /webhook/revenuecat` - RevenueCat webhook handler
- `POST /subscription/provider` - Determines which payment provider to use
- `POST /subscription/revenuecat/restore` - Restore purchases from RevenueCat

#### Modified Services
- **SubscriptionService** - Updated to handle cancellations based on provider
- **Subscription Repository** - Extended queries to include new fields

### Mobile App

#### New Dependencies
- `react-native-purchases` - RevenueCat SDK
- `react-native-purchases-ui` - RevenueCat UI components

#### New Components
- **RevenueCatPaywall** - Native paywall UI for RevenueCat subscriptions
- **useRevenueCat Hook** - Manages RevenueCat SDK initialization and purchases
- **usePaymentProvider Hook** - Determines which payment provider to use

#### Modified Screens
- **Settings Screen** - Conditionally shows Stripe checkout or RevenueCat paywall

## Configuration

### Environment Variables

#### API Backend
```env
# RevenueCat Configuration
REVENUECAT_API_KEY_IOS=your_revenuecat_ios_api_key
REVENUECAT_API_KEY_ANDROID=your_revenuecat_android_api_key
REVENUECAT_WEBHOOK_SECRET=your_revenuecat_webhook_secret
```

#### Mobile App
```env
# RevenueCat Configuration
EXPO_PUBLIC_REVENUECAT_API_KEY_IOS=your-revenuecat-ios-api-key
EXPO_PUBLIC_REVENUECAT_API_KEY_ANDROID=your-revenuecat-android-api-key
```

## Provider Selection Logic

The system determines which payment provider to use based on:

1. **Platform**: Android always uses RevenueCat
2. **Location**: iOS users in the US use Stripe, others use RevenueCat
3. **Web**: Always uses Stripe

```typescript
// Provider selection logic
if (platform === 'android') return 'revenuecat';
if (platform === 'ios' && country !== 'US') return 'revenuecat';
return 'stripe';
```

## RevenueCat Setup

### Prerequisites
1. Create a RevenueCat account
2. Configure iOS and Android apps in RevenueCat dashboard
3. Create products in App Store Connect and Google Play Console
4. Configure products and entitlements in RevenueCat
5. Set up webhook endpoint in RevenueCat dashboard

### Product Configuration
Configure these product IDs in RevenueCat:
- iOS Monthly: `com.decorebator.premium.monthly`
- iOS Annual: `com.decorebator.premium.annual`
- Android Monthly: `premium_monthly`
- Android Annual: `premium_annual`

### Entitlement Configuration
Create a "premium" entitlement that includes all premium products.

### Webhook Configuration
1. In RevenueCat dashboard, add webhook endpoint: `https://your-api.com/webhook/revenuecat`
2. Add Authorization header with your webhook secret
3. Select all relevant events to monitor

## User Flow

### New Subscription Flow
1. User opens settings and taps upgrade
2. App checks payment provider based on platform/location
3. For RevenueCat users:
   - Show RevenueCat paywall
   - User selects plan and completes purchase
   - RevenueCat SDK updates customer info
   - App syncs with backend
4. For Stripe users:
   - Show plan selection
   - Redirect to Stripe checkout
   - Process webhook on completion

### Purchase Restoration
1. User taps "Restore Purchases"
2. RevenueCat SDK restores purchases
3. App syncs restored purchases with backend
4. User's subscription status is updated

## Testing

### Integration Tests
Added integration tests in `api/tests/integration/revenuecat_test.go`:
- Provider detection based on platform/location
- Webhook processing
- Purchase restoration endpoint

### Manual Testing
1. Test with sandbox accounts on iOS/Android
2. Verify provider selection logic
3. Test purchase flow for each provider
4. Verify webhook processing
5. Test purchase restoration

## Migration Notes

- Existing Stripe subscriptions remain unchanged
- No automatic migration of existing subscriptions
- Users must cancel Stripe subscription before purchasing through RevenueCat

## Security Considerations

1. **Webhook Verification**: Validate webhook authorization header
2. **Customer Linking**: Link RevenueCat customers to user accounts
3. **Platform Validation**: Verify platform claims from client
4. **Rate Limiting**: Apply appropriate rate limits to endpoints

## Monitoring

1. **Webhook Events**: Track all RevenueCat events in `revenuecat_events` table
2. **Provider Usage**: Monitor which providers are being used
3. **Failed Purchases**: Track and investigate failed purchase attempts
4. **Subscription Status**: Ensure consistency between RevenueCat and local state

## Future Considerations

1. **Price Localization**: RevenueCat supports local currency pricing
2. **Promotional Offers**: Can be configured in RevenueCat dashboard
3. **Analytics**: RevenueCat provides detailed subscription analytics
4. **A/B Testing**: RevenueCat supports paywall experiments