# Mobile Release Playbook

## 1. Prep Environment
- `npm install`
- `npx expo --version` (should be 53.0.23)
- `eas --version` (must be ≥16.7.0 per `eas.json`)
- `npx expo doctor`

## 2. Bump Version
- `npm run version:patch` (or `version:minor`/`version:major`)
- Confirm `package.json`, `app.json`, `ios/decorebator/Info.plist`, and `android/app/build.gradle` all reflect the new semantic version.

## 3. QA Gates
- `npm run lint`
- `npm run typecheck`
- `npm test`
- `npm run start` (smoke test on device/simulator)

## 4. Cloud Builds
- `npm run build:ios` → wraps `eas build --platform ios --profile production`
- `npm run build:android` → wraps `eas build --platform android --profile production`
- Optional one-step: `eas build --platform ios --profile production --auto-submit`

## 5. Store Submission
- After builds finish: `npm run submit:ios -- --latest`
- Android: `npm run submit:android -- --latest`
- Update release notes/screenshots in App Store Connect & Play Console as needed.

## 6. OTA Updates (optional post-approval)
- `eas update --branch production --message "fix: Improve signin screen" --environment production --clear-cache`

## 7. Utilities
- Local Android production build: `eas build --platform android --profile production --local`
- Wipe Android emulator: `$ANDROID_SDK_ROOT/emulator/emulator -avd Galaxy_S25_Ultra_6_9_inch -wipe-data`
