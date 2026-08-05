package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/security"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareStoreEntitlementTest(
	t *testing.T,
) (*repository.StoreEntitlementRepository, *pgxpool.Pool, context.Context, int64, int64) {
	t.Helper()
	db := setup.CreateTestDB(t)
	t.Cleanup(db.Close)
	require.NoError(t, setup.RunMigrations(db))
	ctx := context.Background()
	_, err := db.Exec(ctx, `TRUNCATE store_purchase_bindings, store_entitlements, store_account_identities RESTART IDENTITY`)
	require.NoError(t, err)
	protector, err := security.NewStoreEvidenceProtector(
		bytes.Repeat([]byte{1}, 32), 1, map[int16][]byte{1: bytes.Repeat([]byte{2}, 32)},
	)
	require.NoError(t, err)
	repo, err := repository.NewStoreEntitlementRepository(db, protector)
	require.NoError(t, err)
	return repo, db, ctx, insertStoreTestUser(t, db, "one"), insertStoreTestUser(t, db, "two")
}

func insertStoreTestUser(t *testing.T, db *pgxpool.Pool, suffix string) int64 {
	t.Helper()
	var userID int64
	err := db.QueryRow(context.Background(), `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES ('Store', 'Test', $1, 'not-a-real-password') RETURNING id
	`, fmt.Sprintf("store-%s-%d@example.com", suffix, time.Now().UnixNano())).Scan(&userID)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = db.Exec(context.Background(), "DELETE FROM users WHERE id=$1", userID) })
	return userID
}

func googleStoreUpdate(
	userID int64,
	account model.GoogleAccountIdentity,
	token string,
	observedAt time.Time,
) repository.StoreEntitlementUpdate {
	periodStart, periodEnd := observedAt.Add(-24*time.Hour), observedAt.Add(30*24*time.Hour)
	acknowledged := true
	return repository.StoreEntitlementUpdate{
		Entitlement: model.StoreEntitlement{
			UserID: userID, Store: model.EntitlementStoreGoogle,
			ProductID: "decorebator_monthly_premium1", Entitlement: "premium",
			Status: model.EntitlementStatusActive, PeriodStart: &periodStart, PeriodEnd: &periodEnd,
			AutoRenewEnabled: true, Environment: model.StoreEnvironmentProduction,
			LastVerifiedAt: observedAt,
		},
		AccountIdentifier: account.ObfuscatedExternalAccountID, ProviderRecordID: token,
		Acknowledged: &acknowledged, SnapshotObservedAt: observedAt, CurrentBinding: true,
	}
}

func TestStoreAccountsAreStableEncryptedAndOwnerBound(t *testing.T) {
	repo, db, ctx, firstUser, secondUser := prepareStoreEntitlementTest(t)
	first, err := repo.GetOrCreateGoogleAccount(ctx, firstUser)
	require.NoError(t, err)
	again, err := repo.GetOrCreateGoogleAccount(ctx, firstUser)
	require.NoError(t, err)
	second, err := repo.GetOrCreateGoogleAccount(ctx, secondUser)
	require.NoError(t, err)
	assert.Equal(t, first, again)
	assert.NotEqual(t, first.ObfuscatedExternalAccountID, second.ObfuscatedExternalAccountID)

	var ciphertext []byte
	require.NoError(t, db.QueryRow(ctx, `
		SELECT encrypted_identifier FROM store_account_identities WHERE user_id=$1 AND store='google'
	`, firstUser).Scan(&ciphertext))
	assert.NotContains(t, string(ciphertext), first.ObfuscatedExternalAccountID)
}

