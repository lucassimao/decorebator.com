# Decorebator - Comprehensive Technical Documentation for Claude

Last Updated: 2026-01-07

## Table of Contents

1. [Project Overview](#project-overview)
2. [Repository Structure](#repository-structure)
3. [Technology Stack](#technology-stack)
4. [API Backend (`/api`)](#api-backend-api)
5. [Mobile App (`/mobile`)](#mobile-app-mobile)
6. [Web App (`/web`)](#web-app-web)
7. [Key Features & Business Logic](#key-features--business-logic)
8. [Development Workflow](#development-workflow)
9. [Testing Strategy](#testing-strategy)
10. [Deployment & Infrastructure](#deployment--infrastructure)
11. [Documentation Index](#documentation-index)

---

## Project Overview

**Decorebator** is an AI-powered vocabulary learning platform designed to help users master languages through scientifically-proven spaced repetition, engaging quizzes, and multi-modal content. The system consists of three main components in a monorepo architecture:

- **API Backend**: Go-based REST API with job queue system
- **Mobile App**: Cross-platform Expo/React Native application (iOS & Android)
- **Web App**: Next.js marketing site (with plans for full learning platform)

### Core Mission

Provide an engaging, AI-powered vocabulary learning experience across 7 languages (English, Spanish, French, German, Italian, Portuguese, Japanese) with premium features like offline mode, real-time speaking practice, and public quiz sharing.

### Active Users & Scale

- **10,000+ active learners** (as mentioned in marketing materials)
- Multi-platform deployment (iOS, Android, Web)
- Production-ready with comprehensive monitoring (Sentry)

---

## Repository Structure

```
decorebator-v2/
├── api/                    # Go 1.25 backend
│   ├── cmd/               # Executable commands (api, workers, migrate, admin, benchmark)
│   ├── internal/          # Internal packages (http, service, repository, model, etc.)
│   ├── tests/             # Integration and unit tests
│   ├── migrations/        # Database migrations
│   ├── scripts/           # Utility scripts
│   ├── Makefile           # Build automation
│   ├── go.mod             # Go dependencies
│   └── docker-compose.yml # Local development services
│
├── mobile/                # Expo SDK 53 / React Native 0.79
│   ├── app/              # Expo Router file-based routing
│   ├── components/       # Reusable UI components
│   ├── api/              # API client layer
│   ├── hooks/            # Custom React hooks
│   ├── contexts/         # React contexts
│   ├── i18n/             # Internationalization
│   ├── theme/            # Design tokens
│   ├── assets/           # Images, fonts, animations
│   ├── android/          # Native Android config
│   ├── ios/              # Native iOS config
│   └── package.json      # NPM dependencies
│
├── web/                  # Next.js 15 App Router
│   ├── src/
│   │   ├── app/         # Next.js routes ([locale], help, privacy, terms, etc.)
│   │   ├── components/  # React components (home, layout, features, seo)
│   │   ├── lib/         # Utilities and helpers
│   │   ├── config/      # Configuration files
│   │   └── styles/      # CSS and animations
│   ├── messages/        # i18n translations (7 languages)
│   ├── public/          # Static assets
│   └── package.json     # NPM dependencies
│
├── docs/                # Comprehensive documentation
│   ├── TESTING_GUIDE.md
│   ├── SUBSCRIPTION_SYSTEM.md
│   ├── REVENUECAT_INTEGRATION.md
│   ├── QA_PRODUCTION_RELEASE_PLAN.md
│   └── [other design docs]
│
├── .github/             # CI/CD workflows
│   └── workflows/       # GitHub Actions
│
├── README.md            # Main project documentation
├── AGENTS.md            # Repository guidelines
└── CLAUDE.md            # This file
```

---

## Technology Stack

### Backend (`api/`)

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Language** | Go | 1.25 | Primary backend language |
| **Web Framework** | Gin | 1.9.1 | HTTP routing and middleware |
| **Database** | PostgreSQL | 15+ | Primary data store |
| **ORM/Driver** | pgx/v5 | 5.7.5 | PostgreSQL driver |
| **Cache** | Redis | 7-alpine | Optional caching layer |
| **Job Queue** | River | 0.23.1 | Background job processing |
| **Object Storage** | MinIO | latest | S3-compatible storage (images, audio) |
| **Migrations** | golang-migrate | 4.18.1 | Database schema versioning |
| **Testing** | testify | 1.10.0 | Testing framework |
| **Validation** | validator/v10 | 10.20.0 | Request validation |
| **JWT** | golang-jwt | 5.2.2 | Authentication tokens |

**AI & External Services:**
- **OpenAI API** (google.golang.org/genai): Definitions, image generation, TTS, realtime chat
- **Stripe** (v82.1.0): Payment processing
- **RevenueCat**: In-app subscriptions (iOS/Android)
- **Resend**: Transactional emails
- **Sentry**: Error tracking and logging

### Mobile (`mobile/`)

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Framework** | Expo | 54.0.30 | Development platform |
| **React Native** | React Native | 0.81.5 | Mobile framework |
| **React** | React | 19.1.0 | UI library |
| **Routing** | Expo Router | 6.0.21 | File-based navigation |
| **State Management** | React Query | 5.36.2 | Server state management |
| **Forms** | React Hook Form | 7.51.4 | Form handling |
| **Validation** | Zod | 3.23.8 | Schema validation |
| **Subscriptions** | RevenueCat | 9.6.12 | In-app purchases |
| **i18n** | i18next | 25.2.1 | Internationalization |
| **WebRTC** | react-native-webrtc | 124.0.7 | Real-time voice chat |
| **Charts** | react-native-chart-kit | 6.12.0 | Analytics visualization |

**Analytics & Monitoring:**
- **Sentry**: Error tracking (~7.2.0)
- **PostHog**: Product analytics (4.17.1)

### Web (`web/`)

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Framework** | Next.js | 16.1.1 | React framework |
| **React** | React | 19.2.3 | UI library |
| **TypeScript** | TypeScript | 5.x | Type safety |
| **Styling** | Tailwind CSS | 4.x | Utility-first CSS |
| **i18n** | next-intl | 4.6.1 | Internationalization |
| **Charts** | Chart.js + React wrapper | 4.5.0 | Data visualization |
| **Analytics** | Vercel Analytics | 1.5.0 | Performance monitoring |

---

## API Backend (`/api`)

### Architecture Overview

```
cmd/
├── api/            # Main HTTP server (port 8080)
├── workers/        # Background job workers (River)
├── migrate/        # Database migration CLI
├── admin/          # Admin utilities
└── benchmark/      # Performance testing

internal/
├── http/           # HTTP handlers and routes
├── service/        # Business logic layer
├── repository/     # Data access layer
├── model/          # Domain models and DTOs
├── middleware/     # HTTP middleware (auth, logging, etc.)
├── common/         # Shared utilities (logger, database, redis, sentry)
├── openai/         # OpenAI API integrations
├── mail/           # Email service (Resend)
├── app/            # Application context
└── testutils/      # Testing utilities
```

### Key Services

#### 1. **User Service** (`internal/service/user.go`)
- User registration and authentication
- JWT token generation and validation
- Password reset functionality
- Email verification

#### 2. **Wordlist Service** (`internal/service/wordlist.go`)
- CRUD operations for wordlists
- Free plan: 1 wordlist, 10 words max
- Premium: unlimited wordlists and words
- Multi-language support (7 languages)

#### 3. **Definition Service** (`internal/service/definition.go`)
- AI-powered word definitions
- Pronunciation systems (IPA, Romaji, Hiragana, Pinyin, Hangul)
- Image generation via OpenAI DALL-E
- Audio generation via OpenAI TTS
- Example sentences with audio
- Error reporting with rate limits

#### 4. **Quiz Service** (`internal/service/quiz.go`)
- 8 quiz modes (Guess Meaning, Word from Meaning, Image, Audio, etc.)
- Deterministic 7-box Leitner spaced repetition
- Fixed intervals: 0h, 6h, 24h, 72h, 168h, 336h, 720h
- Priority-based selection algorithm
- Progress tracking and analytics

#### 5. **Subscription Service** (`internal/service/subscription.go`)
- Dual-provider system (Stripe + RevenueCat)
- Provider routing logic:
  - Android → RevenueCat
  - iOS US → Stripe
  - iOS non-US → RevenueCat
  - Web → Stripe
- Webhook processing (asynchronous via River workers)
- 3-day grace period for failed payments

#### 6. **Analytics Service** (`internal/service/analytics.go`)
- Word mastery tracking
- Learning progress metrics
- Quiz performance history
- Practice time monitoring
- Box distribution visualization
- Redis-backed caching with DB fallback

#### 7. **Real-time Chat Service** (`internal/service/realtime.go`)
- OpenAI Realtime API integration
- WebRTC-based voice sessions
- Premium-only feature
- Telemetry capture

#### 8. **Public Quiz Service** (`internal/service/public_quiz.go`)
- Shareable quiz publishing
- Difficulty levels (easy/medium/hard)
- Time limits (1-15 minutes)
- Leaderboard system
- OG image generation

### Background Workers (River)

All background jobs are processed using **River**, a PostgreSQL-based job queue:

**Worker Queues:**
- `definition-fetch`: Fetch AI-generated definitions
- `image-generation`: Generate word images
- `audio-generation`: Generate pronunciation audio
- `example-audio`: Generate example sentence audio
- `public-quiz-og`: Generate OG images for public quizzes
- `subscription-reminder`: Send renewal reminder emails
- `stripe-webhook`: Process Stripe webhooks asynchronously
- `revenuecat-webhook`: Process RevenueCat webhooks asynchronously

**Key Features:**
- Exponential backoff retry strategy
- Max 5 retries per job
- 30-second job timeout
- Duplicate prevention via unique event IDs
- Comprehensive error logging

### Database Schema Highlights

**Core Tables:**
- `users`: User accounts, subscription plans, platform info
- `wordlists`: Vocabulary collections
- `definitions`: AI-generated word data (definitions, images, audio)
- `user_word_progress`: Spaced repetition tracking
- `subscriptions`: Payment provider subscriptions
- `subscription_events`: Unified webhook event log (Stripe + RevenueCat)
- `revenuecat_events`: RevenueCat-specific event audit trail
- `error_reports`: User-reported content issues
- `public_quizzes`: Published shareable quizzes
- `public_quiz_attempts`: Leaderboard entries
- `analytics_*`: Various analytics aggregation tables

**Key Enums:**
- `subscription_provider`: 'stripe' | 'revenuecat'
- `platform_type`: 'ios' | 'android' | 'web'
- `subscription_status`: 'active' | 'past_due' | 'canceled' | 'incomplete'
- `pronunciation_system`: 'ipa' | 'romaji' | 'hiragana' | 'pinyin' | 'hangul'

### API Endpoints (Key Routes)

**Authentication:**
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/refresh` - Refresh JWT token
- `POST /auth/forgot-password` - Request password reset
- `POST /auth/reset-password` - Reset password with token

**Wordlists:**
- `GET /wordlists` - List user's wordlists
- `POST /wordlists` - Create wordlist
- `GET /wordlists/:id` - Get wordlist details
- `PUT /wordlists/:id` - Update wordlist
- `DELETE /wordlists/:id` - Delete wordlist
- `POST /wordlists/:id/words` - Add word to wordlist

**Quizzes:**
- `GET /wordlists/:id/quiz` - Get next quiz question
- `POST /wordlists/:id/quiz/answer` - Submit quiz answer
- `GET /wordlists/:id/progress` - Get learning progress

**Subscriptions:**
- `POST /subscription/checkout` - Create Stripe checkout session
- `POST /subscription/revenuecat/restore` - Restore RevenueCat purchases
- `GET /subscription` - Get subscription status
- `POST /subscription/cancel` - Cancel subscription

**Webhooks:**
- `POST /webhook/stripe` - Stripe webhook handler
- `POST /webhook/revenuecat` - RevenueCat webhook handler

**Analytics:**
- `GET /analytics/dashboard` - Dashboard stats (premium)
- `GET /analytics/progress` - Learning progress charts

**Public Quizzes:**
- `GET /public-quiz/:id` - Get public quiz
- `POST /public-quiz/:id/submit` - Submit quiz attempt
- `GET /public-quiz/:id/leaderboard` - Get leaderboard

### Configuration

**Environment Variables:** (See `api/.env.example`)
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection (optional)
- `MINIO_*`: Object storage configuration
- `OPENAI_API_KEY`: OpenAI API access
- `STRIPE_*`: Stripe API keys and webhook secrets
- `REVENUECAT_*`: RevenueCat API keys
- `RESEND_API_KEY`: Email service
- `SENTRY_DSN`: Error tracking
- `JWT_SECRET`: Token signing key

### Build & Run

```bash
cd api

# Setup (first time)
make setup

# Start infrastructure
docker compose -f docker-compose.yml up -d

# Run migrations
make migrate-up

# Start API server (port 8080)
make run

# Start background workers (separate terminal)
make workers

# Run tests
make test              # Full integration suite (dockerized)
make test-unit         # Unit tests only
make test-fast         # Local integration tests

# Linting and formatting
make lint
make format
```

---

## Mobile App (`/mobile`)

### Architecture Overview

**File-Based Routing (Expo Router):**
```
app/
├── _layout.tsx                # Root layout with providers
├── index.tsx                  # Landing/auth gate
├── signin.tsx                 # Login screen
├── signup.tsx                 # Registration screen
├── dashboard/                 # Main app screens
│   ├── _layout.tsx           # Dashboard navigation
│   ├── index.tsx             # Wordlist overview
│   ├── wordlist-detail.tsx   # Word management
│   └── [id].tsx              # Dynamic routes
├── quiz.tsx                   # Quiz interface
├── flashcard.tsx              # Flashcard mode
├── analytics.tsx              # Progress tracking (premium)
├── realtime-chat.tsx          # AI speaking practice (premium)
├── settings.tsx               # App settings
├── profileSettings.tsx        # User profile
├── onboarding/                # First-time user flow
└── __tests__/                 # Component tests
```

### Key Components

**Authentication:**
- `components/auth/LoginForm.tsx` - Email/password login
- `components/auth/RegisterForm.tsx` - New user registration
- `components/auth/ForgotPasswordModal.tsx` - Password reset flow

**Dashboard:**
- `components/dashboard/WordlistCard.tsx` - Wordlist preview card
- `components/dashboard/CreateWordlistModal.tsx` - New wordlist creation
- `components/dashboard/WordItem.tsx` - Individual word display

**Quiz:**
- `components/quiz/QuizCard.tsx` - Main quiz interface
- `components/quiz/QuizResult.tsx` - Answer feedback
- `components/quiz/QuizModes.tsx` - 8 quiz type implementations

**Subscription:**
- `components/subscription/RevenueCatPaywall.tsx` - Native paywall modal
- `components/subscription/StripePlanSelection.tsx` - Stripe checkout flow
- `components/subscription/SubscriptionStatus.tsx` - Plan display

**Analytics:**
- `components/analytics/ProgressChart.tsx` - Learning progress visualization
- `components/analytics/BoxDistribution.tsx` - Spaced repetition boxes
- `components/analytics/StatsCard.tsx` - Metric displays

### API Integration (`mobile/api/`)

**API Client Structure:**
```typescript
api/
├── api.ts              # Base Axios client with auth interceptors
├── users.ts            # User management endpoints
├── wordlists.ts        # Wordlist CRUD operations
├── quiz.ts             # Quiz gameplay endpoints
├── subscription.ts     # Stripe integration
├── analytics.ts        # Progress tracking endpoints
├── publicQuiz.ts       # Public quiz endpoints
└── constants.ts        # API configuration
```

**Authentication Flow:**
- JWT tokens stored in `expo-secure-store`
- Automatic token refresh on 401 responses
- Auth context provider for app-wide state

### State Management

**React Query Configuration:**
- Query client with 5-minute stale time
- Automatic background refetching
- Optimistic updates for mutations
- Cache invalidation on user actions

**Key Hooks:**
- `useUserInfo()` - Current user data
- `useWordlists()` - Wordlist management
- `useQuiz()` - Quiz gameplay state
- `useSubscription()` - Subscription status
- `useRevenueCat()` - RevenueCat SDK integration

### Offline Functionality (Premium)

**Features:**
- Download wordlists for offline access
- Local SQLite database for word data
- Sync on reconnection
- Quiz answers NOT synced while offline

**Implementation:**
- `expo-file-system` for data persistence
- `@react-native-async-storage/async-storage` for metadata
- `@react-native-community/netinfo` for connectivity detection

### Internationalization

**Supported Languages:** 7 (en, es, fr, de, it, pt, ja)

**Implementation:**
- i18next with `react-i18next`
- Locale files in `i18n/locales/`
- Dynamic language switching
- Device locale detection

### Native Features

**iOS:**
- App Store In-App Purchases (via RevenueCat)
- Native haptic feedback
- Push notifications (Expo Notifications)
- Face ID / Touch ID support

**Android:**
- Google Play Billing (via RevenueCat)
- Native vibration feedback
- Push notifications
- Biometric authentication

**Cross-Platform:**
- Audio playback (`expo-audio`)
- Image picker (`expo-image-picker`)
- WebRTC voice calls (`react-native-webrtc`)
- Lottie animations (`lottie-react-native`)

### Build & Deployment

**Development:**
```bash
cd mobile
npm install
npm run start          # Start Expo dev server
npm run android        # Run on Android
npm run ios            # Run on iOS
```

**Testing:**
```bash
npm test               # Run Jest tests
npm run lint           # ESLint
npm run typecheck      # TypeScript validation
```

**Production Builds:**
```bash
# Version bump (managed workflow)
npm run version:bump           # Update app.json + package.json version

# EAS Build
npm run build:ios              # iOS production build
npm run build:android          # Android production build
npm run build:all              # Both platforms

# OTA Updates
npm run ota:prod               # Production OTA (EAS update)
npm run ota:preview            # Preview channel
npm run ota:dev                # Development channel
```

**App Store Submission:**
```bash
npm run submit:ios             # Submit to App Store
npm run submit:android         # Submit to Google Play
```

### Configuration

**Environment Variables:** (See `mobile/.env.example`)
- `EXPO_PUBLIC_API_URL`: Backend API endpoint
- `EXPO_PUBLIC_APP_DOMAIN`: Public domain for deep links
- `EXPO_PUBLIC_REVENUECAT_API_KEY_IOS`: RevenueCat iOS key
- `EXPO_PUBLIC_REVENUECAT_API_KEY_ANDROID`: RevenueCat Android key
- `EXPO_PUBLIC_POSTHOG_KEY`: Analytics key
- `EXPO_PUBLIC_SENTRY_DSN`: Error tracking DSN

---

## Web App (`/web`)

### Current State: Marketing Site

**Status:** Production-ready marketing landing page with plans to evolve into full learning platform.

**Implemented Features:**
- ✅ Multi-language support (7 languages)
- ✅ SEO-optimized with structured data (6 JSON-LD schema types)
- ✅ Responsive design (mobile-first)
- ✅ Interactive elements (quiz demo modal, video showcase)
- ✅ Professional animations (CSS keyframes, glassmorphism)
- ✅ Legal pages (privacy policy, terms of service)
- ✅ Password reset flow (partial implementation)
- ✅ Comprehensive favicon set and PWA manifest
- ✅ Security headers and performance optimization

**Not Yet Implemented:**
- ❌ User authentication
- ❌ Dashboard and wordlist management
- ❌ Quiz system
- ❌ Analytics
- ❌ Full subscription integration

### Architecture

**Next.js 15 App Router:**
```
src/app/
├── layout.tsx                    # Root layout (fonts, metadata)
├── page.tsx                      # Home page
├── [locale]/                     # Internationalized routes
│   ├── layout.tsx               # Locale-specific layout
│   ├── help/page.tsx            # Support page
│   ├── privacy/page.tsx         # Privacy policy
│   ├── terms/page.tsx           # Terms of service
│   └── reset-password/page.tsx  # Password reset (partial)
└── og/route.tsx                  # OG image generation
```

**Component Structure:**
```
components/
├── home/                # 13 landing page sections
│   ├── EnhancedHeroSection.tsx
│   ├── FeaturesSection.tsx
│   ├── HowItWorksSection.tsx
│   ├── PricingSection.tsx
│   ├── TestimonialsSection.tsx
│   └── [others]
├── layout/              # Layout components
│   ├── Header.tsx       # Fixed navigation
│   ├── PageLayout.tsx   # Page wrapper
│   └── BackgroundElements.tsx
├── features/            # Feature page templates
├── seo/                 # SEO components
│   ├── StructuredData.tsx
│   └── MetaTags.tsx
└── common/              # Shared utilities
```

### Design System

**Colors:**
- Primary: `#FF7B54` (orange)
- Secondary: `#FFD700` (gold)
- Background: `#FDF6E3` (cream)
- Text: `#2D3436` (dark gray)

**Typography:**
- Font: Geist Sans/Mono (Next.js built-in)
- Responsive scaling (text-4xl lg:text-5xl pattern)

**Animations:**
- Float (6s cycle)
- Pulse Glow (3s cycle)
- Gradient Shift (8s cycle)
- Word Rotation (12s cycle)
- Hover effects (transform, scale, shadow)

**Key Features:**
- Glassmorphism effects (backdrop-blur)
- 3D card transformations
- Smooth transitions (300ms, 600ms timing)
- Reduced motion support

### SEO Implementation

**Structured Data (JSON-LD):**
- Website schema
- Organization schema
- FAQ schema
- Course schema
- Educational schema
- Breadcrumb schema

**Meta Tags:**
- Complete Open Graph tags
- Twitter Card optimization
- Language-specific descriptions
- Canonical URLs
- hreflang tags for 7 languages

**Performance:**
- Next.js Image optimization (WebP/AVIF)
- Security headers (X-Frame-Options, CSP, etc.)
- Resource hints (preconnect, prefetch)
- Automated sitemap generation

**Social Sharing:**
- 1200x630 OG images for all major platforms
- Custom meta descriptions per language
- Twitter Card with summary_large_image

### Internationalization

**next-intl Configuration:**
- 7 supported locales (en, es, fr, de, it, pt, ja)
- Routing: `/[locale]/path` structure
- Automatic locale detection
- Fallback to English

**Message Organization:**
```
messages/
├── en.json          # English
├── es.json          # Spanish
├── fr.json          # French
├── de.json          # German
├── it.json          # Italian
├── pt.json          # Portuguese
└── ja.json          # Japanese
```

### Development Roadmap

**Phase 1: Foundation** (Weeks 1-2)
- Implement authentication (JWT, localStorage)
- Mirror mobile API client structure
- Set up React Query for state management

**Phase 2: Core Features** (Weeks 3-4)
- User dashboard
- Wordlist CRUD operations
- Word management interface

**Phase 3: Learning Features** (Weeks 5-6)
- Quiz system (8 modes)
- Progress tracking and analytics
- Spaced repetition scheduling

**Phase 4: Premium Features** (Weeks 7-8)
- Stripe checkout integration
- Subscription management
- Error reporting system
- Advanced analytics

### Build & Run

```bash
cd web
npm install

# Development (port 4000)
npm run dev

# Production build
npm run build

# Start production server
npm start

# Linting and formatting
npm run lint
npm run format
npm run format:check
```

### Configuration

**Environment Variables:**
- `NEXT_PUBLIC_API_BASE`: API endpoint (default: localhost:8080)
- `NEXT_PUBLIC_STRIPE_PUBLIC_KEY`: Stripe publishable key (future)
- `NEXT_PUBLIC_POSTHOG_KEY`: Analytics key (future)
- `NEXT_PUBLIC_SENTRY_DSN`: Error tracking DSN (future)

---

## Key Features & Business Logic

### Subscription Model

**Free Plan:**
- 1 wordlist maximum
- 10 words per wordlist
- Basic quiz modes
- No offline access
- No real-time chat
- Error reporting: 3/hour, 5/day

**Premium Plan:**
- Unlimited wordlists and words
- All quiz modes
- Offline mode (download wordlists)
- Real-time speaking practice
- Public quiz publishing
- Enhanced analytics
- Error reporting: 10/hour, 30/day

**Pricing:**
- Monthly: $6.99/month
- Annual: $69.90/year (17% savings)

**Payment Provider Routing:**
```
Platform: Android        → RevenueCat (Google Play)
Platform: iOS + US       → Stripe
Platform: iOS + Non-US   → RevenueCat (App Store)
Platform: Web            → Stripe
```

### Spaced Repetition System (Leitner)

**7-Box System:**
- Box 1: 0 hours (new words)
- Box 2: 6 hours
- Box 3: 24 hours (1 day)
- Box 4: 72 hours (3 days)
- Box 5: 168 hours (7 days)
- Box 6: 336 hours (14 days)
- Box 7: 720 hours (30 days)

**Selection Algorithm:**
1. Priority weight (lower boxes = higher priority)
2. Oldest review time within box
3. Definition ID (deterministic tie-breaker)

**Answer Handling:**
- ✅ Correct: Move to next box
- ❌ Incorrect: Reset to Box 1, 10-minute cooldown

### AI Content Pipeline

**Definition Generation:**
1. User adds word to wordlist
2. Word lowercased and trimmed (max 15 chars)
3. Check for existing definition (reuse if found)
4. Queue `definition-fetch` background job
5. OpenAI generates: definition, IPA pronunciation, example sentence

**Image Generation:**
1. After definition created
2. Queue `image-generation` job
3. DALL-E generates contextual image
4. Upload to MinIO object storage

**Audio Generation:**
1. Queue `audio-generation` job (word pronunciation)
2. Queue `example-audio` job (example sentence)
3. OpenAI TTS generates MP3 files
4. Upload to MinIO

### Error Reporting System

**Rate Limits:**
- Free: 3 reports/hour, 5 reports/day per user
- Premium: 10 reports/hour, 30 reports/day per user
- 1-hour cooldown per (user, word, error-type) combination

**Triggers:**
- User reports incorrect definition
- User reports inappropriate image
- User reports audio quality issues

**Actions:**
- Creates `error_report` record
- If threshold met, queues regeneration job
- Admin can manually review and regenerate

### Public Quiz System

**Features:**
- Publish wordlists as shareable quizzes
- Difficulty levels: easy, medium, hard
- Time limits: 1-15 minutes
- Leaderboard with score ranking
- OG image generation for social sharing

**Scoring:**
- Correct answers: +1 point
- Completion time: tie-breaker
- Leaderboard ranks by score DESC, then time ASC

**Access:**
- Public URL: `https://decorebator.com/quiz/[id]`
- No authentication required
- Premium-only creation

### Real-time Speaking Practice

**Implementation:**
- OpenAI Realtime API (WebRTC)
- Premium-only feature
- WebRTC native module (`react-native-webrtc`)
- Session telemetry capture
- In-call manager for audio routing

**Flow:**
1. User starts real-time chat session
2. App establishes WebRTC connection
3. OpenAI Realtime API provides conversational AI
4. Session recorded for analytics

---

## Development Workflow

### Code Style & Conventions

**Go (API):**
- `gofmt` + `goimports` for formatting
- `golangci-lint` for linting
- PascalCase for exports
- camelCase for locals
- SCREAMING_SNAKE_CASE for env vars

**TypeScript (Mobile & Web):**
- Prettier (2-space indent, single quotes)
- ESLint with Next.js/Expo configs
- PascalCase for components
- camelCase for functions/variables
- Tailwind class sorting (web)

### Git Workflow

**Commit Guidelines:**
- Concise, imperative subject line
- Describe intent in body
- Reference issues when applicable

**PR Process:**
- Run lint, format, and tests locally
- Include scope, affected modules
- Add UI screenshots or API notes
- Note environment or schema changes

**Husky Pre-commit:**
- Mobile: Prettier auto-formatting on staged files

### Branch Strategy

**Current Setup:**
- `master`: Main development branch
- Feature branches: Create for significant changes
- No explicit staging/production branches (deploy from main)

### Environment Management

**API:**
- `.env` for local development
- `.env.example` template (tracked)
- `.env.prod` for production values (NOT tracked)
- Docker Compose overrides for local services

**Mobile:**
- `.env.development` for local dev
- `.env.local` for overrides (NOT tracked)
- EAS Secrets for production values
- Platform-specific configs in `app.json`

**Web:**
- `.env.local` for local dev (NOT tracked)
- Vercel env vars for production
- `NEXT_PUBLIC_` prefix for client-side vars

---

## Testing Strategy

### API Testing

**Unit Tests:**
- Location: `api/internal/tests/unit/`
- Run: `make test-unit`
- Coverage target: ≥70%
- Framework: testify

**Integration Tests:**
- Location: `api/tests/integration/`
- Run: `make test` (dockerized)
- Run: `make test-fast` (local)
- Coverage target: ≥80%
- Uses Docker Compose for isolated environment

**Test Script:**
- `scripts/run-tests.sh` orchestrates full suite
- Options: setup, unit, integration, all, report, watch, security, clean

**Tools:**
- `gotestsum`: Structured test output
- `nancy`: Dependency vulnerability scanning
- `govulncheck`: Go vulnerability checking

**CI Integration:**
- GitHub Actions workflow: `.github/workflows/test.yml`
- Runs on PR to `master`, `main`, or `develop`
- Lints only changed files (via `--new-from-rev` flag)
- Security scanning with gosec
- Coverage reports

### Mobile Testing

**Unit Tests:**
- Location: Co-located `*.test.tsx` files
- Framework: Jest + React Native Testing Library
- Run: `npm test`

**Mocks:**
- Location: `__mocks__/`
- Mocks for network, storage, native modules

**Testing Guidelines:**
- Use Testing Library queries (getByText, getByRole)
- Mock network calls (API layer)
- Test user interactions (fireEvent)

**QA Plan:**
- Comprehensive checklist: `docs/QA_PRODUCTION_RELEASE_PLAN.md`
- 91 test cases across 14 categories
- Device matrix testing
- Platform-specific testing (iOS/Android)

### Web Testing

**Current State:**
- No automated test suite
- ESLint for code quality
- Prettier for formatting

**Future:**
- Playwright for E2E tests
- Vitest for unit tests
- Testing Library for component tests

---

## Deployment & Infrastructure

### API Backend

**Hosting:**
- Production: Likely a VPS or cloud provider (not specified in docs)
- Database: PostgreSQL 15+
- Cache: Redis 7
- Object Storage: MinIO (S3-compatible)

**Services:**
- API server (Gin, port 8080)
- Background workers (River)
- Database (PostgreSQL)
- Redis cache
- MinIO object storage

**Docker Setup:**
- `docker-compose.yml`: Local development services
- `docker-compose.test.yml`: Isolated testing environment
- `Dockerfile`: API server container

**Monitoring:**
- **Sentry**: Error tracking and logging
- **Sentry Logs**: Debug/info+ in production
- Performance monitoring
- Error grouping and alerting

### Mobile App

**Distribution:**
- **iOS**: App Store (via EAS Submit)
- **Android**: Google Play Store (via EAS Submit)

**Build System:**
- **EAS Build**: Managed builds (cloud-based)
- Build profiles: production, preview, development

**OTA Updates:**
- **Expo Updates**: JS-only updates without store submission
- Runtime version: `appVersion` policy

**Analytics:**
- Sentry: Error tracking
- PostHog: Product analytics
- Custom events via PostHog

**Push Notifications:**
- Expo Push Notifications
- Token management via API
- Receipt tracking

### Web App

**Hosting:**
- Likely **Vercel** (given Vercel Analytics integration)
- Static generation + server-side rendering
- Edge functions for API routes

**Performance:**
- Vercel Analytics
- Vercel Speed Insights
- Core Web Vitals monitoring

**SEO:**
- Automated sitemap generation
- robots.txt
- Structured data (JSON-LD)

---

## Documentation Index

### Main Docs (`/docs`)

| Document | Description |
|----------|-------------|
| `TESTING_GUIDE.md` | Comprehensive guide to running tests locally and in CI |
| `SUBSCRIPTION_SYSTEM.md` | Complete dual-provider subscription architecture |
| `REVENUECAT_INTEGRATION.md` | RevenueCat setup and integration details |
| `QA_PRODUCTION_RELEASE_PLAN.md` | 91-test production checklist for mobile app |
| `TESTING_BEST_PRACTICES.md` | Testing patterns and conventions |
| `DETERMINISTIC_LEITNER_IMPLEMENTATION.md` | Spaced repetition algorithm details |
| `DB_BENCHMARK.md` | Database performance testing results |

### Web Docs (`/web/docs`)

| Document | Description |
|----------|-------------|
| `WEB_APP_ARCHITECTURE.md` | Complete web app technical overview |
| `WEB_APP_IMPROVEMENT_PLAN.md` | Roadmap for web platform evolution |
| `ANALYTICS_SETUP_GUIDE.md` | Analytics configuration |
| `APP_STORE_CONFIG.md` | App store settings |
| `STATS_CONFIG.md` | Statistics configuration |

### Mobile Docs (`/mobile`)

| Document | Description |
|----------|-------------|
| `README.md` | Mobile app environment setup and OTA update process |

### Root Docs

| Document | Description |
|----------|-------------|
| `README.md` | Main project overview and getting started guide |
| `AGENTS.md` | Repository structure and development guidelines |
| `CLAUDE.md` | This comprehensive technical documentation |

### GitHub Docs (`/.github`)

| Document | Description |
|----------|-------------|
| `LINT_STRATEGY.md` | PR-only linting strategy with 3 approaches |

### Design Guidelines

| Document | Location | Description |
|----------|----------|-------------|
| `DESIGN_GUIDELINES.md` | `/web` | Complete web design system documentation |

---

## Quick Reference Commands

### API

```bash
# Setup
cd api && make setup

# Development
make run                    # Start API (port 8080)
make workers                # Start background workers

# Database
make migrate-up             # Apply migrations
make migrate-down           # Rollback last migration
make migrate-create name=X  # Create new migration

# Testing
make test                   # Full integration suite
make test-unit              # Unit tests only
make test-fast              # Local integration tests
make coverage-html          # Generate HTML coverage report

# Code Quality
make lint                   # Run golangci-lint
make format                 # Format code
make format-check           # Check formatting
```

### Mobile

```bash
# Setup
cd mobile && npm install

# Development
npm run start               # Start Expo dev server
npm run android             # Run on Android
npm run ios                 # Run on iOS

# Testing
npm test                    # Run Jest tests
npm run lint                # ESLint
npm run typecheck           # TypeScript check

# Building
npm run version:bump        # Update app.json + package.json version
npm run build:ios           # Build iOS production
npm run build:android       # Build Android production
npm run submit:ios          # Submit to App Store
npm run submit:android      # Submit to Play Store

# OTA Updates
npm run ota:prod            # Production OTA (EAS update)
npm run ota:preview         # Preview channel
```

### Web

```bash
# Setup
cd web && npm install

# Development
npm run dev                 # Start dev server (port 4000)

# Production
npm run build               # Build for production
npm start                   # Start production server

# Code Quality
npm run lint                # ESLint
npm run format              # Format with Prettier
npm run format:check        # Check formatting
```

---

## Key Learnings & Insights

### Architecture Strengths

1. **Monorepo Benefits**: Shared documentation, consistent tooling, unified versioning
2. **Clear Separation**: API, mobile, and web are independently deployable
3. **Comprehensive Docs**: Extensive markdown documentation for all major features
4. **Production-Ready**: Error tracking, monitoring, and analytics built-in
5. **Type Safety**: TypeScript (mobile/web) and Go (API) provide strong typing

### Technical Highlights

1. **Dual-Provider Subscriptions**: Intelligent routing between Stripe and RevenueCat based on platform/location
2. **Background Job Queue**: River provides robust, PostgreSQL-backed job processing
3. **Deterministic Spaced Repetition**: Well-implemented Leitner system with clear intervals
4. **AI Content Pipeline**: Automated word enrichment with definitions, images, and audio
5. **Offline Support**: Premium feature with local storage and sync capabilities
6. **Real-time Voice**: WebRTC integration with OpenAI Realtime API

### Potential Improvements

1. **Web Platform**: Currently limited to marketing; full learning platform planned
2. **Testing Coverage**: Mobile could benefit from more comprehensive test suite
3. **Documentation**: Some API endpoints not fully documented (use code as source of truth)
4. **Deployment Docs**: Production infrastructure details could be more explicit
5. **Monitoring**: Consider adding application performance monitoring (APM) beyond Sentry

---

## Conclusion

Decorebator is a well-architected, production-ready language learning platform with a solid foundation across all three components. The comprehensive documentation, clear separation of concerns, and thoughtful technical choices make it an excellent project for continued development and scaling.

**For Claude AI:**
- Always check relevant documentation files before making architectural changes
- Respect the established patterns (especially subscription routing logic)
- Maintain backward compatibility for mobile app (consider existing users)
- Follow the testing strategies outlined for each component
- Reference this document when exploring the codebase

**Last Updated:** 2026-01-07
**Repository:** decorebator-v2
**Status:** Active Development
