# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Decorebator is an AI-powered vocabulary learning platform that uses AI-powered enrichment and the Leitner spaced repetition system to help users master new languages effectively. It consists of:
- **API Backend** (Go/Gin) - RESTful API with PostgreSQL database, River queue system, and AI integrations
- **Mobile App** (React Native/Expo) - Cross-platform mobile application with offline support
- **Web Frontend** (Next.js) - Landing page and web application

## Quick Reference Commands

### Most Common Development Tasks
```bash
# Start full development environment
cd api && make watch                    # Terminal 1: API server with auto-reload
cd api && make workers                  # Terminal 2: Background workers
cd mobile && npm start                  # Terminal 3: Mobile app

# Run tests
cd api && make test                     # Integration tests with Docker
cd api && make test-unit               # Fast unit tests only
cd mobile && npm test                  # Mobile Jest tests

# Database operations
cd api && make migrate-up              # Apply new migrations
cd api && make psql                    # Database console

# Database benchmarking
cd api && make db-benchmark             # Run PostgreSQL connection benchmarking
cd api && make db-benchmark ARGS="-pool -max=100"  # Test with connection pooling

# Code quality
cd api && make lint                    # Run golangci-lint
cd web && npm run lint                 # Web frontend linting
cd mobile && npm run lint              # Mobile app linting
cd mobile && npm run typecheck         # TypeScript type checking
```

### Advanced Test Script Commands
```bash
# Comprehensive test script (./scripts/run-tests.sh)
./scripts/run-tests.sh setup           # First-time test environment setup
./scripts/run-tests.sh unit            # Run unit tests only
./scripts/run-tests.sh integration     # Run integration tests
./scripts/run-tests.sh all             # Run all tests
./scripts/run-tests.sh watch           # Watch mode with auto-reload
./scripts/run-tests.sh coverage        # Check coverage thresholds
./scripts/run-tests.sh security        # Security scans (govulncheck)
./scripts/run-tests.sh clean           # Clean test environment
./scripts/run-tests.sh versions        # Tool version verification

# Quality Assurance workflow
cd api && make qa                      # Full QA suite (lint + security + tests)
cd api && make qa-fast                 # Fast QA checks (no integration tests)
cd api && make qa-ci                   # CI-style with JUnit XML output

# Single Test Execution
# API - Run specific test
go test -v -run TestSpecificFunction ./internal/service/

# API - Run specific test file
go test -v ./tests/integration/analytics_word_mastery_test.go

# Mobile - Run specific test  
npm test -- --testNamePattern="specific test name"

# Mobile - Run tests without watch mode
cd mobile && jest

# Mobile - Run specific test file
npm test -- components/quiz/QuizContent.test.tsx

# Mobile - Run tests with coverage
cd mobile && jest --coverage

# Mobile - TypeScript checking
cd mobile && npm run typecheck         # Run once
cd mobile && npm run typecheck:watch   # Watch mode
```

## Common Development Commands

### API Backend (in `/api` directory)

```bash
# Setup development environment
make setup

# Start development API server with auto-reload
make watch

# Start background workers
make workers

# Run tests with coverage
make test              # Run all tests with Docker
make test-unit         # Run unit tests only (fast)
make test-integration  # Run integration tests locally
make test-all          # Run both unit and integration tests
make test-fast         # Run tests without Docker (requires local services)
make test-watch        # Run tests in watch mode (auto-reload)
make coverage-html     # Generate HTML coverage reports
make coverage-threshold # Check if coverage meets thresholds

# Test runner script for comprehensive testing
./scripts/run-tests.sh setup       # First-time test setup
./scripts/run-tests.sh unit        # Run unit tests only
./scripts/run-tests.sh integration # Run integration tests
./scripts/run-tests.sh all         # Run all tests
./scripts/run-tests.sh watch       # Watch mode
./scripts/run-tests.sh coverage    # Check coverage thresholds
./scripts/run-tests.sh clean       # Clean test environment

# Database operations
make migrate-up         # Apply migrations
make migrate-down       # Rollback last migration
make migrate-up-test    # Run migrations on test database
make migrate-down-test  # Rollback test database migrations
make create-migration   # Create new migration file
make psql              # Open PostgreSQL console

# Debug commands
make debug-api         # Debug API with delve
make debug-workers     # Debug workers with delve

# Linting and code quality
make lint              # Run golangci-lint on the codebase
make lint-fix          # Run golangci-lint with automatic fixes
make lint-changed      # Run golangci-lint on changed files only
make lint-watch        # Run linting in watch mode

# Other commands
make build             # Build API binary
make clean             # Remove build artifacts
make help              # Show all available commands
make migrate-drop      # Drop all database tables (destructive)

# Security scanning
make security-scan     # Run govulncheck + gosec
make security-scan-full # Comprehensive security scan with reports

# Additional commands
make admin             # Admin tool commands (usage: make admin CMD=health)
make load-test         # Run load testing script with environment variables
make git-hooks-install # Install git hooks for automated linting
make git-hooks-remove  # Remove git hooks
```

