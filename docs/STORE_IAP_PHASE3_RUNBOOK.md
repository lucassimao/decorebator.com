# Store IAP Phase 3 runbook

Status: automatable preparation in progress; no build, submission, console mutation, purchase, refund, or production change has been performed by this runbook.

## Safety boundary

- `store-test` is the only EAS profile permitted for Phase 3 binaries. Android submits to the Play `internal` track; the production submit profile remains separate.
- A store-test binary must use a nonproduction API configured with `STORE_IAP_CLIENT_ENVIRONMENT=sandbox`. Never test a sandbox/TestFlight purchase against the production API.
- Do not put passwords, receipts, signed transactions, purchase tokens, private keys, full transaction IDs, or Pub/Sub bearer tokens in this repository. Evidence records may use provider-console links, redacted suffixes, timestamps, and backend inbox/result IDs that contain no provider evidence.
- Store console edits, builds, submissions, refund/revoke actions, production smoke purchases, and production data changes require owner authorization at execution time.

## Read-only baseline — 2026-08-05

- EAS authentication resolves to `@lsimaocosta/decorebator` (`882f0434-5dcf-448c-92ac-47e70c9a8d84`).
- App Store Connect app ID `6749329064` and package/bundle ID `com.lsimaocosta.decorebator` match the checked-in app configuration.
- The latest successful EAS store binaries are version `1.1.1`, iOS build `47` and Android version code `42`, from commit `455310298737cefae30f603b1c83abb2a6f26e4c`; they predate native IAP and are not Phase 3 evidence.
- The latest recorded EAS submissions succeeded to App Store Connect and the Google Play production track. No current internal-track release was inferred from that history.
- The ignored Android Publisher service account can read the live subscription catalog when explicitly scoped to `androidpublisher`. It has two active auto-renewing products:
  - `p1m`: base plan `p1m`, period `P4W`, grace `P7D`, listings in `en-US` and `pt-BR`.
  - `p2m`: base plan `p2m`, period `P1Y`, grace `P14D`, listing in `en-US` only.
- The Play catalog has no active offers. Product text is incomplete for the app's German, Spanish, French, Italian, Japanese, Portuguese variants, and annual Portuguese listing; this requires an owner Play Console/API update before Gate 3 metadata sign-off.
- EAS `preview` and `development` environments do not contain `EXPO_PUBLIC_API_URL`. The mobile API resolver now fails closed on the `store-test` channel when that variable is missing, points to production/local HTTP, or is malformed; this blocks usable store-test execution until the owner provisions a sandbox API URL and backend.
- EAS environments still contain RevenueCat public keys. They are rollback-era configuration until Phase 4 and must be removed after the IAP-only release is stable.
- Apple subscription products, App Store Server Notification URLs/version, TestFlight groups, and live review metadata have not been read from App Store Connect and remain owner-console evidence.
- The screenshot generator now rejects legacy provider names, hard-coded currency prices, unsupported trial/discount claims, missing slots, and empty copy before any paid image call. All eight checked-in locale files pass, and their flashcard headline matches the implemented tap-to-flip interaction.

## Prerequisites before either store matrix starts

1. Deploy a nonproduction API with migrations through `000076`, `STORE_IAP_ENABLED=true`, `STORE_IAP_CLIENT_ENVIRONMENT=sandbox`, separate Redis/evidence keys, Apple sandbox credentials, and Google test credentials/catalog; do not reuse production evidence encryption keys or production entitlement rows.
2. Add `EXPO_PUBLIC_API_URL=https://<sandbox-api-host>` to the EAS `preview` environment and verify the resolved Expo public config contains that HTTPS host before building.
3. Confirm backend catalogs exactly map Apple monthly/annual product IDs and Google `p1m`/`p2m` to `premium` with monthly/annual periods.
4. In App Store Connect, verify both auto-renewable products are available to the TestFlight build, have complete localized name/description/review screenshot metadata, and configure separate production and sandbox HTTPS URLs as App Store Server Notifications V2.
5. In Play Console, complete missing localized subscription listings; verify `p1m` and `p2m` are active and available to the internal track; add controlled license testers and have each tester opt in.
6. In Google Cloud/Play Console, verify the RTDN topic, authenticated push subscription, exact OIDC audience/service-account binding, retry policy, dead-letter topic/subscription, and Play Console **Send Test Message** result.
7. Build with `eas build --platform all --profile store-test`. Do not use the production submit profile. Submit iOS to App Store Connect/TestFlight and Android with `eas submit --platform android --profile store-test` only after owner authorization.
8. Confirm the four controlled store-approval accounts in `docs/fixtures/store-approval-accounts.json` are the only test identities excluded from reporting; do not classify by email pattern or display name.

## Apple sandbox/TestFlight matrix

Record one row per case. Use a fresh account or documented reset when prior eligibility/state would affect the result.

