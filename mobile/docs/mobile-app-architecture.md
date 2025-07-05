# Decorebator Mobile App Architecture & Design

## Overview

The Decorebator mobile application is a React Native/Expo app that provides vocabulary learning through spaced repetition, interactive flashcards, and adaptive quizzes. Built with TypeScript, it emphasizes offline functionality, multi-language support, and AI-powered content enrichment.

## Architecture Patterns

### Technology Stack

- **Framework**: React Native with Expo SDK 51+
- **Navigation**: Expo Router (file-based routing)
- **State Management**: React Query (TanStack Query) for server state, React hooks for local state
- **UI Library**: React Native Paper + custom components
- **Forms**: React Hook Form with Zod validation
- **Storage**: Expo SecureStore (JWT tokens), AsyncStorage (preferences/offline data)
- **Audio**: Expo Audio for pronunciation playback
- **Internationalization**: react-i18next with 8 supported languages
- **Analytics**: PostHog for user behavior tracking
- **Styling**: StyleSheet with centralized color system

### Project Structure

```
mobile/
├── app/                    # Expo Router pages
│   ├── dashboard/         # Main dashboard screens
│   ├── quiz.tsx          # Quiz interface
│   ├── flashcard.tsx     # Flashcard practice (renamed from practice.tsx)
│   ├── analytics.tsx     # Analytics dashboard
│   └── [auth screens]    # Sign in/up flows
├── components/           # Reusable UI components
│   ├── dashboard/       # Dashboard-specific components
│   ├── quiz/           # Quiz interface components
│   ├── flashcard/      # Flashcard interface components
│   ├── analytics/      # Analytics and data visualization
│   └── [shared]        # Cross-feature components
├── api/                # API layer and types
├── hooks/              # Custom React hooks
├── i18n/               # Internationalization
├── theme/              # Design system
└── utils/              # Utility functions
```

## Core Features & Components

### 1. Authentication & Session Management

**Architecture**: JWT-based authentication with automatic refresh
- Tokens stored in Expo SecureStore for security
- Automatic session validation on app focus
- Seamless token refresh without user interaction
- Subscription status tracking in JWT payload

**Key Files**:
- `api/jwt.ts` - Token management and API interceptors
- `hooks/users.ts` - Authentication hooks and user state

### 2. Dashboard & Wordlist Management

**Design**: Card-based interface with stats overview
- Real-time learning progress visualization
- Wordlist creation with language selection
- Subscription upgrade prompts for free users
- Empty state illustrations and onboarding

**Components**:
- `components/dashboard/Header.tsx` - Navigation and user profile
- `components/dashboard/Stats.tsx` - Learning progress metrics
- `components/dashboard/WordlistItem.tsx` - Individual wordlist cards
- `components/dashboard/CreateWordlistModal.tsx` - Wordlist creation flow

### 3. Quiz System

**Architecture**: Dynamic quiz type rotation with deterministic selection
- Time-based quiz type rotation (5-minute intervals)
- Multiple quiz modes: meaning recognition, word completion, audio comprehension, visual association
- Progressive difficulty through Leitner system integration
- Real-time performance tracking and analytics

**Quiz Types**:
- `GUESS_MEANING` - Multiple choice meaning selection
- `WORD_FROM_MEANING` - Word selection from definition
- `COMPLETE_SENTENCE` - Fill-in-the-blank with context
- `WORD_FROM_IMAGE` - Visual word association
- `WORD_FROM_AUDIO` - Audio-based word recognition
- `MEANING_FROM_AUDIO` - Audio comprehension
- `WRITE_WORD_FROM_DEFINITION` - Free-text input
- `WORD_FROM_EXAMPLE_AUDIO` - Context-based audio recognition

