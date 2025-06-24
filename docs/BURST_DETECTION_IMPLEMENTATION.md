# Burst Detection Implementation

This document summarizes the comprehensive burst detection and anti-abuse system implemented for the Decorebator API.

## Overview

The burst detection system protects the API from abuse by implementing progressive rate limiting and temporary account suspensions. It uses Redis for tracking violations and implements a fair, deterministic 24-hour blocking system.

## Key Features

### 1. **24-Hour Deterministic Blocking**
- Users are blocked for exactly 24 hours from the time of violation
- No more variable block durations based on time of day
- Consistent user experience regardless of when violations occur

### 2. **Progressive Enforcement**
- **First violation**: Warning (HTTP 429 with `BURST_WARNING`)
- **Second violation**: 24-hour suspension (HTTP 403 with `ACCOUNT_SUSPENDED_BURST`)
- Clear error messages with restoration times

### 3. **Admin Management Endpoints**
- Unblock specific users by ID
- Unblock all users at once
- View currently blocked users
- Protected by static authentication

## Architecture

### Core Components

#### BurstDetector Service (`internal/service/burst_detector.go`)
```go
type BurstDetector struct {
    redis *redis.Client
}

// Thresholds per endpoint per minute
var BurstThresholds = map[string]int64{
    "word_create":     50, // 50 words in 1 minute
    "wordlist_create": 10, // 10 wordlists in 1 minute
    "error_report":    20, // 20 error reports in 1 minute
}
```

**Key Methods:**
- `CheckAndTrackBurst()` - Detects and tracks violations
- `BlockUserFor24Hours()` - Blocks user for exactly 24 hours
- `UnblockUser()` - Admin function to remove blocks
- `UnblockAllUsers()` - Admin function to remove all blocks
- `GetBlockedUsers()` - Returns list of blocked users

#### AntiBurstMiddleware (`internal/http/anti_burst.go`)
- Intercepts protected endpoints
- Checks for existing blocks
- Implements progressive blocking logic
- Sends email notifications and Sentry alerts

### Protected Endpoints
1. `POST /wordlists` - Creating wordlists (10/minute threshold)
2. `POST /wordlists/:id/words` - Adding words (50/minute threshold)
3. `POST /errorReports` - Error reporting (20/minute threshold)

### Redis Key Structure

#### Block Tracking
```
blocked:{userID} -> Hash containing:
  - blocked_at: Unix timestamp when blocked
  - blocked_until: Unix timestamp when block expires
  - TTL: 24 hours
```

#### Violation Tracking
```
violations:{userID}:{YYYY-MM-DD} -> Count of violations today
burst:{userID}:{endpoint} -> Request count in current minute window
```

## Admin Endpoints

### Authentication
All admin endpoints use static authentication via `STATIC_TOKEN` environment variable.

### Available Endpoints

#### 1. Get Blocked Users
```http
GET /static/admin/burst/blocked-users
Authorization: Bearer <STATIC_TOKEN>
```

**Response:**
```json
{
  "blocked_users": [
    {
      "user_id": 123,
      "blocked_until": "2025-06-25T15:40:59-03:00"
    }
  ],
  "total_count": 1
}
```

#### 2. Unblock Specific User
```http
POST /static/admin/burst/unblock/123
Authorization: Bearer <STATIC_TOKEN>
```

**Response:**
```json
{
  "message": "User unblocked successfully",
  "user_id": 123
}
```

#### 3. Unblock All Users
```http
POST /static/admin/burst/unblock-all
Authorization: Bearer <STATIC_TOKEN>
```

**Response:**
```json
{
  "message": "All users unblocked successfully",
  "unblocked_count": 5
}
```

## Email Notifications

When users are suspended, they receive:
- **HTML email** with detailed suspension information
- **Activity type** that triggered the suspension
- **Violation count** for the day
- **Exact restoration time** (24 hours from block)
- **Appeal instructions** for legitimate users

**Template:** `internal/mail/burst_blocked.html`

**Environment Handling:** Emails are only sent in production (`ENV=production`)

## Mobile App Integration

### Error Handling
The mobile app includes specialized handling for burst violations:

```typescript
// Custom error class
class BurstViolationError extends Error {
  public errorCode: "BURST_WARNING" | "ACCOUNT_SUSPENDED_BURST";
  public suspendedUntil?: Date;
  public retryAfter?: number;
}
```

### Localized Messages
Burst error messages are translated into 8 languages:
- English, German, Spanish, French, Italian, Japanese, Portuguese (BR), Portuguese (PT)

