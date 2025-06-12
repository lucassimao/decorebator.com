# Decorebator - AI-Powered Vocabulary Learning Platform

Decorebator is a comprehensive vocabulary learning application that uses AI-powered enrichment and the Leitner spaced repetition system to help users master new languages effectively.

## 🌟 Features

### Core Learning Features
- **Build Vocabulary Lists**: Create and manage multiple wordlists for any language
- **AI-Powered Enrichment**: Automatically generates definitions, images, audio pronunciations, and example sentences
- **Multiple Quiz Modes** (8 Different Types):
  - **Guess Meaning**: Choose the correct meaning for a given word
  - **Word from Meaning**: Select the word that matches a given definition
  - **Word from Image**: Identify words from AI-generated visual associations
  - **Audio Comprehension**: Recognize words and meanings from pronunciation
  - **Sentence Completion**: Complete sentences with the correct word using grammatical context
  - **Write from Definition**: Type the word based on its meaning (active recall)
  - **Example Audio Recognition**: Identify words from contextual example sentence audio
- **Interactive Flashcards**: Study definitions with examples, pronunciation, and grammatical context
- **Spaced Repetition**: Uses the advanced 7-box Leitner system with deterministic selection to optimize learning retention
- **Progress Tracking**: Monitor your learning journey with comprehensive analytics and detailed statistics
- **Error Reporting**: Report issues with AI-generated content (images, audio, definitions) for continuous improvement
- **Offline Support**: Premium users can access wordlists and practice offline

### Subscription Tiers
- **Free Plan**:
  - 1 wordlist
  - Up to 10 words per wordlist
  - Basic quiz modes
  - Online-only access
- **Premium Plans**:
  - **Monthly**: $6.99/month
  - **Annual**: $69.90/year (save $13.98)
  - Unlimited wordlists
  - Unlimited words
  - All quiz modes and flashcards
  - AI-powered content generation
  - Comprehensive analytics and progress tracking
  - Offline support for wordlists and practice
  - Error reporting system
  - Priority support

## 🏗️ Architecture Overview

The project consists of three main components:

### 1. API Backend (Go/Gin)
Located in `/api`, the backend follows a 3-tier layered architecture:
- **HTTP Layer** (`internal/http/`): Request handling, JWT authentication
- **Service Layer** (`internal/service/`): Business logic and orchestration
- **Repository Layer** (`internal/repository/`): Database operations

**Key Technologies**:
- Go with Gin web framework
- PostgreSQL 15+ database with pgx/v5 driver and materialized views
- River queue system for background jobs with PostgreSQL backend
- MinIO for S3-compatible object storage
- SendGrid for email services
- OpenAI API for AI features (DALL-E, TTS, GPT)
- Stripe for subscription payments with webhook integration
- Sentry for error monitoring and logging
- Structured logging with slog

### 2. Mobile Application (React Native/Expo)
Located in `/mobile`, cross-platform mobile app with:
- Expo Router for navigation
- React Query for API state management with offline caching
- React Hook Form with Zod validation
- React Native Paper UI components
- Secure storage for JWT tokens
- Automatic session refresh on focus
- Real-time subscription status updates
- Internationalization (i18n) support for 8 languages
- Offline support for premium users with local storage
- Error reporting modal for AI-generated content issues
- Interactive flashcard system with flip animations

### 3. Web Frontend (Next.js)
Located in `/web`, marketing website and web app with:
- Next.js 15 with App Router
- TailwindCSS for styling
- TypeScript for type safety

## 🚀 Getting Started

### Prerequisites
- Go 1.22+
- Node.js 18+
- PostgreSQL 15+
- Docker & Docker Compose
- MinIO (S3-compatible storage)

### Backend Setup

1. Navigate to the API directory:
```bash
cd api
```

2. Install dependencies and setup tools:
```bash
make setup
```

3. Create `.env` file with required environment variables:
```env
# Database
POSTGRES_USER=your_user
POSTGRES_PASSWORD=your_password
POSTGRES_DB=decorebator
POSTGRES_PORT=5432

# MinIO
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
MINIO_PORT=9000

# API Keys
SENDGRID_API_KEY=your_sendgrid_key
OPENAI_API_KEY=your_openai_key
JWT_SECRET=your_jwt_secret

# Stripe (Required for subscriptions)
STRIPE_API_KEY=your_stripe_api_key
STRIPE_WEBHOOK_SECRET=your_stripe_webhook_secret
STRIPE_PRICE_ID_MONTHLY=price_monthly_id_from_stripe
STRIPE_PRICE_ID_ANNUAL=price_annual_id_from_stripe
STRIPE_SUCCESS_URL=https://yourapp.com/subscription?status=success
STRIPE_CANCEL_URL=https://yourapp.com/subscription?status=cancel
```