func TestStoreEntitlementRejectsOwnershipTransferWithoutPartialWrite(t *testing.T) {
	repo, db, ctx, firstUser, secondUser := prepareStoreEntitlementTest(t)
	firstAccount, err := repo.GetOrCreateGoogleAccount(ctx, firstUser)
	require.NoError(t, err)
	secondAccount, err := repo.GetOrCreateGoogleAccount(ctx, secondUser)
	require.NoError(t, err)
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	_, err = repo.Apply(ctx, googleStoreUpdate(firstUser, firstAccount, "shared-provider-token", now), nil)
	require.NoError(t, err)
	_, err = repo.Apply(ctx, googleStoreUpdate(secondUser, secondAccount, "shared-provider-token", now.Add(time.Minute)), nil)
	require.ErrorIs(t, err, repository.ErrStoreEvidenceOwnershipConflict)

	var entitlementCount, bindingCount int
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM store_entitlements").Scan(&entitlementCount))
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM store_purchase_bindings").Scan(&bindingCount))
	assert.Equal(t, 1, entitlementCount)
	assert.Equal(t, 1, bindingCount)
	resolved, found, err := repo.ResolvePurchaseBinding(
		ctx, model.EntitlementStoreGoogle, model.StoreEnvironmentProduction, "shared-provider-token",
	)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, firstUser, resolved.UserID)
	assert.Equal(t, firstAccount.ObfuscatedExternalAccountID, resolved.AccountIdentifier)
	assert.Equal(t, "shared-provider-token", resolved.ProviderRecordID)
	serialized, err := json.Marshal(resolved)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), "shared-provider-token")
	assert.NotContains(t, string(serialized), firstAccount.ObfuscatedExternalAccountID)
}

func TestEffectiveStoreAccessSeparatesEnvironmentAndRequiresGoogleAcknowledgement(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)

	production := googleStoreUpdate(userID, account, "production-token", now)
	unacknowledged := false
	production.Acknowledged = &unacknowledged
	_, err = repo.Apply(ctx, production, nil)
	require.NoError(t, err)
	grants, err := repo.HasEffectivePremiumAccess(ctx, userID, model.StoreEnvironmentProduction, now)
	require.NoError(t, err)
	assert.False(t, grants, "verified but unacknowledged Google state must not grant")

	sandbox := googleStoreUpdate(userID, account, "sandbox-token", now.Add(time.Minute))
	sandbox.Entitlement.Environment = model.StoreEnvironmentSandbox
	_, err = repo.Apply(ctx, sandbox, nil)
	require.NoError(t, err)
	grants, err = repo.HasEffectivePremiumAccess(ctx, userID, model.StoreEnvironmentSandbox, now.Add(time.Minute))
	require.NoError(t, err)
	assert.True(t, grants)
	grants, err = repo.HasEffectivePremiumAccess(ctx, userID, model.StoreEnvironmentProduction, now.Add(time.Minute))
	require.NoError(t, err)
	assert.False(t, grants, "sandbox access must never leak into production reads")

	acknowledged := true
	production.Acknowledged = &acknowledged
	production.SnapshotObservedAt = now.Add(2 * time.Minute)
	production.Entitlement.LastVerifiedAt = production.SnapshotObservedAt
	_, err = repo.Apply(ctx, production, nil)
	require.NoError(t, err)
	grants, err = repo.HasEffectivePremiumAccess(ctx, userID, model.StoreEnvironmentProduction, now.Add(2*time.Minute))
	require.NoError(t, err)
	assert.True(t, grants)

	bindings, err := repo.ListUserPurchaseBindings(
		ctx, userID, model.EntitlementStoreGoogle, model.StoreEnvironmentProduction, 20,
	)
	require.NoError(t, err)
	require.Len(t, bindings, 1)
	assert.Equal(t, "production-token", bindings[0].ProviderRecordID)
	assert.NotEqual(t, "sandbox-token", bindings[0].ProviderRecordID)

	effective, err := service.NewEffectiveAccessService(
		repo, model.StoreEnvironmentProduction, func() time.Time { return now.Add(2 * time.Minute) },
	)
	require.NoError(t, err)
	t.Setenv("STRIPE_API_KEY", "sk_test_not_real")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_not_real")
	t.Setenv("STRIPE_SUCCESS_URL", "https://example.invalid/success")
	t.Setenv("STRIPE_CANCEL_URL", "https://example.invalid/cancel")
	subscriptions := service.NewSubscriptionService(db, nil, nil)
	subscriptions.SetEffectiveAccess(effective)
	require.NoError(t, subscriptions.CheckSubscriptionLimits(
		ctx, userID, model.UserActionChatSession, &service.SubscriptionCheckOptions{},
	), "canonical store access must pass a real premium-only limit gate")
	users := service.NewUserService(db, subscriptions, nil)
	users.SetEffectiveAccess(effective)
	require.NoError(t, users.ValidateUserEligibilityForWorkers(ctx, userID))
	profile, _, err := users.GetProfile(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, model.PlanMonthly, profile.SubscriptionPlan, "profile must expose the effective compatibility projection")
}

