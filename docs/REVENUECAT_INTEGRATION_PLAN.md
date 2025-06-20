# RevenueCat Integration Plan for iOS Non-US Users

## Overview

This document outlines the plan to integrate RevenueCat as an alternative subscription management system for iOS users outside the United States, while maintaining the existing Stripe integration for US users and all other platforms.

## Current Subscription Architecture

### Backend (Go/Gin)
- **Subscription Service**: `internal/service/subscription.go` - Handles Stripe checkout, webhooks, and subscription management
- **HTTP Routes**: `internal/http/subscription.go` - REST endpoints for subscription operations
- **Database Models**: Plans table with pricing, subscription tracking, Stripe ID mapping
- **Webhook Handler**: `POST /webhook/stripe` - Processes Stripe webhook events
- **Middleware**: Subscription enforcement in various endpoints

### Mobile (React Native/Expo)
- **Subscription API**: `mobile/api/subscription.ts` - TypeScript functions for subscription operations
- **Settings Screen**: `mobile/app/settings.tsx` - Full subscription management UI
- **Upgrade Modals**: Platform-aware subscription limit enforcement
- **Web Browser Flow**: Uses `expo-auth-session` for Stripe checkout

### Current Flow
1. User selects plan → Mobile calls `/subscription/checkout-session`
2. Backend creates Stripe session → Mobile opens web browser checkout
3. User completes payment → Stripe webhooks update backend
4. Mobile refreshes status → JWT updated with subscription plan

## Integration Requirements

### Business Logic
- **Geographic Routing**: iOS users outside US → RevenueCat, all others → Stripe
- **Unified Subscription Model**: Same subscription plans and limits regardless of provider
- **Seamless Migration**: Existing US iOS users continue with Stripe
- **Fallback Strategy**: Stripe as fallback if RevenueCat fails

### Technical Requirements
- **Provider Abstraction**: Clean separation between Stripe and RevenueCat logic
- **Webhook Handling**: Support both Stripe and RevenueCat webhook formats
- **User Identification**: Map RevenueCat customer IDs to Decorebator users
- **Platform Detection**: Accurate iOS + region detection in mobile app
- **Testing Strategy**: Comprehensive testing for both subscription providers

## Implementation Plan

### Phase 1: Backend Infrastructure (Week 1)

#### 1.1 Database Schema Changes
```sql
-- Add provider tracking to subscriptions table
ALTER TABLE subscriptions ADD COLUMN provider VARCHAR(20) DEFAULT 'stripe';
ALTER TABLE subscriptions ADD COLUMN external_customer_id VARCHAR(255);
ALTER TABLE subscriptions ADD COLUMN platform VARCHAR(20);

-- Create RevenueCat webhook events table
CREATE TABLE revenuecat_webhook_events (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    user_id BIGINT REFERENCES users(id),
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    data JSONB NOT NULL
);
```

#### 1.2 Subscription Provider Abstraction
```go
// internal/service/subscription_provider.go
type SubscriptionProvider interface {
    CreateCheckoutSession(userID int64, planType string) (*CheckoutSession, error)
    ProcessWebhook(payload []byte, signature string) error
    GetSubscriptionStatus(customerID string) (*SubscriptionStatus, error)
    CancelSubscription(customerID string) error
}

type StripeProvider struct {
    client *stripe.Client
    config *StripeConfig
}

type RevenueCatProvider struct {
    client *revenuecat.Client
    config *RevenueCatConfig
}
```

#### 1.3 Enhanced Subscription Service
```go
// internal/service/subscription.go
type SubscriptionService struct {
    stripeProvider    SubscriptionProvider
    revenuecatProvider SubscriptionProvider
    userRepo          repository.UserRepository
    subscriptionRepo  repository.SubscriptionRepository
}

func (s *SubscriptionService) CreateCheckoutSession(userID int64, planType string) (*CheckoutSession, error) {
    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        return nil, err
    }
    
    provider := s.selectProvider(user)
    return provider.CreateCheckoutSession(userID, planType)
}

func (s *SubscriptionService) selectProvider(user *models.User) SubscriptionProvider {
    if user.Platform == "ios" && !s.isUSUser(user) {
        return s.revenuecatProvider
    }
    return s.stripeProvider
}
```

#### 1.4 RevenueCat Webhook Handler
```go
// internal/http/revenuecat_webhook.go
func (h *RevenueCatWebhookHandler) HandleWebhook(c *gin.Context) {
    payload, err := io.ReadAll(c.Request.Body)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
        return
    }
    
    event, err := h.revenuecatService.ParseWebhook(payload)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook"})
        return
    }
    
    if err := h.subscriptionService.ProcessRevenueCatEvent(event); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"status": "success"})
}
```

### Phase 2: Mobile App Enhancement (Week 2)

