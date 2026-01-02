# Mobile Release Playbook

## 1. Prep Environment
- `npm install`
- `npx expo --version` (should be 54.x)
- `eas --version` (must be ≥16.7.0 per `eas.json`)
- `npx expo doctor`
- CNG/Prebuild is enabled; `mobile/ios` and `mobile/android` are generated and gitignored.

## 2. Bump Version
- Update `app.json` → `expo.version` (and `package.json` version if you keep it in sync)
- With CNG/Prebuild, native folders are generated, so do not edit `ios/` or `android/` version files directly.
- `runtimeVersion` uses the `fingerprint` policy to keep OTA updates safe across native changes.

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
- Preferred: `npm run ota:prod -- --message "fix: Improve signin screen"`
- Direct: `eas update --channel production --message "fix: Improve signin screen" --environment production --clear-cache`
- Note: channels are the public routing layer; branches are the source of updates. We publish to `production` channel.

## 7. Utilities
- Local Android production build: `eas build --platform android --profile production --local`
- Local Android dev client build: `eas build --platform android --profile development --local`
- Wipe Android emulator: `$ANDROID_SDK_ROOT/emulator/emulator -avd Galaxy_S25_Ultra_6_9_inch -wipe-data`
