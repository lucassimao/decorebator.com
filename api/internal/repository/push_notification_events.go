package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PushNotificationEventRepository struct {
	Db *pgxpool.Pool
}

func (r *PushNotificationEventRepository) InsertEvents(ctx context.Context, userIDs []int64, notificationType string, sentAt time.Time) error {
	if len(userIDs) == 0 {
		return nil
	}
	_, err := r.Db.Exec(ctx, `
		INSERT INTO push_notification_events (user_id, notification_type, sent_at)
		SELECT unnest($1::bigint[]), $2, $3
	`, pgtype.FlatArray[int64](userIDs), notificationType, sentAt)
	return err
}