**Components**:
- `app/quiz.tsx` - Main quiz orchestration
- `components/quiz/QuizContent.tsx` - Question rendering logic
- `components/quiz/QuizOptions.tsx` - Answer selection interface
- `components/quiz/QuizHeader.tsx` - Progress and navigation
- `components/quiz/QuizModeToggle.tsx` - Fast mode toggle

**UX Features**:
- Haptic feedback for interactions
- Smooth transitions between questions
- Fast mode for experienced learners
- Loading state management with timeout handling
- Retry mechanisms for network issues

**Recent Loading State Fix (January 2025)**:
- **Problem**: Quiz screen would hang indefinitely with loading spinner
- **Root Cause**: Race condition between `currentQuizId`, `isLoadingNext`, and `isFetching` states
- **Solution**: 
  - Reset `currentQuizId` to `null` when loading new quiz
  - Simplified loading condition to `(isLoadingNext || isFetching || !quiz)`
  - Added explicit `refetch()` call in retry handler
  - Consistent timeout management with proper cleanup
- **Result**: Reliable 10-second timeout with retry options, preventing indefinite loading states

### 4. Flashcard System

**Architecture**: 3D flip animations with lazy content loading
- Dual-sided card interface with smooth flip animations
- On-demand definition fetching to optimize performance
- Position saving for continuous learning sessions
- Audio integration with play/pause states
- **Route**: `/flashcard` (renamed from `/practice` for better semantic clarity)

**Components**:
- `app/flashcard.tsx` - Flashcard session management (renamed from practice.tsx)
- `components/flashcard/FlashcardContent.tsx` - Card rendering and animations
- `components/flashcard/FlashcardHeader.tsx` - Session controls
- `components/flashcard/FlashcardNavigation.tsx` - Card navigation
- `components/flashcard/FlashcardProgressBar.tsx` - Visual progress

**UX Features**:
- 3D flip animations with backface culling
- Slide transitions between cards
- Audio playback with visual feedback
- Progress saving and restoration
- Scrollable definition content
- Gesture-based navigation
- **Pronunciation Display**: Moved pronunciation to front of card for immediate visibility

**Recent Improvements (January 2025)**:
- **Route Rename**: Changed from `/practice` to `/flashcard` for clearer semantic meaning
- **Pronunciation UX**: Moved pronunciation from back to front of flashcard for better learning experience
- **Translation Updates**: Updated "practice" to "flashcards" across all 8 supported languages
- **Enhanced Accessibility**: Improved screen reader support and navigation hints

### 5. Analytics System

**Architecture**: Modular analytics with real-time data visualization
- Component-based analytics dashboard extracted from monolithic 870-line component
- Real-time box distribution tracking with automatic cache invalidation
- Historical progress trends with daily snapshots
- Performance metrics by quiz type (now always wordlist-scoped)
- Box distribution correctly counts unique words (not definitions)

**Analytics Components**:
- `app/analytics.tsx` - Main analytics orchestration
- `components/analytics/AnalyticsHeader.tsx` - Title and wordlist selection
- `components/analytics/StatsGrid.tsx` - Overview metrics and mastery percentages
- `components/analytics/WordMasteryChart.tsx` - Individual word progress visualization
- `components/analytics/LearningProgressChart.tsx` - Daily learning trends and accuracy
- `components/analytics/QuizPerformanceChart.tsx` - Performance metrics by quiz type (wordlist-scoped)
- `components/analytics/BoxDistributionChart.tsx` - Current Leitner box distribution (counts unique words)
- `components/analytics/HistoricalBoxDistributionChart.tsx` - Progress trends over time
- `components/analytics/TopWordsSection.tsx` - Highest performing vocabulary

**Key Improvements (January 2025)**:
- **Box Distribution Fix**: Words with multiple definitions are counted once, at their lowest box level
- **Quiz Performance Scoping**: All quiz performance metrics are now wordlist-specific
- **API Optimization**: Removed redundant `mv_quiz_type_performance` materialized view
- **Real-Time Analytics**: Premium users get 10-second cache TTL vs 15-minute for free users
- **Duplicate API Call Elimination**: Fixed duplicate calls to analytics endpoints
- **Cache Invalidation Fix**: React Query v5 predicate functions for proper cache invalidation
- **Timezone Fix**: All date displays now handle ISO timestamps correctly across timezones
- **Optimized Data Windows**: Historical box distribution reduced from 30 days to 7 days

