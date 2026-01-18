# Decorebator - AI-Powered Vocabulary Learning Platform

Decorebator is a multi-platform vocabulary learning system that combines AI content generation, deterministic spaced repetition, and premium offline practice. The repo contains a Go API, an Expo mobile app, and a Next.js marketing/public-quiz site.

## 🌟 Features

### Learning & Content
- **Multi-language wordlists (7 languages)**: English, Spanish, French, German, Italian, Portuguese, Japanese.
- **AI enrichment pipeline**: definitions, pronunciations, example sentences, example audio, and images.
- **Pronunciation systems**: IPA (default), plus Romaji/Hiragana (Japanese). Backend supports IPA, Romaji, Hiragana, Pinyin, and Hangul.
- **Interactive flashcards** with examples, pronunciation, and error reporting.

### Quizzes & Spaced Repetition
- **8 quiz modes** (from `api/internal/model/quiz.go`):
  - Guess Meaning
  - Word from Meaning
  - Word from Image
  - Word from Audio
  - Meaning from Audio
  - Complete Sentence
  - Write Word from Definition
  - Word from Example Audio
- **Deterministic 7-box Leitner system** with fixed intervals and priority-based selection.

### Analytics & Insights
- Word mastery, learning progress, quiz performance, practice time, and box distribution history.
- Redis-backed caching when available (falls back to DB if Redis is unavailable).

### Premium Features
- **Offline mode**: download wordlists + definitions for offline practice. (Quiz answers are not synced while offline.)
- **Realtime speaking practice**: OpenAI Realtime API (WebRTC) chat sessions with telemetry capture.
- **Public quizzes**: publish wordlists as shareable quizzes with leaderboard and OG images.

### Quality & Operations
- **Error reporting** with cooldowns/rate limits that trigger regeneration jobs.
- **Background jobs** (River): definition fetch, image generation, audio generation, example-audio generation, public-quiz OG images, subscription reminders, webhook processing.
- **Payments**: Stripe + RevenueCat with webhooks and optimistic client-side subscription updates.
- **Monitoring**: Sentry error tracking + Sentry Logs (debug/info+ in production).

## 🧠 Key Business Logic

- **Free plan limits**: 1 wordlist, 10 words per wordlist (`api/internal/model/subscription.go`).
- **Premium gates**: offline mode, realtime chat sessions, and publishing public quizzes.
- **Word creation rules**:
  - Word names are lowercased + trimmed.
  - Max length is 15 Unicode characters.
  - If a definition already exists for the same word, it is reused; otherwise the AI pipeline is queued.
- **Pronunciation system validation**: only supported systems per language are accepted.
- **Leitner spaced repetition**:
  - 7 boxes with intervals: 0h, 6h, 24h, 72h, 168h, 336h, 720h.
  - Deterministic selection by priority weight, then oldest review, then definition ID.
  - Incorrect answers reset to Box 1 and temporarily skip the item for 10 minutes.
- **Error reporting limits**:
  - Free: 3/hour, 5/day. Premium: 10/hour, 30/day.
  - 1-hour cooldown per (user, word, error-type) before re-reporting.
- **Public quizzes**:
  - Difficulty (easy/medium/hard) + time limit (1–15 minutes).
  - Leaderboard ranks by score, then completion time.
- **Payment provider routing (mobile)**:
  - Android → RevenueCat.
  - iOS US users → Stripe, iOS non‑US → RevenueCat.
  - Web → Stripe.

## 🏗️ Architecture Overview

### API Backend (`/api`)
- Go 1.23 + Gin
- PostgreSQL + pgx/v5
- River job queue
- Redis caching (optional)
- MinIO (S3-compatible object storage)
- OpenAI API (definitions, images, TTS, realtime chat)
- Stripe + RevenueCat billing

### Mobile App (`/mobile`)
- Expo SDK 53, React Native 0.79
- Expo Router
- React Query + Zod + React Hook Form
- RevenueCat for in-app subscriptions
- Offline manager and cache
- Sentry + PostHog

### Web (`/web`)
- Next.js 15 App Router + next-intl
- Tailwind CSS v4
- Public quiz pages and marketing site

## 🚀 Getting Started

### Prerequisites
- Go 1.23
- Node.js 18+
- PostgreSQL 15+
- Docker & Docker Compose

### Backend Setup

```bash
cd api
make setup
cp .env.example .env
```

Start infrastructure:
```bash
docker compose -f docker-compose.yml up -d
```

Run migrations:
```bash
make migrate-up
```

Start API and workers:
```bash
make run
# in another terminal
make workers
```

### Mobile App Setup

```bash
cd mobile
npm install
cp .env.example .env
```

Common env vars (see `mobile/.env.example`):
- `EXPO_PUBLIC_API_URL`
- `EXPO_PUBLIC_TEST_USER_EMAIL`
- `EXPO_PUBLIC_TEST_USER_PASSWORD`
- `EXPO_PUBLIC_POSTHOG_KEY`
- `EXPO_PUBLIC_POSTHOG_HOST`
- `EXPO_PUBLIC_SENTRY_DSN`
- `EXPO_PUBLIC_REVENUECAT_API_KEY_IOS`
- `EXPO_PUBLIC_REVENUECAT_API_KEY_ANDROID`
- `EXPO_PUBLIC_APP_DOMAIN` (used for public quiz share links)

Run the app:
```bash
npm run start
npm run android
npm run ios
```

### Web Setup

```bash
cd web
npm install
```

Run the dev server (port 4000):
```bash
npm run dev
```

Env vars used by web:
- `NEXT_PUBLIC_API_URL` (API base URL for public quizzes)
- `STATIC_AUTHENTICATION` (optional, for sitemap public-quiz fetch)
- `SITE_URL` (sitemap base)

## 🧪 Testing

### API
```bash
cd api
make test         # dockerized integration suite
make test-unit    # unit tests
make test-fast    # local integration tests
```

### Mobile
```bash
cd mobile
npm test
npm run lint
npm run typecheck
```

### Web
```bash
cd web
npm run lint
npm run format:check
```

## 📄 License

Proprietary. All rights reserved.
