## Environment variables

Create a `.env` file in the `mobile/` directory for local development (or configure via EAS secrets) with:

```
EXPO_PUBLIC_APP_DOMAIN=https://decorebator.com
EXPO_PUBLIC_API_URL=https://api.decorebator.com
```

This is used for building shareable links in the app. In local development, if this is not set, the app will fallback to `http://localhost:3000`.

For local development, use `.env.local` or `.env.development` with a local API URL (for example `http://10.0.2.2:3000`). Do not commit `.env.local` or `.env.development`. Production OTA updates should rely on EAS environment variables.

## OTA updates

`npm run ota:prod` publishes a production OTA using `eas update`.

## Runtime policy: appVersion

The app uses:

```json
"runtimeVersion": { "policy": "appVersion" }
```

### Release flow

- **Store release = new runtime.**
  - Bump `expo.version` (e.g., `1.1.1`).
  - Build and submit new binaries.
- **OTA updates stay on the current store version.**
  - Any JS-only changes after release are safe OTAs for that version.

### Suggested flow

1. Decide on a new store version (e.g., `1.1.1`).
2. Run `npm run version:bump` (updates `package.json` and `app.json` only).
3. Build & submit the binaries:
   - `eas build --platform ios --profile production`
   - `eas build --platform android --profile production`
4. After approval, publish JS-only changes via OTA:
   - `npm run ota:prod`

You can override the runtime guard for one-off cases with:

```
ALLOW_RUNTIME_MISMATCH=1 npm run ota:prod
```

You can also customize the update message:

```
OTA_MESSAGE="Fix critical API URL" npm run ota:prod
```

## App Store Review Prompt

The app includes a non-intrusive review prompt system to encourage users to rate the app on the App Store / Play Store.

### How it works

- **Trigger**: Shows after exiting a quiz session with 10+ questions and 70%+ accuracy
- **Soft prompt first**: A custom modal asks "Enjoying Decorebator?" before triggering the native review dialog
- **Rate limiting**:
  - If user taps "Yes, I love it!" → marks as completed, never prompts again
  - If user taps "Not really" → dismisses for this app version, may prompt on next version
  - If user taps "Maybe later" → closes without saving, may prompt again in same version
- **Platform handling**: Uses `expo-store-review` which gracefully handles:
  - Dev builds / simulators (native prompt skipped)
  - OS rate limits (iOS limits to 3 prompts per year)
  - Missing native module (soft prompt still works via OTA)

### Files

- `hooks/useAppReview.ts` - Core logic and AsyncStorage persistence
- `components/common/AppReviewModal.tsx` - Soft prompt UI
- `app/quiz.tsx` - Integration point (intercepts back navigation)
- `i18n/locales/*.json` - Translations under `review.*` keys

### Native dependency

`expo-store-review` is a native module. The soft prompt modal works via OTA, but the native store review dialog requires a new native build.