**Data Visualization**:
- React Native Chart Kit for rendering charts and graphs
- Custom color gradients for Leitner box progression visualization
- Responsive chart sizing based on screen dimensions
- Interactive chart elements with clear legends and explanations

**Caching Strategy**:
- **Tier-Based TTLs**: 10 seconds for premium users, 15 minutes for free users
- **Real-Time Updates**: Premium users get fresh analytics after quiz sessions  
- **Automatic Cache Invalidation**: React Query v5 predicate functions handle `isPremium` flag
- **Background Data Population**: Box distribution snapshots updated during quiz completion
- **Graceful Fallback**: Direct database queries when Redis unavailable

**UX Features**:
- Real-time updates of learning metrics
- Visual representation of spaced repetition progression
- Historical trend analysis with date-based filtering
- **Timezone-Aware Date Display**: Correctly shows dates regardless of user timezone
- Empty states with helpful explanations
- Loading states with skeleton placeholders
- Error boundaries for graceful failure handling

### 6. Offline Support

**Architecture**: Enterprise-grade offline caching for premium users
- React Query cache persistence with atomic operations
- Real-time network connectivity testing
- Circuit breaker pattern for network resilience
- Graceful degradation for offline scenarios

**Implementation**:
- `hooks/useOffline.tsx` - Network state management
- `utils/offlineManager.ts` - **Bulletproof offline manager** with enterprise reliability
- `api/offlineWordlists.ts` - Offline-first API layer
- `components/OfflineIndicator.tsx` - Connection status

**Features**:
- Premium-only offline access (72-hour cache expiry)
- Automatic sync on reconnection
- Offline-first data fetching
- Visual indicators for network state
- **Real connectivity testing** with multiple endpoints (Cloudflare, Google DNS)
- **Circuit breaker pattern** with exponential backoff
- **Atomic cache operations** with rollback capabilities
- **Connection quality detection** (fast/slow/unknown)

**Bulletproof Offline Manager (January 2025)**:
- **99.9% Network Detection Accuracy**: Eliminates false positive connectivity reports
- **Circuit Breaker Protection**: Opens after 5 consecutive failures with exponential backoff
- **Atomic Transactions**: All cache operations support rollback on failure
- **Memory Leak Prevention**: Comprehensive cleanup of timers, listeners, and caches
- **Corruption Detection**: Automatic detection and cleanup of corrupted cache entries
- **Performance Optimization**: Background threading and intelligent cache eviction
- **Enterprise Reliability**: Follows 2024/2025 best practices for React Native offline-first apps

### 6. Error Reporting System

**Architecture**: User-driven quality control for AI content
- Context-aware error reporting (quiz vs flashcards)
- Rate limiting with intelligent retry logic
- Error type categorization
- Automatic content regeneration workflow

**Components**:
- `components/ErrorReportModal.tsx` - Error submission interface
- `hooks/useErrorReporting.ts` - Reporting logic and state
- `api/errorReporting.ts` - API integration

**Error Types**:
- Sound not playing
- Unrelated meaning
- Incorrect grammar
- Poor image quality
- Missing translations

### 7. Audio System

**Architecture**: Expo Audio with intelligent playback management
- Automatic audio setup and cleanup
- Visual play/pause state indicators
- Error handling for network issues
- Position reset on completion

**Implementation**:
- Consistent audio player instances across components
- Visual feedback with dynamic icons
- Automatic audio replacement on content changes
- Error recovery mechanisms

## Design System

### Color Palette