func TestGoogleVoidedPurchaseRevokesOnlyCurrentBindingAndHonorsEventClock(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	update := googleStoreUpdate(userID, account, "current-token", now)
	eventAt := now.Add(time.Minute)
	update.EventOccurredAt = &eventAt
	_, err = repo.Apply(ctx, update, nil)
	require.NoError(t, err)
	binding, found, err := repo.ResolvePurchaseBinding(
		ctx, model.EntitlementStoreGoogle, model.StoreEnvironmentProduction, "current-token",
	)
	require.NoError(t, err)
	require.True(t, found)
	assert.True(t, binding.Current)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	result, err := repo.ApplyGoogleVoidedPurchase(ctx, tx, binding, eventAt, 2)
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))
	assert.Equal(t, model.EntitlementResultRevoked, result.Code)
	var status, reason string
	var autoRenew bool
	require.NoError(t, db.QueryRow(ctx, `
		SELECT status, auto_renew_enabled, revocation_reason
		FROM store_entitlements WHERE id=$1
	`, binding.EntitlementID).Scan(&status, &autoRenew, &reason))
	assert.Equal(t, string(model.EntitlementStatusRevoked), status)
	assert.False(t, autoRenew)
	assert.Equal(t, "google_voided_purchase_refund_type_2", reason)

	tx, err = db.Begin(ctx)
	require.NoError(t, err)
	result, err = repo.ApplyGoogleVoidedPurchase(ctx, tx, binding, eventAt.Add(-time.Second), 1)
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))
	assert.Equal(t, model.EntitlementResultStaleEvent, result.Code)
}

func TestGoogleVoidedHistoricalTokenCannotRevokeReplacement(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	_, err = repo.Apply(ctx, googleStoreUpdate(userID, account, "old-token", now), nil)
	require.NoError(t, err)
	replacement := googleStoreUpdate(userID, account, "new-token", now.Add(time.Minute))
	replacement.LinkedProviderID = "old-token"
	_, err = repo.Apply(ctx, replacement, nil)
	require.NoError(t, err)
	oldBinding, found, err := repo.ResolvePurchaseBinding(
		ctx, model.EntitlementStoreGoogle, model.StoreEnvironmentProduction, "old-token",
	)
	require.NoError(t, err)
	require.True(t, found)
	assert.False(t, oldBinding.Current)
	currentBindings, err := repo.ListUserPurchaseBindings(
		ctx, userID, model.EntitlementStoreGoogle, model.StoreEnvironmentProduction, 20,
	)
	require.NoError(t, err)
	require.Len(t, currentBindings, 1, "historical replacements must not consume the bounded restore set")
	assert.Equal(t, "new-token", currentBindings[0].ProviderRecordID)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	result, err := repo.ApplyGoogleVoidedPurchase(ctx, tx, oldBinding, now.Add(2*time.Minute), 2)
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))
	assert.Equal(t, model.EntitlementResultStaleEvent, result.Code)
	var status string
	require.NoError(t, db.QueryRow(ctx, "SELECT status FROM store_entitlements WHERE id=$1", oldBinding.EntitlementID).Scan(&status))
	assert.Equal(t, string(model.EntitlementStatusActive), status)
}