### Mobile App (in `/mobile` directory)

```bash
# Start development server
npm start

# Platform-specific development
npm run android
npm run ios
npm run web
npm run android-emulator  # Start Android emulator

# Code quality
npm run lint
npm run lint-fix       # Automatic ESLint fixes
npm run typecheck      # TypeScript type checking
npm run typecheck:watch # TypeScript checking in watch mode
npm run test           # Run Jest tests in watch mode
npm run prepare        # Setup Husky git hooks

# Update Expo dependencies
npm run expo:update

# EAS Updates (over-the-air updates)
npm run update:local   # Update local development build
npm run update:prod    # Update production build
```

### Web Frontend (in `/web` directory)

Next.js 15 application with modern stack and internationalization:

```bash
# Start development server (uses Turbopack for fast refresh)
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Generate sitemap
npm run sitemap

# Lint code
npm run lint
```

#### Modern Stack & Features
- **Next.js 15** with Turbopack for fast development
- **Tailwind CSS 4** for styling
- **TypeScript 5** with strict configuration
- **next-intl 4** for internationalization
- **7 Language Support**: EN, DE, ES, FR, IT, JA, PT
- **Dynamic Routing**: `/[locale]/...` pattern with middleware for locale detection
- **SEO Optimized**: Sitemap generation, structured data
- **Performance**: Turbopack dev server, optimized builds

## Architecture Overview

### API Backend Architecture

The API follows a 3-tier layered architecture:

1. **HTTP Layer** (`internal/http/`) - Request handling, authentication, response formatting
2. **Service Layer** (`internal/service/`) - Business logic and orchestration
3. **Repository Layer** (`internal/repository/`) - Database operations

Key architectural decisions:
- **Modern Dependency Injection**: Complete AppContext-based dependency injection with constructor pattern
- **Clean Service Architecture**: All services use explicit dependency injection, no internal construction
- **Centralized Configuration**: AppContext builder pattern manages all service dependencies
- JWT-based authentication with automatic session refresh
- Background job processing using River queue system (PostgreSQL-backed, replaces Redis-based queues)
- MinIO for S3-compatible object storage (images, audio)
- OpenAI API integration (DALL-E, TTS, GPT) for AI-powered content generation
- Dual subscription system: Stripe (web, US iOS) + RevenueCat (Android, non-US iOS)
- SendGrid for email services (subscription notifications)
- Sentry for error monitoring and logging
- Structured logging with `slog` (enhancement planned)
- Redis-based analytics caching with intelligent invalidation strategies

### Background Job Processing

River-based PostgreSQL queue system with dedicated worker types:
- `image_generator` - Generates images using OpenAI DALL-E (max 5 workers)
- `text_to_speech` - Converts text to audio using OpenAI TTS (max 30 workers)
- `definition_fetcher` - Fetches word definitions from external sources (max 50 workers)
- `subscription_reminder` - Sends renewal reminder emails (max 10 workers)
- `example_audio_generator` - Generates audio for example sentences with fair usage distribution (max 20 workers)
- `revenuecat_worker` - Processes RevenueCat webhook events and subscription syncing (max 5 workers)
- `stripe_webhook_worker` - Processes Stripe webhook events asynchronously (max 5 workers)

**Note**: River tables were removed in migration 000052. The system now uses PostgreSQL-native queuing for better transaction safety and simplified architecture.