4. Start infrastructure services:
```bash
docker-compose up -d
```

5. Run database migrations:
```bash
make migrate-up
```

6. Start the API server:
```bash
make watch  # Development with auto-reload
# or
make run    # Production mode
```

7. Start background workers (in another terminal):
```bash
make workers
```

### Mobile App Setup

1. Navigate to the mobile directory:
```bash
cd mobile
```

2. Install dependencies:
```bash
npm install
```

3. Create `.env` file:
```env
EXPO_PUBLIC_API_URL=http://localhost:8080
```

**Supported Languages**: The mobile app includes full internationalization support for:
- English (en)
- German (de) 
- Spanish (es)
- French (fr)
- Italian (it)
- Japanese (ja)
- Portuguese - Brazil (pt-BR)
- Portuguese - Portugal (pt-PT)

**AI Content Languages**: The system supports comprehensive AI-powered content generation in 7 languages:
- English (en) - Complete grammar support with phrasal verbs, "alloy" voice for TTS
- Spanish (es) - Gender-aware definitions with formal/informal variations, "nova" voice for natural pronunciation
- French (fr) - Proper accent handling and liaison considerations, "shimmer" voice for elegant pronunciation
- German (de) - Four-case system support with separable verbs, "echo" voice for clear consonants
- Italian (it) - Verb group conjugations and gender agreement, "fable" voice for expressive speech
- Portuguese (pt) - Brazilian and European variations, "onyx" voice for deep, clear pronunciation
- Japanese (ja) - Hiragana, katakana, and kanji with keigo support, "alloy" voice optimized for Japanese

4. Start the development server:
```bash
npm start
```

5. Run on specific platforms:
```bash
npm run ios     # iOS simulator
npm run android # Android emulator
npm run web     # Web browser
```

### Web Frontend Setup

