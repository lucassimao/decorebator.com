package repository

import (
	"context"
	"encoding/json"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RealtimeTelemetryRepository struct {
	db *pgxpool.Pool
}

func NewRealtimeTelemetryRepository(db *pgxpool.Pool) *RealtimeTelemetryRepository {
	return &RealtimeTelemetryRepository{db: db}
}

func (r *RealtimeTelemetryRepository) InsertRealtimeChatTelemetry(ctx context.Context, telemetry *model.RealtimeChatTelemetry) error {
	turnsJSON, err := json.Marshal(telemetry.Turns)
	if err != nil {
		return err
	}

	rateLimitsJSON, err := json.Marshal(telemetry.RateLimits)
	if err != nil {
		return err
	}

	var deviceJSONString *string
	if telemetry.DeviceInfo != nil {
		deviceJSON, marshalErr := json.Marshal(telemetry.DeviceInfo)
		if marshalErr != nil {
			return marshalErr
		}
		str := string(deviceJSON)
		deviceJSONString = &str
	}

	_, err = r.db.Exec(ctx, `
    INSERT INTO realtime_chat_sessions (
      user_id,
      wordlist_id,
      session_uuid,
      conversation_id,
      language_code,
      model,
      started_at,
      ended_at,
      duration_ms,
      total_audio_tokens_in,
      total_audio_tokens_out,
      total_text_tokens_in,
      total_text_tokens_out,
      total_turns,
      turns,
      rate_limits,
      error_count,
      client_version,
      device_info
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9,
      $10, $11, $12, $13, $14,
      $15::jsonb, $16::jsonb, $17, $18, CASE WHEN $19::text IS NULL THEN NULL ELSE $19::jsonb END
    )
  `,
		telemetry.UserID,
		telemetry.WordlistID,
		telemetry.SessionUUID,
		telemetry.ConversationID,
		telemetry.LanguageCode,
		telemetry.Model,
		telemetry.StartedAt,
		telemetry.EndedAt,
		telemetry.DurationMs,
		telemetry.TotalAudioTokensIn,
		telemetry.TotalAudioTokensOut,
		telemetry.TotalTextTokensIn,
		telemetry.TotalTextTokensOut,
		telemetry.TotalTurns,
		string(turnsJSON),
		string(rateLimitsJSON),
		telemetry.ErrorCount,
		telemetry.ClientVersion,
		deviceJSONString,
	)

	return err
}
