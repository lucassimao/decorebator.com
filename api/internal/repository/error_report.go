package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ErrorReportRepository struct {
	db *pgxpool.Pool
}

func NewErrorReportRepository(db *pgxpool.Pool) *ErrorReportRepository {
	return &ErrorReportRepository{
		db: db,
	}
}

// CountUserReportsInWindow counts error reports for a user within a time window
func (r *ErrorReportRepository) CountUserReportsInWindow(ctx context.Context, userID int64, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM error_reports 
		WHERE user_id = $1 AND reported_at > $2
	`, userID, since).Scan(&count)
	return count, err
}

// GetOldestReportTimeInWindow gets the oldest report time for a user within a window
func (r *ErrorReportRepository) GetOldestReportTimeInWindow(ctx context.Context, userID int64, since time.Time) (time.Time, error) {
	var oldestTime time.Time
	err := r.db.QueryRow(ctx, `
		SELECT MIN(reported_at) 
		FROM error_reports 
		WHERE user_id = $1 AND reported_at > $2
	`, userID, since).Scan(&oldestTime)
	return oldestTime, err
}

// CountUserReportsInWindows counts reports in both hourly and daily windows
func (r *ErrorReportRepository) CountUserReportsInWindows(ctx context.Context, userID int64, hourAgo, dayAgo time.Time) (hourlyCount, dailyCount int, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT 
			COUNT(CASE WHEN reported_at > $2 THEN 1 END) as hourly_count,
			COUNT(CASE WHEN reported_at > $3 THEN 1 END) as daily_count
		FROM error_reports 
		WHERE user_id = $1 AND reported_at > $3
	`, userID, hourAgo, dayAgo).Scan(&hourlyCount, &dailyCount)
	return
}

// UpdateRateLimitTracking updates the rate limit tracking table
func (r *ErrorReportRepository) UpdateRateLimitTracking(ctx context.Context, userID int64, hourlyCount, dailyCount int, now time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO error_report_limits (user_id, hourly_count, daily_count, last_hour_reset, last_day_reset)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			hourly_count = EXCLUDED.hourly_count,
			daily_count = EXCLUDED.daily_count,
			updated_at = NOW()
	`, userID, hourlyCount, dailyCount, now, now)
	return err
}

// GetUserActiveCooldowns gets all active cooldowns for a user
func (r *ErrorReportRepository) GetUserActiveCooldowns(ctx context.Context, userID int64) ([]ErrorReportCooldown, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			COALESCE(word_id, 0) as word_id,
			COALESCE(definition_id, 0) as definition_id,
			error_type,
			cooldown_until
		FROM error_report_cooldowns
		WHERE user_id = $1 AND cooldown_until > NOW()
		ORDER BY cooldown_until DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cooldowns []ErrorReportCooldown
	for rows.Next() {
		var c ErrorReportCooldown
		err := rows.Scan(&c.WordID, &c.DefinitionID, &c.ErrorType, &c.CooldownUntil)
		if err != nil {
			return nil, err
		}
		cooldowns = append(cooldowns, c)
	}

	return cooldowns, rows.Err()
}

// GetErrorReportStats gets statistics for error reports within a time range
func (r *ErrorReportRepository) GetErrorReportStats(ctx context.Context, startTime, endTime time.Time) (*ErrorReportStats, error) {
	stats := &ErrorReportStats{
		ReportsByType: make(map[string]int64),
		PeriodStart:   startTime,
		PeriodEnd:     endTime,
	}

	// Get total reports
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM error_reports 
		WHERE reported_at BETWEEN $1 AND $2
	`, startTime, endTime).Scan(&stats.TotalReports)
	if err != nil {
		return nil, err
	}

	// Get reports by type
	rows, err := r.db.Query(ctx, `
		SELECT error_type, COUNT(*) as count
		FROM error_reports
		WHERE reported_at BETWEEN $1 AND $2
		GROUP BY error_type
		ORDER BY count DESC
	`, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var errorType string
		var count int64
		if scanErr := rows.Scan(&errorType, &count); scanErr != nil {
			return nil, scanErr
		}
		stats.ReportsByType[errorType] = count
	}

	// Get top reporters
	rows, err = r.db.Query(ctx, `
		SELECT 
			er.user_id,
			COUNT(*) as report_count,
			COALESCE(u.subscription_plan, 'free') as subscription_plan
		FROM error_reports er
		JOIN users u ON u.id = er.user_id
		WHERE er.reported_at BETWEEN $1 AND $2
		GROUP BY er.user_id, u.subscription_plan
		ORDER BY report_count DESC
		LIMIT 10
	`, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userStats UserReportStats
		if scanErr := rows.Scan(&userStats.UserID, &userStats.ReportCount, &userStats.SubscriptionPlan); scanErr != nil {
			return nil, scanErr
		}
		stats.ReportsByUser = append(stats.ReportsByUser, userStats)
	}

	// Get recent reports
	rows, err = r.db.Query(ctx, `
		SELECT 
			user_id,
			error_type,
			reported_at,
			COALESCE(word_id, 0) as word_id,
			COALESCE(definition_id, 0) as definition_id
		FROM error_reports
		WHERE reported_at BETWEEN $1 AND $2
		ORDER BY reported_at DESC
		LIMIT 20
	`, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var report RecentErrorReport
		if err := rows.Scan(&report.UserID, &report.ErrorType, &report.ReportedAt, &report.WordID, &report.DefinitionID); err != nil {
			return nil, err
		}
		stats.RecentReports = append(stats.RecentReports, report)
	}

	return stats, nil
}