1. Navigate to the web directory:
```bash
cd web
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

4. Open http://localhost:3000

## 🧪 Testing

### API Tests
```bash
cd api
make test  # Runs integration tests with Docker
```

### Mobile Tests
```bash
cd mobile
npm test
```

## 🔧 Development Commands

### API Backend
```bash
make help              # Show all available commands
make migrate-up        # Apply database migrations
make migrate-down      # Rollback last migration
make create-migration  # Create new migration file
make psql             # Open PostgreSQL console
make debug-workers    # Debug workers with delve
make build            # Build production binary
```

### Mobile App
```bash
npm run lint          # Run ESLint
npm run expo:update   # Update Expo dependencies
npm run test          # Run Jest tests
```

**Code Quality Improvements**:
- Reusable error reporting modal component (`components/ErrorReportModal.tsx`)
- Internationalization support with React i18next
- Consistent TypeScript interfaces and error handling
- Comprehensive test coverage for core features

### Web Frontend
```bash
npm run build         # Build for production
npm run start         # Start production server
npm run lint          # Run linter
```

## 🆕 Recent Features & Improvements

### Analytics System Redesign (January 2025)
- **Solved N×8 API Call Problem**: New batch progress endpoint reduces dashboard calls from N×8 to 1
- **Redis Caching Infrastructure**: 
  - Tier-based TTLs: 10 seconds for premium users, 15 minutes for free users
  - Graceful fallback to database when Redis unavailable
  - Automatic cache invalidation using predicate functions for React Query v5 compatibility
  - Real-time analytics updates for premium users after quiz sessions
- **Performance Optimizations**:
  - 6 new strategic database indexes for sub-second query performance
  - Unified analytics service with interface-based architecture
  - Factory pattern eliminates redundant database connections
  - Historical box distribution optimized for 7-day windows instead of 30 days
- **Developer Experience**:
  - Simplified API with shorter method names (`Stats`, `Progress`, `WordMastery`)
  - Single factory method with configuration flags
  - Per-request service instantiation for clean separation of concerns
  - Fixed cache invalidation key matching for `isPremium` flag in query keys

### Multi-Language Definition Support
- **7 Supported Languages**: English, Spanish, French, German, Italian, Portuguese, Japanese
- **Native Language Processing**: AI generates definitions in the target language with proper grammar
- **Language-Specific Prompts**: ChatGPT receives instructions in the wordlist's language for authentic content
- **Dynamic Part-of-Speech Validation**: Grammar rules adapted for each language (e.g., German cases, Spanish gender)
- **Automatic Language Detection**: System detects wordlist language and adapts content generation accordingly
- **Comprehensive Verb Systems**: Language-appropriate verb tenses and inflections (presente, passé composé, Präteritum, etc.)
- **Cultural Linguistic Accuracy**: Regional variations supported (Brazilian vs European Portuguese)
- **Language-Optimized Audio**: Voice selection optimized per language for natural pronunciation
- **Culturally-Aware Images**: Image generation prompts in native language for cultural relevance

### Enhanced Grammar Support for Verbs
- **Verb Inflection System**: Automatic generation of verb forms (past tense, present tense, gerund, participle)
- **Contextual Examples**: Smart example sentences showing verbs in different grammatical contexts
- **Part-of-Speech Intelligence**: Enhanced definition fetching with proper grammatical categorization
- **Quiz Integration**: Sentence completion quizzes now use verb inflections for realistic practice

### Advanced Error Reporting System
- **User-Driven Quality Control**: Users can report issues with AI-generated content
- **5 Error Types Supported**:
  - Image doesn't match word meaning
  - Image not loading properly
  - Wrong or inaccurate definition
  - Irrelevant example sentences
  - Audio pronunciation not playing
- **Automatic Content Regeneration**: Reported errors trigger background jobs to fix issues
- **Temporary Skip Logic**: Problematic content is temporarily removed from quiz rotation
- **Multi-Language Support**: Error reporting available in 8 languages

### Interactive Flashcard System
- **Immersive Learning Experience**: Full-screen flashcards with smooth flip animations
- **Rich Content Display**: Definitions, pronunciations, examples, and part-of-speech information
- **Verb-Specific Features**: Special handling for verb inflections and tense examples
- **Progress Tracking**: Integrated with Leitner system for spaced repetition
- **Error Reporting Integration**: Report issues directly from flashcard interface

### Comprehensive Analytics Platform (January 2025 Redesign)
- **Word Mastery Tracking**: Individual word progress with accuracy calculations
- **Learning Progress Visualization**: Daily statistics and progress charts
- **Quiz Performance Analysis**: Performance metrics by quiz type (now wordlist-scoped)
- **Box Distribution Insights**: 
  - Current distribution correctly counts unique words (not definitions)
  - Historical snapshots of Leitner system progression
  - Word considered at its lowest box level across all definitions
- **High-Performance Architecture**: 
  - Redis caching with tier-based TTLs (1 hour for free users, 1 minute for premium)
  - Batch progress API reduces dashboard calls from N×8 to 1
  - Strategic database indexes for sub-second query performance
  - Real-time cache invalidation for premium users after quiz completion
- **Scalable Design**: Built to handle millions of users with efficient caching strategy
- **Modern Service Architecture**:
  - Interface-based design for flexible cached/non-cached implementations
  - Factory pattern with per-request service instantiation
  - Graceful Redis fallback to direct database queries
  - Unified analytics interface with simplified method names

### Offline Support for Premium Users
- **Local Data Caching**: Wordlists and definitions cached for offline access
- **Seamless Synchronization**: Automatic sync when connection is restored
- **Offline Quiz Support**: Complete quiz functionality without internet
- **Progress Preservation**: Offline learning progress saved and synced

### Internationalization (i18n)
- **8 Language Support**: English, German, Spanish, French, Italian, Japanese, Portuguese (BR), Portuguese (PT)
- **Dynamic Language Switching**: Real-time language changes without app restart
- **Comprehensive Translation Coverage**: All UI elements, error messages, and features translated
- **Cultural Localization**: Currency symbols, date formats, and cultural adaptations

### Enhanced Leitner System
- **Improved Box Progression**: Refined algorithm prevents immediate repetition of failed words
- **Intelligent Quiz Type Selection**: Dynamic quiz types based on word difficulty and available content
- **Pronunciation Integration**: Word pronunciation display in quiz and flashcard interfaces
- **Error Recovery**: Automatic handling of content validation and fallback mechanisms

### Modular Analytics Architecture (December 2024)
- **Component-Based Refactoring**: Extracted analytics screen from 870-line monolithic component into 7 focused, reusable components
- **Real-Time Box Distribution**: Live tracking of word progression across Leitner system boxes with automatic cache invalidation
- **Historical Progress Tracking**: Daily snapshots of learning progress with visual trend analysis
- **Component Modularity**:
  - `AnalyticsHeader`: Title and wordlist selection controls
  - `StatsGrid`: Overview metrics with mastery percentages
  - `WordMasteryChart`: Individual word progress visualization
  - `LearningProgressChart`: Daily learning trends and accuracy rates
  - `QuizPerformanceChart`: Performance metrics by quiz type
  - `BoxDistributionChart`: Current word distribution across Leitner boxes
  - `HistoricalBoxDistributionChart`: Progress trends over time
  - `TopWordsSection`: Highest performing vocabulary words
- **Background Data Population**: Automatic box distribution snapshots during quiz completion
- **Cache Strategy Optimization**: Infinite cache for current data, 5-minute cache for historical trends
- **Maintainable Architecture**: Clear separation of concerns for analytics features

### Modular Flashcard Architecture (December 2024)
- **Component-Based Refactoring**: Broke down 962-line flashcard component into 4 focused, reusable components
- **Smart Content Filtering**: API enhancement to only fetch words with definitions, preventing broken flashcard experiences
- **Improved Data Integrity**: New `onlyWithDefinitions` API parameter ensures flashcards always have content to display
- **Component Modularity**: 
  - `FlashcardHeader`: Title, progress counter, and error reporting controls
  - `FlashcardProgressBar`: Visual learning progress indicator
  - `FlashcardContent`: Complex card flip animations and rich content display
  - `FlashcardNavigation`: Previous/next navigation with keyboard support
- **Enhanced User Experience**: Eliminates empty flashcard states caused by async definition processing
- **Maintainable Codebase**: Easier to modify and extend individual flashcard features
- **Backward Compatible**: Existing API endpoints maintain full compatibility

### Quiz Loading State Fix (January 2025)
- **Fixed Loading Forever Issue**: Resolved race condition causing quiz screen to hang indefinitely
- **Improved State Management**: 
  - Reset `currentQuizId` when loading new quiz
  - Simplified loading condition to `(isLoadingNext || isFetching || !quiz)`
  - Added explicit `refetch()` call in retry handler
- **Enhanced User Experience**: 
  - Reliable 10-second timeout with retry options
  - Clear loading states during quiz transitions
  - Proper initial loading screen with timeout handling
- **Root Cause**: Complex state dependencies between `isLoadingNext`, `isFetching`, and `currentQuizId` were getting out of sync

### Timezone & Date Handling Fix (January 2025)
- **Fixed Date Display Issues**: Resolved timezone conversion problems affecting all analytics charts
- **Root Cause**: Backend sends ISO timestamps like "2025-06-11T00:00:00Z" which JavaScript interpreted as UTC, causing dates to display incorrectly in non-UTC timezones
- **Components Fixed**:
  - `HistoricalBoxDistributionChart.tsx`: Chart labels now show correct dates
  - `PracticeTimeChart.tsx`: Practice time chart displays accurate date labels  
  - `LearningProgressChart.tsx`: Learning progress trends show proper dates
- **Technical Solution**: 
  - Extract date part from ISO timestamps using `split('T')[0]`
  - Parse date components manually as local time instead of UTC
  - Prevents timezone conversion that was shifting dates by one day
- **User Impact**: Charts now correctly show "Today", "Yesterday", and proper MM/DD format instead of incorrect dates

## 📊 Database Schema

Key tables with performance indexes:
- `users`: User accounts, authentication, subscription status, and profile data (profile picture, country, date of birth, preferred language)
- `subscriptions`: Subscription history and details with Stripe integration
- `subscription_events`: Stripe webhook event audit trail & email notification tracking
- `wordlists`: User's vocabulary lists with language field and word counts
- `words`: Individual words in wordlists with audio URLs and learning status
- `definitions`: AI-generated definitions with multimedia, sources, and example sentences
- `definition_images`: Images associated with definitions
- `definition_example_audio`: Audio files for example sentences with fair usage tracking
- `example_audio_usage`: Usage tracking for example audio to ensure variety in quizzes
- `leitner_system_tracking`: Spaced repetition progress with temporary skip functionality
- `error_reports`: User-reported errors for definitions and AI-generated content
- `quiz_performance`: Individual quiz attempts with performance metrics
- `word_mastery`: Overall mastery tracking for each word per user
- `learning_progress`: Daily aggregated learning statistics
- `quiz_type_analytics`: Performance metrics grouped by quiz type
- `box_distribution_snapshot`: Daily snapshots of word distribution across Leitner boxes
- `river_job`: Background job queue

## 💳 Subscription System

### Features
- **Stripe Integration**: Secure payment processing with webhook event tracking
- **Flexible Plans**: Monthly ($6.99) and Annual ($69.90) subscriptions
- **Free Plan**: 1 wordlist with 10 words maximum
- **Automatic Limits**: Enforced at API level with subscription-aware middleware
- **Webhook Processing**: Real-time subscription updates with event deduplication
- **Seamless Activation**: Automatic session refresh after checkout
- **Instant Access**: Premium features unlock immediately upon payment
- **Customer Management**: Automatic Stripe customer creation and tracking

### Subscription Limits
Free users are limited to:
- 1 wordlist maximum
- 10 words per wordlist

Premium users get:
- Unlimited wordlists
- Unlimited words per wordlist
- All quiz modes
- Priority support

### Setting up Stripe
1. Create products and prices in Stripe Dashboard
2. Configure webhook endpoint: `https://yourapi.com/webhook/stripe`
3. Add webhook events to monitor:
   - `checkout.session.completed`
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.payment_succeeded`
   - `invoice.payment_failed`
4. Set environment variables:
   - `STRIPE_API_KEY`: Your secret API key
   - `STRIPE_WEBHOOK_SECRET`: Webhook endpoint secret for verification
   - `STRIPE_PRICE_ID_MONTHLY`: Price ID for monthly plan
   - `STRIPE_PRICE_ID_ANNUAL`: Price ID for annual plan
   - `STRIPE_SUCCESS_URL`: Success redirect URL
   - `STRIPE_CANCEL_URL`: Cancel redirect URL

### Subscription Flow
1. User initiates checkout from mobile app
2. Stripe checkout session created with user metadata
3. User completes payment on Stripe hosted page
4. Webhook updates subscription in database
5. User returns to app → automatic session refresh
6. New JWT issued with updated subscription plan
7. Premium features instantly available

### Email Notifications
Subscribers receive automated emails for:
- **Welcome Email**: Sent when subscription is activated
- **Payment Confirmation**: Sent when subscription renews successfully
- **Renewal Reminder**: Sent 3 days before subscription renewal
- **Cancellation Confirmation**: Sent when subscription is cancelled
- **Payment Failed Alert**: Sent when payment method fails

Email notifications are powered by SendGrid and include:
- Personalized content with user's name
- Clear subscription details (plan, amount, dates)
- Professional HTML templates
- Actionable buttons for account management

## 🔄 Background Jobs

The system uses River (PostgreSQL-based queue) for processing:
- **Image Generation**: Creates relevant images using DALL-E with language-specific prompts for cultural accuracy
- **Text-to-Speech**: Generates pronunciation audio using OpenAI TTS with language-optimized voice selection
- **Definition Fetching**: Multi-language AI-powered definition generation with native language prompts and grammar rules
- **Example Audio Generation**: Creates contextual audio for example sentences with smart selection (longest examples for verbs)
- **Subscription Renewal Reminders**: Automatically sends email reminders 3 days before renewal

Workers can be scaled independently with configurable concurrency:
- Image Generator: Max 5 workers
- Text-to-Speech: Max 30 workers
- Definition Fetcher: Max 50 workers
- Example Audio Generator: Max 20 workers
- Subscription Reminder: Max 10 workers

### Multi-Language Worker Features

**Text-to-Speech Worker**:
- **Language-Specific Voices**: Automatically selects optimal OpenAI TTS voice per language
- **Voice Mapping**: English (alloy), Spanish (nova), French (shimmer), German (echo), Italian (fable), Portuguese (onyx), Japanese (alloy)
- **Natural Pronunciation**: Each language uses voices optimized for that language's phonetics

**Image Generator Worker**:
- **Native Language Prompts**: DALL-E receives image generation instructions in the target language
- **Cultural Context**: Language-specific prompts ensure culturally appropriate visual representations
- **7-Language Support**: Full prompt templates in English, Spanish, French, German, Italian, Portuguese, Japanese

**Definition Fetcher Worker**:
- **Automatic Language Detection**: Detects wordlist language and adapts all AI processing accordingly
- **Native Grammar Rules**: Uses language-specific part-of-speech lists and verb tense systems
- **Fallback Mechanisms**: Robust error handling with English fallbacks for unsupported languages

**Example Audio Generator Worker**:
- **Smart Example Selection**: Cost-optimized processing - generates audio only for longest examples of verbs/phrasal verbs (60-80% TTS cost savings)
- **Language-Aware Voice Selection**: Automatically selects appropriate TTS voice based on wordlist language
- **Fair Distribution**: Tracks usage of each example audio to ensure variety in quiz selection
- **Batch Processing**: Processes multiple examples per definition efficiently
- **Inflection Support**: Handles both main examples and verb inflection examples with proper categorization

The system includes periodic jobs:
- **Daily Renewal Reminder Check**: Runs daily to identify subscriptions renewing in 3-4 days
- Worker processes include error reporting and retry logic for failed AI generations

### Queue Management
- River-based PostgreSQL queue system with transaction safety
- Job retry logic with exponential backoff
- Queue-specific worker limits for optimal resource usage
- Background job monitoring and error handling

## 📊 Analytics & Performance Tracking

The system includes comprehensive analytics for tracking learning progress:

### Learning Analytics
- **Word Mastery Tracking**: Individual word mastery levels with streak counting
- **Daily Progress**: Aggregated daily learning statistics per wordlist
- **Quiz Performance**: Performance metrics by quiz type (word meaning, visual, audio, contextual)
- **Box Distribution**: Historical snapshots of word distribution across Leitner boxes
- **Response Time Analysis**: Tracks average response times for performance optimization

### Analytics Features
- Real-time mastery calculation with accuracy rates
- Materialized views for optimized query performance
- Historical trend analysis for learning progress
- Dashboard statistics with comprehensive user metrics
- Automated analytics updates triggered by quiz completion

### Advanced Error Reporting System
- **User-Driven Quality Control**: Comprehensive error reporting for all AI-generated content
- **5 Structured Error Types**: Image mismatch, loading issues, wrong definitions, irrelevant examples, audio problems
- **Multi-Modal Support**: Error reporting available in both quiz and flashcard interfaces
- **Automatic Content Regeneration**: Background jobs triggered by error reports to fix problematic content
- **Temporary Skip Logic**: Problematic definitions temporarily removed from quiz rotation during error resolution
- **Error Resolution Workflow**: Complete lifecycle tracking from report to resolution
- **Multi-Language Support**: Error reporting interface available in 8 languages
- **Reusable Component Architecture**: Shared error reporting modal across quiz and flashcard features

### Error Report Rate Limiting & Abuse Prevention

To prevent abuse and control OpenAI API costs, the error reporting system implements comprehensive rate limiting:

#### Rate Limits by Subscription Tier

| Tier | Hourly Limit | Daily Limit |
|------|--------------|-------------|
| **Free** | 3 reports | 5 reports |
| **Premium** | 10 reports | 30 reports |

#### Protection Mechanisms

1. **Dual-Layer Rate Limiting**:
   - **Global Rate Limits**: Enforced at middleware level before requests reach handlers
   - **Per-Error Cooldowns**: 1-hour cooldown for reporting the same error on the same word
   - **PostgreSQL-Based**: No Redis dependency, uses efficient database queries with sliding windows

2. **Smart Cooldown System**:
   - Prevents spam reporting of the same issue
   - Tracks cooldowns per user + word + error type combination
   - Returns precise retry times to users
   - Automatically clears after 1 hour

3. **User Experience Features**:
   - Clear error messages with retry times
   - Status endpoint to check remaining quota
   - Differentiated responses for rate limits vs cooldowns
   - Localized error messages in all 8 supported languages

4. **Monitoring & Analytics**:
   - Comprehensive logging of all rate limit hits
   - Admin endpoint for viewing error report statistics
   - Tracks patterns by user, error type, and subscription tier
   - Identifies potential abuse patterns

#### Implementation Architecture

```
Request Flow:
1. POST /errorReports → RateLimitErrorReports() middleware
2. Check hourly/daily limits → Block if exceeded (429 status)
3. If passed → ErrorReportsRoutes.Create handler
4. Check cooldowns → Block if too recent (429 status)
5. If passed → Process error report & trigger regeneration
```

#### Database Schema Additions

- `error_report_cooldowns`: Tracks per-error cooldowns with expiration
- `error_report_limits`: Backup tracking for rate limit analytics
- `last_regenerated_at`: Timestamps on words/definitions for cooldown grace periods

#### API Response Formats

**Rate Limit Exceeded**:
```json
{
  "error": "Daily limit exceeded. You can report 5 errors per day.",
  "retryAfter": 3600,
  "limit": 5,
  "remaining": 0,
  "windowType": "daily"
}
```

**Cooldown Active**:
```json
{
  "error": "Please wait before reporting this error again. You can retry in 45 minutes.",
  "cooldownUntil": 1234567890,
  "retryAfter": 2700,
  "windowType": "cooldown"
}
```

#### Mobile App Integration

The mobile app gracefully handles rate limiting with:
- **Automatic error type detection**: Distinguishes between rate limits and cooldowns using `windowType` field
- **Differentiated error messages**:
  - Rate limits: "Too many reports submitted. Please try again later."
  - Cooldowns: "You've already reported this error recently. Please wait a moment before reporting again."
- **Proper 429 status code handling**: Both rate limits and cooldowns return 429 with different response data
- **Smart retry time display**: Shows remaining minutes calculated from `retryAfter` field
- **Fallback for offline mode**: Displays appropriate offline message when network unavailable
- **Localized messages**: All error messages available in 8 supported languages

The implementation in both quiz and flashcard screens (`quiz.tsx` and `practice.tsx`) properly handles the `ErrorReportRateLimitError` class, checking the `windowType` to provide context-appropriate feedback to users.

This system ensures legitimate users can report genuine issues while preventing abuse that could lead to excessive API costs.

## 🧠 Leitner System for Spaced Repetition (Deterministic Implementation)

Decorebator implements a sophisticated deterministic Leitner system for optimal vocabulary retention through spaced repetition.

### How the Leitner System Works

The system uses **7 boxes** with increasing review intervals:

| Box | Review Interval | Purpose |
|-----|----------------|---------|
| **Box 1** | Immediate (always available) | New words, failed reviews |
| **Box 2** | 1 hour | Recent learning |
| **Box 3** | 1 day (24 hours) | Short-term retention |
| **Box 4** | 3 days (72 hours) | Medium-term retention |
| **Box 5** | 1 week (168 hours) | Long-term retention |
| **Box 6** | 2 weeks (336 hours) | Extended retention |
| **Box 7** | 1 month (720 hours) | Mastered words |

### Progression Rules
- **Correct Answer**: Word moves to the next box (up to Box 7)
- **Incorrect Answer**: Word resets to Box 1 regardless of current box
- **Box 7**: Words stay in Box 7 when answered correctly (mastered state)
- **Temporary Skip**: Incorrect answers trigger 10-minute skip to prevent immediate repetition

### Deterministic Selection Algorithm

The system uses a **100% deterministic priority-based selection** that guarantees word availability:

1. **Progress Calculation**: Each definition has a progress ratio (time_elapsed / target_interval)
2. **Priority Buckets**:
   - Priority 1: Overdue definitions (100%+ of interval elapsed)
   - Priority 2: Due soon definitions (80-100% of interval)
   - Priority 3: Available definitions (50-80% of interval)
   - Priority 4: All other definitions (sorted by progress %)
3. **Selection**: Always picks the highest priority definition deterministically
4. **Benefits**:
   - No randomness - same data always produces same result
   - Guaranteed selection - never "no words available" errors
   - Respects Leitner intervals precisely
   - Easy to debug and test

### Quiz Type Progression

Different quiz types are introduced based on the word's box level with minimal overlap:

| Box | Quiz Types | Learning Focus |
|-----|------------|----------------|
| **1** | Guess Meaning | Recognition |
| **2** | Word from Meaning | Basic recall |
| **3** | Word from Image | Visual association |
| **4** | Complete Sentence | Contextual understanding |
| **5** | Write Word from Definition | Active recall |
| **6** | Word from Audio, Word from Example Audio | Audio recognition |
| **7** | Meaning from Audio, Word from Image, Write Word, Complete Sentence | Mastery mix |

### Intelligent Quiz Selection

The system uses **deterministic quiz type rotation** based on:

- **Content Availability**: Validates audio, images, examples exist
- **Time-Based Rotation**: 5-minute cycles ensure all types get used
- **Fallback Logic**: Always falls back to "Guess Meaning" if content unavailable
- **Fair Distribution**: Definition ID + time rotation ensures variety

### Quiz Generation Features

1. **Definition Selection**: 
   - Deterministic priority-based algorithm
   - Respects box intervals without probability calculations
   - Excludes temporarily skipped definitions (error reports)
   - Logs selection details for monitoring

2. **Quiz Type Selection**:
   - Deterministic rotation based on definition ID and time
   - Content validation before selection
   - Ensures all available types cycle fairly

3. **Multiple Choice Generation**:
   - 3 distractor options from same part of speech
   - Excludes same word family to avoid confusion
   - Length filtering (< 50 characters) for readability
   - Random answer positioning

4. **Fair Example Distribution**:
   - Tracks example usage to prevent repetition
   - Prioritizes unused examples, then least recently used
   - 24-hour usage window for variety

### Error Reporting Integration

- **Temporary Skip**: Problematic definitions are skipped for 1 hour when errors reported
- **Error Resolution**: Once errors are resolved, definitions return to normal rotation
- **Content Quality**: Ensures users aren't repeatedly shown faulty AI-generated content

### Analytics Integration

The Leitner system feeds into comprehensive analytics:

- **Word Mastery Calculation**: Based on box progression and accuracy
- **Learning Progress Tracking**: Daily statistics per wordlist
- **Response Time Analysis**: Tracks improvement over time
- **Box Distribution**: Historical snapshots of learning progress

### Flashcard Integration

The Leitner system is seamlessly integrated with the interactive flashcard feature:

- **Study Mode**: Flashcards provide passive review without affecting box progression
- **Rich Content Display**: Shows definitions, pronunciations, part-of-speech, and contextual examples
- **Verb Inflection Support**: Special handling for verb forms with tense-specific examples
- **Error Reporting**: Direct reporting of content issues from flashcard interface
- **Progress Tracking**: View word mastery and learning statistics while studying

### Advanced Features

- **Deterministic Behavior**: 100% reproducible quiz generation for testing and debugging
- **High Box 7 Concentration Handling**: Monitors and logs when 80%+ words reach mastery
- **Transaction Safety**: All quiz results and box updates are atomic operations
- **Content Validation**: Pre-validates all content before quiz generation
- **Performance Optimization**: Single SQL query with strategic indexes for sub-second response
- **Error Recovery**: Automatic handling of problematic content with temporary skip
- **Multi-Modal Learning**: Supports both active testing (quizzes) and passive review (flashcards)
- **Debug Logging**: Comprehensive logging of selection decisions for monitoring

### Implementation Benefits

This deterministic implementation ensures:
- **Predictable Learning**: Users always have words to practice
- **Fair Scheduling**: Words are reviewed at scientifically optimal intervals
- **No Stagnation**: Even with all words in Box 7, the system continues functioning
- **Easy Testing**: Deterministic behavior simplifies unit and integration tests
- **Performance**: Efficient single-query selection with proper database indexes

## 🔐 Security

- JWT-based authentication with secure token storage
- Password hashing with bcrypt
- CORS configuration for cross-origin requests
- Environment-based configuration
- Secure storage for mobile tokens (Keychain/Keystore)
- Stripe webhook signature verification
- Subscription status enforcement at API level
- Automatic token refresh with updated subscription data

## 🚢 Deployment

### Backend Deployment
1. Build the Docker image:
```bash
cd api
docker build -t decorebator-api .
```

2. Set production environment variables
3. Run migrations on production database
4. Deploy container to your cloud provider

### Mobile Deployment
1. Build for production:
```bash
expo build:ios
expo build:android
```

2. Submit to app stores following platform guidelines

### Web Deployment
The Next.js app can be deployed to Vercel, Netlify, or any Node.js hosting:
```bash
npm run build
npm start
```

## 🔌 API Endpoints

### Authentication
- `POST /users` - Create new account with validation
- `POST /login` - User login with JWT token
- `GET /logout` - User logout
- `POST /password/send-reset-email` - Request password reset
- `PATCH /password/reset` - Reset password with token
- `GET /users` - Get user profile (authenticated)
- `PATCH /users` - Update user profile (authenticated)
- `DELETE /users` - Delete user account (authenticated)

### Subscription Management
- `POST /subscription/checkout-session` - Create Stripe checkout session
- `GET /subscription/checkout-redirect` - Handle Stripe checkout redirects
- `GET /subscription/status` - Get current subscription status
- `POST /subscription/cancel` - Cancel subscription at period end
- `GET /subscription/history` - View subscription history
- `POST /webhook/stripe` - Stripe webhook handler (no auth required)

### Wordlists & Words
- `GET /wordlists` - Get user's wordlists with stats
- `GET /wordlists/stats` - Get user's wordlist statistics
- `POST /wordlists` - Create new wordlist (subscription check)
- `GET /wordlists/:id` - Get specific wordlist
- `PUT /wordlists/:id` - Update wordlist
- `DELETE /wordlists/:id` - Delete wordlist
- `GET /wordlists/:id/words` - Get words in wordlist
- `GET /wordlists/:id/words?onlyWithDefinitions=true` - Get words with definitions only (flashcard optimization)
- `POST /wordlists/:id/words` - Add word (subscription check)
- `PUT /wordlists/:id/words/:wordId` - Update word
- `DELETE /wordlists/:id/words/:wordId` - Delete word
- `GET /wordlists/:id/words/:wordId/definitions` - Get word definitions

### Quiz & Learning
- `POST /wordlists/:id/quizzes` - Create new quiz with Leitner system
- `PATCH /wordlists/:id/quizzes` - Save quiz progress and performance

### Analytics
- `GET /analytics/wordlists/:id/mastery` - Word mastery statistics
- `GET /analytics/wordlists/:id/progress` - Learning progress (daily)
- `GET /analytics/wordlists/:id/distribution` - Box distribution history
- `GET /analytics/wordlists/:id/current-distribution` - Current box distribution (counts unique words)
- `GET /analytics/wordlists/:id/quiz-performance` - Quiz type performance stats (wordlist-scoped)
- `GET /analytics/wordlists/:id/practice-time` - Practice time tracking by day
- `GET /analytics/wordlists/:id/stats` - Comprehensive wordlist statistics
- `GET /analytics/dashboard` - Overall dashboard statistics (deprecated - use progress-summary)
- `GET /analytics/progress-summary` - Batch endpoint for all wordlists progress (reduces N×8 calls to 1)

### Error Reporting & Content Quality
- `POST /errorReports` - Report errors in AI-generated content with structured error types (rate limited)
- `GET /errorReports/status` - Get current rate limit status and active cooldowns

### Admin Endpoints (Static Authentication)
- `GET /static/admin/errorReports/stats` - Get error reporting statistics and patterns

### Worker Management (Static Authentication)
- `POST /static/workers/imageGenerator/:definitionId` - Trigger image generation
- `POST /static/workers/textToAudio/:wordId` - Trigger audio generation
- `POST /static/workers/retry/:jobId` - Retry failed job

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is proprietary software. All rights reserved.

## 🆘 Support

For issues and feature requests, please contact the development team or create an issue in the project repository.

---

Built with ❤️ using Go, React Native, and Next.js