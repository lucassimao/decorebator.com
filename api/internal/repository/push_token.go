package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PushTokenRepository struct {
	Db *pgxpool.Pool
}

type UpsertPushTokenInput struct {
	UserID    int64
	ExpoToken string
	Platform  string
	DeviceID  *string
	Timezone  string
	Locale    *string
}

type DailyReminderCandidate struct {
	UserID    int64
	ExpoToken string
	Timezone  string
	Locale    *string
	Language  *string
}

func (r *PushTokenRepository) Upsert(ctx context.Context, input UpsertPushTokenInput) error {
	_, err := r.Db.Exec(ctx, `
		INSERT INTO push_tokens (
			user_id, expo_token, platform, device_id, timezone, locale,
			is_active, last_seen_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW(), NOW())
		ON CONFLICT (expo_token)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			platform = EXCLUDED.platform,
			device_id = EXCLUDED.device_id,
			timezone = EXCLUDED.timezone,
			locale = EXCLUDED.locale,
			is_active = true,
			last_seen_at = NOW(),
			updated_at = NOW()
	`, input.UserID, input.ExpoToken, input.Platform, input.DeviceID, input.Timezone, input.Locale)
	return err
}

func (r *PushTokenRepository) Deactivate(ctx context.Context, userID int64, expoToken string) error {
	_, err := r.Db.Exec(ctx, `
		UPDATE push_tokens
		SET is_active = false,
			updated_at = NOW()
		WHERE user_id = $1 AND expo_token = $2
	`, userID, expoToken)
	return err
}

func (r *PushTokenRepository) DeactivateByTokens(ctx context.Context, expoTokens []string) error {
	if len(expoTokens) == 0 {
		return nil
	}
	tokenArray := pgtype.FlatArray[string](expoTokens)
	_, err := r.Db.Exec(ctx, `
		UPDATE push_tokens
		SET is_active = false,
			updated_at = NOW()
		WHERE expo_token = ANY($1)
	`, tokenArray)
	return err
}

func (r *PushTokenRepository) MarkNotified(ctx context.Context, expoTokens []string, notifiedAt time.Time) error {
	if len(expoTokens) == 0 {
		return nil
	}
	tokenArray := pgtype.FlatArray[string](expoTokens)
	_, err := r.Db.Exec(ctx, `
		UPDATE push_tokens
		SET last_notified_at = $2,
			updated_at = NOW()
		WHERE expo_token = ANY($1)
	`, tokenArray, notifiedAt)
	return err
}

func (r *PushTokenRepository) FindDailyReminderCandidates(ctx context.Context, now time.Time) ([]DailyReminderCandidate, error) {
	rows, err := r.Db.Query(ctx, `
		SELECT pt.user_id, pt.expo_token, pt.timezone, pt.locale, u.preferred_language
		FROM push_tokens pt
		JOIN users u ON u.id = pt.user_id
		WHERE pt.is_active = true
			AND u.notifications_enabled = true
			AND (u.last_practice_at IS NULL OR u.last_practice_at < $1 - INTERVAL '24 hours')
			AND date_part('hour', $1 AT TIME ZONE pt.timezone) = 11
			AND (
				pt.last_notified_at IS NULL OR
				(pt.last_notified_at AT TIME ZONE pt.timezone)::date < ($1 AT TIME ZONE pt.timezone)::date
			)
	`, now)
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
