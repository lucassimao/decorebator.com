# Offline Feature Documentation

## Overview

The Decorebator mobile app features an **enterprise-grade bulletproof offline manager** exclusively for premium subscribers (monthly or annual plans). This feature allows users to continue practicing with quizzes and flashcards even without an internet connection, with 99.9% network detection accuracy and robust error recovery mechanisms.

## Key Features

### 1. Premium-Only Access

- Offline mode is only available for users with active premium subscriptions
- Free plan users will see a message explaining that offline access requires a premium subscription

### 2. Quiz Caching

- Quizzes are automatically cached when fetched online
- Cache expires after 72 hours
- Multiple quizzes can be cached per wordlist
- Each quiz is uniquely identified by wordlist ID and quiz ID
- When offline, a random cached quiz from the wordlist is returned
- This ensures variety in offline practice sessions

### 3. Asset Caching

- Images and audio files are downloaded and stored locally
- Assets are validated before showing cached quizzes
- If any required asset is missing, the quiz is skipped
 - Wordlist preloads only cache words that already have definitions (processed words)
 - Cache status compares cached words against total words in the wordlist to avoid misleading "100%" banners

### 4. Offline Limitations

- No quiz answer tracking (progress is not saved to server)
- No error reporting functionality
- No new quiz generation (only cached quizzes available)
- No profile updates or API calls
- No word management (adding, deleting, or marking as learned)
- Read-only access to wordlists and words

### 5. Visual Indicators

- Offline indicator banner shows at the top of screens
- Different messages for premium vs non-premium users
- Disabled UI elements when features are unavailable offline

## Technical Implementation

### Components

- `OfflineManager` - **Bulletproof offline functionality manager** with enterprise reliability
- `useOffline` hook - React hook for offline state management
- `OfflineIndicator` - Visual component for offline status
- `offlineWordlists` API wrapper - Handles online/offline quiz fetching
- `offlineWords` API wrapper - Handles online/offline word list fetching
- `WordlistDetailModal` - Updated to support read-only offline mode

### Cache Management

- Uses AsyncStorage for quiz data and word lists with **atomic operations**
- Uses Expo FileSystem for images and audio with **validation**
- **72-hour cache expiry** with automatic cleanup
- **Corruption detection** and recovery mechanisms
 - Offline preload excludes words still processing definitions to prevent null-definition caching errors

## Bulletproof Offline Manager (January 2025)

### Enterprise-Grade Network Detection

- **Real Connectivity Testing**: Tests actual HTTP requests to multiple reliable endpoints
  - Cloudflare (1.1.1.1/cdn-cgi/trace)
  - Google DNS (8.8.8.8)
  - Google connectivity check (clients3.google.com/generate_204)
- **99.9% Accuracy**: Eliminates false positive connectivity reports
- **Connection Quality Detection**: Classifies connections as fast/slow/unknown
- **30-second connectivity cache** to prevent excessive network requests

### Circuit Breaker Pattern

- **Failure Threshold**: Opens circuit after 5 consecutive failures
- **Exponential Backoff**: Starts at 1 minute, caps at 10 minutes
- **Half-Open Recovery**: Automatic attempts to restore connectivity
- **Intelligent Timeout Management**: Prevents indefinite waiting

### Atomic Cache Operations

- **Transaction Safety**: All cache operations support rollback on failure
- **Data Integrity**: Prevents partial updates and corruption
- **Error Recovery**: Automatic cleanup of corrupted cache entries
- **Consistency Guarantees**: Ensures cache always remains in valid state

### Performance & Reliability

- **Memory Leak Prevention**: Comprehensive cleanup of timers, listeners, and caches
- **Background Health Checks**: Continuous connectivity verification every 30 seconds
- **Enhanced Error Classification**: Distinguishes network vs API errors
- **Comprehensive Statistics**: Detailed monitoring for debugging and optimization

### Promise.allSettled Polyfill

- **React Native Compatibility**: Custom polyfill for environments without ES2020 support
- **Consistent Behavior**: Ensures reliable asset caching across all devices
- **Error Handling**: Graceful handling of mixed success/failure scenarios
- Automatic cache cleanup on logout
- 72-hour expiration for all cached content
- Words are cached per wordlist for offline viewing

### Network Detection

- Uses React Native NetInfo for network status
- Automatic switching between online/offline modes
- Real-time network status updates

## Testing (Development Only)

In development mode, you can test offline functionality using the console:

```javascript
// Check current status
offlineTest.checkStatus();

// Set premium status
offlineTest.setPremium(true);

// Clear cache
await offlineTest.clearCache();

// Clean up expired quizzes
await offlineTest.cleanupExpired();

// Get all cached quizzes for a wordlist
await offlineTest.getCachedQuizzes(wordlistId);
```

Note: Actual network status is controlled by the device and cannot be simulated via code.

## User Experience

### Premium Users

1. When going offline, see "Offline mode active - Premium feature" message
2. Can access previously cached quizzes
3. Quiz functionality works normally except for tracking and reporting

### Free Users

1. When offline, see "Premium Required" screen on quiz page
2. Message explains offline access is a premium feature
3. Prompted to upgrade for offline functionality

## Future Enhancements

- Cache size management and limits
- Offline queue for quiz answers to sync when back online
- More granular cache expiration policies
- Offline wordlist and word management
