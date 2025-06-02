# Offline Feature Documentation

## Overview
The Decorebator mobile app now supports offline mode exclusively for premium subscribers (monthly or annual plans). This feature allows users to continue practicing with quizzes even without an internet connection.

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
- `OfflineManager` - Core offline functionality manager
- `useOffline` hook - React hook for offline state management
- `OfflineIndicator` - Visual component for offline status
- `offlineWordlists` API wrapper - Handles online/offline quiz fetching
- `offlineWords` API wrapper - Handles online/offline word list fetching
- `WordlistDetailModal` - Updated to support read-only offline mode

### Cache Management
- Uses AsyncStorage for quiz data and word lists
- Uses Expo FileSystem for images and audio
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
offlineTest.checkStatus()

// Set premium status
offlineTest.setPremium(true)

// Clear cache
await offlineTest.clearCache()

// Clean up expired quizzes
await offlineTest.cleanupExpired()

// Get all cached quizzes for a wordlist
await offlineTest.getCachedQuizzes(wordlistId)
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