**Recent Improvements**:
- **Analytics Caching Layer**: Redis-based caching with intelligent invalidation for sub-second response times
- **Batch Analytics Endpoint**: New `/analytics/progress-summary` reduces API calls from N×8 to 1 for mobile dashboard
- **Asynchronous Webhook Processing**: Both Stripe and RevenueCat webhooks processed via River workers

Workers run as a separate process and include retry logic, rate limiting, and error handling.

### Mobile App Architecture

- Expo Router for navigation
- React Query for API state management with offline caching
- React Hook Form with Zod validation
- React Native Paper for UI components
- Secure storage for JWT tokens (Keychain/Keystore)
- Automatic session refresh on focus
- Real-time subscription status updates
- Internationalization (i18n) support for 8 languages
- PostHog integration for analytics tracking
- Offline support for premium users with local storage
- Error reporting modal for AI-generated content issues
- Interactive flashcard system with flip animations

#### Mobile App Code Quality Standards

**TypeScript Configuration**:
- Strict TypeScript enabled with comprehensive type checking
- Use `npm run typecheck` before committing mobile changes
- Fix all TypeScript errors - the build should have zero type issues
- Common patterns: Proper typing for React components, hooks, and API interfaces

**Testing Patterns**:
- Jest with Expo preset for unit testing
- React Native Testing Library for component testing
- Mock all external dependencies (NetInfo, AsyncStorage, RevenueCat, etc.)
- Tests located alongside components in `__tests__/` directories
- Use descriptive test names: `describe("Component") { it("should do something when condition", () => {}) }`

**Component Architecture**:
- **Responsive Design System**: Uses `useTheme()` hook with responsive utility functions
- **Form Components**: Progressive vs Traditional signup forms based on screen size
- **Modular Analytics**: 9 separate analytics components for dashboard modularity
- **Error Handling**: Centralized error reporting modal component
- **Theme Support**: Light/dark theme support with responsive spacing system

**Key Files & Patterns**:
- `contexts/ThemeContext.tsx`: Responsive design system with spacing utilities
- `components/auth/`: Progressive and traditional auth forms with proper TypeScript types
- `hooks/useI18n.tsx`: Language detection and backend synchronization
- `utils/optimisticSubscription.ts`: RevenueCat subscription state management
- Test mocking in `__mocks__/` and individual test files

#### Detailed Component Organization

**Authentication Components** (Enhanced):
- `ProgressiveSignupForm` - Multi-step signup for small devices
- `TraditionalSignupForm` - Single-step signup for larger screens
- `EmailInput`, `PasswordInput` - Reusable form inputs with validation
- Language detection integration with backend user profile synchronization

**Analytics Components** (7 modular components):
- `AnalyticsHeader`, `StatsGrid`, `WordMasteryChart`, `LearningProgressChart`
- `QuizPerformanceChart`, `BoxDistributionChart`, `HistoricalBoxDistributionChart`, `TopWordsSection`

**Quiz Components**: Modular quiz interface with `QuizContent`, `QuizOptions`, `QuizHeader`, `QuizModeToggle`

**Flashcard Components**: 4-component flashcard system with flip animations and navigation

**Offline Components**: `OfflineManager`, `OfflineIndicator`, `OfflinePreloader` for premium offline support

#### Performance & UX Patterns
- **Loading State Architecture**: 10-second timeout detection with retry mechanisms
- **Error Handling**: Network errors, timeout errors, offline errors with graceful degradation
- **Memory Management**: Bounded cache, animation cleanup, audio management, timer cleanup
- **Design System**: Card-based design with warm gradients, 8px base grid, touch-friendly targets

#### State Management
- **Server State**: React Query with intelligent caching and background updates
- **Local State**: React hooks with custom hook extraction for shared logic
- **Persistence**: Secure storage for tokens, AsyncStorage for preferences/offline data

### Dependency Injection Architecture

The API uses a modern, centralized dependency injection system built around the **AppContext** pattern:

#### **AppContext Builder Pattern**
```go
// Initialize all services with clean dependency injection
appCtx, err := app.NewContext().
    WithDatabase(db).
    WithEnvironment("development").
    Build()
```

