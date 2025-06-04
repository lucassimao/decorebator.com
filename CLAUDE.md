# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Decorebator is a vocabulary learning application with spaced repetition using the Leitner system. It consists of:
- **API Backend** (Go/Gin) - RESTful API with PostgreSQL database
- **Mobile App** (React Native/Expo) - Cross-platform mobile application
- **Web Frontend** (Next.js) - Landing page and web application

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
make test

# Database operations
make migrate-up         # Apply migrations
make migrate-down       # Rollback last migration
make create-migration   # Create new migration file
make psql              # Open PostgreSQL console

# Other commands
make build             # Build API binary
make debug-workers     # Debug workers with delve
make debug-api         # Debug API with delve
make clean            # Remove build artifacts
make help             # Show all available commands
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

# Update Expo dependencies
npm run expo:update

# EAS Updates
npm run update:local   # Update local development build
npm run update:prod    # Update production build
```

### Web Frontend (in `/web` directory)

```bash
# Start development server with Turbopack
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Lint code
npm run lint
```

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
- MinIO for object storage (images, audio)
- Stripe integration for subscription management
- Structured logging with `slog` (enhancement planned)

### Background Job Processing

Four worker queues process asynchronous tasks:
- `image_generator` - Generates images using OpenAI DALL-E (max 5 workers)
- `text_to_speech` - Converts text to audio using OpenAI TTS (max 30 workers)
- `definition_fetcher` - Fetches word definitions from external sources (max 50 workers)
- `subscription_reminder` - Sends renewal reminder emails (max 10 workers)

Workers run as a separate process and include retry logic, rate limiting, and error handling.

### Mobile App Architecture

- Expo Router for navigation
- React Query for API state management
- React Hook Form with Zod validation
- React Native Paper for UI components
- Secure storage for JWT tokens
- Automatic session refresh on focus
- Real-time subscription status updates

### Database Schema

Key tables:
- `users` - User accounts, authentication, and subscription status
- `subscriptions` - Subscription history and details
- `subscription_events` - Stripe webhook event audit trail & email tracking
- `analytics` - User analytics and engagement metrics
- `error_reports` - Application error tracking and reporting
- `wordlists` - User's vocabulary lists with language field
- `words` - Individual words in wordlists
- `definitions` - Word definitions with images and audio
- `definition_images` - Images associated with definitions
- `leitner_system_tracking` - Spaced repetition tracking
- `leitner_system_history` - Learning history
- `river_job` - Background job queue
- `error_reports` - Application error tracking

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
- Coverage reports generated with `go test -cover`
- Run single test: `go test -v -run TestName`
- Run tests with coverage: `make test` (containerized with HTML report)

### Mobile Tests
- Jest with Expo preset
- Run with `npm test` (runs in watch mode)
- Run single test: `npm test -- --testNamePattern="test name"`

## External Services

- **PostgreSQL** - Primary database
- **MinIO** - S3-compatible object storage
- **Redis** - Caching (configured but usage unclear)
- **SendGrid** - Email delivery
- **OpenAI API** - Image generation and text-to-speech
- **Stripe** - Payment processing and subscription management

## Development Workflow

1. API server requires `.env` file with database and service credentials
2. Run `docker-compose up` to start PostgreSQL, MinIO, and Redis
3. Apply database migrations before starting the API
4. Start workers separately when testing background jobs
5. Mobile app connects to API (configure API URL in `mobile/api/constants.ts`)
6. For subscription testing, configure Stripe webhook endpoint and test keys
7. Use `make watch` for API development with auto-reload
8. Run `make test` before committing to ensure tests pass

## Important Notes

- Authentication uses JWT tokens stored securely on mobile devices
- Subscription limits are enforced at the API level
- Background jobs use River queue (PostgreSQL-backed) instead of traditional queue systems
- Email templates are located in `internal/mail/` directory
- API endpoints are documented in `doc/words.http` and `doc/words.prod.http`
- Recent architecture improvements planned in `api/DEPENDENCY_INJECTION_MODERNIZATION_PLAN.md` and `api/LOGGING_IMPROVEMENT_PLAN.md`
# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.