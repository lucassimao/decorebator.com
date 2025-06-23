# Analytics Setup Guide for Decorebator Web App

This guide provides step-by-step instructions for setting up comprehensive analytics tracking on the Decorebator website.

## Google Analytics 4 Setup

### 1. Create GA4 Property
1. Go to [Google Analytics](https://analytics.google.com/)
2. Create a new GA4 property for `decorebator.com`
3. Set up data streams for the website
4. Copy the Measurement ID (format: G-XXXXXXXXXX)

### 2. Add GA4 to Next.js
Create `/src/lib/analytics.ts`:
```typescript
declare global {
  interface Window {
    gtag: (...args: any[]) => void;
  }
}

export const GA_TRACKING_ID = process.env.NEXT_PUBLIC_GA_ID || '';

// Log page views
export const pageview = (url: string) => {
  if (typeof window !== 'undefined' && window.gtag) {
    window.gtag('config', GA_TRACKING_ID, {
      page_path: url,
    });
  }
};

// Log events
export const event = ({ action, category, label, value }: {
  action: string;
  category: string;
  label?: string;
  value?: number;
}) => {
  if (typeof window !== 'undefined' && window.gtag) {
    window.gtag('event', action, {
      event_category: category,
      event_label: label,
      value: value,
    });
  }
};

// Enhanced ecommerce events
export const trackConversion = (eventName: string, parameters: any = {}) => {
  if (typeof window !== 'undefined' && window.gtag) {
    window.gtag('event', eventName, {
      currency: 'USD',
      ...parameters,
    });
  }
};
```

### 3. Add Google Analytics Script
In `/src/app/layout.tsx`:
```typescript
import Script from 'next/script';
import { GA_TRACKING_ID } from '@/lib/analytics';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <head>
        {GA_TRACKING_ID && (
          <>
            <Script
              strategy="afterInteractive"
              src={`https://www.googletagmanager.com/gtag/js?id=${GA_TRACKING_ID}`}
            />
            <Script
              id="google-analytics"
              strategy="afterInteractive"
              dangerouslySetInnerHTML={{
                __html: `
                  window.dataLayer = window.dataLayer || [];
                  function gtag(){dataLayer.push(arguments);}
                  gtag('js', new Date());
                  gtag('config', '${GA_TRACKING_ID}', {
                    page_path: window.location.pathname,
                  });
                `,
              }}
            />
          </>
        )}
      </head>
      <body>{children}</body>
    </html>
  );
}
```

## Key Events to Track

### User Engagement Events
- **Quiz Demo Started**: When user opens quiz modal
- **Video Demo Played**: When user clicks watch demo
- **App Store Click**: Clicks on iOS/Android download buttons
- **Feature Card Hover**: Engagement with feature cards
- **FAQ Expansion**: Which FAQ items are most viewed
- **Language Switch**: Track language preference changes

### Conversion Events
- **CTA Clicks**: All call-to-action button clicks
- **Newsletter Signup**: Email capture conversions
- **Download Intent**: App store button clicks
- **Contact Form**: Help/support form submissions

### Custom Dimensions
- **User Language**: Track which language version is used
- **Page Section**: Which section of the page generates most engagement
- **Device Type**: Mobile vs desktop behavior differences
- **Traffic Source**: Organic, paid, social, direct

## Google Search Console Setup

### 1. Verify Website Ownership
1. Go to [Google Search Console](https://search.google.com/search-console/)
2. Add property for `decorebator.com`
3. Verify ownership using HTML tag method
4. Add the verification meta tag to the website head

### 2. Submit Sitemap
1. Once verified, go to Sitemaps section
2. Submit `https://decorebator.com/sitemap.xml`
3. Monitor indexing status

### 3. Monitor Key Metrics
- **Search Performance**: Queries, impressions, clicks, CTR
- **Index Coverage**: Ensure all important pages are indexed
- **Core Web Vitals**: Monitor LCP, FID, CLS scores
- **Mobile Usability**: Check for mobile-specific issues

## Conversion Tracking Setup

### 1. Goal Configuration in GA4
Set up the following conversions:
- **App Download Intent**: App store button clicks
- **Demo Engagement**: Quiz or video demo completion
- **Email Signup**: Newsletter or lead magnet downloads
- **High-Value Page Views**: Pricing page, feature pages

### 2. Enhanced Ecommerce for Subscriptions
Track subscription funnel:
```typescript
// Track subscription plan selection
trackConversion('select_item', {
  item_id: 'monthly_premium',
  item_name: 'Monthly Premium Plan',
  item_category: 'subscription',
  item_variant: 'monthly',
  value: 6.99,
});

// Track checkout initiation
trackConversion('begin_checkout', {
  currency: 'USD',
  value: 6.99,
  items: [{
    item_id: 'monthly_premium',
    item_name: 'Monthly Premium Plan',
    item_category: 'subscription',
    quantity: 1,
    price: 6.99,
  }],
});
```

## Privacy Compliance

### 1. Cookie Consent
Implement cookie consent banner:
- Analytics cookies (GA4, tracking)
- Functional cookies (preferences, session)
- Marketing cookies (if using ads)

### 2. Data Retention
Configure GA4 data retention:
- Set appropriate retention period (14 months recommended)
- Enable data deletion for user requests
- Document data processing in privacy policy

## Performance Monitoring

### 1. Core Web Vitals Tracking
```typescript
// Track Web Vitals
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

function sendToAnalytics(metric: any) {
  window.gtag('event', metric.name, {
    event_category: 'Web Vitals',
    event_label: metric.id,
    value: Math.round(metric.name === 'CLS' ? metric.value * 1000 : metric.value),
    non_interaction: true,
  });
}

getCLS(sendToAnalytics);
getFID(sendToAnalytics);
getFCP(sendToAnalytics);
getLCP(sendToAnalytics);
getTTFB(sendToAnalytics);
```