#### **Service Dependencies**
All services follow constructor-based dependency injection:
- **No internal service construction** - All dependencies are explicitly injected
- **Clean testing** - Every dependency can be mocked for unit tests
- **Explicit dependency graph** - Constructor signatures show all requirements
- **SOLID compliance** - Dependency Inversion Principle properly implemented

#### **Key Service Relationships**
```
Database → JobService → Core Services → Composite Services
- DefinitionService (db)
- LeitnerTrackingService (db)
- WordService (db, DefinitionService, ModerationService, JobService, LeitnerTrackingService)
- ErrorReportService (db, DefinitionService, WordService, LeitnerTrackingService, JobService)
- UserService (db, SubscriptionRepository, ErrorReportService)
```

#### **JobService Interface**
Clean, consistent method naming for background job scheduling:
- `ScheduleImageJob()` - Image generation via DALL-E
- `ScheduleAudioJob()` - Text-to-speech processing
- `ScheduleDefinitionJob()` - Definition fetching and processing
- `ScheduleExampleAudioJob()` - Example sentence audio generation
- `ScheduleStripeWebhookJob()` / `ScheduleRevenueCatWebhookJob()` - Webhook processing

### Database Schema

Key tables with recent enhancements:
- `users` - User accounts, authentication, subscription status, and profile data with `preferred_language` field
- `subscriptions` - Subscription history and details with Stripe integration
- `subscription_events` - Stripe webhook event audit trail & email notification tracking
- `wordlists` - User's vocabulary lists with language field and word counts
- `words` - Individual words with audio URLs, learning status, and `part_of_speech_normalized` for language-agnostic grammar
- `definitions` - AI-generated definitions with multimedia, sources, example sentences, and `is_verb_type` virtual column
- `definition_images` - Images associated with definitions
- `leitner_system_tracking` - Spaced repetition progress with temporary skip functionality
- `error_reports` - User-reported errors for definitions and AI-generated content
- `quiz_performance` - Individual quiz attempts with performance metrics
- `word_mastery` - Overall mastery tracking for each word per user
- `learning_progress` - Daily aggregated learning statistics
- `quiz_type_analytics` - Performance metrics grouped by quiz type
- `box_distribution_snapshot` - Daily snapshots of word distribution across Leitner boxes
- `definition_example_audio` - Audio files for example sentences
- `example_audio_usage` - Tracks audio generation for fair usage distribution
- `example_usage` - Ensures variety in quiz example selection
- `revenuecat_events` - RevenueCat webhook event audit trail
- Materialized views refreshed hourly:
  - `mv_word_mastery_current` - Current word mastery state
  - `mv_quiz_type_performance` - Quiz performance analytics

## Environment Setup

### Multi-Environment Configuration

**Development Environment:**
1. Copy the environment template:
```bash
cp .env.example .env
```

2. Configure required services in `.env`:
- PostgreSQL connection details
- MinIO credentials for object storage
- OpenAI API key for AI features
- Stripe API keys and webhook secret (for web and US iOS)
- RevenueCat API keys and webhook secret (for Android and non-US iOS)
- SendGrid API key for emails
- JWT secret for authentication
- Sentry DSN for error monitoring (optional)

**Test Environment:**
1. Copy the test environment template:
```bash
cp .env.test.example .env.test
```

2. Test services run on isolated ports:
- PostgreSQL: 5433 (vs 5432 for dev)
- MinIO: 9001 (vs 9000 for dev)
- Redis: 6380 (vs 6379 for dev)

3. Use isolated test services:
```bash
docker-compose -f docker-compose.test.yml up -d
```

### Docker Volume Management
```bash
# Create required Docker volumes for persistence
docker volume create api_postgres_data
docker volume create api_minio-data
```

### Tool Version Alignment
All development tools match GitHub Actions for consistency:
- Go: 1.23
- PostgreSQL: 15
- Redis: 7-alpine
- gotestsum: latest (structured test output)
- golangci-lint: v1.62.2
- govulncheck: latest (security scanning)

**Version Verification:**
```bash
make check-versions                    # Compare installed vs expected versions
make setup                            # Install all tools with correct versions
./scripts/run-tests.sh versions      # Detailed version report
```

## Subscription System

