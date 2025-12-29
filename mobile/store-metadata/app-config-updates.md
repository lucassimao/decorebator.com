# App Configuration Updates for App Store Submission

## app.json Updates Required

### Current Issues to Address

1. **Missing App Store metadata fields**
2. **Incomplete iOS configuration**
3. **Missing Android optimization**
4. **No privacy policy URLs**
5. **Missing app store descriptions**

### Recommended app.json Updates

```json
{
  "expo": {
    "name": "Decorebator - AI Vocabulary Learning",
    "slug": "decorebator",
    "version": "1.0.0",
    "orientation": "portrait",
    "icon": "./assets/images/icon.png",
    "scheme": "decorebator",
    "userInterfaceStyle": "automatic",
    "newArchEnabled": true,
    "description": "Master vocabulary with AI-powered flashcards and intelligent spaced repetition. Learn 7+ languages effectively with offline support.",
    "githubUrl": "https://github.com/decorebator/mobile",
    "assetBundlePatterns": ["**/*"],

    "ios": {
      "supportsTablet": true,
      "bundleIdentifier": "com.lsimaocosta.decorebator",
      "buildNumber": "1",
      "infoPlist": {
        "UIBackgroundModes": ["audio"],
        "NSMicrophoneUsageDescription": "This app uses the microphone for pronunciation practice and voice input features.",
        "NSPhotoLibraryUsageDescription": "This app accesses your photos to let you set your profile picture.",
        "CFBundleDisplayName": "Decorebator",
        "CFBundleShortVersionString": "1.0.0",
        "CFBundleVersion": "1",
        "LSApplicationCategoryType": "public.app-category.education"
      },
      "config": {
        "usesNonExemptEncryption": false
      },
      "associatedDomains": ["applinks:decorebator.com"],
      "privacyManifests": {
        "NSPrivacyCollectedDataTypes": [
          {
            "NSPrivacyCollectedDataType": "NSPrivacyCollectedDataTypeEmailAddress",
            "NSPrivacyCollectedDataTypeLinked": true,
            "NSPrivacyCollectedDataTypeTracking": false,
            "NSPrivacyCollectedDataTypePurposes": [
              "NSPrivacyCollectedDataTypePurposeAppFunctionality"
            ]
          },
          {
            "NSPrivacyCollectedDataType": "NSPrivacyCollectedDataTypeUserContent",
            "NSPrivacyCollectedDataTypeLinked": true,
            "NSPrivacyCollectedDataTypeTracking": false,
            "NSPrivacyCollectedDataTypePurposes": [
              "NSPrivacyCollectedDataTypePurposeAppFunctionality"
            ]
          }
        ]
      }
    },

    "android": {
      "adaptiveIcon": {
        "foregroundImage": "./assets/images/adaptive-icon.png",
        "backgroundColor": "#ffffff"
      },
      "permissions": [
        "android.permission.RECORD_AUDIO",
        "android.permission.MODIFY_AUDIO_SETTINGS",
        "android.permission.INTERNET",
        "android.permission.ACCESS_NETWORK_STATE"
      ],
      "package": "com.lsimaocosta.decorebator",
      "versionCode": 1,
      "intentFilters": [
        {
          "action": "VIEW",
          "data": [
            {
              "scheme": "https",
              "host": "decorebator.com"
            }
          ],
          "category": ["BROWSABLE", "DEFAULT"]
        }
      ],
      "googleServicesFile": "./google-services.json"
    },

    "web": {
      "bundler": "metro",
      "output": "static",
      "favicon": "./assets/images/favicon.png",
      "name": "Decorebator - AI Vocabulary Learning",
      "shortName": "Decorebator",
      "description": "Master vocabulary with AI-powered flashcards and intelligent spaced repetition",
      "themeColor": "#FF7B54",
      "backgroundColor": "#ffffff"
    },

    "plugins": [
      "expo-router",
      "expo-font",
      "expo-audio",
      "expo-secure-store",
      "expo-mail-composer",
      [
        "expo-image-picker",
        {
          "photosPermission": "The app accesses your photos to let you set your profile picture."
        }
      ],
      "expo-localization",
      [
        "expo-tracking-transparency",
        {
          "userTrackingUsageDescription": "This app would like to access your device's advertising identifier to provide personalized content and better ad experiences. You can always change this in your device settings."
        }
      ]
    ],

    "experiments": {
      "typedRoutes": true
    },

    "extra": {
      "router": {},
      "eas": {
        "projectId": "882f0434-5dcf-448c-92ac-47e70c9a8d84"
      }
    },

    "owner": "lsimaocosta",
    "runtimeVersion": {
      "policy": "appVersion"
    },
    "updates": {
      "url": "https://u.expo.dev/882f0434-5dcf-448c-92ac-47e70c9a8d84"
    },

    "locales": {
      "en": "./store-metadata/locales/en.json",
      "es": "./store-metadata/locales/es.json",
      "fr": "./store-metadata/locales/fr.json",
      "de": "./store-metadata/locales/de.json",
      "it": "./store-metadata/locales/it.json",
      "pt-BR": "./store-metadata/locales/pt-BR.json",
      "pt-PT": "./store-metadata/locales/pt-PT.json",
      "ja": "./store-metadata/locales/ja.json"
    }
  }
}
```

## EAS Build Configuration (eas.json)

