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
npm run test

# Update Expo dependencies
npm run expo:update
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
- Manual dependency injection without frameworks
- JWT-based authentication
- River queue system for background jobs (PostgreSQL-backed)
- MinIO for object storage (images, audio)

### Background Job Processing

Three worker queues process asynchronous tasks:
- `image_generator` - Generates images using OpenAI DALL-E
- `text_to_speech` - Converts text to audio using OpenAI TTS
- `definition_fetcher` - Fetches word definitions from external sources

Workers run as a separate process and include retry logic, rate limiting, and error handling.

### Mobile App Architecture

- Expo Router for navigation
- React Query for API state management
- React Hook Form with Zod validation
- React Native Paper for UI components
- Secure storage for JWT tokens

### Database Schema

Key tables:
- `users` - User accounts and authentication
- `wordlists` - User's vocabulary lists
- `words` - Individual words in wordlists
- `definitions` - Word definitions with images and audio
- `leitner_system_tracking` - Spaced repetition tracking
- `river_job` - Background job queue

## Testing Strategy

### API Tests
- Integration tests using `httpexpect`
- Test database with Docker Compose (`docker-compose.test.yml`)
- Coverage reports generated with `go test -cover`

### Mobile Tests
- Jest with Expo preset
- Run with `npm test`

## External Services

- **PostgreSQL** - Primary database
- **MinIO** - S3-compatible object storage
- **Redis** - Caching (configured but usage unclear)
- **SendGrid** - Email delivery
- **OpenAI API** - Image generation and text-to-speech

## Development Workflow

1. API server requires `.env` file with database and service credentials
2. Run `docker-compose up` to start PostgreSQL, MinIO, and Redis
3. Apply database migrations before starting the API
4. Start workers separately when testing background jobs
5. Mobile app connects to API (configure API URL in constants)