# Offline Flash Cards Implementation

This document describes the offline capabilities added to the Decorebator mobile app to support flash cards for premium users.

## Overview

The offline flash cards feature allows premium users to download wordlists and their associated definitions for offline study. This ensures a seamless learning experience even without an internet connection.

## Architecture

### 1. Offline Manager Extensions (`utils/offlineManager.ts`)

**New Features Added:**

- **Definition Caching**: Cache word definitions with metadata (timestamp, wordlist ID, word ID)
- **Asset Management**: Download and cache audio files referenced in definitions
- **Bulk Operations**: Preload entire wordlists for offline use
- **Cache Statistics**: Track cache status and completion percentage
- **Cache Validation**: Ensure offline assets are available and update URLs to local paths

**Key Methods:**

- `cacheDefinitions()`: Cache definitions for a specific word
- `getCachedDefinitions()`: Retrieve cached definitions with asset validation
- `preloadWordlistForOffline()`: Download all words and definitions for a wordlist
- `isWordlistCachedForOffline()`: Check if a wordlist is fully available offline
- `getWordlistCacheStats()`: Get cache completion statistics

### 2. Offline API Layer (`api/offlineWordlists.ts`)

**New Functions:**

- `getWords()`: Fetch words with offline fallback
- `getWordDefinitions()`: Fetch definitions with offline fallback
- `preloadWordlistForOffline()`: Bulk download wordlist data
- `isWordlistAvailableOffline()`: Check offline availability
- `getWordlistCacheStats()`: Get cache statistics

**Behavior:**

- **Online Mode**: Fetch from API and cache automatically
- **Network Failure**: Fall back to cached data if available
- **Offline Mode**: Use cached data exclusively

### 3. Flash Cards UI Updates (`app/practice.tsx`)

**Enhanced Features:**

- Uses offline-aware API calls
- Displays offline indicator
- Shows appropriate error messages for offline users without premium
- Handles network failures gracefully
- Maintains full functionality when offline

**Error Handling:**

- Premium check for offline access
- Graceful degradation when cache is unavailable
- User-friendly error messages

### 4. Offline Preloader Component (`components/OfflinePreloader.tsx`)

**Features:**

- Visual indicator of cache status
- One-click wordlist download
- Progress tracking
- Premium user validation
- Automatic cache status updates

**States:**

- **Not Cached**: Shows download button (requires online connection)
- **Partially Cached**: Shows progress percentage
- **Fully Cached**: Shows "Available offline" status
- **Downloading**: Shows loading state

## Cache Management

### Data Structure

```typescript
interface CachedDefinitions {
  definitions: Definition[];
  timestamp: number;
  wordlistId: number;
  wordId: number;
}
```

### Storage Strategy

- **AsyncStorage**: For definition metadata and content
- **FileSystem**: For audio files and other assets
- **Cache Keys**: `decorebator_offline_definitions_{wordlistId}-{wordId}`
- **Asset Directory**: `{documentDirectory}decorebator_assets/`

### Expiration

- **Cache Lifetime**: 72 hours (configurable)
- **Automatic Cleanup**: Removes expired entries
- **Asset Validation**: Ensures files exist and are accessible

## Premium Features

### Access Control

- Only premium subscribers can access offline features
- Free users see upgrade prompts when offline
- Cache operations are disabled for non-premium users

### User Experience

- Seamless transition between online/offline modes
- Visual indicators for cache status
- Progress feedback during downloads
- Offline indicator in app header

## Benefits

### For Users

- **Uninterrupted Learning**: Study flash cards without internet
- **Data Savings**: Reduce mobile data usage
- **Better Performance**: Faster loading from local cache
- **Travel-Friendly**: Learn while traveling or in areas with poor connectivity

### For Premium Value

- **Exclusive Feature**: Differentiates premium subscription
- **Enhanced Experience**: Provides additional value for paying users
- **Retention**: Increases user engagement and satisfaction

## Technical Implementation Details

### Caching Strategy

1. **Eager Caching**: Cache data automatically when fetched online
2. **Preloading**: Allow users to explicitly download wordlists
3. **Lazy Loading**: Cache definitions only when needed for flash cards
4. **Asset Management**: Download and validate audio files

### Offline Detection

- Uses `@react-native-community/netinfo` for network status
- Provides hooks for components to react to connectivity changes
- Handles edge cases like slow/unreliable connections

### Error Handling

- Graceful fallback to cached data
- Clear error messages for users
- Retry mechanisms for failed downloads
- Asset validation before use

### Performance Considerations

- **Throttling**: Limit concurrent downloads to avoid overwhelming the server
- **Background Processing**: Cache operations don't block UI
- **Memory Management**: Efficient cleanup of expired cache entries
- **Storage Monitoring**: Track cache size and usage

## Future Enhancements

### Potential Improvements

1. **Selective Sync**: Allow users to choose which wordlists to cache
2. **Smart Caching**: Prioritize frequently accessed content
3. **Compression**: Reduce storage requirements for cached data
4. **Sync Conflicts**: Handle data conflicts when coming back online
5. **Cache Analytics**: Track cache hit rates and usage patterns

### Advanced Features

1. **Offline Quiz Generation**: Generate new quizzes from cached data
2. **Offline Progress Tracking**: Store learning progress locally
3. **Background Sync**: Update cache automatically when online
4. **Cross-Device Sync**: Synchronize cache across user devices

## Usage Examples

### Basic Usage

```typescript
// Check if wordlist is available offline
const isAvailable = await offlineWordlistsApi.isWordlistAvailableOffline(123);

// Preload wordlist for offline use
await offlineWordlistsApi.preloadWordlistForOffline(123);

// Use words/definitions with automatic offline fallback
const words = await offlineWordlistsApi.getWords(123);
const definitions = await offlineWordlistsApi.getWordDefinitions(123, 456);
```

### Component Integration

```tsx
// Add offline preloader to wordlist screen
<OfflinePreloader
  wordlistId={wordlistId}
  wordlistName={wordlistName}
  onPreloadComplete={() => showSuccessMessage()}
  onPreloadError={(error) => showErrorMessage(error)}
/>
```

This implementation provides a robust offline experience for premium users while maintaining the existing online functionality for all users.