**Key translations:**
- Warning messages for first violations
- Suspension messages with restoration times
- Updated to reflect 24-hour blocking instead of "midnight" restoration

## Test Configuration

### Problem Solved
The original burst detector tests included a 61-second sleep that made the entire test suite slow and affected other tests.

### Solution: Configurable Burst Detection

#### 1. Config Flag
```go
type Config struct {
    // ... other fields
    DisableBurstDetector bool // Skip burst detection (useful for tests)
}
```

#### 2. Conditional Middleware
```go
// Package-level variable for clean override
var antiBurstMiddleware = func(endpoint string) gin.HandlerFunc {
    return AntiBurstMiddleware(endpoint)
}

// In SetupRoutes - override for tests
if config != nil && config.DisableBurstDetector {
    antiBurstMiddleware = func(endpoint string) gin.HandlerFunc {
        return func(c *gin.Context) {
            c.Next() // No-op middleware
        }
    }
}
```

#### 3. Test Server Configuration
```go
// NewTestServer defaults to disabled burst detection
testConfig := &httphandlers.Config{
    Database:             db,
    DisableBurstDetector: true, // Fast tests by default
}
```

### Benefits
- ✅ **All tests run fast** (no burst detection overhead)
- ✅ **Burst functionality still testable** when specifically enabled
- ✅ **Clean, maintainable code** (no conditional function wrappers)
- ✅ **Easy to enable/disable** per test case

## Implementation Timeline

### Initial Implementation
- 24-hour deterministic blocking system
- Progressive violation tracking
- Email notifications with proper templating
- Mobile app error handling with localization

### Admin Endpoints
- Unblock specific users by ID
- Unblock all users functionality
- List currently blocked users
- Static authentication protection

### Test Optimization
- Configurable burst detection for tests
- Clean variable override pattern
- Fast test suite without affecting functionality

## Files Modified

### Core Implementation
- `api/internal/service/burst_detector.go` - Core detection logic
- `api/internal/http/anti_burst.go` - Middleware implementation
- `api/internal/http/setup.go` - Route configuration and test overrides

### Admin Endpoints
- `api/internal/http/burst_admin.go` - Admin endpoint handlers
- `api/doc/burst-admin.http` - API documentation

### Email System
- `api/internal/mail/burst_blocked.html` - Email template
- `api/internal/mail/mail.go` - Email sending logic

### Mobile Integration
- `mobile/api/burst.ts` - Error handling classes
- `mobile/api/api.ts` - API client integration
- `mobile/components/dashboard/` - UI error handling
- `mobile/i18n/locales/*.json` - Localized messages (8 languages)

### Testing
- `api/tests/integration/burst_detector_test.go` - Comprehensive tests
- `api/tests/integration/setup/test_server.go` - Test configuration

## Configuration

### Environment Variables
- `STATIC_TOKEN` - Admin endpoint authentication
- `ENV` - Controls email sending (production only)
- `REDIS_URL` - Redis connection for tracking
- `DATABASE_URL` - PostgreSQL for persistent data

### Redis Dependencies
- Used for violation tracking and temporary blocking
- Graceful degradation if Redis is unavailable
- Automatic key expiration for cleanup

## Monitoring

### Sentry Integration
Burst violations are logged to Sentry with:
- User information (ID, email)
- Violation details (endpoint, count)
- Block duration and expiry time
- Alert classification for monitoring

### Logging
Comprehensive logging throughout the system:
- Burst detection events
- Block/unblock operations
- Email sending status
- Admin endpoint usage

## Security Considerations

1. **Rate Limiting**: Prevents API abuse and resource exhaustion
2. **Progressive Enforcement**: Gives users a chance to correct behavior
3. **Admin Controls**: Allow manual intervention when needed
4. **Email Notifications**: Users are informed of actions taken
5. **Audit Trail**: All actions logged for monitoring
6. **Environment Isolation**: Test mode doesn't send emails

## Future Enhancements

### Potential Improvements
1. **Whitelist Support**: Allow certain users to bypass burst detection
2. **Custom Thresholds**: Per-user or per-plan rate limits
3. **Analytics Dashboard**: Visual monitoring of burst patterns
4. **Appeal System**: Automated review process for false positives
5. **Graduated Penalties**: Longer blocks for repeat offenders

### Monitoring Additions
1. **Metrics Collection**: Track burst violation patterns
2. **Alerting**: Notify admins of unusual abuse patterns
3. **Reporting**: Regular summaries of abuse prevention effectiveness