// CheckCooldown checks if a user is in cooldown for a specific error report
func (r *ErrorReportRepository) CheckCooldown(ctx context.Context, userID int64, wordID int64, definitionID *int64, errorType string) (*time.Time, error) {
	return checkCooldown(ctx, r.db, userID, wordID, definitionID, errorType)
}

// CheckCooldownTx performs the authoritative cooldown check inside the caller's
// transaction. Error reporting uses this after locking the owned word so two
// concurrent requests cannot both pass the initial advisory check.
func (r *ErrorReportRepository) CheckCooldownTx(ctx context.Context, tx pgx.Tx, userID int64, wordID int64, definitionID *int64, errorType string) (*time.Time, error) {
	return checkCooldown(ctx, tx, userID, wordID, definitionID, errorType)
}

type cooldownQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func checkCooldown(ctx context.Context, querier cooldownQuerier, userID int64, wordID int64, definitionID *int64, errorType string) (*time.Time, error) {
	var cooldownUntil *time.Time

	// Convert pointer to interface{} for SQL parameter
	var defIDParam interface{}
	if definitionID != nil {
		defIDParam = *definitionID
	} else {
		defIDParam = nil
	}

	query := `
		SELECT cooldown_until 
		FROM error_report_cooldowns 
		WHERE user_id = $1 
			AND word_id = $2
			AND definition_id IS NOT DISTINCT FROM $3
			AND error_type = $4
			AND cooldown_until > NOW()
		LIMIT 1
	`

	err := querier.QueryRow(ctx, query, userID, wordID, defIDParam, errorType).Scan(&cooldownUntil)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return cooldownUntil, err
}

// SetCooldown sets a cooldown period for a specific error report
func (r *ErrorReportRepository) SetCooldown(ctx context.Context, tx pgx.Tx, userID int64, wordID int64, definitionID *int64, errorType string, cooldownUntil time.Time) error {
	// Convert pointer to interface{} for SQL parameter
	var defIDParam interface{}
	if definitionID != nil {
		defIDParam = *definitionID
	} else {
		defIDParam = nil
	}

	query := `
		INSERT INTO error_report_cooldowns (user_id, word_id, definition_id, error_type, cooldown_until)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, COALESCE(word_id, -1), COALESCE(definition_id, -1), error_type)
		DO UPDATE SET 
			cooldown_until = EXCLUDED.cooldown_until,
			updated_at = NOW()
	`

	_, err := tx.Exec(ctx, query, userID, wordID, defIDParam, errorType, cooldownUntil)
	return err
}