func TestGoogleVoidedLinkedTokenCannotRevokeCurrentPrimaryBinding(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	update := googleStoreUpdate(userID, account, "new-token", now)
	update.LinkedProviderID = "unknown-old-token"
	_, err = repo.Apply(ctx, update, nil)
	require.NoError(t, err)
	linkedBinding, found, err := repo.ResolvePurchaseBinding(
		ctx, model.EntitlementStoreGoogle, model.StoreEnvironmentProduction, "unknown-old-token",
	)
	require.NoError(t, err)
	require.True(t, found)
	assert.True(t, linkedBinding.Current)
	assert.False(t, linkedBinding.MatchedPrimary)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	result, err := repo.ApplyGoogleVoidedPurchase(ctx, tx, linkedBinding, now.Add(time.Minute), 2)
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))
	assert.Equal(t, model.EntitlementResultStaleEvent, result.Code)
	var status string
	require.NoError(t, db.QueryRow(ctx, "SELECT status FROM store_entitlements WHERE id=$1", linkedBinding.EntitlementID).Scan(&status))
	assert.Equal(t, string(model.EntitlementStatusActive), status)
}

func TestStoreEntitlementCallerTransactionRollsBackOwnershipConflictToSavepoint(t *testing.T) {
	repo, db, ctx, firstUser, secondUser := prepareStoreEntitlementTest(t)
	firstAccount, err := repo.GetOrCreateGoogleAccount(ctx, firstUser)
	require.NoError(t, err)
	secondAccount, err := repo.GetOrCreateGoogleAccount(ctx, secondUser)
	require.NoError(t, err)
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	_, err = repo.Apply(ctx, googleStoreUpdate(firstUser, firstAccount, "owned-token", now), nil)
	require.NoError(t, err)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	_, err = repo.Apply(ctx, googleStoreUpdate(secondUser, secondAccount, "owned-token", now.Add(time.Minute)), tx)
	require.ErrorIs(t, err, repository.ErrStoreEvidenceOwnershipConflict)
	require.NoError(t, tx.Commit(ctx))
	var secondEntitlements int
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM store_entitlements WHERE user_id=$1", secondUser).Scan(&secondEntitlements))
	assert.Zero(t, secondEntitlements)
}

func TestStoreEntitlementConcurrentOwnershipRaceCreatesOneOwner(t *testing.T) {
	repo, db, ctx, firstUser, secondUser := prepareStoreEntitlementTest(t)
	firstAccount, err := repo.GetOrCreateGoogleAccount(ctx, firstUser)
	require.NoError(t, err)
	secondAccount, err := repo.GetOrCreateGoogleAccount(ctx, secondUser)
	require.NoError(t, err)
	now := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	updates := []repository.StoreEntitlementUpdate{
		googleStoreUpdate(firstUser, firstAccount, "raced-token", now),
		googleStoreUpdate(secondUser, secondAccount, "raced-token", now),
	}
	start := make(chan struct{})
	errorsChannel := make(chan error, 2)
	var wait sync.WaitGroup
	for _, update := range updates {
		wait.Add(1)
		go func(value repository.StoreEntitlementUpdate) {
			defer wait.Done()
			<-start
			_, applyErr := repo.Apply(ctx, value, nil)
			errorsChannel <- applyErr
		}(update)
	}
	close(start)
	wait.Wait()
	close(errorsChannel)
	var successes, ownershipConflicts int
	for applyErr := range errorsChannel {
		if applyErr == nil {
			successes++
		} else if errors.Is(applyErr, repository.ErrStoreEvidenceOwnershipConflict) {
			ownershipConflicts++
		} else {
			require.NoError(t, applyErr)
		}
	}
	assert.Equal(t, 1, successes)
	assert.Equal(t, 1, ownershipConflicts)
	var entitlements, bindings int
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM store_entitlements").Scan(&entitlements))
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM store_purchase_bindings").Scan(&bindings))
	assert.Equal(t, 1, entitlements)
	assert.Equal(t, 1, bindings)
}

