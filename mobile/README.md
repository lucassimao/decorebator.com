## Environment variables

Create a `.env` file in the `mobile/` directory for local development (or configure via EAS secrets) with:

```
EXPO_PUBLIC_APP_DOMAIN=https://decorebator.com
EXPO_PUBLIC_API_URL=https://api.decorebator.com
```

This is used for building shareable links in the app. In local development, if this is not set, the app will fallback to `http://localhost:3000`.

For local development, use `.env.local` or `.env.development` with a local API URL (for example `http://10.0.2.2:3000`). Do not commit `.env.local` or `.env.development`. Production OTA updates should rely on EAS environment variables.

## OTA updates

`npm run ota:prod` publishes a production OTA using `eas update`. OTA publishing
is fail-closed until `NATIVE_RUNTIME_READY_VERSION` exactly matches the current
app version. Set that value only after matching iOS and Android binaries have
been built and made available on the target channel.

The generic form requires its target explicitly, for example
`npm run ota -- preview`. A message may be supplied with `--message "text"` or
the `OTA_MESSAGE` environment variable.

## Runtime policy: appVersion

The app uses:

```json
"runtimeVersion": { "policy": "appVersion" }
```

### Release flow

- **Store release = new runtime.**
  - Bump `expo.version` (e.g., `1.1.2`).
  - Build and submit new binaries.
- **OTA updates stay on the current store version.**
  - Any JS-only changes after release are safe OTAs for that version.

### Suggested flow

1. Decide on a new store version (e.g., `1.1.2`).
2. Run `npm run version:bump` (updates `package.json` and `app.json` only).
3. Build & submit the binaries:
   - `eas build --platform ios --profile production`
   - `eas build --platform android --profile production`
4. After both matching native binaries are available on the target channel,
   publish JS-only changes via OTA:
   - `NATIVE_RUNTIME_READY_VERSION=1.1.2 npm run ota:prod`

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
