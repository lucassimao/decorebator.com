# Decorebator Web Application Architecture

## Overview

The Decorebator web application is a Next.js 15 application that serves both as a marketing landing page and the foundation for a full-featured web learning platform. Currently in **marketing phase**, with architecture designed to seamlessly evolve into a complete learning application.

## Current Architecture

### Technology Stack

- **Framework**: Next.js 15.3.2 with App Router
- **UI Framework**: Tailwind CSS v4
- **Internationalization**: next-intl with 7 language support
- **Type Safety**: TypeScript 5
- **State Management**: React hooks (ready for React Query integration)
- **Styling**: Tailwind CSS with custom animations
- **Icons**: Heroicons v2

### Project Structure

```
web/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── [locale]/          # Internationalized routes
│   │   │   ├── features/      # Feature showcase pages (10 pages)
│   │   │   ├── help/          # Support page
│   │   │   ├── privacy/       # Privacy policy
│   │   │   ├── terms/         # Terms of service
│   │   │   ├── signup/        # User registration (placeholder)
│   │   │   └── reset-password/ # Password reset
│   │   ├── globals.css        # Global styles + animations
│   │   └── layout.tsx         # Root layout with fonts
│   ├── components/
│   │   ├── common/            # Shared utilities
│   │   ├── features/          # Feature page templates
│   │   ├── home/              # Landing page sections (13 components)
│   │   ├── layout/            # Layout components
│   │   ├── policy/            # Legal page components
│   │   └── tos/               # Terms components
│   ├── styles/
│   │   └── animations.css     # Animation definitions
│   ├── utils/
│   │   └── featureMetadata.ts # SEO metadata
│   ├── middleware.ts          # i18n routing
│   └── types.ts              # TypeScript interfaces
├── messages/                  # i18n message files (7 languages)
├── public/                    # Static assets
├── next.config.ts            # Next.js configuration
├── i18n.ts                   # Internationalization config
└── package.json              # Dependencies
```

### Current Functionality

#### ✅ Marketing Site Features
- **Landing Page**: Complete marketing site with hero, features, pricing, testimonials
- **Feature Showcase**: 10 dedicated pages explaining core features
- **Internationalization**: Full support for 7 languages (en, es, fr, de, it, pt, ja)
- **Responsive Design**: Mobile-first design with professional animations
- **SEO Optimization**: Meta tags, structured data, semantic HTML
- **Legal Pages**: Privacy policy, terms of service, help documentation

