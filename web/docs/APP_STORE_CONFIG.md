# App Store Configuration Guide

This document explains how to manage app store download buttons and the "coming soon" modal throughout the web application.

## Configuration

All app store-related settings are centralized in `/src/config/appStoreConfig.ts`:

```typescript
export const appStoreConfig = {
  // Toggle this to false to disable the "coming soon" modal globally
  showPendingModal: true,
  
  // Set these to actual store URLs when apps are approved
  appStoreUrl: null,  // iOS App Store URL
  playStoreUrl: null, // Google Play Store URL
}
```

## How It Works

### When `showPendingModal: true` and URLs are `null`:
- All download buttons show the "Coming Soon!" modal
- Modal displays internationalized message about waiting for app store approval

### When URLs are set:
- Download buttons link directly to the respective app stores
- Modal is never shown
- "Download App" buttons in header scroll to the download section

### When `showPendingModal: false`:
- No modal is shown even if URLs are null
- Buttons become no-op (do nothing when clicked)

## Updating for App Launch

When your apps are approved, simply update the configuration:

```typescript
export const appStoreConfig = {
  showPendingModal: false, // Optional: disable modal
  
  // Add your actual app store URLs
  appStoreUrl: 'https://apps.apple.com/app/decorebator/id123456789',
  playStoreUrl: 'https://play.google.com/store/apps/details?id=com.decorebator.app',
}
```

## Components

### AppStoreButton
Used for app store badge buttons (Apple/Google Play images).
```tsx
<AppStoreButton store="apple" />
<AppStoreButton store="google" />
```

### DownloadAppButton
Used for text-based download buttons (e.g., in navigation).
```tsx
<DownloadAppButton className="button-styles">
  Download App
</DownloadAppButton>
```

### SmartDownloadButton
Intelligent download button that adapts based on app store configuration.
```tsx
<SmartDownloadButton 
  size="medium"
  onClick={() => console.log('Custom action')}
>
  Download App
</SmartDownloadButton>
```

**Smart behavior:**
- **If store URLs exist**: Scrolls to download section or opens preferred store
- **If no URLs + modal enabled**: Shows "Coming Soon!" modal
- **If no URLs + modal disabled**: Does nothing (no-op)

## Locations Updated

Download buttons appear in:
- **Header** - Navigation "Download App" button
- **Hero Section** - App store badges under main CTA
- **CTA Section** - "Ready to Transform" section
- **Help Center** - Getting Started section and subscription plans

All locations now use the centralized configuration, ensuring consistent behavior across the entire application.

## Internationalization

The modal messages are internationalized in all 7 supported languages:
- English, Spanish, French, German, Italian, Japanese, Portuguese

Messages are stored in `/messages/[locale].json` under `common.appStorePending`.