### Dual-Provider Architecture
The platform uses an intelligent dual-provider system:
- **Stripe**: Web users and US iOS users (direct payment processing)
- **RevenueCat**: Android users and non-US iOS users (native app store integration)

### Features
- Dual payment provider support (Stripe + RevenueCat)
- Monthly ($6.99) and Annual ($69.9) plans
- Free plan with limits (1 wordlist, 10 words max)
- **Asynchronous webhook processing** via River workers for both providers
- **Unified subscription event tracking** across all payment providers
- Automatic email notifications for subscription events
- Purchase restoration for RevenueCat subscriptions

### Provider Selection Logic
```
- Android users → RevenueCat (Google Play Store)
- US iOS users → Stripe (direct payment)  
- Non-US iOS users → RevenueCat (App Store)
- Web users → Stripe (direct payment)
```

### Stripe Subscription Flow
1. User initiates checkout from mobile app
2. Stripe checkout session created with user metadata
3. User completes payment on Stripe hosted page
4. Webhook updates subscription in database
5. User returns to app → automatic session refresh
6. New JWT issued with updated subscription plan
7. Premium features instantly available

### RevenueCat Subscription Flow
1. User opens native paywall in app
2. RevenueCat displays available packages from App Store/Play Store
3. User completes purchase using platform payment method
4. RevenueCat webhook notifies backend
5. Subscription status synced to database
6. Premium features instantly available

### RevenueCat Error Handling
**IMPORTANT**: RevenueCat error handling uses string error codes, not enum constants:
- Error codes are strings: "1" (user cancelled), "2" (store problem), "3" (purchase not allowed), etc.
- Check `error.userCancelled` property for user cancellation
- Provide specific error messages instead of generic "Purchases are not allowed on this device"
- All error messages are translated in 8 supported languages

### Grace Period Implementation

The platform implements a **3-day grace period** for subscription billing issues, providing users continued access while payment problems are resolved:

**Grace Period Logic:**
- **Trigger**: When subscription status changes to `past_due` due to billing errors
- **Duration**: 3 days from the subscription's original `current_period_end` date
- **Access**: Users maintain full premium features during grace period
- **Implementation**: Mathematical calculation using `IsActive()` method in subscription model

**Key Features:**
```go
// Grace period calculation in subscription.IsActive()
if s.Status == StatusPastDue {
    gracePeriodEnd := s.CurrentPeriodEnd.Add(GracePeriodDays * 24 * time.Hour)
    return time.Now().Before(gracePeriodEnd)
}
```

**Backend Behavior:**
- `GetProfile()` checks for expired grace periods and automatically downgrades users to free plan
- Grace period expiration triggers immediate subscription plan downgrade
- JWT tokens are refreshed with updated subscription status
- Mobile app receives cache invalidation for seamless UX updates

**Mobile App Integration:**
- Automatic cache invalidation when subscription status changes during grace period
- Real-time subscription status updates via React Query
- Proper handling of subscription downgrades with user notification

## Testing Strategy

### API Tests
- Integration tests using `httpexpect`
- Test database with Docker Compose (`docker-compose.test.yml`)
- Coverage requirements: 70% unit, 80% integration
- AAA (Arrange-Act-Assert) test structure
- Test naming: `Test[Subject]_[Scenario]_[Expected]`
- Mock servers for external services (OpenAI, Stripe, SendGrid)
- Transaction-based test isolation
- Performance benchmarking support
- Security testing patterns

### Test Coverage Targets
- **Unit Tests**: 70% minimum coverage
- **Integration Tests**: 80% minimum coverage  
- **Critical paths**: 95% coverage (auth, subscriptions, core business logic)

### Test Naming Convention
Format: `Test[Subject]_[Scenario]_[Expected]`
```go
func TestUserRegistration_WithValidData_ReturnsCreatedUser(t *testing.T)
func TestCreateWordlist_WhenFreePlanLimitExceeded_Returns403(t *testing.T)
```

### Test Structure (AAA Pattern)
- **Arrange**: Setup test data and dependencies
- **Act**: Execute the operation being tested  
- **Assert**: Verify expected outcomes

### Integration Test Patterns
- Transaction-based isolation with automatic rollback
- Mock external services (OpenAI, Stripe, SendGrid) using httptest
- Test data fixtures for complex scenarios