#### ✅ Design System
- **Color Palette**: Orange (#FF7B54) primary with warm gradient system
- **Typography**: Geist Sans/Mono fonts with responsive scaling
- **Components**: Glassmorphism effects, card-based layouts, hover animations
- **Animations**: CSS keyframes for floating elements, gradients, reveals

#### ⚠️ Placeholder Features (Non-functional)
- **User Registration**: Forms simulate API calls with setTimeout
- **Contact Forms**: UI only, no backend integration
- **Password Reset**: Partial implementation, limited functionality

## API Integration Status

### Current State: **Minimal Integration**

**Only Implemented:**
- Password reset endpoint integration
- Basic API configuration in `.env.local`

**Missing Integrations:**
- User authentication and session management
- Wordlist and vocabulary data
- Quiz system and progress tracking
- Subscription and billing
- Analytics and user metrics

### Mobile App Reference Implementation

The mobile app (`/mobile/api/`) provides complete API integration patterns:

```typescript
// Mobile API Structure (Reference)
mobile/api/
├── api.ts              # Base API client with auth
├── users.ts            # User management & auth
├── wordlists.ts        # Vocabulary management
├── subscription.ts     # Stripe integration
├── analytics.ts        # Progress tracking
└── constants.ts        # API configuration
```

## Internationalization Architecture

### Language Support
- **Supported Locales**: en, es, fr, de, it, pt, ja
- **Routing**: `/[locale]/path` structure with automatic detection
- **Fallback**: English as default locale
- **Message Structure**: Hierarchical JSON with feature-specific namespaces

### Implementation
```typescript
// i18n Configuration
export const routing = defineRouting({
  locales: ['en', 'es', 'fr', 'de', 'it', 'pt', 'ja'],
  defaultLocale: 'en',
  pathnames: {
    '/': '/',
    '/signup': '/signup',
    '/help': '/help'
    // ... other routes
  }
});
```

### Message Organization
```json
{
  "common": { "buttons", "navigation" },
  "hero": { "titles", "descriptions" },
  "features": { "feature-specific content" },
  "featurePages": {
    "aiContent": { "page-specific translations" },
    "spacedRepetition": { "page-specific translations" }
  }
}
```

## Component Architecture

### Layout Hierarchy
1. **Root Layout** (`app/layout.tsx`): Base HTML, fonts, metadata
2. **Locale Layout** (`app/[locale]/layout.tsx`): i18n provider
3. **Page Layout** (`components/layout/PageLayout.tsx`): Header, footer, background
4. **Feature Layout** (`components/features/FeaturePageLayout.tsx`): Feature page template

### Component Categories

#### Home Page Components (13 components)
```typescript
// Landing page sections
EnhancedHeroSection      // Hero with phone mockup
FeaturesSection          // Main features grid
HowItWorksSection        // Process explanation
PricingSection          // Subscription plans
TestimonialsSection     // Social proof
AppShowcaseSection      // App store badges
AnalyticsSection        // Progress tracking preview
CTASection             // Conversion points
ContactSection         // Support information
FooterSection          // Site footer
```

#### Shared Components
```typescript
// Layout components
Header                  // Navigation with scroll effects
BackgroundElements      // Animated floating decorations
PageLayout             // Consistent page wrapper

// Common utilities
LanguageSwitcher       // Locale selection
VideoModal            // Feature demonstrations
```

### Design Patterns

#### Responsive Design
- **Mobile-First**: Tailwind breakpoints (sm:640px, md:768px, lg:1024px)
- **Flexible Grid**: CSS Grid with responsive column counts
- **Typography Scaling**: Responsive font sizes with `text-4xl lg:text-5xl`

#### Animation System
- **CSS Keyframes**: Defined in `styles/animations.css`
- **Performance**: GPU-accelerated with `transform` and `opacity`
- **Accessibility**: `prefers-reduced-motion` support
- **Timing**: Consistent durations (300ms, 600ms, 3s, 6s, 8s)

## Environment Configuration

### Current Environment Variables
```bash
# .env.local (basic)
NEXT_PUBLIC_API_BASE="http://localhost:8080"

# Required for full integration
NEXT_PUBLIC_STRIPE_PUBLIC_KEY=""
NEXT_PUBLIC_POSTHOG_KEY=""
NEXT_PUBLIC_SENTRY_DSN=""
```

### API Configuration
```typescript
// Current minimal configuration
const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080';
```

## Technical Debt & Missing Components

### Critical Missing Features

#### 1. Authentication System
```typescript
// Needed implementation
src/lib/auth/
├── context.tsx          # Auth context provider
├── hooks.ts            # useAuth, useSession hooks
├── storage.ts          # Token management (localStorage)
└── middleware.ts       # Route protection
```

#### 2. API Integration Layer
```typescript
// Required API client structure
src/lib/api/
├── client.ts           # Base API client with auth headers
├── auth.ts            # Authentication endpoints
├── users.ts           # User management
├── wordlists.ts       # Vocabulary operations
├── subscriptions.ts   # Billing integration
└── analytics.ts       # Progress tracking
```

#### 3. State Management
```typescript
// React Query setup needed
src/lib/query/
├── client.ts          # QueryClient configuration
├── keys.ts           # Query key factories
└── hooks.ts          # Custom query hooks
```

#### 4. Application Features
```typescript
// Missing functional pages
src/app/[locale]/
├── dashboard/         # User dashboard
├── wordlists/         # Vocabulary management
├── quiz/             # Learning interface
├── analytics/        # Progress tracking
└── settings/         # User preferences
```

## Development Roadmap

### Phase 1: Foundation (Weeks 1-2)
**Goal**: Transform from marketing site to functional application

#### Authentication Implementation
- Set up authentication context and hooks
- Implement localStorage-based token management
- Add protected route middleware
- Create sign-in/sign-up forms with real API integration

#### API Integration
- Mirror mobile app's API structure
- Set up React Query for state management
- Configure environment variables
- Implement error handling and loading states

### Phase 2: Core Features (Weeks 3-4)
**Goal**: Basic learning functionality

#### User Dashboard
- Create authenticated user interface
- Implement wordlist overview
- Add basic navigation and user profile

#### Wordlist Management
- CRUD operations for wordlists
- Word addition and editing interface
- Vocabulary organization tools
- Search and filtering capabilities

### Phase 3: Learning Features (Weeks 5-6)
**Goal**: Complete learning experience

#### Quiz System
- Implement quiz modes from mobile app
- Progress tracking and analytics
- Spaced repetition scheduling
- Performance metrics

#### Analytics Integration
- Learning progress visualization
- Achievement tracking
- Performance charts and insights

### Phase 4: Premium Features (Weeks 7-8)
**Goal**: Revenue and advanced features

#### Subscription Integration
- Stripe checkout integration
- Subscription status management
- Plan upgrade/downgrade flows
- Billing history and invoices

#### Advanced Features
- Error reporting system
- Offline preparation (service worker)
- Social features and sharing
- Advanced analytics and insights

## Performance Considerations

### Current Optimizations
- **Next.js 15**: Latest features and optimizations
- **Turbopack**: Fast refresh in development
- **CSS Optimization**: Tailwind CSS with minimal bundle
- **Image Optimization**: Next.js Image component ready

### Planned Optimizations
- **Code Splitting**: Dynamic imports for large features
- **Caching Strategy**: React Query with intelligent cache management
- **Service Worker**: Offline capability for PWA features
- **Bundle Analysis**: Webpack bundle analyzer integration

## Security Considerations

### Current Security
- **Next.js Security**: Built-in CSRF protection
- **Environment Variables**: Proper client/server separation
- **Type Safety**: TypeScript for runtime error prevention

### Required Security Enhancements
- **JWT Validation**: Client-side token validation
- **Route Protection**: Authenticated route middleware
- **CORS Configuration**: Proper API communication
- **XSS Prevention**: Input sanitization for user content

## Deployment Strategy

### Current Setup
- **Static Generation**: Full static export capability
- **Vercel Ready**: Optimized for Vercel deployment
- **Environment**: Development and production configurations

### Production Requirements
- **CDN Integration**: Static asset optimization
- **Monitoring**: Error tracking with Sentry
- **Analytics**: User behavior tracking
- **Performance**: Core Web Vitals optimization

## Future Architecture Considerations

### Scalability
- **Micro-frontends**: Potential feature-based splitting
- **API Gateway**: Centralized API management
- **Caching Strategy**: Redis integration for performance
- **Database**: Read replicas for analytics queries

### Advanced Features
- **Real-time Updates**: WebSocket integration
- **Progressive Web App**: Full offline capability
- **A/B Testing**: Feature flag system
- **Content Management**: Dynamic content system

## Conclusion

The Decorebator web application has a solid foundation with excellent internationalization, professional design, and modern Next.js architecture. The primary next step is implementing authentication and API integration to unlock the full learning platform potential, using the mobile app as a reference implementation.

The modular component architecture and comprehensive design system provide a strong foundation for rapid feature development once the core integration layer is established.