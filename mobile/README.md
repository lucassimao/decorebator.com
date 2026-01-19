## Environment variables

Create a `.env` file in the `mobile/` directory for local development (or configure via EAS secrets) with:

```
EXPO_PUBLIC_APP_DOMAIN=https://decorebator.com
EXPO_PUBLIC_API_URL=https://api.decorebator.com
```

This is used for building shareable links in the app. In local development, if this is not set, the app will fallback to `http://localhost:3000`.

For local development, use `.env.local` or `.env.development` with a local API URL (for example `http://10.0.2.2:3000`). Do not commit `.env.local` or `.env.development`. Production OTA updates should rely on EAS environment variables.

## OTA preflight checks

`npm run ota:prod` runs a single script (`scripts/ota-release.js`) that:

- Verifies `EXPO_PUBLIC_API_URL` (prefers EAS production env vars).
- Computes the local runtime fingerprint and compares it with the latest deployed
  production runtimes for iOS and Android.
- Publishes the OTA when runtimes match.
- If runtimes don't match, it prompts to publish targeted OTAs for the deployed
  runtimes instead (or aborts).

The script exists because runtimeVersion policy `"fingerprint"` can produce a
new runtime even for JS-only changes, which would prevent OTAs from reaching
existing store builds. The planned migration is to switch to `"appVersion"`
policy on the next store release to keep OTAs stable per app version.

## Planned migration to appVersion policy

Today we use `runtimeVersion: { "policy": "fingerprint" }`, which can create a
new runtime hash even for JS-only changes. That means an OTA may publish to a
runtime that **no installed binary has**, and users won't receive it.

The planned change is to switch to:

```json
"runtimeVersion": { "policy": "appVersion" }
```

### What this changes in the release flow

- **Store release = new runtime.**
  - Bump `expo.version` (e.g., `1.1.1`).
  - Build and submit new binaries.
- **OTA updates stay on the current store version.**
  - Any JS-only changes after release are safe OTAs for that version.
- **No more accidental runtime mismatches** from fingerprint changes.

### Suggested flow once we switch

1. Decide on a new store version (e.g., `1.1.1`).
2. Update `expo.version` in `app.json`.
3. Build & submit the binaries:
   - `eas build --platform ios --profile production`
   - `eas build --platform android --profile production`
4. After approval, publish JS-only changes via OTA:
   - `npm run ota:prod`

You can override the runtime guard for one-off cases with:

```
ALLOW_RUNTIME_MISMATCH=1 npm run ota:prod
```

When a runtime mismatch is detected, the script will prompt to publish updates
targeting the deployed runtimes.

You can also customize the update message:

```
OTA_MESSAGE="Fix critical API URL" PUBLISH_DEPLOYED=1 npm run ota:prod
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