func TestStoreEntitlementConcurrentUpgradeKeepsNewestSnapshotCurrent(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	base := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	_, err = repo.Apply(ctx, googleStoreUpdate(userID, account, "old-token", base), nil)
	require.NoError(t, err)

	updates := []repository.StoreEntitlementUpdate{
		googleStoreUpdate(userID, account, "newer-token", base.Add(2*time.Minute)),
		googleStoreUpdate(userID, account, "middle-token", base.Add(time.Minute)),
	}
	for index := range updates {
		updates[index].LinkedProviderID = "old-token"
	}
	start := make(chan struct{})
	errorsChannel := make(chan error, len(updates))
	var wait sync.WaitGroup
	for _, update := range updates {
		wait.Add(1)
		go func(value repository.StoreEntitlementUpdate) {
			defer wait.Done()
			<-start
			_, applyErr := repo.Apply(ctx, value, nil)
			errorsChannel <- applyErr
		}(update)
	}
	close(start)
	wait.Wait()
	close(errorsChannel)
	for applyErr := range errorsChannel {
		require.NoError(t, applyErr)
	}

	var currentCount int
	var currentVerifiedAt time.Time
	require.NoError(t, db.QueryRow(ctx, `
		SELECT count(*), max(last_verified_at) FROM store_purchase_bindings WHERE is_current
	`).Scan(&currentCount, &currentVerifiedAt))
	assert.Equal(t, 1, currentCount)
	assert.True(t, currentVerifiedAt.Equal(base.Add(2*time.Minute)))
}

func TestStoreEntitlementHistoricalGoogleTokenCannotOverwriteCurrentLifecycleOrBindingMetadata(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	base := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	current := googleStoreUpdate(userID, account, "current-token", base.Add(2*time.Minute))
	current.ProviderVersion = "new-etag"
	_, err = repo.Apply(ctx, current, nil)
	require.NoError(t, err)

	historical := googleStoreUpdate(userID, account, "old-token", base.Add(3*time.Minute))
	historical.Entitlement.Status = model.EntitlementStatusExpired
	historical.Entitlement.AutoRenewEnabled = false
	historical.CurrentBinding = false
	historical.HistoricalBinding = true
	historical.LinkedProviderID = "unseen-linked-token"
	historical.ProviderVersion = "old-etag"
	acknowledged := false
	historical.Acknowledged = &acknowledged
	_, err = repo.Apply(ctx, historical, nil)
	require.NoError(t, err)

	staleCurrent := googleStoreUpdate(userID, account, "current-token", base.Add(time.Minute))
	staleCurrent.ProviderVersion = "stale-etag"
	staleCurrent.Acknowledged = &acknowledged
	_, err = repo.Apply(ctx, staleCurrent, nil)
	require.NoError(t, err)
	var status, version string
	var storedAcknowledged bool
	require.NoError(t, db.QueryRow(ctx, `
		SELECT entitlement.status, binding.provider_version, binding.acknowledged
		FROM store_entitlements entitlement
		JOIN store_purchase_bindings binding ON binding.entitlement_id=entitlement.id AND binding.is_current
		WHERE entitlement.user_id=$1
	`, userID).Scan(&status, &version, &storedAcknowledged))
	assert.Equal(t, string(model.EntitlementStatusActive), status)
	assert.Equal(t, "new-etag", version)
	assert.True(t, storedAcknowledged)
	linked, found, err := repo.ResolvePurchaseBinding(
		ctx, model.EntitlementStoreGoogle, model.StoreEnvironmentProduction, "unseen-linked-token",
	)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, userID, linked.UserID)
	assert.Equal(t, "unseen-linked-token", linked.ProviderRecordID)
}

