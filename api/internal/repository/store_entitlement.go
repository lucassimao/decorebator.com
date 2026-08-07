package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/security"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrStoreEvidenceOwnershipConflict = errors.New("store evidence is already bound to another user")

const storeIdentityGenerationAttempts = 3

type StoreEntitlementRepository struct {
	db        *pgxpool.Pool
	protector *security.StoreEvidenceProtector
}

type StoreEntitlementUpdate struct {
	Entitlement              model.StoreEntitlement
	AccountIdentifier        string `json:"-"`
	ProviderRecordID         string `json:"-"`
	LinkedProviderID         string `json:"-"`
	ProviderVersion          string
	Acknowledged             *bool
	SnapshotObservedAt       time.Time
	ProviderSnapshotSignedAt *time.Time
	EventOccurredAt          *time.Time
	CurrentBinding           bool
	HistoricalBinding        bool
	CancellationReason       *string
}

type StoreEntitlementApplyResult struct {
	Operation   model.EntitlementOperationResult
	Entitlement model.StoreEntitlement
}

type StorePurchaseBinding struct {
	ID                int64
	EntitlementID     int64
	UserID            int64
	Store             model.EntitlementStore
	Environment       model.StoreEnvironment
	AccountIdentifier string `json:"-"`
	ProviderRecordID  string `json:"-"`
	Current           bool
	MatchedPrimary    bool
	Acknowledged      *bool
}

type StoreAcknowledgementRetryCandidate struct {
	BindingID int64
	CreatedAt time.Time
}

type lockedStoreEntitlement struct {
	Entitlement            model.StoreEntitlement
	CancellationReason     *string
	LastSnapshotObservedAt *time.Time
	LastEventOccurredAt    *time.Time
	Created                bool
	Exists                 bool
}

type providerBindingState struct {
	Exists                  bool
	PostRevocationCandidate bool
}

func NewStoreEntitlementRepository(
	db *pgxpool.Pool,
	protector *security.StoreEvidenceProtector,
) (*StoreEntitlementRepository, error) {
	if db == nil || protector == nil {
		return nil, fmt.Errorf("store entitlement database and evidence protector are required")
	}
	return &StoreEntitlementRepository{db: db, protector: protector}, nil
}

func (r *StoreEntitlementRepository) GetOrCreateAppleAccount(
	ctx context.Context,
	userID int64,
) (model.AppleAccountIdentity, error) {
	identifier, err := r.getOrCreateAccount(ctx, userID, model.EntitlementStoreApple)
	if err != nil {
		return model.AppleAccountIdentity{}, err
	}
	token, err := uuid.Parse(identifier)
	if err != nil {
		return model.AppleAccountIdentity{}, fmt.Errorf("decrypt Apple account identity: invalid stored identifier")
	}
	return model.AppleAccountIdentity{UserID: userID, AppAccountToken: token}, nil
}

func (r *StoreEntitlementRepository) GetOrCreateGoogleAccount(
	ctx context.Context,
	userID int64,
) (model.GoogleAccountIdentity, error) {
	identifier, err := r.getOrCreateAccount(ctx, userID, model.EntitlementStoreGoogle)
	if err != nil {
		return model.GoogleAccountIdentity{}, err
	}
	identity := model.GoogleAccountIdentity{UserID: userID, ObfuscatedExternalAccountID: identifier}
	if err = identity.Validate(); err != nil {
		return model.GoogleAccountIdentity{}, fmt.Errorf("decrypt Google account identity: invalid stored identifier")
	}
	return identity, nil
}

func (r *StoreEntitlementRepository) getOrCreateAccount(
	ctx context.Context,
	userID int64,
	store model.EntitlementStore,
) (string, error) {
	if userID <= 0 || !store.Valid() {
		return "", fmt.Errorf("valid user and store are required")
	}
	if existing, found, err := r.readAccount(ctx, r.db, userID, store); err != nil || found {
		return existing, err
	}
	for range storeIdentityGenerationAttempts {
		identifier, err := generateStoreAccountIdentifier(store)
		if err != nil {
			return "", err
		}
		lookup, err := r.protector.LookupDigest(store, nil, security.StoreEvidenceAccountIdentifier, identifier)
		if err != nil {
			return "", err
		}
		protected, err := r.protector.Encrypt(security.StoreEvidenceContext{
			UserID: userID, Store: store, Kind: security.StoreEvidenceAccountIdentifier, Lookup: lookup,
		}, identifier)
		if err != nil {
			return "", err
		}
		command, err := r.db.Exec(ctx, `
			INSERT INTO store_account_identities (
				user_id, store, lookup_digest, encrypted_identifier, encryption_key_version
			) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING
		`, userID, store, lookup, protected.Ciphertext, protected.KeyVersion)
		if err != nil {
			return "", err
		}
		if command.RowsAffected() == 1 {
			return identifier, nil
		}
		if existing, found, readErr := r.readAccount(ctx, r.db, userID, store); readErr != nil || found {
			return existing, readErr
		}
	}
	return "", fmt.Errorf("generate unique store account identity")
}

