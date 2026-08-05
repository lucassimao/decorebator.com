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

The existing call sites below are legacy inventory until the activation-event audit migrates or explicitly retires them. Do not add a second event for the same semantic action during that migration.

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
  - Properties: `email` (string)
- `user_signed_up`
  - Properties: `email` (string)
- `user_signed_in`
  - Properties: `email` (string)

## Dashboard

- `dashboard_viewed`
  - Fired when the dashboard screen mounts.

## Wordlists

- `wordlist_created`
  - Properties: `wordlistId` (number), `wordlistName` (string), `language` (string)

## Quizzes

- `quiz_completed`
  - Properties: `wordlistId` (number), `quizId` (number | null), `quizType` (string | undefined), `correct` (boolean), `responseTimeMs` (number)

## Metrics Mapping

- **Signup completion rate**: `signup_started` → `signup_completed` funnel.
- **Time to first wordlist**: time between `signup_completed` and `wordlist_created`.
- **D1 retention**: cohort by `signup_completed`, return via `dashboard_viewed`.
- **First quiz completion**: first occurrence of `quiz_completed`.
