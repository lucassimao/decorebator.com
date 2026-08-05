package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"decorebator.com/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const providerEventSemanticConstraint = "provider_event_inbox_semantic_key_unique"

type ProviderEventInboxRepository struct {
	db *pgxpool.Pool
}

type ProviderEventInboxInput struct {
	Store              model.EntitlementStore
	Environment        *model.StoreEnvironment
	IdempotencyKey     model.ProviderEventKey
	SemanticKey        *model.ProviderEventKey
	ProviderEventType  string
	ProviderOccurredAt time.Time
}

type ProviderEventClaim struct {
	ID         int64
	LeaseToken uuid.UUID
	Input      ProviderEventInboxInput
}

type ProviderEventInboxResult struct {
	Claim     *ProviderEventClaim
	Duplicate bool
	Outcome   model.EntitlementOperationOutcome
	Code      model.EntitlementResultCode
}

type ProviderEventInboxHealth struct {
	OverdueRetryable       int64
	ExpiredProcessing      int64
	OldestRetryableDueAt   *time.Time
	OldestProcessingExpiry *time.Time
}

type ProviderEventProcessor func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error)

func NewProviderEventInboxRepository(db *pgxpool.Pool) *ProviderEventInboxRepository {
	return &ProviderEventInboxRepository{db: db}
}