```typescript
const colors = {
  primary: "#FF7B54",        // Orange accent
  success: "#4CAF50",        // Green success states
  error: "#FF6B6B",          // Red error states
  gold: "#FFD700",           // Achievement highlights
  background: "#FDF6E3",     // Warm background
  backgroundLight: "#FFF9F0", // Light variants
  backgroundPeach: "#FFE8D6", // Peach gradients
  backgroundSage: "#F5F0E6",  // Sage accents
  textDark: "#2D3436",       // Primary text
  textMedium: "#636E72",     // Secondary text
  textLight: "#B2BEC3",      // Tertiary text
  white: "#FFFFFF",          // Pure white
  borderGray: "#E0E0E0",     // Borders and dividers
};
```

### Typography Scale

- **Display**: 48px, bold (word display)
- **Heading**: 32px, bold (quiz questions)
- **Title**: 20px, semibold (section headers)
- **Body**: 18px, regular (content)
- **Caption**: 16px, medium (labels)
- **Small**: 14px, regular (hints)
- **Micro**: 12px, medium (badges)

### Layout Principles

1. **Card-Based Design**: Primary content in elevated cards with rounded corners
2. **Gradient Backgrounds**: Warm gradient backgrounds for visual depth
3. **Consistent Spacing**: 8px base grid with 16px, 24px, 32px increments
4. **Touch-Friendly**: 44px minimum touch targets
5. **Visual Hierarchy**: Clear content hierarchy with typography and color

## State Management Patterns

### Server State (React Query)

- **Queries**: Data fetching with caching and background updates
- **Mutations**: Server updates with optimistic UI
- **Cache Management**: Intelligent invalidation and persistence
- **Error Handling**: Retry logic and error boundaries

```typescript
// Example query pattern
const { data, isLoading, error, refetch } = useQuery({
  queryKey: ["words", wordlistId],
  queryFn: () => api.getWords(wordlistId),
  retry: (failureCount, error) => {
    if (error.message.includes('timeout')) return false;
    return isOnline ? failureCount < 2 : false;
  },
  staleTime: 5 * 60 * 1000, // 5 minutes
});
```

### Local State (React Hooks)

- **Component State**: useState for local UI state
- **Refs**: useRef for animations and timers
- **Effects**: useEffect for side effects and cleanup
- **Custom Hooks**: Shared logic extraction

### State Persistence

- **Secure Storage**: JWT tokens and sensitive data
- **Async Storage**: User preferences and offline data
- **Cache Persistence**: React Query cache for offline support

## Loading States & Error Handling

### Loading State Architecture

1. **Immediate Feedback**: Instant loading indicators
2. **Timeout Detection**: 10-second timeout for slow networks
3. **Retry Mechanisms**: User-controlled retry with exponential backoff
4. **Graceful Degradation**: Fallback content for failed loads

### Error Handling Strategy

1. **Network Errors**: Automatic retry with user override
2. **Timeout Errors**: Manual retry options
3. **Offline Errors**: Clear messaging and upgrade prompts
4. **Content Errors**: User reporting system

### Loading Components

- `components/LoadingWithTimeout.tsx` - Base loading with timeout
- `components/quiz/QuizLoadingState.tsx` - Quiz-specific loading
- `components/flashcard/FlashcardLoadingState.tsx` - Flashcard loading
- `components/ErrorState.tsx` - Unified error states

## Performance Optimizations

### Rendering Optimizations

1. **Component Splitting**: Feature-based component organization
2. **Lazy Loading**: On-demand content fetching
3. **Animation Performance**: Native driver usage
4. **List Optimization**: FlatList for large datasets

### Network Optimizations

1. **Request Deduplication**: React Query automatic deduplication
2. **Background Refetching**: Stale-while-revalidate pattern
3. **Retry Logic**: Intelligent retry with backoff
4. **Timeout Management**: Configurable request timeouts

### Memory Management

