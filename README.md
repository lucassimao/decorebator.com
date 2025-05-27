# Decorebator - AI-Powered Vocabulary Learning Platform

Decorebator is a comprehensive vocabulary learning application that uses AI-powered enrichment and the Leitner spaced repetition system to help users master new languages effectively.

## 🌟 Features

### Core Learning Features
- **Build Vocabulary Lists**: Create and manage multiple wordlists for any language
- **AI-Powered Enrichment**: Automatically generates definitions, images, audio pronunciations, and example sentences
- **Multiple Quiz Modes**:
  - Word ↔ Meaning matching
  - Visual word association (identify words from images)
  - Audio comprehension (guess words from pronunciation)
  - Contextual learning (complete sentences)
- **Spaced Repetition**: Uses the Leitner system to optimize learning retention
- **Progress Tracking**: Monitor your learning journey with detailed statistics

### Subscription Tiers
- **Free Plan**:
  - 1 wordlist
  - Up to 10 words per wordlist
  - Basic quiz modes
- **Premium Plans**:
  - **Monthly**: $6.99/month
  - **Annual**: $69.9/year (save $13.98)
  - Unlimited wordlists
  - Unlimited words
  - All quiz modes
  - Priority support
  - Early access to features

## 🏗️ Architecture Overview

The project consists of three main components:

### 1. API Backend (Go/Gin)
Located in `/api`, the backend follows a 3-tier layered architecture:
- **HTTP Layer** (`internal/http/`): Request handling, JWT authentication
- **Service Layer** (`internal/service/`): Business logic and orchestration
- **Repository Layer** (`internal/repository/`): Database operations

**Key Technologies**:
- Go with Gin web framework
- PostgreSQL database with pgx driver
- River queue system for background jobs
- MinIO for S3-compatible object storage
- SendGrid for email services
- OpenAI API for AI features
- Stripe for subscription payments

### 2. Mobile Application (React Native/Expo)
Located in `/mobile`, cross-platform mobile app with:
- Expo Router for navigation
- React Query for API state management
- React Hook Form with Zod validation
- React Native Paper UI components
- Secure storage for JWT tokens
- Automatic session refresh on focus
- Real-time subscription status updates

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

### Web Frontend
```bash
npm run build         # Build for production
npm run start         # Start production server
npm run lint          # Run linter
```

## 📊 Database Schema

Key tables:
- `users`: User accounts, authentication, and subscription status
- `subscriptions`: Subscription history and details
- `subscription_events`: Stripe webhook event audit trail & email notification tracking
- `wordlists`: User's vocabulary lists
- `words`: Individual words in wordlists
- `definitions`: AI-generated definitions with multimedia
- `definition_images`: Images associated with definitions
- `leitner_system_tracking`: Spaced repetition progress
- `leitner_system_history`: Learning history
- `river_job`: Background job queue

## 💳 Subscription System

### Features
- **Stripe Integration**: Secure payment processing
- **Flexible Plans**: Monthly ($12) and Annual ($100) subscriptions
- **Free Trial**: 7-day trial for new subscribers
- **Automatic Limits**: Enforced at API level
- **Webhook Processing**: Real-time subscription updates
- **Seamless Activation**: Automatic session refresh after checkout
- **Instant Access**: Premium features unlock immediately upon payment

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
4. Set redirect URLs in environment variables:
   - Success URL: `https://yourapp.com/subscription?status=success`
   - Cancel URL: `https://yourapp.com/subscription?status=cancel`

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
- **Image Generation**: Creates relevant images using DALL-E
- **Text-to-Speech**: Generates pronunciation audio
- **Definition Fetching**: Retrieves word definitions from external sources
- **Subscription Renewal Reminders**: Automatically sends email reminders 3 days before renewal

Workers can be scaled independently with configurable concurrency:
- Image Generator: Max 5 workers
- Text-to-Speech: Max 30 workers
- Definition Fetcher: Max 50 workers
- Subscription Reminder: Max 10 workers

The system also includes periodic jobs:
- **Daily Renewal Reminder Check**: Runs daily to identify subscriptions renewing in 3-4 days

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
- `POST /users` - Create new account
- `POST /login` - User login
- `GET /logout` - User logout
- `POST /password/send-reset-email` - Request password reset
- `PATCH /password/reset` - Reset password
- `POST /auth/refresh` - Refresh JWT with updated user data

### Subscription Management
- `POST /subscription/checkout-session` - Create Stripe checkout session
- `POST /subscription/checkout-redirect` - Handle Stripe checkout redirects
- `GET /subscription/status` - Get current subscription status
- `POST /subscription/cancel` - Cancel subscription
- `GET /subscription/history` - View subscription history
- `POST /webhook/stripe` - Stripe webhook handler (no auth)

### Wordlists & Words
- `GET /wordlists` - Get user's wordlists
- `POST /wordlists` - Create new wordlist (subscription check)
- `GET /wordlists/:id` - Get specific wordlist
- `PUT /wordlists/:id` - Update wordlist
- `DELETE /wordlists/:id` - Delete wordlist
- `GET /wordlists/:id/words` - Get words in wordlist
- `POST /wordlists/:id/words` - Add word (subscription check)
- `PUT /wordlists/:id/words/:wordId` - Update word
- `DELETE /wordlists/:id/words/:wordId` - Delete word

### Quiz & Learning
- `POST /wordlists/:id/quizzes` - Create new quiz
- `PATCH /wordlists/:id/quizzes` - Save quiz progress

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