```json
{
  "cli": {
    "version": ">= 5.4.0"
  },
  "build": {
    "development": {
      "developmentClient": true,
      "distribution": "internal",
      "ios": {
        "resourceClass": "m-medium"
      },
      "android": {
        "resourceClass": "medium"
      }
    },
    "preview": {
      "distribution": "internal",
      "ios": {
        "resourceClass": "m-medium",
        "simulator": true
      },
      "android": {
        "resourceClass": "medium",
        "buildType": "apk"
      }
    },
    "production": {
      "ios": {
        "resourceClass": "m-medium"
      },
      "android": {
        "resourceClass": "medium"
      }
    },
    "production-ios": {
      "extends": "production",
      "platform": "ios",
      "ios": {
        "autoIncrement": "buildNumber"
      }
    },
    "production-android": {
      "extends": "production",
      "platform": "android",
      "android": {
        "autoIncrement": "versionCode"
      }
    }
  },
  "submit": {
    "production": {},
    "production-ios": {
      "platform": "ios",
      "ios": {
        "appleId": "your-apple-id@example.com",
        "ascAppId": "TBD",
        "appleTeamId": "TBD"
      }
    },
    "production-android": {
      "platform": "android",
      "android": {
        "serviceAccountKeyPath": "./service-account-key.json",
        "track": "production"
      }
    }
  }
}
```

## Store Metadata Locales

### English (en.json)

```json
{
  "name": "Decorebator - AI Vocabulary Learning",
  "subtitle": "Master languages with AI-powered flashcards",
  "description": "Transform your vocabulary learning with Decorebator - the AI-powered trainer that adapts to your style. Combines cutting-edge AI with proven spaced repetition to help you master vocabulary in 7+ languages more effectively than ever.",
  "keywords": "vocabulary,flashcards,language learning,spaced repetition,AI,education,study,pronunciation,multilingual",
  "marketingUrl": "https://decorebator.com",
  "privacyPolicyUrl": "https://decorebator.com/privacy",
  "supportUrl": "https://decorebator.com/support"
}
```

### Spanish (es.json)

```json
{
  "name": "Decorebator - Vocabulario IA",
  "subtitle": "Domina idiomas con tarjetas inteligentes",
  "description": "Transforma tu aprendizaje de vocabulario con Decorebator - el entrenador con IA que se adapta a tu estilo. Combina IA avanzada con repetición espaciada para dominar vocabulario en más de 7 idiomas.",
  "keywords": "vocabulario,tarjetas memoria,aprendizaje idiomas,repetición espaciada,IA,educación,estudio,pronunciación",
  "marketingUrl": "https://decorebator.com/es",
  "privacyPolicyUrl": "https://decorebator.com/es/privacy",
  "supportUrl": "https://decorebator.com/es/support"
}
```

## App Store Connect Configuration

### App Information

- **Primary Language**: English (U.S.)
- **Category**: Education
- **Secondary Category**: Reference
- **Content Rights**: Does not contain third-party content
- **Age Rating**: 4+

### Pricing and Availability

- **Price**: Free (with in-app purchases)
- **Availability**: All territories
- **Educational Discount**: Yes

### In-App Purchases

```json
{
  "monthly_premium": {
    "productId": "decorebator_monthly_premium",
    "type": "auto-renewable-subscription",
    "price": "$6.99",
    "duration": "1 month",
    "familyShareable": true
  },
  "annual_premium": {
    "productId": "decorebator_annual_premium",
    "type": "auto-renewable-subscription",
    "price": "$69.90",
    "duration": "1 year",
    "familyShareable": true
  }
}
```

### Subscription Groups

- **Group Name**: Premium Learning Features
- **Reference Name**: premium_features
- **Products**: Monthly Premium, Annual Premium

## Google Play Console Configuration

### Store Listing

- **App Category**: Education
- **Content Rating**: Everyone
- **Target Audience**: 13+
- **Ads**: No (app does not contain ads)

### Content Rating Questionnaire

- **Violence**: None
- **Sexual Content**: None
- **Profanity**: None
- **Drugs**: None
- **Gambling**: None
- **User Generated Content**: Yes (vocabulary wordlists)

### Data Safety

- **Data Collection**: Yes
- **Data Sharing**: No (with third parties for advertising/marketing)
- **Encryption**: Yes (data encrypted in transit)
- **Data Types**:
  - Email address (for account creation)
  - App activity (learning progress)
  - App info and performance (analytics)

## Required Assets

### App Icons

- **iOS**: 1024x1024 PNG (App Store)
- **Android**: 512x512 PNG (Google Play)
- **Adaptive Icon**: Foreground + Background
- **Various Sizes**: Generated by build process

### Screenshots

- **iPhone**: 6.7", 6.5", 5.5" displays
- **iPad**: 12.9", 11" displays
- **Android Phone**: 1080x1920 minimum
- **Android Tablet**: 1800x2560 minimum

### Privacy Manifest (iOS)

Required for iOS apps using certain APIs:

- User analytics
- Device information
- Network requests

## Pre-Submission Checklist

### Technical Requirements

- [ ] App builds successfully for production
- [ ] All required permissions properly configured
- [ ] Privacy policy and terms of service accessible
- [ ] In-app purchases configured and tested
- [ ] Deep linking/universal links working
- [ ] Accessibility features tested
- [ ] App works offline as advertised

### Content Requirements

- [ ] All screenshots meet technical specifications
- [ ] App description accurately represents functionality
- [ ] Age rating appropriate for content
- [ ] Localized content reviewed by native speakers
- [ ] No copyrighted content without permission
- [ ] Educational claims are accurate

### Legal Requirements

- [ ] Privacy policy complies with regional laws
- [ ] Terms of service cover all app functionality
- [ ] Data collection practices clearly disclosed
- [ ] Subscription terms clearly explained
- [ ] App complies with children's privacy laws (COPPA, GDPR-K)

### Store Optimization

- [ ] Keywords researched and optimized
- [ ] Competitive analysis completed
- [ ] A/B testing plan for screenshots
- [ ] Release notes prepared
- [ ] Marketing materials ready

This comprehensive metadata package provides everything needed for successful App Store and Google Play Store submission.