### Running Tests
```bash
# Run single test
go test -v -run TestName

# Run tests in specific package
go test -v ./internal/service/...

# Run with race detection
go test -race ./...

# Generate coverage report
make coverage-html
```

### Mobile Tests
- Jest with Expo preset configured in `package.json`
- Run with `npm test` (runs in watch mode with `--watchAll`)
- Run single test: `npm test -- --testNamePattern="test name"`
- To run tests without watch mode: `jest` (directly)
- Test files located in component directories alongside source files

### Mobile Testing Requirements
**CRITICAL**: All mobile changes must pass TypeScript checking before committing:
```bash
cd mobile && npm run typecheck
```

**Common Testing Patterns**:
- Mock all native modules (`NetInfo`, `AsyncStorage`, `RevenueCat`, etc.)
- Use `__mocks__/` directory for global mocks
- Mock components that require native functionality
- Test authentication flows include `preferredLanguage` parameter
- Mock responsive theme utilities in tests

**Required Mocks for Mobile Tests**:
```javascript
// Essential mocks that should be in every test file touching these modules
jest.mock("@react-native-community/netinfo")
jest.mock("@react-native-async-storage/async-storage")
jest.mock("posthog-react-native")
jest.mock("react-native-purchases")
jest.mock("@/contexts/ThemeContext")
```

### Database Query Testing (CRITICAL REQUIREMENT)

**ALL database query changes MUST be tested directly in PostgreSQL before committing code.**

#### Why Direct Database Testing is Required
- **PostgreSQL Version Compatibility**: Ensures queries work on our specific PostgreSQL 15+ version
- **Syntax Validation**: Catches PostgreSQL-specific syntax errors that Go compilation cannot detect
- **Performance Verification**: Validates query execution plans and performance characteristics
- **Data Type Compatibility**: Prevents runtime type mismatch errors (e.g., date vs timestamp issues)
- **Complex Query Validation**: Essential for recursive CTEs, complex JOINs, and advanced SQL features

#### Testing Process for Database Changes
1. **Extract SQL Queries**: Copy exact SQL from Go repository files
2. **Test with Real Parameters**: Replace `$1`, `$2` placeholders with actual test values
3. **Verify on Target Database**: Test on same PostgreSQL version used in production
4. **Check Execution Plans**: Use `EXPLAIN` to verify query performance
5. **Test Edge Cases**: Verify with NULL values, empty results, and boundary conditions

#### Database Connection for Testing
```bash
# Use development database for testing
psql "postgresql://user:pass@localhost:5432/decorebator?sslmode=disable"

# Test query syntax with EXPLAIN
EXPLAIN (FORMAT TEXT) SELECT ...;

# Test with actual data
SELECT ... WHERE user_id = 137 AND wordlist_id = 100;
```

**NEVER commit database query changes without direct PostgreSQL testing verification.**

## CI/CD Pipeline

### GitHub Actions Workflow
- **5 Parallel Jobs**: unit-tests, integration-tests, lint-checks, build-verification, security-scan
- **Enhanced Test Reporting**: JUnit XML + JSON reports with detailed failure analysis
- **Structured Output**: Individual test failure names, error messages, and execution timing
- **Security Integration**: gosec SARIF reports, govulncheck scanning
- **Coverage Enforcement**: 70% unit, 80% integration coverage thresholds
- **Artifact Management**: Test results, coverage reports, security scans, detailed failure logs
- **Quality Checks**: golangci-lint, gosec, go vet, race condition detection

#### Enhanced Test Failure Reporting (January 2025)
- **Specific Test Names**: Shows exact test functions that failed instead of just counts
- **Error Messages**: Displays actual failure messages and assertion details
- **Execution Timing**: Reports test duration and performance metrics
- **Structured Summaries**: Markdown tables with test status, duration, and failure reasons
- **Multiple Formats**: Both XML (JUnit) and JSON output for comprehensive analysis
- **Gotestsum Integration**: Uses `pkgname-and-test-fails` format for enhanced failure visibility

#### 5-Job Parallel CI/CD Pipeline
- **unit-tests**: gotestsum with JUnit XML + JSON output, specific failure details
- **integration-tests**: Full service stack (PostgreSQL, Redis, MinIO)
- **lint-checks**: golangci-lint on changed files (PR only)
- **build-verification**: Multi-binary builds + dependency verification
- **security-scan**: gosec SARIF + govulncheck + go vet

