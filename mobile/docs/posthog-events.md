# PostHog Events (Mobile)

This document lists the PostHog events currently emitted by the mobile app and the canonical activation events being introduced by the revamp.

## Canonical activation contract

The typed contract lives in `utils/activationEvents.ts`. New funnel instrumentation must use these names exactly and must include `eventVersion: 1`:

- `user_signed_up`
- `wordlist_created`
- `word_added`
- `quiz_session_started`
- `quiz_session_completed`
- `quiz_answered`
- `practice_cta_opened`

Allowed properties are scalar values only: `source`, `entryPoint`, `appVersion`, `platform`, `userId`, `wordlistId`, `language`, `sessionId`, `quizMode`, `quizType`, `correct`, `answeredCount`, `correctCount`, `wordCount`, `durationMs`, `responseTimeMs`, `outcome`, and `errorCode`.

Do not send email addresses, names, raw words, definitions, examples, wordlist names, audio URLs, or other user/content payloads. A dry-run sink is available to validate event names and properties without sending to PostHog.

Session semantics:

- `quiz_session_started` is emitted once when a learning session begins.
- `quiz_answered` is emitted once per submitted answer and carries the session ID.
- `quiz_session_completed` is emitted once when the session reaches its completion state; it must not be emitted for an individual answer.

The remaining events below are legacy inventory outside the activation funnel. Do not add a second event for a semantic action already covered by the canonical contract.

## Onboarding

- `onboarding_start`
  - When the user taps Continue on the welcome step.
- `onboarding_skipped`
  - Properties: `at_step` (string)
- `onboarding_feature_viewed`
  - Properties: `slide` (string)
- `onboarding_completed`
  - When the user completes the final onboarding step.

## Auth

- `signup_started`
  - Properties: `source` (string) — currently `"signup_screen"`.
- `signup_completed`
  - Retired by the activation contract; do not emit this duplicate signup-success event.
- `user_signed_up`
  - Properties: `eventVersion` (number), `source` (string)
- `user_signed_in`
  - Properties: `source` (string)

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