| Case | Required observation | Client evidence | Backend evidence | Result |
|---|---|---|---|---|
| New monthly purchase | Native sheet → verified active entitlement → transaction finished | Screen/state timestamp; redacted transaction suffix | One sandbox entitlement/binding; applied inbox/result | Pending |
| New annual purchase | Correct annual localized terms and renewal period | Screen/state timestamp; redacted transaction suffix | One sandbox entitlement/binding | Pending |
| Restore on same account | Access restored without creating a duplicate entitlement | Restore outcome counts | Existing binding resolved to same user | Pending |
| Renewal | Period/version advances once | Store sandbox history | One entitlement update; no duplicate binding | Pending |
| Cancellation | Access remains until provider expiry | Manage-subscription state | Auto-renew off; lifecycle refresh retained access | Pending |
| Billing retry/grace | Grace is shown and access follows provider status | Paywall/settings state | Canonical `grace_period` with provider period end | Pending |
| Refund/revocation | Access is removed after signed provider state | Settings refresh | Canonical `revoked`; no legacy access bridge | Pending |
| Invalid transaction | No access and no transaction finish | Recoverable error | Rejected result; no entitlement mutation | Pending |
| Duplicate notification | Terminal duplicate is acknowledged without reapply | N/A | One semantic effect; duplicate inbox disposition | Pending |
| Out-of-order notification | Newer snapshot wins | N/A | Stale event recorded; entitlement does not regress | Pending |

Also request an Apple V2 `TEST` notification and use Apple's test-notification status/history to prove endpoint delivery independently of purchase events.

## Google Play internal-track matrix

Install from Google Play with the exact license-tester account that opted into the internal test. A test-track user who is not a license tester can incur a real charge.

| Case | Required observation | Client evidence | Backend evidence | Result |
|---|---|---|---|---|
| New monthly purchase | `p1m` localized terms → verified/acknowledged access | Play test-payment state | Sandbox entitlement; current binding acknowledged | Pending |
| New annual purchase | `p2m` localized annual terms | Play test-payment state | Sandbox entitlement; current binding acknowledged | Pending |
| Pending purchase | No access and local transaction remains unfinished | Pending UI then resumed outcome | Canonical pending/no grant | Pending |
| Restore/query | Same account recovers access without duplicate rows | Restore outcome counts | Existing purchase token binding | Pending |
| Renewal | Accelerated test renewal advances once | Play Billing Lab/history | New provider version/period | Pending |
| Cancellation | Access persists through paid period | Play Subscription Center | Auto-renew off; active until period end | Pending |
| Decline/grace/hold | State follows Play Billing Lab and current API state | Grace/hold UI | `grace_period` then `on_hold`; hold has no access | Pending |
| Refund/revocation | Refund/revoke removes access | Play order state | Voided RTDN plus authoritative refresh | Pending |
| Invalid token | No access and bounded error | Recoverable error | Rejected result; no entitlement mutation | Pending |
| Duplicate RTDN | Redelivery has one semantic effect | N/A | Transport/semantic duplicate evidence | Pending |
| Out-of-order RTDN | Older event cannot regress state | N/A | Stale result with latest entitlement preserved | Pending |
| Acknowledgement repair | Transient ack failure retries before refund boundary | Pending then active | Durable retry and current binding acknowledged | Pending |

Use Play Console **Send Test Message** to prove topic publication, then prove the authenticated push delivery and terminal backend disposition separately.

## Store-review and policy checklist

- [ ] Paywall shows the store-authored localized title, price, billing period, introductory terms, recurring terms, automatic renewal, free path/dismiss action, restore action, privacy policy, and terms.
- [ ] Settings provides store subscription management/cancellation and does not route to Stripe or web checkout.
- [ ] In-app account deletion is reachable and warns that deleting the account does not cancel store billing.
- [ ] Public `/<locale>/delete-account` resource explains in-app deletion and offers a no-app-access request route.
- [ ] App Store privacy declarations and Google Data safety answers match actual OpenAI, Sentry, PostHog, notification, profile, user-content, and store-purchase processing.
- [ ] Store descriptions/screenshots contain no hard-coded prices, unsupported free-trial claim, Stripe/RevenueCat mention, fabricated review count, or unsupported offline claim.
- [ ] App Review credentials use only the controlled registry and instructions identify Settings → subscription/paywall → restore.
- [ ] Account deletion is tested with and without an active store subscription; store cancellation remains a separate explicit action.

## Immediate pre-release evidence

1. Run `./docs/queries/run-release-entitlement-audit.sh` against production in the release window. Every unclassified count must be zero; approved test counts must reconcile to the controlled registry.
2. Record whether one refundable production purchase per store is authorized. If waived, record Lucas's explicit waiver and the residual risk that production notification credentials/plumbing remain unproven.
3. Record exact build IDs, app version/build numbers, backend release SHA, store-test matrix timestamps, notification endpoint test results, and the reviewer account key—never its credentials.
4. Only after Gate 3 is complete may Phase 4 cutover/removal begin.

## Owner actions currently required

1. Provide or authorize deployment of the separate sandbox API and add its URL to EAS `preview` as `EXPO_PUBLIC_API_URL`.
2. Authorize store-test builds/submissions and provide access to a physical iOS device plus an Android device/emulator signed into controlled tester accounts.
3. Complete/authorize the missing Play subscription localizations and confirm Apple product metadata, notification URLs, review metadata, and TestFlight tester group.
4. Choose whether to perform one refundable production smoke purchase per store or explicitly waive it with the residual risk recorded.