func (r *StoreEntitlementRepository) readAccount(
	ctx context.Context,
	executor interface {
		QueryRow(context.Context, string, ...any) pgx.Row
	},
	userID int64,
	store model.EntitlementStore,
) (string, bool, error) {
	var lookup string
	var ciphertext []byte
	var keyVersion int16
	err := executor.QueryRow(ctx, `
		SELECT lookup_digest, encrypted_identifier, encryption_key_version
		FROM store_account_identities
		WHERE user_id = $1 AND store = $2
	`, userID, store).Scan(&lookup, &ciphertext, &keyVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	plaintext, err := r.protector.Decrypt(security.StoreEvidenceContext{
		UserID: userID, Store: store, Kind: security.StoreEvidenceAccountIdentifier, Lookup: lookup,
	}, security.ProtectedStoreEvidence{Ciphertext: ciphertext, KeyVersion: keyVersion})
	if err != nil {
		return "", false, err
	}
	return plaintext, true, nil
}

// GetCurrentEntitlement returns the canonical row for one exact environment.
// Sandbox and production rows are never merged on this read path.
func (r *StoreEntitlementRepository) GetCurrentEntitlement(
	ctx context.Context,
	userID int64,
	store model.EntitlementStore,
	environment model.StoreEnvironment,
) (model.StoreEntitlement, bool, error) {
	if userID <= 0 || !store.Valid() || !environment.Valid() {
		return model.StoreEntitlement{}, false, fmt.Errorf("valid entitlement lookup is required")
	}
	var entitlement model.StoreEntitlement
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, store, product_id, entitlement, status, period_start,
		       period_end, auto_renew_enabled, canceled_at, revoked_at,
		       revocation_reason, environment, last_verified_at, created_at, updated_at
		FROM store_entitlements
		WHERE user_id=$1 AND store=$2 AND environment=$3 AND entitlement='premium'
	`, userID, store, environment).Scan(
		&entitlement.ID, &entitlement.UserID, &entitlement.Store, &entitlement.ProductID,
		&entitlement.Entitlement, &entitlement.Status, &entitlement.PeriodStart,
		&entitlement.PeriodEnd, &entitlement.AutoRenewEnabled, &entitlement.CanceledAt,
		&entitlement.RevokedAt, &entitlement.RevocationReason, &entitlement.Environment,
		&entitlement.LastVerifiedAt, &entitlement.CreatedAt, &entitlement.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.StoreEntitlement{}, false, nil
	}
	return entitlement, err == nil, err
}

func (r *StoreEntitlementRepository) CurrentGoogleBindingAcknowledged(
	ctx context.Context,
	entitlementID int64,
) (bool, error) {
	if entitlementID <= 0 {
		return false, fmt.Errorf("valid entitlement ID is required")
	}
	var acknowledged bool
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(acknowledged, FALSE)
		FROM store_purchase_bindings
		WHERE entitlement_id=$1 AND store='google' AND is_current
	`, entitlementID).Scan(&acknowledged)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return acknowledged, err
}

func (r *StoreEntitlementRepository) GetPurchaseBindingByID(
	ctx context.Context,
	bindingID int64,
) (StorePurchaseBinding, bool, error) {
	if bindingID <= 0 {
		return StorePurchaseBinding{}, false, fmt.Errorf("valid binding ID is required")
	}
	var binding StorePurchaseBinding
	var providerLookup, accountLookup string
	var providerCiphertext, accountCiphertext []byte
	var providerVersion, accountVersion int16
	err := r.db.QueryRow(ctx, `
		SELECT binding.id, binding.entitlement_id, binding.user_id, binding.store,
		       binding.environment, binding.is_current, binding.acknowledged,
		       binding.provider_record_digest, binding.encrypted_provider_record,
		       binding.encryption_key_version, account.lookup_digest,
		       account.encrypted_identifier, account.encryption_key_version
		FROM store_purchase_bindings binding
		JOIN store_account_identities account ON account.id=binding.account_identity_id
		WHERE binding.id=$1
	`, bindingID).Scan(
		&binding.ID, &binding.EntitlementID, &binding.UserID, &binding.Store,
		&binding.Environment, &binding.Current, &binding.Acknowledged,
		&providerLookup, &providerCiphertext, &providerVersion,
		&accountLookup, &accountCiphertext, &accountVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return StorePurchaseBinding{}, false, nil
	}
	if err != nil {
		return StorePurchaseBinding{}, false, err
	}
	binding.ProviderRecordID, err = r.protector.Decrypt(security.StoreEvidenceContext{
		UserID: binding.UserID, Store: binding.Store, Environment: &binding.Environment,
		Kind: security.StoreEvidenceProviderRecord, Lookup: providerLookup,
	}, security.ProtectedStoreEvidence{Ciphertext: providerCiphertext, KeyVersion: providerVersion})
	if err != nil {
		return StorePurchaseBinding{}, false, err
	}
	binding.AccountIdentifier, err = r.protector.Decrypt(security.StoreEvidenceContext{
		UserID: binding.UserID, Store: binding.Store,
		Kind: security.StoreEvidenceAccountIdentifier, Lookup: accountLookup,
	}, security.ProtectedStoreEvidence{Ciphertext: accountCiphertext, KeyVersion: accountVersion})
	if err != nil {
		return StorePurchaseBinding{}, false, err
	}
	return binding, true, nil
}

func (r *StoreEntitlementRepository) ListUnacknowledgedGoogleBindings(
	ctx context.Context,
	environment model.StoreEnvironment,
	limit int,
) ([]StoreAcknowledgementRetryCandidate, error) {
	if !environment.Valid() || limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("valid acknowledgement retry lookup is required")
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, created_at
		FROM store_purchase_bindings
		WHERE store='google' AND environment=$1 AND is_current AND acknowledged IS FALSE
		ORDER BY last_verified_at ASC, id ASC
		LIMIT $2
	`, environment, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	candidates := make([]StoreAcknowledgementRetryCandidate, 0)
	for rows.Next() {
		var candidate StoreAcknowledgementRetryCandidate
		if err = rows.Scan(&candidate.BindingID, &candidate.CreatedAt); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}

// HasEffectivePremiumAccess requires a current acknowledged Google binding;
// Apple has no analogous acknowledgement state.
func (r *StoreEntitlementRepository) HasEffectivePremiumAccess(
	ctx context.Context,
	userID int64,
	environment model.StoreEnvironment,
	now time.Time,
) (bool, error) {
	if userID <= 0 || !environment.Valid() || now.IsZero() {
		return false, fmt.Errorf("valid access lookup is required")
	}
	var grants bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM store_entitlements entitlement
			WHERE entitlement.user_id=$1
			  AND entitlement.environment=$2
			  AND entitlement.entitlement='premium'
			  AND entitlement.status IN ('active', 'grace_period')
			  AND entitlement.period_end > $3
			  AND (
				entitlement.store='apple'
				OR EXISTS (
					SELECT 1 FROM store_purchase_bindings binding
					WHERE binding.entitlement_id=entitlement.id
					  AND binding.is_current AND binding.acknowledged IS TRUE
				)
			  )
		)
	`, userID, environment, now.UTC()).Scan(&grants)
	return grants, err
}

