# Decorebator Web Application

The Decorebator web application is a Next.js 15 marketing site and future learning platform for the AI-powered vocabulary learning system. Currently serves as a comprehensive marketing presence with plans to evolve into a full-featured web learning application.

## 🌟 Current Features

### Marketing Site
- **Landing Page**: Professional marketing site with hero section, features, pricing, and testimonials
- **Feature Showcase**: 10 dedicated pages showcasing core platform features
- **Internationalization**: Full i18n support for 7 languages (English, Spanish, French, German, Italian, Portuguese, Japanese)
- **Responsive Design**: Mobile-first design with professional animations and glassmorphism effects
- **SEO Optimized**: Meta tags, structured data, and semantic HTML for search engines

### Technical Stack
- **Framework**: Next.js 15.3.2 with App Router
- **Styling**: Tailwind CSS v4 with custom animations
- **Charts**: Chart.js 4 with react-chartjs-2 for interactive analytics demos
- **Internationalization**: next-intl with automatic locale detection
- **Type Safety**: TypeScript 5 with strict configuration
- **Icons**: Heroicons v2
- **Performance**: Turbopack for fast development refresh

## 🚀 Getting Started

### Prerequisites
- Node.js 18+ 
- npm, yarn, pnpm, or bun

### Development Setup

1. **Install dependencies**:
```bash
npm install
```

2. **Set up environment variables**:
```bash
# Create .env.local file
NEXT_PUBLIC_API_BASE="http://localhost:8080"
```

3. **Run the development server**:
```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

4. **Open your browser**:
Navigate to [http://localhost:3000](http://localhost:3000)

### Available Scripts

```bash
npm run dev      # Start development server with Turbopack
npm run build    # Build for production
npm run start    # Start production server
npm run lint     # Run ESLint
```

## 🏗️ Architecture

### Project Structure
```
web/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── [locale]/          # Internationalized routes
│   │   │   ├── features/      # Feature showcase pages
│   │   │   ├── help/          # Support documentation
│   │   │   ├── privacy/       # Privacy policy
│   │   │   ├── terms/         # Terms of service
│   │   │   └── signup/        # User registration
│   │   └── globals.css        # Global styles
│   ├── components/
│   │   ├── home/              # Landing page components (13 sections)
│   │   ├── layout/            # Layout components
│   │   ├── common/            # Shared utilities
│   │   └── features/          # Feature page templates
│   ├── styles/
│   │   └── animations.css     # Animation definitions
│   └── utils/
├── messages/                  # i18n translations (7 languages)
├── docs/                      # Documentation
└── public/                    # Static assets
```

### Key Components

#### Landing Page Sections
- **EnhancedHeroSection**: Hero with animated background and rotating text
- **FeaturesSection**: Main features grid with hover effects
- **PricingSection**: Subscription plans and pricing
- **TestimonialsSection**: Social proof and user reviews
- **AnalyticsSection**: Progress tracking preview
- **HowItWorksSection**: Learning process explanation

#### Layout Components
- **Header**: Fixed navigation with scroll effects and mobile menu
- **PageLayout**: Consistent wrapper with background elements
- **BackgroundElements**: Animated floating decorations

## 🌍 Internationalization

### Supported Languages
- **English** (en) - Default
- **Spanish** (es)
- **French** (fr) 
- **German** (de)
- **Italian** (it)
- **Portuguese** (pt)
- **Japanese** (ja)

### Route Structure
- **Localized**: `/[locale]/path` (e.g., `/es/features/ai-content`)
- **Auto-detection**: Automatic locale detection and redirection
- **Fallback**: English as default for unsupported locales

### Message Organization
```json
{
  "common": { "navigation", "buttons", "actions" },
  "hero": { "titles", "descriptions", "cta" },
  "features": { "feature descriptions" },
  "featurePages": {
    "aiContent": { "page-specific content" },
    "spacedRepetition": { "page-specific content" }
  }
}
```

## 🎨 Design System

### Color Palette
- **Primary**: #FF7B54 (Orange)
- **Secondary**: #FFD700 (Gold)
- **Background**: #FDF6E3 (Cream)
- **Text**: #2D3436 (Dark Gray)

### Typography
- **Font Family**: Geist Sans/Mono
- **Responsive Scaling**: Mobile-first with lg: breakpoint scaling
- **Hierarchy**: Clear heading structure with gradient effects

### Animations
- **Float Effects**: 6-second cycle for background elements
- **Hover States**: Scale and translate effects
- **Page Transitions**: Slide-in animations for reveals
- **Performance**: GPU-accelerated with `transform` and `opacity`

## 📊 Current Status

### ✅ Implemented
- Complete marketing site with professional design
- Multi-language support with automatic detection
- Responsive design optimized for all devices
- SEO optimization with structured data
- Feature showcase pages explaining all platform capabilities
- Legal pages (Privacy Policy, Terms of Service)

### ⚠️ Placeholder Implementation
- User registration forms (UI only, no backend integration)
- Contact forms (no email sending capability)
- Password reset (partial implementation)

### ❌ Missing (Future Implementation)
- User authentication and session management
- API integration with backend services
- User dashboard and account management
- Wordlist creation and management interface
- Quiz system and learning features
- Progress tracking and analytics
- Subscription management and billing
- Real-time features and offline support

## 🔄 Future Development

### Phase 1: Authentication & API Integration
- Implement JWT-based authentication
- Connect to existing Go backend API
- Set up React Query for state management
- Add protected routes and user sessions

### Phase 2: Core Learning Features
- User dashboard with wordlist overview
- Vocabulary management interface
- Basic quiz functionality
- Progress tracking and analytics

### Phase 3: Advanced Features
- Complete quiz system with all modes
- Subscription and billing integration
- Advanced analytics and insights
- Offline capability preparation

### Phase 4: Premium Features
- Real-time collaboration features
- Advanced learning analytics
- Social features and sharing
- Progressive Web App capabilities

## 📚 Documentation

- **[Architecture Guide](./docs/WEB_APP_ARCHITECTURE.md)**: Detailed technical architecture
- **[Design Guidelines](./DESIGN_GUIDELINES.md)**: Complete design system documentation
- **[API Integration Plan](../docs/WEB_API_INTEGRATION_PLAN.md)**: Future API integration roadmap

## 🚀 Deployment

### Production Build
```bash
npm run build
npm start
```

### Environment Configuration
```bash
# Required environment variables
NEXT_PUBLIC_API_BASE=""              # API backend URL
NEXT_PUBLIC_STRIPE_PUBLIC_KEY=""     # Stripe integration
NEXT_PUBLIC_POSTHOG_KEY=""           # Analytics tracking
NEXT_PUBLIC_SENTRY_DSN=""            # Error monitoring
```

### Deployment Platforms
- **Vercel**: Optimized deployment (recommended)
- **Netlify**: Static site deployment
- **Custom**: Node.js hosting with PM2

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow the existing code style and component patterns
4. Test across multiple languages and screen sizes
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📄 License

This project is proprietary software. All rights reserved.

---

**Note**: This web application is currently a marketing site with plans to evolve into a full learning platform. See the architecture documentation for detailed technical specifications and development roadmap.