// Claim creates or reclaims a durable processing lease in its own short
// transaction. Provider I/O happens after this returns and before Complete,
// so slow network calls never hold database locks or pool connections.
func (r *ProviderEventInboxRepository) Claim(
	ctx context.Context,
	input ProviderEventInboxInput,
	now time.Time,
	leaseDuration time.Duration,
) (ProviderEventInboxResult, error) {
	if err := validateProviderEventInboxInput(input, now, leaseDuration); err != nil {
		return ProviderEventInboxResult{}, err
	}
	if r == nil || r.db == nil {
		return ProviderEventInboxResult{}, fmt.Errorf("provider event inbox database is required")
	}
	leaseToken := uuid.New()
	leaseExpiresAt := now.Add(leaseDuration)
	var eventID int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO provider_event_inbox (
			store, environment, idempotency_key, semantic_key,
			provider_event_type, provider_occurred_at, state,
			attempt_count, lease_token, lease_expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'processing', 1, $7, $8)
		ON CONFLICT (idempotency_key) DO UPDATE SET
			state = 'processing',
			outcome = NULL,
			result_code = NULL,
			attempt_count = provider_event_inbox.attempt_count + 1,
			lease_token = EXCLUDED.lease_token,
			lease_expires_at = EXCLUDED.lease_expires_at,
			next_attempt_at = NULL,
			updated_at = $9
		WHERE (provider_event_inbox.state = 'processing' AND provider_event_inbox.lease_expires_at <= $9)
		   OR (provider_event_inbox.state = 'retryable' AND provider_event_inbox.next_attempt_at <= $9)
		RETURNING id
	`, input.Store, input.Environment, input.IdempotencyKey.String(), providerEventKeyString(input.SemanticKey),
		input.ProviderEventType, input.ProviderOccurredAt.UTC(), leaseToken, leaseExpiresAt.UTC(), now.UTC()).Scan(&eventID)
	if err == nil {
		claim := &ProviderEventClaim{ID: eventID, LeaseToken: leaseToken, Input: input}
		return ProviderEventInboxResult{Claim: claim}, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return r.readDuplicate(ctx, input)
	}
	if semanticConflict(err) {
		return r.readSemanticDuplicate(ctx, input)
	}
	return ProviderEventInboxResult{}, err
}

// Complete locks the active lease, runs only the local database mutation, and
// commits that mutation with the final inbox result atomically.
func (r *ProviderEventInboxRepository) Complete(
	ctx context.Context,
	claim ProviderEventClaim,
	processor ProviderEventProcessor,
) (ProviderEventInboxResult, error) {
	if r == nil || r.db == nil || claim.ID <= 0 || claim.LeaseToken == uuid.Nil || processor == nil {
		return ProviderEventInboxResult{}, fmt.Errorf("valid provider event claim and processor are required")
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ProviderEventInboxResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var eventID int64
	err = tx.QueryRow(ctx, `
		SELECT id
		FROM provider_event_inbox
		WHERE id = $1 AND state = 'processing' AND lease_token = $2 AND lease_expires_at > NOW()
		FOR UPDATE
	`, claim.ID, claim.LeaseToken).Scan(&eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProviderEventInboxResult{}, fmt.Errorf("provider event lease is no longer active")
	}
	if err != nil {
		return ProviderEventInboxResult{}, err
	}

	operation, err := processor(ctx, tx)
	if err != nil {
		return ProviderEventInboxResult{}, err
	}
	if validationErr := validateCompletedOperation(operation); validationErr != nil {
		return ProviderEventInboxResult{}, validationErr
	}
	state := "completed"
	if operation.Outcome == model.EntitlementOutcomeRejected {
		state = "rejected"
	}
	command, err := tx.Exec(ctx, `
		UPDATE provider_event_inbox
		SET state = $2, outcome = $3, result_code = $4,
			lease_token = NULL, lease_expires_at = NULL, next_attempt_at = NULL, updated_at = NOW()
		WHERE id = $1 AND lease_token = $5
	`, eventID, state, operation.Outcome, operation.Code, claim.LeaseToken)
	if err != nil {
		return ProviderEventInboxResult{}, err
	}
	if command.RowsAffected() != 1 {
		return ProviderEventInboxResult{}, fmt.Errorf("provider event completion lost its lease")
	}
	if err = tx.Commit(ctx); err != nil {
		return ProviderEventInboxResult{}, err
	}
	return ProviderEventInboxResult{Outcome: operation.Outcome, Code: operation.Code}, nil
}

func (r *ProviderEventInboxRepository) MarkRetryable(
	ctx context.Context,
	claim ProviderEventClaim,
	nextAttemptAt time.Time,
) error {
	if r == nil || r.db == nil || claim.ID <= 0 || claim.LeaseToken == uuid.Nil || nextAttemptAt.IsZero() {
		return fmt.Errorf("valid provider event claim and retry time are required")
	}
	command, err := r.db.Exec(ctx, `
		UPDATE provider_event_inbox
		SET state = 'retryable', outcome = 'retry', result_code = 'retryable_provider_error',
			lease_token = NULL, lease_expires_at = NULL, next_attempt_at = $3, updated_at = NOW()
		WHERE id = $1 AND state = 'processing' AND lease_token = $2
	`, claim.ID, claim.LeaseToken, nextAttemptAt.UTC())
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return fmt.Errorf("provider event retry lost its lease")
	}
	return nil
}

// StrandedHealth returns aggregate, non-sensitive alert data. Inbox payloads
// are intentionally not stored, so this method never attempts replay or state
// mutation; provider redelivery and reconciliation remain the recovery paths.
func (r *ProviderEventInboxRepository) StrandedHealth(
	ctx context.Context,
	now time.Time,
	gracePeriod time.Duration,
) (ProviderEventInboxHealth, error) {
	if r == nil || r.db == nil || now.IsZero() || gracePeriod < 0 {
		return ProviderEventInboxHealth{}, fmt.Errorf("valid provider inbox health inputs are required")
	}
	cutoff := now.UTC().Add(-gracePeriod)
	var health ProviderEventInboxHealth
	err := r.db.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE state = 'retryable' AND next_attempt_at <= $1),
			count(*) FILTER (WHERE state = 'processing' AND lease_expires_at <= $1),
			min(next_attempt_at) FILTER (WHERE state = 'retryable' AND next_attempt_at <= $1),
			min(lease_expires_at) FILTER (WHERE state = 'processing' AND lease_expires_at <= $1)
		FROM provider_event_inbox
	`, cutoff).Scan(
		&health.OverdueRetryable,
		&health.ExpiredProcessing,
		&health.OldestRetryableDueAt,
		&health.OldestProcessingExpiry,
	)
	return health, err
}