#### 2.1 Platform and Region Detection
```typescript
// mobile/utils/subscriptionProvider.ts
export interface UserLocation {
  country: string;
  isUS: boolean;
}

export interface PlatformInfo {
  platform: 'ios' | 'android' | 'web';
  isNativeApp: boolean;
}

export function shouldUseRevenueCat(
  platformInfo: PlatformInfo, 
  userLocation: UserLocation
): boolean {
  return platformInfo.platform === 'ios' && 
         platformInfo.isNativeApp && 
         !userLocation.isUS;
}

export async function detectUserLocation(): Promise<UserLocation> {
  // Implementation using expo-localization or IP-based detection
  const locale = Localization.getLocales()?.[0];
  const region = locale?.region || 'US';
  
  return {
    country: region,
    isUS: region === 'US'
  };
}
```

#### 2.2 Enhanced Subscription API
```typescript
// mobile/api/subscription.ts
export interface SubscriptionProvider {
  createCheckoutSession(planType: PlanType): Promise<CheckoutSession>;
  getSubscriptionStatus(): Promise<SubscriptionStatus>;
  cancelSubscription(): Promise<void>;
}

class StripeSubscriptionProvider implements SubscriptionProvider {
  async createCheckoutSession(planType: PlanType): Promise<CheckoutSession> {
    const response = await api.post('/subscription/checkout-session', {
      planType,
      provider: 'stripe'
    });
    return response.data;
  }
}

class RevenueCatSubscriptionProvider implements SubscriptionProvider {
  async createCheckoutSession(planType: PlanType): Promise<CheckoutSession> {
    // RevenueCat uses native iOS store interface
    const offerings = await Purchases.getOfferings();
    const package = offerings.current?.availablePackages.find(
      p => p.identifier === planType
    );
    
    if (!package) throw new Error('Plan not available');
    
    const purchaseResult = await Purchases.purchasePackage(package);
    return {
      provider: 'revenuecat',
      customerInfo: purchaseResult.customerInfo
    };
  }
}

export function createSubscriptionProvider(
  platformInfo: PlatformInfo,
  userLocation: UserLocation
): SubscriptionProvider {
  if (shouldUseRevenueCat(platformInfo, userLocation)) {
    return new RevenueCatSubscriptionProvider();
  }
  return new StripeSubscriptionProvider();
}
```

#### 2.3 Updated Settings Screen
```typescript
// mobile/app/settings.tsx
export default function SettingsScreen() {
  const [platformInfo] = useState<PlatformInfo>({
    platform: Platform.OS as 'ios' | 'android',
    isNativeApp: Constants.appOwnership !== 'expo'
  });
  
  const [userLocation, setUserLocation] = useState<UserLocation | null>(null);
  const [subscriptionProvider, setSubscriptionProvider] = useState<SubscriptionProvider | null>(null);
  
  useEffect(() => {
    (async () => {
      const location = await detectUserLocation();
      setUserLocation(location);
      
      const provider = createSubscriptionProvider(platformInfo, location);
      setSubscriptionProvider(provider);
    })();
  }, []);
  
  const handleUpgrade = async (planType: PlanType) => {
    if (!subscriptionProvider) return;
    
    try {
      if (shouldUseRevenueCat(platformInfo!, userLocation!)) {
        // Native iOS purchase flow
        await subscriptionProvider.createCheckoutSession(planType);
        // RevenueCat automatically updates backend via webhooks
        queryClient.invalidateQueries(['subscription-status']);
      } else {
        // Web-based Stripe checkout
        const session = await subscriptionProvider.createCheckoutSession(planType);
        await WebBrowser.openBrowserAsync(session.url);
      }
    } catch (error) {
      console.error('Subscription error:', error);
      Alert.alert('Error', 'Failed to process subscription');
    }
  };
}
```

### Phase 3: Configuration and Environment Setup (Week 2)

#### 3.1 Environment Variables
```bash
# .env additions
REVENUECAT_API_KEY=rc_live_xxxxx
REVENUECAT_WEBHOOK_SECRET=xxxxx
REVENUECAT_PUBLIC_KEY=xxxxx

# Product IDs for RevenueCat
REVENUECAT_MONTHLY_PRODUCT_ID=decorebator_monthly
REVENUECAT_ANNUAL_PRODUCT_ID=decorebator_annual
```

#### 3.2 Mobile Environment Configuration
```typescript
// mobile/config/subscription.ts
export const SUBSCRIPTION_CONFIG = {
  stripe: {
    publicKey: process.env.EXPO_PUBLIC_STRIPE_KEY!,
  },
  revenuecat: {
    apiKey: process.env.EXPO_PUBLIC_REVENUECAT_KEY!,
    products: {
      monthly: 'decorebator_monthly',
      annual: 'decorebator_annual'
    }
  }
};
```

### Phase 4: Testing Strategy (Week 3)