func TestStoreEntitlementHistoricalFirstEvidenceCreatesNoCanonicalState(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	historical := googleStoreUpdate(
		userID, account, "historical-only-token", time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC),
	)
	historical.Entitlement.Status = model.EntitlementStatusExpired
	historical.Entitlement.AutoRenewEnabled = false
	historical.CurrentBinding = false
	historical.HistoricalBinding = true
	result, err := repo.Apply(ctx, historical, nil)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementResultStaleEvent, result.Operation.Code)
	var entitlements, bindings int
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM store_entitlements").Scan(&entitlements))
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM store_purchase_bindings").Scan(&bindings))
	assert.Zero(t, entitlements)
	assert.Zero(t, bindings)
}

func TestStoreEntitlementPendingRepurchaseCandidateCanActivateAfterRevocation(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	base := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	_, err = repo.Apply(ctx, googleStoreUpdate(userID, account, "revoked-token", base), nil)
	require.NoError(t, err)
	revoked := googleStoreUpdate(userID, account, "revoked-token", base.Add(time.Minute))
	revoked.Entitlement.Status = model.EntitlementStatusRevoked
	revoked.Entitlement.RevokedAt = timePointer(base.Add(time.Minute))
	revoked.Entitlement.AutoRenewEnabled = false
	_, err = repo.Apply(ctx, revoked, nil)
	require.NoError(t, err)

	pending := googleStoreUpdate(userID, account, "pending-repurchase-token", base.Add(2*time.Minute))
	pending.Entitlement.Status = model.EntitlementStatusPending
	pending.Entitlement.PeriodStart = nil
	pending.Entitlement.PeriodEnd = nil
	pending.Entitlement.AutoRenewEnabled = false
	pending.CurrentBinding = false
	retained, err := repo.Apply(ctx, pending, nil)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementResultRevocationRetained, retained.Operation.Code)
	var candidate bool
	require.NoError(t, db.QueryRow(ctx, `
		SELECT post_revocation_candidate FROM store_purchase_bindings
		WHERE user_id=$1 AND provider_record_digest <> (
			SELECT provider_record_digest FROM store_purchase_bindings WHERE user_id=$1 AND is_current
		)
	`, userID).Scan(&candidate))
	assert.True(t, candidate)

	activated := googleStoreUpdate(userID, account, "pending-repurchase-token", base.Add(3*time.Minute))
	activated.Entitlement.PeriodStart = timePointer(base.Add(2 * time.Minute))
	result, err := repo.Apply(ctx, activated, nil)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementResultApplied, result.Operation.Code)
	assert.Equal(t, model.EntitlementStatusActive, result.Entitlement.Status)
	require.NoError(t, db.QueryRow(ctx, `
		SELECT post_revocation_candidate FROM store_purchase_bindings WHERE user_id=$1 AND is_current
	`, userID).Scan(&candidate))
	assert.False(t, candidate)
}

