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

# Code quality
cd api && make lint                    # Run golangci-lint
cd web && npm run lint                 # Web frontend linting
cd mobile && npm run lint              # Mobile app linting
```

### Single Test Execution
```bash
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
make lint-watch        # Run linting in watch mode

# Other commands
make build             # Build API binary
make clean             # Remove build artifacts
make help              # Show all available commands
make migrate-drop      # Drop all database tables (destructive)
```

### Mobile App (in `/mobile` directory)

```bash
# Start development server
npm start

# Platform-specific development
npm run android
npm run ios
npm run web

# Code quality
npm run lint
npm run test          # Run Jest tests in watch mode
npm run prepare       # Setup Husky git hooks

# Update Expo dependencies
npm run expo:update

# EAS Updates (over-the-air updates)
npm run update:local   # Update local development build
npm run update:prod    # Update production build
```

### Web Frontend (in `/web` directory)

Next.js 15 application with internationalization:

```bash
# Start development server (uses Turbopack for fast refresh)
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Lint code
npm run lint
```

#### Internationalization Setup
- Uses `next-intl` for i18n with 7 supported languages
- Dynamic routing with locale prefixes (`/[locale]/...`)
- Middleware for automatic locale detection
- Localized feature showcase pages
- Messages stored in `/messages/*.json`

## Architecture Overview

### API Backend Architecture

The API follows a 3-tier layered architecture:

1. **HTTP Layer** (`internal/http/`) - Request handling, authentication, response formatting
2. **Service Layer** (`internal/service/`) - Business logic and orchestration
3. **Repository Layer** (`internal/repository/`) - Database operations

Key architectural decisions:
- Singleton pattern for database connections using `sync.Once`
- Manual dependency injection without frameworks (modernization planned)
- JWT-based authentication with automatic session refresh
- River queue system for background jobs (PostgreSQL-backed)
- MinIO for S3-compatible object storage (images, audio)
- OpenAI API integration (DALL-E, TTS, GPT) for AI-powered content generation
- Stripe integration for subscription management with webhook processing
- SendGrid for email services (subscription notifications)
- Sentry for error monitoring and logging
- Structured logging with `slog` (enhancement planned)
- Redis-based analytics caching with intelligent invalidation strategies

### Background Job Processing

Five worker queues process asynchronous tasks:
- `image_generator` - Generates images using OpenAI DALL-E (max 5 workers)
- `text_to_speech` - Converts text to audio using OpenAI TTS (max 30 workers)
- `definition_fetcher` - Fetches word definitions from external sources (max 50 workers)
- `subscription_reminder` - Sends renewal reminder emails (max 10 workers)
- `example_audio_generator` - Generates audio for example sentences with fair usage distribution (max 20 workers)

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

#### Detailed Component Organization

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

### Database Schema

Key tables with recent enhancements:
- `users` - User accounts, authentication, subscription status, and profile data
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
- `river_job` - Background job queue
- Materialized views refreshed hourly:
  - `mv_word_mastery_current` - Current word mastery state
  - `mv_quiz_type_performance` - Quiz performance analytics

## Environment Setup

### Development Environment

1. Copy the environment template:
```bash
cp .env.example .env
```

2. Configure required services in `.env`:
- PostgreSQL connection details
- MinIO credentials for object storage
- OpenAI API key for AI features
- Stripe API keys and webhook secret
- SendGrid API key for emails
- JWT secret for authentication
- Sentry DSN for error monitoring (optional)

### Test Environment

1. Copy the test environment template:
```bash
cp .env.test.example .env.test
```

2. Test services run on different ports:
- PostgreSQL: 5433 (vs 5432 for dev)
- MinIO: 9001 (vs 9000 for dev)
- Redis: 6380 (vs 6379 for dev)

3. Use `docker-compose.test.yml` for isolated test services:
```bash
docker-compose -f docker-compose.test.yml up -d
```

## Subscription System

### Features
- Stripe integration for payment processing
- Monthly ($6.99) and Annual ($69.9) plans
- Free plan with limits (1 wordlist, 10 words max)
- Webhook processing for real-time updates
- Automatic email notifications for subscription events

### Subscription Flow
1. User initiates checkout from mobile app
2. Stripe checkout session created with user metadata
3. User completes payment on Stripe hosted page
4. Webhook updates subscription in database
5. User returns to app → automatic session refresh
6. New JWT issued with updated subscription plan
7. Premium features instantly available

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

#### Recent Database Query Fixes (January 2025)
- **Type Mismatch Errors**: Fixed `(sd.practice_date - INTERVAL '1 day')::date` casting in recursive CTEs
- **Recursive CTE Syntax**: Corrected `WITH RECURSIVE` placement in multi-CTE queries
- **Comprehensive Testing**: All 15 analytics repository queries verified working correctly

#### Files Requiring Special Attention
- `internal/repository/analytics/` - Complex analytical queries with CTEs and recursive logic
- Any repository files with raw SQL strings - Syntax must be PostgreSQL-compatible
- Migration files - Must be tested on exact target PostgreSQL version

**NEVER commit database query changes without direct PostgreSQL testing verification.**

## CI/CD Pipeline

### GitHub Actions Workflow
- Automated testing on push/PR to master, main, develop branches
- Unit tests, integration tests, linting, and security checks
- Coverage threshold enforcement (70% unit, 80% integration)
- Build verification for all binaries
- Uses `golangci-lint` for code quality
- `gosec` for security scanning
- Race condition detection in tests

## External Services

- **PostgreSQL 15+** - Primary database with materialized views and pgx/v5 driver
- **MinIO** - S3-compatible object storage for images and audio
- **Redis** - Analytics caching layer with automatic invalidation
- **SendGrid** - Email delivery for subscription notifications
- **OpenAI API** - Image generation (DALL-E), text-to-speech (TTS), and AI content generation (GPT)
- **Stripe** - Payment processing and subscription management with webhook integration
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

## Known Architecture Issues & Modernization Plans

### Critical Issues Requiring Attention
1. **Global State Anti-Patterns**: Service layer uses global variables with `init()` functions containing `os.Exit(1)` calls that break testing
2. **Connection Pool Inefficiency**: Creating new service instances repeatedly instead of reusing connections

### Recent Performance Improvements
- **Analytics Caching**: Implemented Redis-based caching layer (`analytics_cached.go`) with automatic invalidation
- **Database Indexes**: Added performance indexes for analytics queries (migration 000044)
- **SQL Injection Fixes**: Resolved SQL injection vulnerabilities in analytics queries

### Analytics System Improvements (January 2025)

**Batch Analytics Endpoint**: 
- New `/analytics/progress-summary` endpoint reduces API calls from N×8 to 1
- Single request returns all analytics data for dashboard
- Optimized for mobile app performance

### Analytics Data Quality Fixes (January 2025)
- **Fixed Critical Streak Calculation Bug**: Corrected broken gap-and-islands logic in `GetWordlistCurrentStreak` and `GetAllWordlistsProgress`
  - **Issue**: `ROW_NUMBER() OVER (ORDER BY date DESC)` caused consecutive dates to have different group identifiers
  - **Solution**: Implemented recursive CTE approach that accurately counts backwards from most recent practice
  - **Files**: `learning_progress.go:176-212`, `batch_progress.go:73-112`
- **Corrected Box Distribution Logic**: Fixed inconsistent word progress tracking across analytics
  - **Issue**: Used `MIN(box_id)` showing worst performance instead of best progress per word
  - **Solution**: Changed to `MAX(box_id)` for consistency with word mastery analytics showing highest achievement
  - **Files**: `box_distribution.go:45-64`, `box_distribution.go:109-130`
- **Enhanced Response Time Filtering**: Added consistent outlier detection across all analytics functions
  - **Practice Time**: Filters to 200ms-30s range excluding accidental clicks and timeouts
  - **Quiz Performance**: Added NULL handling with COALESCE protection and realistic range filtering
  - **Files**: `dashboard_stats.go:48-72`, `quiz_performance.go:72-81`, `quiz_performance.go:127-150`
- **Fixed Word Count Inconsistencies**: Corrected mastery statistics to count all words in wordlists
  - **Issue**: Only counted words with existing mastery data, not total wordlist size
  - **Solution**: Use LEFT JOIN from words table to include all active learning words
  - **Files**: `word_mastery.go:152-162`
- **Database Schema Cleanup**: Removed unused `study_time_seconds` column that was never maintained
  - **Migration**: `000046_remove_unused_study_time_seconds.up.sql`
  - **Code Cleanup**: Updated models and queries to remove dead code references
  - **Files**: `learning_progress.go:88`, `analytics_types.go:73-81`
- **Added Consistent Learned Words Filtering**: All analytics now exclude completed words (`w.learned = FALSE`)
  - **Impact**: Analytics focus on active learning progress, consistent across all functions

### Planned Modernization (Priority Order)
1. **Week 1-2**: Remove `init()` functions, create repository interfaces
2. **Week 3-4**: Convert services to constructor-based DI
3. **Week 5-6**: Implement dependency container, refactor HTTP layer

**Files requiring immediate attention:**
- `internal/service/word.go`, `internal/service/user.go` - Remove global repositories
- `internal/common/database.go` - Remove `os.Exit(1)` calls

## Important Notes

- Authentication uses JWT tokens stored securely on mobile devices
- Subscription limits are enforced at the API level
- Background jobs use River queue (PostgreSQL-backed) instead of traditional queue systems
- Email templates are located in `internal/mail/` directory
- API endpoints are documented in `doc/words.http` and `doc/words.prod.http`
- Recent architecture improvements planned in:
  - `api/docs/DEPENDENCY_INJECTION_MODERNIZATION_PLAN.md` - DI framework adoption
  - `api/docs/LOGGING_IMPROVEMENT_PLAN.md` - Enhanced structured logging
  - `api/docs/ANALYTICS_REVIEW_REPORT.md` - Performance and bug fixes for analytics
  - `api/docs/PROBABILISTIC_LEITNER_IMPLEMENTATION.md` - Advanced Leitner system solution
  - `api/docs/DETERMINISTIC_LEITNER_IMPLEMENTATION.md` - Deterministic priority-based Leitner algorithm
  - `api/docs/LEITNER_ALGORITHM_OPTIMIZATION_REPORT.md` - Comprehensive Leitner system optimization analysis
  - `api/docs/TESTING_BEST_PRACTICES.md` - Detailed testing guidelines and patterns
  - `api/docs/WORKER_ABUSE_PREVENTION.md` - Background worker abuse prevention for free users

## Multi-Language Support

The application provides comprehensive multi-language support:

### Mobile App Internationalization
- 8 supported languages: English, German, Spanish, French, Italian, Japanese, Portuguese (Brazil), Portuguese (Portugal)
- Real-time language switching without app restart
- All UI elements, error messages, and features fully translated
- Cultural localization including currency symbols and date formats

### AI Content Generation
- 7 languages for AI-powered content: English, Spanish, French, German, Italian, Portuguese, Japanese
- Native language processing with language-specific grammar rules and verb systems
- Culturally-aware image generation prompts in target language
- Language-optimized voice selection for text-to-speech (alloy, nova, shimmer, echo, fable, onyx)
- Dynamic part-of-speech validation per language
- Automatic language detection and content adaptation
- Multi-modal AI features: definitions, images, audio, and example sentences

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
- Box 2: 1 hour
- Box 3: 6 hours  
- Box 4: 1 day
- Box 5: 3 days
- Box 6: 7 days
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

## Memories
- read README.md for more additional context on decorebator project
- read api/docs/ANALYTICS_PERFORMANCE_SCALABILITY_REPORT.md for analytics system architecture and future plans
- read mobile/docs/mobile-app-architecture.md for detailed mobile app patterns and design system
- read api/docs/ANALYTICS_TESTING_IMPLEMENTATION.md for comprehensive analytics testing patterns
- Update README.md right after introducing major features or refactorings
- Update relevant documentation after making architectural changes
- Fixed critical analytics bugs in January 2025: streak calculations, box distribution logic, response time filtering, word count inconsistencies, and removed unused database fields
- Analytics repository functions reviewed and corrected: GetWordlistCurrentStreak, GetAllWordlistsProgress, GetCurrentBoxDistribution, GetPracticeTime, GetQuizTypePerformance, GetWordlistMasteryStats
- Implemented comprehensive analytics integration tests covering all edge cases and data scenarios
- Added batch analytics endpoint `/analytics/progress-summary` reducing mobile API calls from 8 to 1

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.