#### 4.1 Backend Testing
```go
// tests/integration/subscription_providers_test.go
func TestSubscriptionProviders_CreateCheckoutSession(t *testing.T) {
    tests := []struct {
        name     string
        userCountry string
        platform string
        expectedProvider string
    }{
        {"US iOS User", "US", "ios", "stripe"},
        {"EU iOS User", "DE", "ios", "revenuecat"},
        {"US Android User", "US", "android", "stripe"},
        {"EU Android User", "DE", "android", "stripe"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### 4.2 Mobile Testing
```typescript
// mobile/__tests__/subscription/providerSelection.test.ts
describe('Subscription Provider Selection', () => {
  it('should use RevenueCat for non-US iOS users', () => {
    const platformInfo = { platform: 'ios', isNativeApp: true };
    const userLocation = { country: 'DE', isUS: false };
    
    expect(shouldUseRevenueCat(platformInfo, userLocation)).toBe(true);
  });
  
  it('should use Stripe for US iOS users', () => {
    const platformInfo = { platform: 'ios', isNativeApp: true };
    const userLocation = { country: 'US', isUS: true };
    
    expect(shouldUseRevenueCat(platformInfo, userLocation)).toBe(false);
  });
});
```

### Phase 5: Deployment and Monitoring (Week 4)

#### 5.1 Feature Flags
```go
// internal/config/feature_flags.go
type FeatureFlags struct {
    RevenueCatEnabled bool `env:"REVENUECAT_ENABLED" envDefault:"false"`
    RevenueCatRegions []string `env:"REVENUECAT_REGIONS" envDefault:"EU,CA,AU"`
}
```

#### 5.2 Monitoring and Analytics
```go
// internal/analytics/subscription_analytics.go
func (a *Analytics) TrackSubscriptionCreated(userID int64, provider string, planType string) {
    a.track("subscription_created", map[string]interface{}{
        "user_id": userID,
        "provider": provider,
        "plan_type": planType,
        "timestamp": time.Now(),
    })
}
```

## Migration Strategy

### Gradual Rollout
1. **Week 1**: Backend infrastructure ready, feature flag disabled
2. **Week 2**: Mobile app updated, RevenueCat SDK integrated
3. **Week 3**: Enable for beta users in select EU countries
4. **Week 4**: Full rollout to all non-US iOS users

### Risk Mitigation
- **Fallback Logic**: If RevenueCat fails, fallback to Stripe web checkout
- **Monitoring**: Comprehensive logging for subscription events
- **Testing**: Extensive testing in sandbox environments
- **Gradual Rollout**: Feature flags for controlled deployment

## API Changes

### New Endpoints
```
POST /subscription/revenuecat/webhook - RevenueCat webhook handler
GET /subscription/provider-config - Get subscription provider configuration
POST /subscription/validate-receipt - iOS receipt validation (if needed)
```

### Modified Endpoints
```
POST /subscription/checkout-session - Add provider parameter
GET /subscription/status - Include provider information
```

## Dependencies

### Backend
```go
// go.mod additions
github.com/revenuecat/revenuecat-go v1.0.0
```

### Mobile
```json
// package.json additions
{
  "react-native-purchases": "^6.0.0",
  "@react-native-async-storage/async-storage": "^1.19.0"
}
```

## Success Metrics

### Technical Metrics
- **Provider Selection Accuracy**: 100% correct provider selection based on user location/platform
- **Webhook Processing**: <1s average webhook processing time
- **Error Rate**: <1% subscription creation failure rate
- **Fallback Usage**: <5% fallback to Stripe for non-US iOS users

### Business Metrics
- **Conversion Rate**: Monitor subscription conversion rates by provider
- **Revenue Impact**: Track revenue changes from iOS international users
- **User Experience**: Monitor support tickets related to subscription issues

## Rollback Plan

### Immediate Rollback
1. Set feature flag `REVENUECAT_ENABLED=false`
2. All users route to Stripe web checkout
3. Monitor for any subscription disruptions

### Data Cleanup
1. RevenueCat webhook events remain in database for audit
2. Subscription records maintain provider information
3. No data loss during rollback

## Security Considerations

### RevenueCat Integration
- **Webhook Signature Validation**: Verify all incoming webhooks
- **API Key Management**: Secure storage of RevenueCat API keys
- **Receipt Validation**: Server-side validation of iOS receipts
- **Customer ID Mapping**: Secure mapping between RevenueCat and Decorebator users

### Privacy Compliance
- **GDPR Compliance**: Ensure RevenueCat integration meets EU privacy requirements
- **Data Minimization**: Only collect necessary subscription data
- **User Consent**: Clear communication about subscription provider selection

## Documentation Updates

### Developer Documentation
- Update API documentation with new endpoints
- Add RevenueCat integration guide
- Update deployment documentation

### User Documentation
- Update subscription FAQ
- Add platform-specific subscription guides
- Update troubleshooting documentation

---

**Estimated Timeline**: 4 weeks
**Risk Level**: Medium (new payment provider integration)
**Dependencies**: RevenueCat account setup, App Store Connect configuration
**Team Requirements**: Backend developer, Mobile developer, QA engineer

This plan provides a comprehensive roadmap for integrating RevenueCat while maintaining the existing Stripe functionality and ensuring a smooth user experience across all platforms and regions.