func TestStoreEntitlementCancellationCyclesAndRevocationDominatesClockRaces(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	base := time.Date(2026, 8, 5, 15, 0, 0, 0, time.UTC)
	initial := googleStoreUpdate(userID, account, "token", base)
	_, err = repo.Apply(ctx, initial, nil)
	require.NoError(t, err)

	firstCancel := googleStoreUpdate(userID, account, "token", base.Add(time.Minute))
	firstCancel.Entitlement.AutoRenewEnabled = false
	firstCancel.Entitlement.CanceledAt = timePointer(base.Add(time.Minute))
	_, err = repo.Apply(ctx, firstCancel, nil)
	require.NoError(t, err)
	repeated := firstCancel
	repeated.SnapshotObservedAt, repeated.Entitlement.LastVerifiedAt = base.Add(2*time.Minute), base.Add(2*time.Minute)
	repeated.Entitlement.CanceledAt = timePointer(base.Add(2 * time.Minute))
	_, err = repo.Apply(ctx, repeated, nil)
	require.NoError(t, err)

	reactivated := googleStoreUpdate(userID, account, "token", base.Add(3*time.Minute))
	_, err = repo.Apply(ctx, reactivated, nil)
	require.NoError(t, err)
	secondCancel := googleStoreUpdate(userID, account, "token", base.Add(4*time.Minute))
	secondCancel.Entitlement.AutoRenewEnabled = false
	secondCancel.Entitlement.CanceledAt = timePointer(base.Add(4 * time.Minute))
	_, err = repo.Apply(ctx, secondCancel, nil)
	require.NoError(t, err)

	revoked := googleStoreUpdate(userID, account, "token", base.Add(2*time.Minute))
	revoked.Entitlement.Status = model.EntitlementStatusRevoked
	revoked.Entitlement.RevokedAt = timePointer(base.Add(5 * time.Minute))
	revoked.Entitlement.AutoRenewEnabled = false
	revoked.Entitlement.CanceledAt = nil
	eventAt := base.Add(5 * time.Minute)
	revoked.EventOccurredAt = &eventAt
	_, err = repo.Apply(ctx, revoked, nil)
	require.NoError(t, err)
	newerNonterminal := googleStoreUpdate(userID, account, "token", base.Add(6*time.Minute))
	retained, err := repo.Apply(ctx, newerNonterminal, nil)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementResultRevocationRetained, retained.Operation.Code)

	var status string
	var canceledAt, eventCursor *time.Time
	require.NoError(t, db.QueryRow(ctx, `
		SELECT status, canceled_at, last_event_occurred_at FROM store_entitlements WHERE user_id=$1
	`, userID).Scan(&status, &canceledAt, &eventCursor))
	assert.Equal(t, string(model.EntitlementStatusRevoked), status)
	assert.Nil(t, canceledAt)
	require.NotNil(t, eventCursor)
	assert.True(t, eventCursor.Equal(eventAt))

	repurchase := googleStoreUpdate(userID, account, "repurchase-token", base.Add(7*time.Minute))
	repurchase.Entitlement.PeriodStart = timePointer(base.Add(6 * time.Minute))
	repurchase.Entitlement.PeriodEnd = timePointer(base.Add(30 * 24 * time.Hour))
	result, err := repo.Apply(ctx, repurchase, nil)
	require.NoError(t, err)
	assert.Equal(t, model.EntitlementResultApplied, result.Operation.Code)
	assert.Equal(t, model.EntitlementStatusActive, result.Entitlement.Status)
}

func TestStoreEvidenceReencryptionMustFinishBeforeRetirement(t *testing.T) {
	repo, db, ctx, userID, _ := prepareStoreEntitlementTest(t)
	account, err := repo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	_, err = repo.Apply(ctx, googleStoreUpdate(userID, account, "token", time.Now().UTC()), nil)
	require.NoError(t, err)

	rotatedProtector, err := security.NewStoreEvidenceProtector(
		bytes.Repeat([]byte{1}, 32), 2,
		map[int16][]byte{1: bytes.Repeat([]byte{2}, 32), 2: bytes.Repeat([]byte{3}, 32)},
	)
	require.NoError(t, err)
	rotatedRepo, err := repository.NewStoreEntitlementRepository(db, rotatedProtector)
	require.NoError(t, err)
	before, err := rotatedRepo.CountEvidenceAtVersion(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), before)
	updated, err := rotatedRepo.ReencryptEvidenceBatch(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, updated)
	after, err := rotatedRepo.CountEvidenceAtVersion(ctx, 1)
	require.NoError(t, err)
	assert.Zero(t, after)
	restored, err := rotatedRepo.GetOrCreateGoogleAccount(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, account, restored)
}

func timePointer(value time.Time) *time.Time { return &value }
