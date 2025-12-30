# Push Notifications

This document describes the current push notification implementation, the architecture, and tradeoffs to guide future work.

## Goals

- Daily reminder at 11:00 local time for users who have not practiced in the last 24 hours.
- Simple on/off toggle for users.
- Backend-driven sending to enable future notification types.
- Device-aware delivery with Expo push receipts to clean up invalid tokens.

## Architecture Overview

### Backend

**Database**
- `users.notifications_enabled` (default true) controls user opt-in.
- `users.last_practice_at` updated on quiz completion to determine inactivity.
- `push_tokens` stores device tokens, timezone, locale, and last-notified timestamp.
- `push_receipts` stores Expo receipt IDs and delivery outcomes for cleanup.

**HTTP Endpoints**
- `POST /push/register`
  - Registers or updates a device token.
  - Requires `expoPushToken`, `platform`, `timezone`, `locale`, optional `deviceId`.
- `POST /push/unregister`
  - Deactivates a device token for the current user.
- `PATCH /users`
  - Accepts `notificationsEnabled` to toggle reminders on/off.

**Services and Workers**
- `PushNotificationService`
  - Selects eligible candidates per device timezone.
  - Sends Expo push notifications in batches.
  - Marks tokens as notified for the day.
  - Inserts receipt IDs for later verification.
  - Checks receipts and deactivates invalid tokens.
- River periodic jobs
  - Daily reminder job every 15 minutes (to catch the 11:00 local window).
  - Receipt check job every 1 hour.

**Eligibility Logic**
A device is eligible if all conditions match:
- `notifications_enabled = true` on the user.
- `last_practice_at` is null or older than 24 hours.
- Current time in the device timezone is 11:xx.
- `last_notified_at` is not already today (local day).

### Mobile

**Permission Flow**
- Default: users are opted-in at the backend, but devices are only registered after permission.
- On the first quiz completion:
  - Prompt for notification permission once per device.
  - If granted, register Expo token.
- On subsequent launches:
  - If notifications are enabled and permission already granted, register silently.
- Settings toggle:
  - On: prompt for permission and register token.
  - Off: update user preference and unregister token.

**Local Storage**
- `expoPushToken` stored in AsyncStorage to unregister later.
- `pushPromptedAfterFirstQuiz` stored to avoid repeated prompting.

## Tradeoffs and Rationale

- **Opt-in default (backend)** keeps rollout simple and enables future notifications without user migration, but requires device registration for delivery.
- **Prompt after first quiz** respects platform best practices: ask in context after value is shown rather than on first launch.
- **Timezone-based scheduling** uses device-provided timezone, allowing a single global job to send at local 11:00.
- **Receipt checks hourly** reduces load while still removing invalid tokens; can be increased if delivery issues are high.
- **Expo token storage per device** supports multi-device users and isolates failures per token.

## Key Files

### Backend
- Migrations:
  - `api/cmd/migrate/migrations/000069_add_push_notifications.*.sql`
  - `api/cmd/migrate/migrations/000070_add_push_receipts.*.sql`
- Repos:
  - `api/internal/repository/push_token.go`
  - `api/internal/repository/push_receipt.go`
- Services/Workers:
  - `api/internal/service/push_notification_service.go`
  - `api/internal/service/daily_practice_reminder_worker.go`
  - `api/internal/service/push_receipt_worker.go`
  - `api/internal/service/river.go`
- HTTP:
  - `api/internal/http/push_notifications.go`
  - `api/internal/http/quiz.go` (updates `last_practice_at`)

### Mobile
- API client:
  - `mobile/api/pushNotifications.ts`
  - `mobile/api/users.ts` (adds `notificationsEnabled`)
- Utilities:
  - `mobile/utils/pushNotifications.ts`
- UI and flows:
  - `mobile/app/settings.tsx`
  - `mobile/app/quiz.tsx`
  - `mobile/hooks/useUserSession.ts`

## Future Improvements

- Add localized notification copy from a shared translation source.
- Add retry/backoff for transient Expo errors.
- Add richer notification types (streaks, milestones) using the same backend pipeline.
- Add analytics for prompt acceptance and notification open rates.