// ListUserPurchaseBindings decrypts a bounded internal restore set. Callers
// must not log or serialize the returned identifiers.
func (r *StoreEntitlementRepository) ListUserPurchaseBindings(
	ctx context.Context,
	userID int64,
	store model.EntitlementStore,
	environment model.StoreEnvironment,
	limit int,
) ([]StorePurchaseBinding, error) {
	if userID <= 0 || !store.Valid() || !environment.Valid() || limit <= 0 || limit > 20 {
		return nil, fmt.Errorf("valid bounded binding lookup is required")
	}
	rows, err := r.db.Query(ctx, `
		SELECT binding.id, binding.entitlement_id, binding.user_id, binding.store,
		       binding.environment, binding.is_current, binding.acknowledged,
		       binding.provider_record_digest, binding.encrypted_provider_record,
		       binding.encryption_key_version, account.lookup_digest,
		       account.encrypted_identifier, account.encryption_key_version
		FROM store_purchase_bindings binding
		JOIN store_account_identities account ON account.id=binding.account_identity_id
		WHERE binding.user_id=$1 AND binding.store=$2 AND binding.environment=$3 AND binding.is_current
		ORDER BY binding.is_current DESC, binding.last_verified_at DESC, binding.id DESC
		LIMIT $4
	`, userID, store, environment, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bindings := make([]StorePurchaseBinding, 0)
	for rows.Next() {
		var binding StorePurchaseBinding
		var providerLookup, accountLookup string
		var providerCiphertext, accountCiphertext []byte
		var providerVersion, accountVersion int16
		if err = rows.Scan(
			&binding.ID, &binding.EntitlementID, &binding.UserID, &binding.Store,
			&binding.Environment, &binding.Current, &binding.Acknowledged,
			&providerLookup, &providerCiphertext, &providerVersion,
			&accountLookup, &accountCiphertext, &accountVersion,
		); err != nil {
			return nil, err
		}
		binding.ProviderRecordID, err = r.protector.Decrypt(security.StoreEvidenceContext{
			UserID: userID, Store: store, Environment: &environment,
			Kind: security.StoreEvidenceProviderRecord, Lookup: providerLookup,
		}, security.ProtectedStoreEvidence{Ciphertext: providerCiphertext, KeyVersion: providerVersion})
		if err != nil {
			return nil, err
		}
		binding.AccountIdentifier, err = r.protector.Decrypt(security.StoreEvidenceContext{
			UserID: userID, Store: store, Kind: security.StoreEvidenceAccountIdentifier, Lookup: accountLookup,
		}, security.ProtectedStoreEvidence{Ciphertext: accountCiphertext, KeyVersion: accountVersion})
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, rows.Err()
}

func (r *StoreEntitlementRepository) Apply(
	ctx context.Context,
	update StoreEntitlementUpdate,
	tx pgx.Tx,
) (StoreEntitlementApplyResult, error) {
	if err := validateStoreEntitlementUpdate(update); err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	if tx != nil {
		if _, err := tx.Exec(ctx, "SAVEPOINT store_entitlement_apply"); err != nil {
			return StoreEntitlementApplyResult{}, err
		}
		result, err := r.applyInTx(ctx, tx, update)
		if err != nil {
			if _, rollbackErr := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT store_entitlement_apply"); rollbackErr != nil {
				return StoreEntitlementApplyResult{}, fmt.Errorf("store entitlement apply failed and savepoint rollback failed: %w", rollbackErr)
			}
			_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT store_entitlement_apply")
			return StoreEntitlementApplyResult{}, err
		}
		if _, err = tx.Exec(ctx, "RELEASE SAVEPOINT store_entitlement_apply"); err != nil {
			return StoreEntitlementApplyResult{}, err
		}
		return result, nil
	}
	managed, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	defer common.RollbackTx(ctx, managed, "store entitlement apply")
	result, err := r.applyInTx(ctx, managed, update)
	if err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	if err = managed.Commit(ctx); err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	return result, nil
}

// ReencryptEvidenceBatch rewrites a bounded number of rows that still use an
// older encryption key. Operators must observe CountEvidenceAtVersion == 0
// before removing that key from the configured decryption keyring.
func (r *StoreEntitlementRepository) ReencryptEvidenceBatch(
	ctx context.Context,
	fromVersion int16,
	limit int,
) (int, error) {
	if fromVersion <= 0 || fromVersion == r.protector.ActiveVersion() || limit <= 0 || limit > 1000 {
		return 0, fmt.Errorf("valid non-active evidence key version and batch limit are required")
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer common.RollbackTx(ctx, tx, "store entitlement re-encryption")
	updated, err := r.reencryptAccounts(ctx, tx, fromVersion, limit)
	if err != nil {
		return 0, err
	}
	if updated < limit {
		bindingUpdates, bindingErr := r.reencryptBindings(ctx, tx, fromVersion, limit-updated)
		if bindingErr != nil {
			return 0, bindingErr
		}
		updated += bindingUpdates
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return updated, nil
}

func (r *StoreEntitlementRepository) CountEvidenceAtVersion(
	ctx context.Context,
	version int16,
) (int64, error) {
	if version <= 0 {
		return 0, fmt.Errorf("evidence key version must be positive")
	}
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM store_account_identities WHERE encryption_key_version = $1) +
			(SELECT count(*) FROM store_purchase_bindings
			 WHERE encryption_key_version = $1 OR linked_encryption_key_version = $1)
	`, version).Scan(&count)
	return count, err
}

// ResolvePurchaseBinding performs an internal blind-index lookup for restore
// and provider-notification processing. Its plaintext fields must never be
// serialized or logged.
func (r *StoreEntitlementRepository) ResolvePurchaseBinding(
	ctx context.Context,
	store model.EntitlementStore,
	environment model.StoreEnvironment,
	providerRecordID string,
) (StorePurchaseBinding, bool, error) {
	lookup, err := r.protector.LookupDigest(
		store, &environment, security.StoreEvidenceProviderRecord, providerRecordID,
	)
	if err != nil {
		return StorePurchaseBinding{}, false, err
	}
	var binding StorePurchaseBinding
	var matchedCiphertext, accountCiphertext []byte
	var matchedVersion, accountVersion int16
	var matchedPrimary bool
	var accountLookup string
	err = r.db.QueryRow(ctx, `
		SELECT binding.id, binding.entitlement_id, binding.user_id, binding.store, binding.environment,
		       binding.is_current, binding.acknowledged,
		       binding.provider_record_digest = $3,
		       CASE WHEN binding.provider_record_digest = $3 THEN binding.encrypted_provider_record ELSE binding.encrypted_linked_record END,
		       CASE WHEN binding.provider_record_digest = $3 THEN binding.encryption_key_version ELSE binding.linked_encryption_key_version END,
		       account.lookup_digest, account.encrypted_identifier, account.encryption_key_version
		FROM store_purchase_bindings binding
		JOIN store_account_identities account ON account.id = binding.account_identity_id
		WHERE binding.store=$1 AND binding.environment=$2
		  AND (binding.provider_record_digest=$3 OR binding.linked_record_digest=$3)
		ORDER BY (binding.provider_record_digest=$3) DESC, binding.is_current DESC, binding.last_verified_at DESC
		LIMIT 1
	`, store, environment, lookup).Scan(
		&binding.ID, &binding.EntitlementID, &binding.UserID, &binding.Store, &binding.Environment,
		&binding.Current, &binding.Acknowledged,
		&matchedPrimary, &matchedCiphertext, &matchedVersion,
		&accountLookup, &accountCiphertext, &accountVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return StorePurchaseBinding{}, false, nil
	}
	if err != nil {
		return StorePurchaseBinding{}, false, err
	}
	matchedKind := security.StoreEvidenceLinkedRecord
	if matchedPrimary {
		matchedKind = security.StoreEvidenceProviderRecord
	}
	binding.MatchedPrimary = matchedPrimary
	binding.ProviderRecordID, err = r.protector.Decrypt(security.StoreEvidenceContext{
		UserID: binding.UserID, Store: binding.Store, Environment: &binding.Environment,
		Kind: matchedKind, Lookup: lookup,
	}, security.ProtectedStoreEvidence{Ciphertext: matchedCiphertext, KeyVersion: matchedVersion})
	if err != nil {
		return StorePurchaseBinding{}, false, err
	}
	binding.AccountIdentifier, err = r.protector.Decrypt(security.StoreEvidenceContext{
		UserID: binding.UserID, Store: binding.Store,
		Kind: security.StoreEvidenceAccountIdentifier, Lookup: accountLookup,
	}, security.ProtectedStoreEvidence{Ciphertext: accountCiphertext, KeyVersion: accountVersion})
	if err != nil {
		return StorePurchaseBinding{}, false, err
	}
	return binding, true, nil
}

// ApplyGoogleVoidedPurchase revokes only the currently bound entitlement for
// an authenticated RTDN void event. Historical replacement tokens cannot
// revoke a newer subscription cycle.
func (r *StoreEntitlementRepository) ApplyGoogleVoidedPurchase(
	ctx context.Context,
	tx pgx.Tx,
	binding StorePurchaseBinding,
	occurredAt time.Time,
	refundType int32,
) (model.EntitlementOperationResult, error) {
	if tx == nil || binding.EntitlementID <= 0 || binding.UserID <= 0 ||
		binding.Store != model.EntitlementStoreGoogle || !binding.Environment.Valid() ||
		occurredAt.IsZero() || refundType < 0 {
		return model.EntitlementOperationResult{}, fmt.Errorf("valid Google voided-purchase mutation is required")
	}
	if !binding.Current || !binding.MatchedPrimary {
		return model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeUnchanged,
			Code:    model.EntitlementResultStaleEvent,
		}, nil
	}
	var status model.EntitlementStatus
	var lastEventOccurredAt *time.Time
	err := tx.QueryRow(ctx, `
		SELECT status, last_event_occurred_at
		FROM store_entitlements
		WHERE id=$1 AND user_id=$2 AND store=$3 AND environment=$4
		FOR UPDATE
	`, binding.EntitlementID, binding.UserID, binding.Store, binding.Environment).Scan(
		&status, &lastEventOccurredAt,
	)
	if err != nil {
		return model.EntitlementOperationResult{}, err
	}
	occurredAt = occurredAt.UTC()
	// Equal-time void events remain restrictive. Google can emit cancellation
	// and void notifications for one billing action at the same millisecond;
	// dropping the latter would preserve refunded access.
	if lastEventOccurredAt != nil && occurredAt.Before(lastEventOccurredAt.UTC()) {
		return model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeUnchanged,
			Status:  status,
			Code:    model.EntitlementResultStaleEvent,
		}, nil
	}
	reason := fmt.Sprintf("google_voided_purchase_refund_type_%d", refundType)
	command, err := tx.Exec(ctx, `
		UPDATE store_entitlements
		SET status='revoked', auto_renew_enabled=FALSE,
		    revoked_at=COALESCE(revoked_at, $5),
		    revocation_reason=COALESCE(revocation_reason, $6),
		    last_event_occurred_at=$5, updated_at=NOW()
		WHERE id=$1 AND user_id=$2 AND store=$3 AND environment=$4
	`, binding.EntitlementID, binding.UserID, binding.Store, binding.Environment, occurredAt, reason)
	if err != nil {
		return model.EntitlementOperationResult{}, err
	}
	if command.RowsAffected() != 1 {
		return model.EntitlementOperationResult{}, fmt.Errorf("Google voided-purchase mutation lost its entitlement")
	}
	if status == model.EntitlementStatusRevoked {
		return model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeUnchanged,
			Status:  model.EntitlementStatusRevoked,
			Code:    model.EntitlementResultRevocationRetained,
		}, nil
	}
	return model.EntitlementOperationResult{
		Outcome: model.EntitlementOutcomeApplied,
		Status:  model.EntitlementStatusRevoked,
		Code:    model.EntitlementResultRevoked,
	}, nil
}

func (r *StoreEntitlementRepository) reencryptAccounts(
	ctx context.Context,
	tx pgx.Tx,
	fromVersion int16,
	limit int,
) (int, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, user_id, store, lookup_digest, encrypted_identifier
		FROM store_account_identities
		WHERE encryption_key_version = $1
		ORDER BY id
		FOR UPDATE SKIP LOCKED
		LIMIT $2
	`, fromVersion, limit)
	if err != nil {
		return 0, err
	}
	type accountRow struct {
		id, userID int64
		store      model.EntitlementStore
		lookup     string
		ciphertext []byte
	}
	accounts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (accountRow, error) {
		var value accountRow
		scanErr := row.Scan(&value.id, &value.userID, &value.store, &value.lookup, &value.ciphertext)
		return value, scanErr
	})
	if err != nil {
		return 0, err
	}
	for _, account := range accounts {
		context := security.StoreEvidenceContext{
			UserID: account.userID, Store: account.store,
			Kind: security.StoreEvidenceAccountIdentifier, Lookup: account.lookup,
		}
		plaintext, decryptErr := r.protector.Decrypt(context, security.ProtectedStoreEvidence{
			Ciphertext: account.ciphertext, KeyVersion: fromVersion,
		})
		if decryptErr != nil {
			return 0, decryptErr
		}
		protected, encryptErr := r.protector.Encrypt(context, plaintext)
		if encryptErr != nil {
			return 0, encryptErr
		}
		if _, err = tx.Exec(ctx, `
			UPDATE store_account_identities
			SET encrypted_identifier=$2, encryption_key_version=$3, updated_at=NOW()
			WHERE id=$1 AND encryption_key_version=$4
		`, account.id, protected.Ciphertext, protected.KeyVersion, fromVersion); err != nil {
			return 0, err
		}
	}
	return len(accounts), nil
}

