# PostHog Events (Mobile)

This document lists the privacy-safe PostHog contract used by the mobile revamp. The typed allowlist lives in `utils/activationEvents.ts`; raw `posthog.capture` calls are retained only for the two legacy inventory events identified below.

## Canonical activation contract

The typed contract lives in `utils/activationEvents.ts`. New funnel instrumentation must use these names exactly and must include `eventVersion: 1`:

- `user_signed_up`
- `wordlist_created`
- `word_added`
- `quiz_session_started`
- `quiz_session_completed`
- `quiz_answered`
- `practice_cta_opened`
- `user_signed_in`
- `onboarding_started`
- `onboarding_skipped`
- `onboarding_feature_viewed`
- `onboarding_completed`
- `paywall_impression`
- `paywall_plan_selected`
- `purchase_pending`
- `purchase_succeeded`
- `purchase_failed`
- `notification_opened`
- `restore_completed`

Allowed properties are scalar values only: `source`, `entryPoint`, `appVersion`, `platform`, `userId`, `wordlistId`, `language`, `sessionId`, `quizMode`, `quizType`, `correct`, `answeredCount`, `correctCount`, `wordCount`, `durationMs`, `responseTimeMs`, `outcome`, `errorCode`, `step`, `slide`, `destination`, `store`, `productId`, `billingPeriod`, `notificationType`, and `restoreStatus`.

Do not send email addresses, names, raw words, definitions, examples, wordlist names, notification titles/bodies, receipts, transaction IDs, purchase tokens, audio URLs, or other user/content/provider evidence. Unknown notification types collapse to `unknown`. A dry-run sink is available to validate event names and properties without sending to PostHog.

Identity semantics:

- Before authentication, PostHog owns an anonymous distinct ID.
- After successful signup or sign-in, the app calls `identify` with only the positive numeric server user ID. This joins subsequent events to the authenticated user without adding email, name, or profile properties.
- The dashboard repeats the same id-only identification after profile hydration so a transient JWT decode failure can recover.
- Central auth cleanup requests `reset` through the mounted analytics bridge before routing onward, so manual sign-out, token expiry, account deletion, and storage-cleanup failures all leave the next person with a fresh anonymous identity.
- Sentry default PII collection is disabled; authenticated Sentry context contains only the opaque server ID.

Session semantics:

- `quiz_session_started` is emitted once when a learning session begins.
- `quiz_answered` is emitted once per submitted answer and carries the session ID.
- `quiz_session_completed` is emitted once when the session reaches its completion state; it must not be emitted for an individual answer.

The remaining events below are legacy inventory outside the activation funnel. Do not add a second event for a semantic action already covered by the canonical contract.

## Onboarding

- `onboarding_started`
  - When the user taps Continue on the welcome step.
- `onboarding_skipped`
  - Properties: `eventVersion` (number), `step` (string)
- `onboarding_feature_viewed`
  - Properties: `eventVersion` (number), `slide` (string)
- `onboarding_completed`
  - When the user completes the final onboarding step.
  - Properties: `eventVersion` (number), `destination` (`signup` or `signin`)

## Auth

- `signup_started`
  - Properties: `source` (string) — currently `"signup_screen"`.
- `signup_completed`
  - Retired by the activation contract; do not emit this duplicate signup-success event.
- `user_signed_up`
  - Properties: `eventVersion` (number), `source` (string)
- `user_signed_in`
  - Properties: `eventVersion` (number), `source` (string)

## Notifications

- `notification_opened`
  - Emitted for warm and cold notification responses, deduplicated by the local request identifier.
  - Properties: `eventVersion` (number), `source` (`push_notification`), `notificationType` (`daily_practice_reminder`, `due_items_reminder`, or `unknown`).

## Native store purchase funnel

These names and safe property requirements are defined now. Their production call sites belong to the dependent native-IAP paywall/purchase item so a legacy Stripe/RevenueCat action is never mislabeled as an App Store or Play Store purchase.

- `paywall_impression`: `source`, `store`
- `paywall_plan_selected`: `source`, `store`, `billingPeriod`
- `purchase_pending`: `store`, `productId`, `billingPeriod`
- `purchase_succeeded`: `store`, `productId`, `billingPeriod`
- `purchase_failed`: `store`, `productId`, `errorCode`
- `restore_completed`: `store`, `restoreStatus`

`store` is only `apple` or `google`; prices and localized product copy stay in the store SDK and are not analytics properties. Failure codes and restore statuses must come from bounded application enums, never raw provider messages.

## Dashboard

- `dashboard_viewed`
  - Fired when the dashboard screen mounts.

## Wordlists

- `wordlist_created`
  - Properties: `eventVersion` (number), `wordlistId` (number), `language` (string), `source` (string)
- `word_added`
  - Properties: `eventVersion` (number), `wordlistId` (number), `wordCount` (number), `source` (string)

## Quizzes

- `quiz_session_started`
  - Properties: `eventVersion` (number), `sessionId` (string), `wordlistId` (number), `quizMode` (string), `source` (string)
- `quiz_answered`
  - Properties: `eventVersion` (number), `sessionId` (string), `wordlistId` (number), `quizType` (string), `correct` (boolean), `responseTimeMs` (number)
- `quiz_session_completed`
  - Properties: `eventVersion` (number), `sessionId` (string), `wordlistId` (number), `answeredCount` (number), `correctCount` (number), `durationMs` (number), `outcome` (string)

## Metrics Mapping

- **Signup completion rate**: `signup_started` → `user_signed_up` funnel.
- **Time to first wordlist**: time between `user_signed_up` and `wordlist_created`.
- **D1 retention**: cohort by `user_signed_up`, return via `dashboard_viewed`.
- **First quiz completion**: first `quiz_session_completed` with `outcome=success`.
