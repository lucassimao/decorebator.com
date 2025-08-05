# Statistics Configuration Guide

This document explains how to manage and control the display of marketing statistics and social proof elements throughout the web application.

## Configuration File

All statistics are centrally managed in `/src/config/statsConfig.ts`:

```typescript
export const statsConfig = {
  // Global visibility toggles
  showStats: {
    userCount: false, // "10,000+ active learners"
    starRating: false, // "4.9/5" rating display
    languageCount: false, // "7 native AI languages"
  },

  // Configurable values (when stats are enabled)
  values: {
    userCount: '10,000+',
    starRating: '4.9/5',
    languageCount: '7',
    totalLanguages: '7 languages with native AI',
  },

  // Locations where stats appear
  locations: {
    heroSection: {
      showSocialProof: false, // Hide entire social proof section
    },
    ctaSection: {
      showUserCount: false, // Hide user count in CTA text
    },
    testimonials: {
      showStats: false, // Hide stats in testimonials section
    },
  },
}
```

## What Gets Hidden/Shown

### When All Stats Are Disabled (Current State):

**Hero Section:**

- ⭐ Star rating (4.9/5) - **HIDDEN**
- 👥 User count (10,000+ active learners) - **HIDDEN**
- 🌍 Language count (7 native AI languages) - **HIDDEN**
- Entire social proof section is hidden

**CTA Section:**

- User count in subtitle text - **HIDDEN**
- Text changes from "Join 10,000+ learners..." to "Master new languages..."

**Testimonials Section:**

- Success metrics section - **HIDDEN**
- All stats (Words Learned, Retention Rate, AI Languages, App Rating) - **HIDDEN**

**SEO/Metadata:**

- Meta description user count - **HIDDEN**
- Structured data rating and user count - **HIDDEN**
- Language count in feature lists - **HIDDEN**

## How to Enable Stats

To show statistics when your app launches and you have real data:

```typescript
export const statsConfig = {
  showStats: {
    userCount: true, // Show user counts
    starRating: true, // Show app ratings
    languageCount: true, // Show language counts
  },

  // Update with real values
  values: {
    userCount: '50,000+', // Your actual user count
    starRating: '4.8/5', // Real app store rating
    languageCount: '7', // Actual language count
    totalLanguages: '7 languages with native AI',
  },

  locations: {
    heroSection: {
      showSocialProof: true, // Show social proof section
    },
    ctaSection: {
      showUserCount: true, // Include user count in CTA
    },
    testimonials: {
      showStats: true, // Show testimonials stats
    },
  },
}
```

## Components Updated

The following components now use the centralized configuration:

1. **EnhancedHeroSection** - Social proof section with conditional rendering
2. **CTASection** - User count in subtitle text
3. **TestimonialsSection** - Success metrics section
4. **Layout Metadata** - SEO description with user count
5. **StructuredData** - JSON-LD schema with ratings and counts

## Benefits

- **Single source of truth** - Change stats in one place
- **Easy to disable** - Hide all marketing stats before launch
- **Flexible control** - Show/hide individual stat types
- **SEO-friendly** - Updates metadata and structured data automatically
- **Consistent branding** - Ensures same values across all locations

## Before vs After

**Before Launch (Current):**

- No user counts or ratings displayed anywhere
- Clean, minimal presentation
- No misleading statistics

**After Launch:**

- Real user counts and ratings throughout the site
- Social proof elements drive conversions
- SEO benefits from structured data
