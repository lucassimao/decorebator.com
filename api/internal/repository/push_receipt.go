package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PushReceiptRepository struct {
	Db *pgxpool.Pool
}

type PushReceiptInsert struct {
	TicketID  string
	ExpoToken string
}

type PushReceiptRow struct {
	TicketID  string
	ExpoToken string
}

func (r *PushReceiptRepository) InsertTickets(ctx context.Context, inserts []PushReceiptInsert) error {
	if len(inserts) == 0 {
		return nil
	}

	tx, err := r.Db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	for _, insert := range inserts {
		_, err := tx.Exec(ctx, `
			INSERT INTO push_receipts (expo_ticket_id, expo_token, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (expo_ticket_id) DO NOTHING
		`, insert.TicketID, insert.ExpoToken)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PushReceiptRepository) FetchPending(ctx context.Context, limit int) ([]PushReceiptRow, error) {
	rows, err := r.Db.Query(ctx, `
		SELECT expo_ticket_id, expo_token
		FROM push_receipts
		WHERE checked_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PushReceiptRow
	for rows.Next() {
		var row PushReceiptRow
		if err := rows.Scan(&row.TicketID, &row.ExpoToken); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *PushReceiptRepository) UpdateReceipt(ctx context.Context, ticketID string, status, message string, details map[string]interface{}, checkedAt time.Time) error {
	var detailsJSON []byte
	if details != nil {
		raw, err := json.Marshal(details)
		if err != nil {
			return err
		}
		detailsJSON = raw
	}

	_, err := r.Db.Exec(ctx, `
		UPDATE push_receipts
		SET status = $2,
			message = $3,
			details = $4,
			checked_at = $5,
			updated_at = NOW()
		WHERE expo_ticket_id = $1
	`, ticketID, status, message, detailsJSON, checkedAt)
	return err
}