// UpdateLastRegeneratedAt updates the last regenerated timestamp for words or definitions
func (r *ErrorReportRepository) UpdateWordAudioLastRegeneratedAt(ctx context.Context, tx pgx.Tx, wordID int64) error {
	_, err := tx.Exec(ctx,
		"UPDATE words SET audio_last_regenerated_at = NOW() WHERE id = $1",
		wordID)
	return err
}

func (r *ErrorReportRepository) UpdateDefinitionLastRegeneratedAt(ctx context.Context, tx pgx.Tx, definitionID int64) error {
	_, err := tx.Exec(ctx,
		"UPDATE definitions SET last_regenerated_at = NOW() WHERE id = $1",
		definitionID)
	return err
}

// UpsertErrorReport updates or inserts an error report with content snapshot
func (r *ErrorReportRepository) UpsertErrorReport(ctx context.Context, tx pgx.Tx, userID int64, definitionID *int64, wordID int64, errorType string, contentSnapshot map[string]any) error {
	// Convert pointer to interface{} for SQL parameter
	var defIDParam any
	if definitionID != nil {
		defIDParam = *definitionID
	} else {
		defIDParam = nil
	}

	// Convert content snapshot to JSON
	var snapshotJSON any
	if contentSnapshot != nil {
		snapshotJSON = contentSnapshot
	} else {
		snapshotJSON = nil
	}

	// First, try to update an existing row
	tag, err := tx.Exec(ctx, `
		UPDATE error_reports
		SET
			error_type = $4,
			reported_at = NOW(),
			status = 'pending',
			content_snapshot = $5
		WHERE
			user_id = $1
			AND definition_id IS NOT DISTINCT FROM $2
			AND word_id = $3
			AND status = 'pending'
			AND CASE
				WHEN error_type IN ('_unrelated_meaning', '_unrelated_example') THEN 'definition'
				WHEN error_type IN ('_unrelated_image', '_missing_image') THEN 'image'
				ELSE error_type
			END = CASE
				WHEN $4 IN ('_unrelated_meaning', '_unrelated_example') THEN 'definition'
				WHEN $4 IN ('_unrelated_image', '_missing_image') THEN 'image'
				ELSE $4
			END
	`, userID, defIDParam, wordID, errorType, snapshotJSON)
	if err != nil {
		return err
	}

	// If no row was updated, insert a new one
	if tag.RowsAffected() == 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO error_reports
				(user_id, definition_id, word_id, error_type, reported_at, status, content_snapshot)
			VALUES
				($1, $2, $3, $4, NOW(), 'pending', $5)
		`, userID, defIDParam, wordID, errorType, snapshotJSON)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteUserErrorReports deletes all error reports for a user
func (r *ErrorReportRepository) DeleteUserErrorReports(ctx context.Context, userID int64) (int64, error) {
	result, err := r.db.Exec(ctx, `DELETE FROM error_reports WHERE user_id=$1`, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// Types used by the repository

type ErrorReportCooldown struct {
	WordID        int64
	DefinitionID  int64
	ErrorType     string
	CooldownUntil time.Time
}

type ErrorReportStats struct {
	TotalReports  int64
	ReportsByType map[string]int64
	ReportsByUser []UserReportStats
	RecentReports []RecentErrorReport
	RateLimitHits int64
	PeriodStart   time.Time
	PeriodEnd     time.Time
}

type UserReportStats struct {
	UserID           int64
	ReportCount      int64
	SubscriptionPlan string
}

type RecentErrorReport struct {
	UserID       int64
	ErrorType    string
	ReportedAt   time.Time
	WordID       int64
	DefinitionID int64
}