func (r *StoreEntitlementRepository) reencryptBindings(
	ctx context.Context,
	tx pgx.Tx,
	fromVersion int16,
	limit int,
) (int, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, user_id, store, environment, provider_record_digest,
		       encrypted_provider_record, encryption_key_version,
		       linked_record_digest, encrypted_linked_record, linked_encryption_key_version
		FROM store_purchase_bindings
		WHERE encryption_key_version = $1 OR linked_encryption_key_version = $1
		ORDER BY id
		FOR UPDATE SKIP LOCKED
		LIMIT $2
	`, fromVersion, limit)
	if err != nil {
		return 0, err
	}
	type bindingRow struct {
		id, userID         int64
		store              model.EntitlementStore
		environment        model.StoreEnvironment
		providerLookup     string
		providerCiphertext []byte
		providerVersion    int16
		linkedLookup       *string
		linkedCiphertext   []byte
		linkedVersion      *int16
	}
	bindings, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (bindingRow, error) {
		var value bindingRow
		scanErr := row.Scan(
			&value.id, &value.userID, &value.store, &value.environment,
			&value.providerLookup, &value.providerCiphertext, &value.providerVersion,
			&value.linkedLookup, &value.linkedCiphertext, &value.linkedVersion,
		)
		return value, scanErr
	})
	if err != nil {
		return 0, err
	}
	for _, binding := range bindings {
		providerCiphertext, providerVersion := binding.providerCiphertext, binding.providerVersion
		if binding.providerVersion == fromVersion {
			context := security.StoreEvidenceContext{
				UserID: binding.userID, Store: binding.store, Environment: &binding.environment,
				Kind: security.StoreEvidenceProviderRecord, Lookup: binding.providerLookup,
			}
			plaintext, decryptErr := r.protector.Decrypt(context, security.ProtectedStoreEvidence{
				Ciphertext: binding.providerCiphertext, KeyVersion: binding.providerVersion,
			})
			if decryptErr != nil {
				return 0, decryptErr
			}
			protected, encryptErr := r.protector.Encrypt(context, plaintext)
			if encryptErr != nil {
				return 0, encryptErr
			}
			providerCiphertext, providerVersion = protected.Ciphertext, protected.KeyVersion
		}
		linkedCiphertext, linkedVersion := binding.linkedCiphertext, binding.linkedVersion
		if binding.linkedVersion != nil && *binding.linkedVersion == fromVersion {
			context := security.StoreEvidenceContext{
				UserID: binding.userID, Store: binding.store, Environment: &binding.environment,
				Kind: security.StoreEvidenceLinkedRecord, Lookup: *binding.linkedLookup,
			}
			plaintext, decryptErr := r.protector.Decrypt(context, security.ProtectedStoreEvidence{
				Ciphertext: binding.linkedCiphertext, KeyVersion: *binding.linkedVersion,
			})
			if decryptErr != nil {
				return 0, decryptErr
			}
			protected, encryptErr := r.protector.Encrypt(context, plaintext)
			if encryptErr != nil {
				return 0, encryptErr
			}
			linkedCiphertext = protected.Ciphertext
			value := protected.KeyVersion
			linkedVersion = &value
		}
		if _, err = tx.Exec(ctx, `
			UPDATE store_purchase_bindings
			SET encrypted_provider_record=$2, encryption_key_version=$3,
			    encrypted_linked_record=$4, linked_encryption_key_version=$5, updated_at=NOW()
			WHERE id=$1
		`, binding.id, providerCiphertext, providerVersion, linkedCiphertext, linkedVersion); err != nil {
			return 0, err
		}
	}
	return len(bindings), nil
}

func (r *StoreEntitlementRepository) applyInTx(
	ctx context.Context,
	tx pgx.Tx,
	update StoreEntitlementUpdate,
) (StoreEntitlementApplyResult, error) {
	accountID, err := r.lockAccount(ctx, tx, update)
	if err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	providerLookup, err := r.protector.LookupDigest(
		update.Entitlement.Store, &update.Entitlement.Environment,
		security.StoreEvidenceProviderRecord, update.ProviderRecordID,
	)
	if err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	var linkedLookup *string
	if update.LinkedProviderID != "" {
		lookup, lookupErr := r.protector.LookupDigest(
			update.Entitlement.Store, &update.Entitlement.Environment,
			security.StoreEvidenceProviderRecord, update.LinkedProviderID,
		)
		if lookupErr != nil {
			return StoreEntitlementApplyResult{}, lookupErr
		}
		linkedLookup = &lookup
	}
	bindingState, err := r.rejectForeignBindings(ctx, tx, update, providerLookup, linkedLookup)
	if err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	locked, err := r.lockOrCreateEntitlement(ctx, tx, update)
	if err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	if !locked.Exists {
		return StoreEntitlementApplyResult{Operation: model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeUnchanged,
			Code:    model.EntitlementResultStaleEvent,
		}}, nil
	}
	merged, cancellationReason, lifecycleApplied, eventAdvanced, revocationRetained :=
		mergeStoreEntitlement(locked, update, bindingState)
	if err = r.persistEntitlement(ctx, tx, merged, cancellationReason, update, lifecycleApplied, eventAdvanced); err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	if err = r.persistBinding(
		ctx, tx, merged, accountID, providerLookup, linkedLookup, update,
		lifecycleApplied, revocationRetained, bindingState,
	); err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	persisted, err := r.readEntitlementByID(ctx, tx, merged.ID)
	if err != nil {
		return StoreEntitlementApplyResult{}, err
	}
	operation := entitlementOperation(persisted.Status, lifecycleApplied, revocationRetained)
	return StoreEntitlementApplyResult{Operation: operation, Entitlement: persisted}, nil
}

func (r *StoreEntitlementRepository) lockAccount(
	ctx context.Context,
	tx pgx.Tx,
	update StoreEntitlementUpdate,
) (int64, error) {
	lookup, err := r.protector.LookupDigest(
		update.Entitlement.Store, nil, security.StoreEvidenceAccountIdentifier, update.AccountIdentifier,
	)
	if err != nil {
		return 0, err
	}
	var accountID, ownerID int64
	err = tx.QueryRow(ctx, `
		SELECT id, user_id
		FROM store_account_identities
		WHERE store = $1 AND lookup_digest = $2
		FOR SHARE
	`, update.Entitlement.Store, lookup).Scan(&accountID, &ownerID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && ownerID != update.Entitlement.UserID) {
		return 0, ErrStoreEvidenceOwnershipConflict
	}
	if err != nil {
		return 0, err
	}
	return accountID, nil
}

func (r *StoreEntitlementRepository) rejectForeignBindings(
	ctx context.Context,
	tx pgx.Tx,
	update StoreEntitlementUpdate,
	providerLookup string,
	linkedLookup *string,
) (providerBindingState, error) {
	lookups := []string{providerLookup}
	if linkedLookup != nil && *linkedLookup != providerLookup {
		lookups = append(lookups, *linkedLookup)
	}
	sort.Strings(lookups)
	providerState := providerBindingState{}
	for _, lookup := range lookups {
		var ownerID int64
		var postRevocationCandidate bool
		err := tx.QueryRow(ctx, `
			SELECT user_id, post_revocation_candidate
			FROM store_purchase_bindings
			WHERE store = $1 AND environment = $2 AND provider_record_digest = $3
			FOR UPDATE
		`, update.Entitlement.Store, update.Entitlement.Environment, lookup).Scan(&ownerID, &postRevocationCandidate)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return providerBindingState{}, err
		}
		if ownerID != update.Entitlement.UserID {
			return providerBindingState{}, ErrStoreEvidenceOwnershipConflict
		}
		if lookup == providerLookup {
			providerState.Exists = true
			providerState.PostRevocationCandidate = postRevocationCandidate
		}
	}
	return providerState, nil
}

func (r *StoreEntitlementRepository) lockOrCreateEntitlement(
	ctx context.Context,
	tx pgx.Tx,
	update StoreEntitlementUpdate,
) (lockedStoreEntitlement, error) {
	e := update.Entitlement
	if update.HistoricalBinding {
		return r.lockExistingEntitlement(ctx, tx, update)
	}
	command, err := tx.Exec(ctx, `
		INSERT INTO store_entitlements (
			user_id, store, environment, product_id, entitlement, status,
			period_start, period_end, auto_renew_enabled, canceled_at,
			cancellation_reason, revoked_at, revocation_reason, last_verified_at,
			last_snapshot_observed_at, last_event_occurred_at,
			last_provider_snapshot_signed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (user_id, store, environment, entitlement) DO NOTHING
	`, e.UserID, e.Store, e.Environment, e.ProductID, e.Entitlement, e.Status,
		e.PeriodStart, e.PeriodEnd, e.AutoRenewEnabled, e.CanceledAt,
		update.CancellationReason, e.RevokedAt, e.RevocationReason, e.LastVerifiedAt,
		update.SnapshotObservedAt, update.EventOccurredAt, update.ProviderSnapshotSignedAt)
	if err != nil {
		return lockedStoreEntitlement{}, err
	}
	locked := lockedStoreEntitlement{Created: command.RowsAffected() == 1, Exists: true}
	err = tx.QueryRow(ctx, `
		SELECT id, user_id, store, product_id, entitlement, status, period_start,
		       period_end, auto_renew_enabled, canceled_at, revoked_at,
		       revocation_reason, environment, last_verified_at, created_at, updated_at,
		       cancellation_reason, last_snapshot_observed_at, last_event_occurred_at
		FROM store_entitlements
		WHERE user_id=$1 AND store=$2 AND environment=$3 AND entitlement=$4
		FOR UPDATE
	`, e.UserID, e.Store, e.Environment, e.Entitlement).Scan(
		&e.ID, &e.UserID, &e.Store, &e.ProductID, &e.Entitlement, &e.Status,
		&e.PeriodStart, &e.PeriodEnd, &e.AutoRenewEnabled, &e.CanceledAt, &e.RevokedAt,
		&e.RevocationReason, &e.Environment, &e.LastVerifiedAt, &e.CreatedAt, &e.UpdatedAt,
		&locked.CancellationReason, &locked.LastSnapshotObservedAt, &locked.LastEventOccurredAt,
	)
	locked.Entitlement = e
	return locked, err
}

func (r *StoreEntitlementRepository) lockExistingEntitlement(
	ctx context.Context,
	tx pgx.Tx,
	update StoreEntitlementUpdate,
) (lockedStoreEntitlement, error) {
	e := update.Entitlement
	locked := lockedStoreEntitlement{}
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, store, product_id, entitlement, status, period_start,
		       period_end, auto_renew_enabled, canceled_at, revoked_at,
		       revocation_reason, environment, last_verified_at, created_at, updated_at,
		       cancellation_reason, last_snapshot_observed_at, last_event_occurred_at
		FROM store_entitlements
		WHERE user_id=$1 AND store=$2 AND environment=$3 AND entitlement=$4
		FOR UPDATE
	`, e.UserID, e.Store, e.Environment, e.Entitlement).Scan(
		&e.ID, &e.UserID, &e.Store, &e.ProductID, &e.Entitlement, &e.Status,
		&e.PeriodStart, &e.PeriodEnd, &e.AutoRenewEnabled, &e.CanceledAt, &e.RevokedAt,
		&e.RevocationReason, &e.Environment, &e.LastVerifiedAt, &e.CreatedAt, &e.UpdatedAt,
		&locked.CancellationReason, &locked.LastSnapshotObservedAt, &locked.LastEventOccurredAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return locked, nil
	}
	if err != nil {
		return lockedStoreEntitlement{}, err
	}
	locked.Entitlement, locked.Exists = e, true
	return locked, nil
}