func (r *ProviderEventInboxRepository) readDuplicate(
	ctx context.Context,
	input ProviderEventInboxInput,
) (ProviderEventInboxResult, error) {
	return r.readExisting(ctx, "idempotency_key", input.IdempotencyKey.String())
}

func (r *ProviderEventInboxRepository) readSemanticDuplicate(
	ctx context.Context,
	input ProviderEventInboxInput,
) (ProviderEventInboxResult, error) {
	if input.SemanticKey == nil {
		return ProviderEventInboxResult{}, fmt.Errorf("provider event semantic conflict has no semantic key")
	}
	return r.readExisting(ctx, "semantic_key", input.SemanticKey.String())
}

func (r *ProviderEventInboxRepository) readExisting(
	ctx context.Context,
	column string,
	key string,
) (ProviderEventInboxResult, error) {
	if column != "idempotency_key" && column != "semantic_key" {
		return ProviderEventInboxResult{}, fmt.Errorf("unsupported provider event lookup")
	}
	var state string
	var outcome *model.EntitlementOperationOutcome
	var code *model.EntitlementResultCode
	query := "SELECT state, outcome, result_code FROM provider_event_inbox WHERE " + column + " = $1"
	err := r.db.QueryRow(ctx, query, key).Scan(&state, &outcome, &code)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProviderEventInboxResult{}, fmt.Errorf("provider event conflict is not yet visible; retry claim")
	}
	if err != nil {
		return ProviderEventInboxResult{}, err
	}
	if outcome == nil || code == nil {
		return ProviderEventInboxResult{
			Duplicate: true, Outcome: model.EntitlementOutcomeRetry, Code: model.EntitlementResultRetryableProvider,
		}, nil
	}
	return ProviderEventInboxResult{Duplicate: true, Outcome: *outcome, Code: *code}, nil
}

func validateProviderEventInboxInput(input ProviderEventInboxInput, now time.Time, leaseDuration time.Duration) error {
	providerTime := input.ProviderOccurredAt.UTC()
	if !input.Store.Valid() || input.IdempotencyKey == "" || len(input.IdempotencyKey) > model.ProviderKeyMaxLength ||
		strings.TrimSpace(input.ProviderEventType) == "" || providerTime.Before(time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)) ||
		now.IsZero() || leaseDuration <= 0 {
		return fmt.Errorf("invalid provider event inbox input")
	}
	if input.Environment != nil && !input.Environment.Valid() {
		return fmt.Errorf("invalid provider event environment")
	}
	if input.Store == model.EntitlementStoreApple && input.Environment == nil {
		return fmt.Errorf("Apple provider event environment is required")
	}
	if input.SemanticKey != nil && (*input.SemanticKey == "" || len(*input.SemanticKey) > model.ProviderKeyMaxLength) {
		return fmt.Errorf("invalid provider semantic event key")
	}
	return nil
}

func validateCompletedOperation(operation model.EntitlementOperationResult) error {
	if operation.Outcome == "" || operation.Code == "" ||
		operation.Outcome == model.EntitlementOutcomeDuplicate || operation.Outcome == model.EntitlementOutcomeRetry {
		return fmt.Errorf("processor returned an invalid completion result")
	}
	return nil
}

func semanticConflict(err error) bool {
	var pgError *pgconn.PgError
	return errors.As(err, &pgError) && pgError.Code == "23505" && pgError.ConstraintName == providerEventSemanticConstraint
}

func providerEventKeyString(key *model.ProviderEventKey) *string {
	if key == nil {
		return nil
	}
	value := key.String()
	return &value
}
