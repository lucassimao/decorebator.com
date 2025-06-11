# Image Loading Improvements in Quiz Component

## Overview
Enhanced the quiz component to gracefully handle image loading failures with retry and skip options, improving user experience when network issues or image loading problems occur.

## Changes Made

### 1. Enhanced Error Handling
- Added dedicated error state for image loading failures
- Replaced simple alert with interactive error UI
- Added retry counter to track retry attempts

### 2. User Actions
When an image fails to load, users now have two options:
- **Retry**: Reload the image with cache-busting timestamp
- **Skip Question**: Move to the next quiz question

### 3. Loading Timeout
- Added 15-second timeout for image loading
- Prevents indefinite loading states on slow connections
- Automatically shows error state after timeout

### 4. Visual Improvements
- Clear error state with icon and message
- Action buttons for retry and skip
- Retry attempt counter display
- Smooth loading and error transitions

### 5. Implementation Details

#### State Management
```typescript
const [imageError, setImageError] = useState(false);
const [imageRetryCount, setImageRetryCount] = useState(0);
const imageLoadTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
```

#### Retry Logic
- Appends timestamp query parameter to force reload
- Preserves existing URL parameters
- Increments retry counter for tracking

#### Error UI
- Displays placeholder icon when image fails
- Shows localized error message
- Provides two action buttons (retry/skip)
- Shows retry attempt count

### 6. Internationalization
Added translation key `retryAttempts` to all supported languages:
- English: "Retry attempts: {{count}}"
- Spanish: "Intentos de reintento: {{count}}"
- German: "Wiederholungsversuche: {{count}}"
- French: "Tentatives de réessai: {{count}}"
- Italian: "Tentativi di riprovare: {{count}}"
- Japanese: "再試行回数: {{count}}"
- Portuguese: "Tentativas de recarregar: {{count}}"

## Benefits
1. **Better UX**: Users aren't stuck when images fail to load
2. **Graceful Degradation**: Quiz can continue even with image issues
3. **Network Resilience**: Handles slow/unstable connections better
4. **User Control**: Users can decide whether to retry or skip
5. **Transparency**: Shows retry attempts for debugging
6. **No Lost Progress**: Users can skip problematic images without losing quiz progress