func mergeStoreEntitlement(
	locked lockedStoreEntitlement,
	update StoreEntitlementUpdate,
	bindingState providerBindingState,
) (model.StoreEntitlement, *string, bool, bool, bool) {
	current := locked.Entitlement
	incoming := update.Entitlement
	incoming.ID, incoming.CreatedAt, incoming.UpdatedAt = current.ID, current.CreatedAt, current.UpdatedAt
	eventAdvanced := update.EventOccurredAt != nil &&
		(locked.LastEventOccurredAt == nil || update.EventOccurredAt.After(*locked.LastEventOccurredAt))
	if locked.Created {
		return incoming, update.CancellationReason, !update.HistoricalBinding, eventAdvanced, false
	}
	snapshotNewer := locked.LastSnapshotObservedAt == nil || update.SnapshotObservedAt.After(*locked.LastSnapshotObservedAt)
	if update.HistoricalBinding {
		return current, locked.CancellationReason, false, eventAdvanced, false
	}
	if current.Status == model.EntitlementStatusRevoked && incoming.Status != model.EntitlementStatusRevoked {
		newPurchaseAfterRevocation := (!bindingState.Exists || bindingState.PostRevocationCandidate) &&
			snapshotNewer && incoming.PeriodStart != nil &&
			current.RevokedAt != nil && incoming.PeriodStart.After(*current.RevokedAt)
		if !newPurchaseAfterRevocation {
			return current, locked.CancellationReason, false, eventAdvanced, true
		}
	}
	if incoming.Status == model.EntitlementStatusRevoked {
		return incoming, update.CancellationReason, true, eventAdvanced, false
	}
	if !snapshotNewer {
		merged, reason, applied := mergeStaleStoreEntitlement(current, incoming, locked, update)
		return merged, reason, applied, eventAdvanced, false
	}
	merged, reason := mergeFreshStoreEntitlement(current, incoming, locked, update)
	return merged, reason, true, eventAdvanced, false
}

