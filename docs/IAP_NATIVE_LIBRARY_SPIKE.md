# Native IAP library decision and SDK 54 spike

Status: selected for implementation on 2026-08-04

Decision: use [`expo-iap` 5.0.1](https://www.npmjs.com/package/expo-iap/v/5.0.1) for the App Store and Google Play purchase client.

This decision covers only the native client boundary. The API remains the authority for product allowlisting, purchase verification, entitlement state, and access. A native purchase callback must never grant premium access by itself.

## Candidate comparison

| Criterion | `expo-iap` 5.0.1 | `react-native-iap` 16.0.2 |
| --- | --- | --- |
| Expo integration | Expo Module with a config plugin | React Native module with Expo config support |
| SDK 54 / New Architecture fit | Supports Expo SDK 53+, React Native 0.79+, and the New Architecture | Supports the New Architecture but also requires the Nitro Modules runtime peer |
| Extra native runtime | No Nitro runtime peer | Requires `react-native-nitro-modules` |
| Store account binding | Apple `appAccountToken`; Google `obfuscatedAccountId` | Equivalent store parameters are available |
| Transaction completion control | Explicit `finishTransaction`; automatic Apple completion can be disabled | Explicit completion is available |
| Repository fit | Smallest native dependency surface for the existing Expo development-client workflow | Viable fallback, but adds an avoidable runtime/native layer |

`expo-iap` is selected because it is implemented as an Expo Module, directly supports the repository's Expo SDK 54 and New Architecture setup, exposes the account-binding values required by the server contract, and avoids adding Nitro Modules solely for purchases. The public API follows OpenIAP's cross-platform purchase model.

Primary references:

- [Expo IAP setup](https://www.openiap.dev/docs/setup/expo)
- [OpenIAP purchase request](https://www.openiap.dev/docs/apis/request-purchase)
- [OpenIAP purchase request types](https://www.openiap.dev/docs/types/request-purchase-props)
- [Expo SDK 54 release](https://expo.dev/changelog/sdk-54)
- [Expo development builds](https://docs.expo.dev/develop/development-builds/introduction/)

## Contract mapping

The compile-checked spike is in `mobile/spikes/expoIap.ts` and proves the API surface needed by Phase 2:

- initialize and close the store connection;
- fetch subscription products while preserving store-localized product details;
- initiate Apple purchases with the server-issued app-account token;
- initiate Google purchases with the server-issued obfuscated account identifier and an optional store offer token;
- query available purchases for restore/status refresh;
- finish a non-consumable subscription transaction only after the backend returns its authoritative entitlement envelope.

Purchase updates and store errors are event-driven. Phase 2 must register and clean up listeners at the application purchase boundary, also handle immediate `requestPurchase` promise rejection for validation/not-prepared failures, send purchase evidence to the API, refresh on foreground, map pending/cancelled/retryable states into the mobile response contract, and call `finishTransaction` only after successful backend verification. Product identifiers supplied to the module must come from the server-allowlisted catalog.

## Native compatibility evidence

Repository configuration at the time of the spike:

- Expo `54.0.32`, React Native `0.81.5`, and `newArchEnabled: true`;
- `expo-dev-client` `6.0.20`;
- EAS `development` profile has `developmentClient: true`, internal distribution, and the `development` update channel;
- native directories are generated and ignored, so the build check ran in an isolated copy rather than changing local generated projects.

Commands and results:

1. `npx expo prebuild --no-install` succeeded for Android and iOS with the `expo-iap` config plugin.
2. The generated Android project used compile/target SDK 36, minimum SDK 24, Kotlin 2.1.20, OpenIAP Google `3.0.1`, and included `com.android.vending.BILLING`.
3. `./gradlew :app:assembleDebug --no-daemon` completed all 609 tasks successfully and produced a 227 MB development-client APK. In particular, the `expo-iap` Kotlin and Java compilation tasks passed with the repository's generated SDK 54 toolchain.
4. The generated iOS Podfile and Xcode project target iOS 15.1 and include the plugin's OpenIAP CocoaPods integration.
5. `npx tsc --noEmit --pretty false` and `npx eslint spikes/expoIap.ts` pass against the installed 5.0.1 typings.

The build emitted existing dependency deprecation warnings plus D8 stack-map warnings from the plugin's disabled Amazon Appstore dependency; they did not fail compilation or alter the selected Google Play path. Recheck these warnings when upgrading `expo-iap`.

## Development and release constraints

- This module does not run in Expo Go. Adding or upgrading it requires a new development-client/store binary; an OTA update alone is insufficient.
- Linux validation proves Android native compilation and iOS project/plugin generation, not iOS compilation. Xcode compilation must run on EAS/macOS before Phase 2 is accepted.
- Product fetching and transactions require store-configured products and a build installed through an appropriate App Store sandbox/TestFlight or Google Play test track. Those real-store flows remain the Phase 3 matrix.
- Google refunds purchases that remain unacknowledged for three days. Phase 2 must monitor verification/acknowledgement latency and retry durably inside that deadline; it must not bypass backend verification to meet it. Unfinished Apple transactions can replay on later launches and must remain idempotent.
- The current RevenueCat package and code remain untouched during Phase 1. Phase 2 replaces reachable purchase behavior behind the server contract; the removal milestone later deletes RevenueCat and Stripe artifacts after parity is proven.
- Pin `expo-iap` to `5.0.1`. Review release notes and rerun prebuild plus both platform builds before changing the version.

## Reconsideration triggers

Re-evaluate the selection if a supported Expo SDK upgrade cannot compile, account-binding parameters regress, backend-verified completion is no longer controllable, App Store/Play test flows expose a blocking defect, or the package becomes materially unmaintained. `react-native-iap` remains the first alternative, with its Nitro Modules cost included in that future evaluation.