## External Services

- **PostgreSQL 15+** - Primary database with materialized views and pgx/v5 driver
- **MinIO** - S3-compatible object storage for images and audio
- **Redis** - Analytics caching layer with automatic invalidation
- **SendGrid** - Email delivery for subscription notifications
- **OpenAI API** - Image generation (DALL-E), text-to-speech (TTS), and AI content generation (GPT)
- **Stripe** - Payment processing and subscription management with webhook integration
- **RevenueCat** - Native app store subscription management (Android, non-US iOS)
- **Sentry** - Error monitoring and logging

## Development Workflow

1. API server requires `.env` file with database and service credentials
2. Run `docker-compose up` to start PostgreSQL, MinIO, and Redis
3. Apply database migrations before starting the API
4. Start workers separately when testing background jobs
5. Mobile app connects to API (configure API URL in `mobile/api/constants.ts`)
6. For subscription testing, configure Stripe webhook endpoint and test keys
7. Use `make watch` for API development with auto-reload
8. Run `make test` before committing to ensure tests pass
9. **CRITICAL**: Run `npm run typecheck` in mobile directory before committing mobile changes

## Error Reporting System

### Rate Limiting
To prevent abuse and control API costs, error reporting implements comprehensive rate limiting:

| Tier | Hourly Limit | Daily Limit |
|------|--------------|-------------|
| **Free** | 3 reports | 5 reports |
| **Premium** | 10 reports | 30 reports |

- 1-hour cooldown for reporting the same error on the same word
- PostgreSQL-based rate limiting without Redis dependency
- Clear error messages with retry times
- Status endpoint to check remaining quota

## Multi-Language Support & Internationalization

### Mobile App Internationalization
- **8 supported languages**: English, German, Spanish, French, Italian, Japanese, Portuguese (Brazil), Portuguese (Portugal)
- **Real-time language switching** without app restart
- **User Language Detection**: Mobile app detects UI language and sends it during signup to backend
- **Backend Synchronization**: User's `preferred_language` stored in database matches detected UI language
- **Language Mapping**: `useI18n.tsx` provides `backendLanguageMap` for converting i18n codes to backend format
- **Cultural localization** including currency symbols and date formats

### AI Content Generation
- **7 languages for AI-powered content**: English, Spanish, French, German, Italian, Portuguese, Japanese
- **Native language processing** with language-specific grammar rules and verb systems
- **Culturally-aware image generation** prompts in target language
- **Language-optimized voice selection** for text-to-speech (alloy, nova, shimmer, echo, fable, onyx)
- **Dynamic part-of-speech validation** per language
- **Automatic language detection** and content adaptation
- **Multi-modal AI features**: definitions, images, audio, and example sentences

### Language Implementation Details
**Signup Flow Language Detection**:
- Mobile app detects current UI language using `i18n.language`
- Converts to backend format using `backendLanguageMap` in `hooks/useI18n.tsx`
- Sends `preferredLanguage` parameter during user registration
- Backend stores in `users.preferred_language` field
- Prevents language mismatch between UI and user profile

## Important Notes

- Authentication uses JWT tokens stored securely on mobile devices
- Subscription limits are enforced at the API level
- Background jobs use River queue (PostgreSQL-backed) instead of traditional queue systems
- Email templates are located in `internal/mail/` directory
- API endpoints are documented in `doc/words.http` and `doc/words.prod.http`
- Dual subscription system: Existing Stripe subscriptions remain unchanged; new subscriptions route to appropriate provider
- RevenueCat integration documented in `REVENUECAT_INTEGRATION.md`
- Recent architecture improvements documented in:
  - `docs/DETERMINISTIC_LEITNER_IMPLEMENTATION.md` - Pure Priority Leitner algorithm with clean 3-tier tiebreaking
  - `docs/TESTING_BEST_PRACTICES.md` - Detailed testing guidelines and patterns
  - `docs/SUBSCRIPTION_SYSTEM.md` - Complete subscription system documentation
  - `docs/WORKER_ABUSE_PREVENTION.md` - Background job abuse prevention strategies