func mergeStaleStoreEntitlement(
	current model.StoreEntitlement,
	incoming model.StoreEntitlement,
	locked lockedStoreEntitlement,
	update StoreEntitlementUpdate,
) (model.StoreEntitlement, *string, bool) {
	if incoming.AutoRenewEnabled || incoming.CanceledAt == nil || !current.AutoRenewEnabled {
		return current, locked.CancellationReason, false
	}
	current.AutoRenewEnabled = false
	current.CanceledAt = incoming.CanceledAt
	if incoming.LastVerifiedAt.After(current.LastVerifiedAt) {
		current.LastVerifiedAt = incoming.LastVerifiedAt
	}
	return current, update.CancellationReason, true
}

func mergeFreshStoreEntitlement(
	current model.StoreEntitlement,
	incoming model.StoreEntitlement,
	locked lockedStoreEntitlement,
	update StoreEntitlementUpdate,
) (model.StoreEntitlement, *string) {
	if incoming.AutoRenewEnabled {
		incoming.CanceledAt = nil
		return incoming, nil
	}
	if current.CanceledAt != nil && !current.AutoRenewEnabled {
		incoming.CanceledAt = current.CanceledAt
		return incoming, locked.CancellationReason
	}
	return incoming, update.CancellationReason
}

func (r *StoreEntitlementRepository) persistEntitlement(
	ctx context.Context,
	tx pgx.Tx,
	entitlement model.StoreEntitlement,
	cancellationReason *string,
	update StoreEntitlementUpdate,
	lifecycleApplied bool,
	eventAdvanced bool,
) error {
	if !lifecycleApplied && !eventAdvanced {
		return nil
	}
	command, err := tx.Exec(ctx, `
		UPDATE store_entitlements SET
			product_id = CASE WHEN $2 THEN $3 ELSE product_id END,
			status = CASE WHEN $2 THEN $4 ELSE status END,
			period_start = CASE WHEN $2 THEN $5 ELSE period_start END,
			period_end = CASE WHEN $2 THEN $6 ELSE period_end END,
			auto_renew_enabled = CASE WHEN $2 THEN $7 ELSE auto_renew_enabled END,
			canceled_at = CASE WHEN $2 THEN $8 ELSE canceled_at END,
			cancellation_reason = CASE WHEN $2 THEN $9 ELSE cancellation_reason END,
			revoked_at = CASE WHEN $2 THEN $10 ELSE revoked_at END,
			revocation_reason = CASE WHEN $2 THEN $11 ELSE revocation_reason END,
			last_verified_at = CASE WHEN $2 THEN GREATEST(last_verified_at, $12) ELSE last_verified_at END,
			last_snapshot_observed_at = CASE
				WHEN $2 AND last_snapshot_observed_at IS NULL THEN $13
				WHEN $2 THEN GREATEST(last_snapshot_observed_at, $13)
				ELSE last_snapshot_observed_at
			END,
			last_provider_snapshot_signed_at = CASE WHEN $2 THEN $15 ELSE last_provider_snapshot_signed_at END,
			last_event_occurred_at = CASE
				WHEN $14::timestamptz IS NOT NULL AND (last_event_occurred_at IS NULL OR $14 > last_event_occurred_at) THEN $14
				ELSE last_event_occurred_at
			END,
			updated_at = NOW()
		WHERE id = $1
	`, entitlement.ID, lifecycleApplied, entitlement.ProductID, entitlement.Status,
		entitlement.PeriodStart, entitlement.PeriodEnd, entitlement.AutoRenewEnabled,
		entitlement.CanceledAt, cancellationReason, entitlement.RevokedAt,
		entitlement.RevocationReason, entitlement.LastVerifiedAt, update.SnapshotObservedAt,
		update.EventOccurredAt, update.ProviderSnapshotSignedAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return fmt.Errorf("store entitlement update affected no rows")
	}
	return nil
}