### 2. Custom Performance Metrics
- Page load times by section
- Image load performance
- Video playback metrics
- Mobile vs desktop performance differences

## A/B Testing Setup

### 1. Google Optimize Integration
1. Create Google Optimize account
2. Link to GA4 property
3. Set up experiments for:
   - CTA button colors and text
   - Hero section layouts
   - Pricing presentation
   - Feature description variations

### 2. Test Ideas
- **Hero CTA**: "Try Quick Quiz" vs "Start Learning" vs "Download App"
- **Pricing Display**: Monthly vs Annual emphasis
- **Social Proof**: Testimonials vs statistics vs badges
- **Feature Order**: AI-first vs Spaced Repetition first

## Social Media Analytics

### 1. Facebook Pixel (if using ads)
```javascript
!function(f,b,e,v,n,t,s)
{if(f.fbq)return;n=f.fbq=function(){n.callMethod?
n.callMethod.apply(n,arguments):n.queue.push(arguments)};
if(!f._fbq)f._fbq=n;n.push=n;n.loaded=!0;n.version='2.0';
n.queue=[];t=b.createElement(e);t.async=!0;
t.src=v;s=b.getElementsByTagName(e)[0];
s.parentNode.insertBefore(t,s)}(window, document,'script',
'https://connect.facebook.net/en_US/fbevents.js');
fbq('init', 'YOUR_PIXEL_ID');
fbq('track', 'PageView');
```

### 2. Twitter Conversion Tracking
Set up Twitter conversion tracking for:
- App downloads
- Newsletter signups
- Demo engagements

## Heatmap and Session Recording

### 1. Hotjar Setup (Recommended)
1. Create Hotjar account
2. Install tracking code
3. Set up heatmaps for:
   - Homepage (full page)
   - Pricing section
   - Feature sections
   - FAQ section

### 2. Session Recording
- Record user sessions for UX insights
- Focus on mobile vs desktop behavior
- Identify friction points in conversion funnel

## Email Marketing Analytics

### 1. Newsletter Performance
Track email signup conversions:
- Source page (where signup occurred)
- Lead magnet effectiveness
- Email open and click rates
- Conversion from email to app download

### 2. Drip Campaign Tracking
- Welcome series performance
- Feature introduction emails
- Retention campaigns
- Re-engagement sequences

## Regular Reporting

### 1. Weekly Reports
- Traffic overview (sessions, users, pageviews)
- Conversion metrics (app downloads, signups)
- Top performing content
- Mobile vs desktop performance

### 2. Monthly Analysis
- SEO performance (rankings, organic traffic)
- Social media referrals
- Email marketing ROI
- A/B test results and insights

### 3. Quarterly Reviews
- Goal achievement analysis
- User behavior pattern changes
- Competitive analysis impact
- Technical performance trends

## Implementation Checklist

### Phase 1: Basic Analytics ✅ PARTIALLY COMPLETED
- [x] **Vercel Analytics integrated** - User behavior tracking active
- [x] **Vercel Speed Insights integrated** - Core Web Vitals monitoring active
- [x] **Search engine verification ready** - Meta tags configured for Google/Yandex/Yahoo
- [x] **Automated sitemap generation** - Multi-language sitemap with next-sitemap
- [ ] Set up GA4 property (pending)
- [ ] Install GA4 tracking code (pending)
- [ ] Configure basic events (page views, clicks)
- [ ] Set up Google Search Console verification
- [ ] Submit sitemap to search engines

### Phase 2: Advanced Tracking ✅ FOUNDATION READY
- [x] **Core Web Vitals monitoring** - Vercel Speed Insights tracking LCP, FID, CLS
- [x] **Performance monitoring foundation** - Security headers, caching strategies implemented
- [x] **Structured data for conversions** - Schema markup ready for enhanced tracking
- [ ] Implement conversion goals (GA4 setup required)
- [ ] Set up custom events for key interactions
- [ ] Configure enhanced ecommerce tracking

### Phase 3: Optimization Tools
- [ ] Install heatmap tool (Hotjar)
- [ ] Set up A/B testing platform
- [ ] Implement session recording
- [ ] Add performance monitoring

### Phase 4: Attribution & ROI
- [ ] Multi-channel funnel analysis
- [ ] Customer journey mapping
- [ ] ROI calculation setup
- [ ] Attribution modeling

## Environment Variables

Add to `.env.local`:
```
# Analytics
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX
NEXT_PUBLIC_HOTJAR_ID=XXXXXXX
NEXT_PUBLIC_FB_PIXEL_ID=XXXXXXX

# Search Engine Verification (configured in layout.tsx)
GOOGLE_SITE_VERIFICATION=your-verification-code
YANDEX_SITE_VERIFICATION=your-verification-code
YAHOO_SITE_VERIFICATION=your-verification-code

# Current Vercel Integration (already active)
# VERCEL_ANALYTICS_ID - automatically provided by Vercel
# VERCEL_SPEED_INSIGHTS_ID - automatically provided by Vercel
```

## Data Studio Dashboard

Create a comprehensive dashboard showing:
- Traffic overview and trends
- Conversion funnel analysis
- User behavior flow
- Performance metrics (Core Web Vitals)
- SEO performance
- Social media impact

This setup provides comprehensive visibility into user behavior, conversion performance, and technical metrics to optimize the Decorebator website for maximum effectiveness.