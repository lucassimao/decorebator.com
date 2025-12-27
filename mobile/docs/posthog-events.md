# PostHog Events (Mobile)

This document lists the PostHog events currently emitted by the mobile app and the properties sent with them.

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