func (r *StoreEntitlementRepository) readEntitlementByID(
	ctx context.Context,
	tx pgx.Tx,
	entitlementID int64,
) (model.StoreEntitlement, error) {
	var entitlement model.StoreEntitlement
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, store, product_id, entitlement, status, period_start,
		       period_end, auto_renew_enabled, canceled_at, revoked_at,
		       revocation_reason, environment, last_verified_at, created_at, updated_at
		FROM store_entitlements WHERE id=$1
	`, entitlementID).Scan(
		&entitlement.ID, &entitlement.UserID, &entitlement.Store, &entitlement.ProductID,
		&entitlement.Entitlement, &entitlement.Status, &entitlement.PeriodStart,
		&entitlement.PeriodEnd, &entitlement.AutoRenewEnabled, &entitlement.CanceledAt,
		&entitlement.RevokedAt, &entitlement.RevocationReason, &entitlement.Environment,
		&entitlement.LastVerifiedAt, &entitlement.CreatedAt, &entitlement.UpdatedAt,
	)
	return entitlement, err
}

func (r *StoreEntitlementRepository) persistBinding(
	ctx context.Context,
	tx pgx.Tx,
	entitlement model.StoreEntitlement,
	accountID int64,
	providerLookup string,
	linkedLookup *string,
	update StoreEntitlementUpdate,
	lifecycleApplied bool,
	revocationRetained bool,
	bindingState providerBindingState,
) error {
	providerContext := security.StoreEvidenceContext{
		UserID: entitlement.UserID, Store: entitlement.Store, Environment: &entitlement.Environment,
		Kind: security.StoreEvidenceProviderRecord, Lookup: providerLookup,
	}
	protected, err := r.protector.Encrypt(providerContext, update.ProviderRecordID)
	if err != nil {
		return err
	}
	var linkedCiphertext []byte
	var linkedVersion *int16
	if linkedLookup != nil {
		linked, encryptErr := r.protector.Encrypt(security.StoreEvidenceContext{
			UserID: entitlement.UserID, Store: entitlement.Store, Environment: &entitlement.Environment,
			Kind: security.StoreEvidenceLinkedRecord, Lookup: *linkedLookup,
		}, update.LinkedProviderID)
		if encryptErr != nil {
			return encryptErr
		}
		linkedCiphertext = linked.Ciphertext
		linkedVersion = &linked.KeyVersion
	}
	postRevocationCandidate := revocationRetained &&
		(!bindingState.Exists || bindingState.PostRevocationCandidate)
	makeCurrent := update.CurrentBinding && lifecycleApplied
	if makeCurrent {
		_, err = tx.Exec(ctx, `UPDATE store_purchase_bindings SET is_current = FALSE, updated_at = NOW() WHERE entitlement_id = $1 AND is_current`, entitlement.ID)
		if err != nil {
			return err
		}
	}
	command, err := tx.Exec(ctx, `
		INSERT INTO store_purchase_bindings (
			entitlement_id, account_identity_id, user_id, store, environment,
			provider_record_digest, encrypted_provider_record, encryption_key_version,
			linked_record_digest, encrypted_linked_record, linked_encryption_key_version,
			provider_version, acknowledged, post_revocation_candidate, is_current, last_verified_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (store, environment, provider_record_digest) DO UPDATE SET
			account_identity_id = CASE WHEN $17 THEN EXCLUDED.account_identity_id ELSE store_purchase_bindings.account_identity_id END,
			entitlement_id = CASE WHEN $17 THEN EXCLUDED.entitlement_id ELSE store_purchase_bindings.entitlement_id END,
			encrypted_provider_record = EXCLUDED.encrypted_provider_record,
			encryption_key_version = EXCLUDED.encryption_key_version,
			linked_record_digest = CASE WHEN $17 THEN EXCLUDED.linked_record_digest ELSE store_purchase_bindings.linked_record_digest END,
			encrypted_linked_record = CASE WHEN $17 THEN EXCLUDED.encrypted_linked_record ELSE store_purchase_bindings.encrypted_linked_record END,
			linked_encryption_key_version = CASE WHEN $17 THEN EXCLUDED.linked_encryption_key_version ELSE store_purchase_bindings.linked_encryption_key_version END,
			provider_version = CASE WHEN $17 THEN EXCLUDED.provider_version ELSE store_purchase_bindings.provider_version END,
			acknowledged = CASE WHEN $17 THEN EXCLUDED.acknowledged ELSE store_purchase_bindings.acknowledged END,
			post_revocation_candidate = CASE WHEN $17 THEN FALSE ELSE (store_purchase_bindings.post_revocation_candidate OR EXCLUDED.post_revocation_candidate) END,
			is_current = CASE WHEN $17 THEN EXCLUDED.is_current ELSE store_purchase_bindings.is_current END,
			last_verified_at = GREATEST(store_purchase_bindings.last_verified_at, EXCLUDED.last_verified_at),
			updated_at = NOW()
		WHERE store_purchase_bindings.user_id = EXCLUDED.user_id
	`, entitlement.ID, accountID, entitlement.UserID, entitlement.Store, entitlement.Environment,
		providerLookup, protected.Ciphertext, protected.KeyVersion,
		linkedLookup, linkedCiphertext, linkedVersion, nullString(update.ProviderVersion),
		update.Acknowledged, postRevocationCandidate, makeCurrent,
		update.Entitlement.LastVerifiedAt, lifecycleApplied)
	if ownershipConstraint(err) {
		return ErrStoreEvidenceOwnershipConflict
	}
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return ErrStoreEvidenceOwnershipConflict
	}
	return nil
}

func validateStoreEntitlementUpdate(update StoreEntitlementUpdate) error {
	if err := update.Entitlement.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(update.AccountIdentifier) == "" || strings.TrimSpace(update.ProviderRecordID) == "" ||
		update.SnapshotObservedAt.IsZero() || !update.Entitlement.LastVerifiedAt.Equal(update.SnapshotObservedAt) {
		return fmt.Errorf("store entitlement evidence and snapshot observation are required")
	}
	if update.EventOccurredAt != nil && update.EventOccurredAt.IsZero() {
		return fmt.Errorf("provider event occurrence is invalid")
	}
	if update.Entitlement.Store == model.EntitlementStoreApple {
		if update.LinkedProviderID != "" || update.Acknowledged != nil || update.HistoricalBinding ||
			update.ProviderSnapshotSignedAt == nil || update.ProviderSnapshotSignedAt.IsZero() {
			return fmt.Errorf("Apple persistence input contains Google-only evidence")
		}
	} else if update.Acknowledged == nil || update.ProviderSnapshotSignedAt != nil {
		return fmt.Errorf("Google acknowledgement state is required")
	}
	if update.HistoricalBinding && update.CurrentBinding {
		return fmt.Errorf("historical store binding cannot be current")
	}
	if update.ProviderSnapshotSignedAt != nil &&
		update.ProviderSnapshotSignedAt.After(update.SnapshotObservedAt.Add(5*time.Minute)) {
		return fmt.Errorf("provider snapshot signature is too far in the future")
	}
	return nil
}

func entitlementOperation(
	status model.EntitlementStatus,
	lifecycleApplied bool,
	revocationRetained bool,
) model.EntitlementOperationResult {
	if revocationRetained {
		return model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeUnchanged,
			Status:  model.EntitlementStatusRevoked,
			Code:    model.EntitlementResultRevocationRetained,
		}
	}
	if !lifecycleApplied {
		return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeUnchanged, Code: model.EntitlementResultStaleEvent}
	}
	if status == model.EntitlementStatusPending {
		return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomePending, Status: status, Code: model.EntitlementResultPurchasePending}
	}
	if status == model.EntitlementStatusRevoked {
		return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied, Status: status, Code: model.EntitlementResultRevoked}
	}
	return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeApplied, Status: status, Code: model.EntitlementResultApplied}
}

func generateStoreAccountIdentifier(store model.EntitlementStore) (string, error) {
	if store == model.EntitlementStoreApple {
		return uuid.NewString(), nil
	}
	if store != model.EntitlementStoreGoogle {
		return "", fmt.Errorf("unsupported entitlement store")
	}
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate Google account identity: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func ownershipConstraint(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505" &&
		postgresError.ConstraintName == "store_purchase_bindings_provider_record_unique"
}