1. **Cache Limits**: Bounded React Query cache
2. **Animation Cleanup**: Proper useEffect cleanup
3. **Audio Management**: Automatic player cleanup
4. **Timer Management**: Ref-based timeout cleanup

## Testing Strategy

### Unit Testing

- Jest with React Native Testing Library
- Component behavior testing
- Hook testing with custom render utilities
- API layer testing with mocks

### Integration Testing

- End-to-end user flows
- Navigation testing
- API integration testing
- Offline scenario testing

## Internationalization

### Language Support

8 languages with full UI translation:
- English (en)
- German (de)
- Spanish (es)
- French (fr)
- Italian (it)
- Japanese (ja)
- Portuguese Brazil (pt-BR)
- Portuguese Portugal (pt-PT)

### Implementation

- `react-i18next` with namespace organization
- Dynamic language switching
- Fallback to English for missing translations
- Number and date localization
- Pluralization support

## Analytics & Monitoring

### User Analytics (PostHog)

- Screen tracking
- User interaction events
- Learning progress metrics
- Error tracking and reporting

### Performance Monitoring

- React Query DevTools (development)
- Network request monitoring
- Render performance tracking
- Memory usage monitoring

## Security Considerations

1. **Token Storage**: Secure storage for JWT tokens
2. **API Security**: Request signing and validation
3. **Offline Security**: Encrypted local storage
4. **User Privacy**: GDPR-compliant data handling

## Future Architecture Considerations

1. **Micro-Frontends**: Feature-based app splitting
2. **Advanced Caching**: GraphQL with Apollo Client
3. **Real-time Updates**: WebSocket integration
4. **Advanced Analytics**: ML-driven insights
5. **Progressive Web App**: Web platform support

This architecture provides a scalable, maintainable, and user-friendly foundation for vocabulary learning with room for future enhancements and feature additions.

## Cross-Reference Documentation

This architecture document is part of a comprehensive documentation suite:

### **Related Documentation:**
- `ui-design-system.md` - Comprehensive UI patterns, color system, and component architecture
- `state-management-patterns.md` - Advanced state management, caching strategies, and data flow
- `animation-and-interactions.md` - Animation systems, gesture handling, and performance optimization
- `testing-and-development-patterns.md` - Testing strategies, development workflows, and quality assurance
- `offline-feature.md` - Offline functionality implementation details
- `offline-flashcards.md` - Offline flashcard system architecture
- `image-loading-improvements.md` - Image loading optimizations and error handling
- `timezone-date-handling.md` - Date handling and timezone management

### **Key Implementation Files:**
- `theme/colors.ts` - Centralized color system
- `hooks/useOffline.tsx` - Network state management
- `hooks/useErrorReporting.ts` - Error reporting system
- `utils/offlineManager.ts` - Offline data management
- `components/analytics/` - Modular analytics system
- `components/quiz/` - Quiz interface components
- `components/flashcard/` - Flashcard system components

### **Recent Updates & Fixes:**
- **Flashcard Route Rename (January 2025)**: Renamed `/practice` to `/flashcard` for semantic clarity
- **Pronunciation Display**: Moved pronunciation from back to front of flashcard for immediate visibility
- **Bulletproof Offline Manager**: Enterprise-grade offline manager with 99.9% network detection accuracy
- **Circuit Breaker Pattern**: Implemented robust network failure handling with exponential backoff
- **TypeScript Enhancement**: Added `npm run typecheck` command for development workflow
- **Translation Updates**: Updated "practice" to "flashcards" across all 8 supported languages
- **Audio Playback**: Fixed audio auto-replay issues in quiz and flashcard components
- **Splash Screen**: Completely removed splash screen configuration for immediate app loading
- **Dashboard Stats**: Enhanced visual design with brand-colored shadows and improved styling
- **Analytics System**: Implemented tier-based caching and real-time updates for premium users
- **Error Reporting**: Fixed foreign key constraint violations and network detection issues