## Core Features

- **Multiple Quiz Modes**: Guess meaning, word from meaning, visual association, audio comprehension, sentence completion, active recall
- **Interactive Flashcards**: Full-screen study mode with rich content display and flip animations
- **Advanced Leitner System**: 7-box spaced repetition with intelligent quiz type progression
- **Error Reporting System**: User-driven quality control for AI-generated content with automatic regeneration
- **Comprehensive Analytics**: Word mastery tracking, learning progress visualization, and performance metrics
- **Offline Support**: Premium users can access wordlists and practice offline with seamless sync
- **Subscription Tiers**: Free plan (1 wordlist, 10 words) and Premium plans ($6.99/month, $69.90/year)

## Advanced Leitner System - Deterministic Implementation

The system uses a deterministic priority-based algorithm for word selection, ensuring consistent and predictable learning experiences.

**Box Review Intervals:**
- Box 1: Immediate (new/failed words)
- Box 2: 6 hours
- Box 3: 1 day  
- Box 4: 3 days
- Box 5: 1 week
- Box 6: 2 weeks
- Box 7: 1 month

**Selection Algorithm:**
1. Words due for review (past their interval)
2. Words temporarily skipped due to error reports
3. Quiz type progression tied to box level
4. 100% deterministic - same conditions produce same results

**Features:**
- Prevents Box 7 stagnation with intelligent scheduling
- Temporary skip for reported errors (auto-resumed after regeneration)
- Progressive quiz difficulty matching box level
- Consistent user experience across sessions

## Documentation Architecture

**Specialized Documentation by Domain:**
```bash
docs/                                   # Architecture & best practices
├── DEPENDENCY_INJECTION_MODERNIZATION_PLAN.md  # Critical refactoring roadmap
├── TESTING_BEST_PRACTICES.md         # Comprehensive testing guide
├── DETERMINISTIC_LEITNER_IMPLEMENTATION.md     # Core algorithm specs
├── SUBSCRIPTION_SYSTEM.md            # Payment system architecture
├── ANALYTICS_TESTING_IMPLEMENTATION.md # Analytics test patterns
└── WORKER_ABUSE_PREVENTION.md       # Background job security

api/docs/                              # API-specific documentation
└── *.http files                      # API endpoint examples

mobile/docs/                           # Mobile app documentation
├── mobile-app-architecture.md        # App structure & patterns
├── state-management-patterns.md      # Data flow & caching
└── offline-feature.md               # Offline functionality

web/docs/                             # Web frontend documentation
├── WEB_APP_ARCHITECTURE.md          # Next.js 15 architecture
└── WEB_APP_IMPROVEMENT_PLAN.md      # Enhancement roadmap
```

**Documentation Standards:**
- All major features documented in `/docs/` before implementation
- API changes require updating `.http` files for examples
- Architecture decisions documented with rationale and migration plans

## Important Development Notes

### Code References
When referencing specific functions or pieces of code, include the pattern `file_path:line_number` to allow easy navigation to the source code location.

**Example**: Clients are marked as failed in the `connectToServer` function in `src/services/process.ts:712`.

### Critical Development Requirements
- **NEVER commit changes unless explicitly asked** - Only commit when the user specifically requests it
- **ALWAYS run `npm run typecheck` in mobile directory** before committing mobile changes
- **Database query changes MUST be tested directly in PostgreSQL** before committing code
- **ALL mobile changes must pass TypeScript checking** before committing: `cd mobile && npm run typecheck`
- **Run `make test` before committing** to ensure API tests pass
- **Use `make qa` for comprehensive quality assurance** before major releases

### Memory Guidelines
- Read README.md for additional context on the Decorebator project
- Read all files inside the docs folder for comprehensive documentation and patterns
- Update README.md right after introducing major features or refactorings
- Use api/Makefile and api/scripts/run-tests.sh as the source for main automations and commands in this monorepo
- Follow the testing patterns documented in `docs/TESTING_BEST_PRACTICES.md`
- Refer to `docs/SUBSCRIPTION_SYSTEM.md` for payment system details
- Check `docs/DETERMINISTIC_LEITNER_IMPLEMENTATION.md` for learning algorithm specifications