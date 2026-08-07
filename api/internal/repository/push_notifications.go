package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PushNotificationRepository struct {
	Db *pgxpool.Pool
}

type DailyReminderCandidate struct {
	UserID    int64
	ExpoToken string
	Timezone  string
	Locale    *string
	Language  *string
}

type DueItemsReminderCandidate struct {
	UserID       int64
	ExpoToken    string
	Timezone     string
	Locale       *string
	Language     *string
	DueCount     int
	WordlistID   int64
	WordlistName string
}

func (r *PushNotificationRepository) FindDailyReminderCandidates(ctx context.Context, now time.Time) ([]DailyReminderCandidate, error) {
	rows, err := r.Db.Query(ctx, `
		-- Daily reminders are a fallback: exclude users with any due items so we don't double-notify.
		WITH due_users AS (
			SELECT DISTINCT lst.user_id
			FROM leitner_system_tracking lst
			JOIN words w ON w.id = lst.word_id
			WHERE lst.next_review_at IS NOT NULL
				AND w.user_id = lst.user_id
				AND lst.next_review_at <= $1
				AND w.learned = FALSE
				AND (lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < $1)
		),
		weekly_caps AS (
			SELECT user_id, COUNT(*) AS sent_count
			FROM push_notification_events
			WHERE sent_at >= $1 - INTERVAL '7 days'
				AND notification_type IN ('daily_practice_reminder', 'due_items_reminder')
			GROUP BY user_id
		),
		local AS (
			SELECT pt.user_id,
				pt.expo_token,
				pt.timezone,
				pt.locale,
				u.preferred_language,
				($1 AT TIME ZONE pt.timezone) AS local_now,
				(pt.last_notified_at AT TIME ZONE pt.timezone)::date AS last_notified_date,
				COALESCE(weekly_caps.sent_count, 0) AS sent_count
			FROM push_tokens pt
			JOIN users u ON u.id = pt.user_id
			LEFT JOIN due_users du ON du.user_id = pt.user_id
			LEFT JOIN weekly_caps ON weekly_caps.user_id = pt.user_id
			WHERE pt.is_active = true
				AND u.notifications_enabled = true
				AND (u.last_practice_at IS NULL OR u.last_practice_at < ($1 - INTERVAL '24 hours'))
				AND du.user_id IS NULL
		)
		SELECT user_id, expo_token, timezone, locale, preferred_language
		FROM local
		WHERE date_part('hour', local_now) = 11
			AND (last_notified_date IS NULL OR last_notified_date < local_now::date)
			AND sent_count < 2
	`, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []DailyReminderCandidate
	for rows.Next() {
		var candidate DailyReminderCandidate
		if err := rows.Scan(&candidate.UserID, &candidate.ExpoToken, &candidate.Timezone, &candidate.Locale, &candidate.Language); err != nil {
			return nil, err
		}
		results = append(results, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *PushNotificationRepository) FindDueItemsReminderCandidates(ctx context.Context, now time.Time) ([]DueItemsReminderCandidate, error) {
	rows, err := r.Db.Query(ctx, `
		-- Compute due counts per user+wordlist and pick the most-due wordlist per user.
		WITH due_items AS (
			SELECT lst.user_id,
				w.wordlist_id,
				wl.name AS wordlist_name,
				COUNT(*) AS due_count
			FROM leitner_system_tracking lst
			JOIN words w ON w.id = lst.word_id
			JOIN wordlists wl ON wl.id = w.wordlist_id AND wl.user_id = w.user_id
			WHERE lst.next_review_at IS NOT NULL
				AND w.user_id = lst.user_id
				AND lst.next_review_at <= $1
				AND w.learned = FALSE
				AND (lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < $1)
			GROUP BY lst.user_id, w.wordlist_id, wl.name
		),
		selected_due AS (
			SELECT DISTINCT ON (user_id)
				user_id,
				wordlist_id,
				wordlist_name,
				due_count
			FROM due_items
			ORDER BY user_id, due_count DESC, wordlist_id ASC
		),
		weekly_caps AS (
			SELECT user_id, COUNT(*) AS sent_count
			FROM push_notification_events
			WHERE sent_at >= $1 - INTERVAL '7 days'
				AND notification_type IN ('daily_practice_reminder', 'due_items_reminder')
			GROUP BY user_id
		),
		local AS (
			SELECT pt.user_id,
				pt.expo_token,
				pt.timezone,
				pt.locale,
				u.preferred_language,
				($1 AT TIME ZONE pt.timezone) AS local_now,
				(pt.last_notified_at AT TIME ZONE pt.timezone)::date AS last_notified_date,
				(u.last_practice_at AT TIME ZONE pt.timezone)::date AS last_practice_date,
				selected_due.due_count,
				selected_due.wordlist_id,
				selected_due.wordlist_name,
				COALESCE(weekly_caps.sent_count, 0) AS sent_count
			FROM push_tokens pt
			JOIN users u ON u.id = pt.user_id
			JOIN selected_due ON selected_due.user_id = pt.user_id
			LEFT JOIN weekly_caps ON weekly_caps.user_id = pt.user_id
			WHERE pt.is_active = true
				AND u.notifications_enabled = true
		)
		SELECT user_id, expo_token, timezone, locale, preferred_language, due_count, wordlist_id, wordlist_name
		FROM local
		WHERE date_part('hour', local_now) BETWEEN 10 AND 19
			AND (last_practice_date IS NULL OR last_practice_date < local_now::date)
			AND (last_notified_date IS NULL OR last_notified_date < local_now::date)
			AND due_count > 0
			AND sent_count < 2
	`, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []DueItemsReminderCandidate
	for rows.Next() {
		var candidate DueItemsReminderCandidate
		if err := rows.Scan(&candidate.UserID, &candidate.ExpoToken, &candidate.Timezone, &candidate.Locale, &candidate.Language, &candidate.DueCount, &candidate.WordlistID, &candidate.WordlistName); err != nil {
			return nil, err
		}
		results = append(results, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
