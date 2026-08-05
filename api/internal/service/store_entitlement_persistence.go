package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
)

type StoreEntitlementWriter interface {
	Apply(context.Context, repository.StoreEntitlementUpdate, pgx.Tx) (repository.StoreEntitlementApplyResult, error)
}

type StoreEntitlementPersistence struct {
	writer StoreEntitlementWriter
}

func NewStoreEntitlementPersistence(writer StoreEntitlementWriter) (*StoreEntitlementPersistence, error) {
	if writer == nil {
		return nil, fmt.Errorf("store entitlement writer is required")
	}
	return &StoreEntitlementPersistence{writer: writer}, nil
}

func (p *StoreEntitlementPersistence) PersistApple(
	ctx context.Context,
	tx pgx.Tx,
	verified VerifiedAppleSubscription,
	eventOccurredAt *time.Time,
) (repository.StoreEntitlementApplyResult, error) {
	if verified.Identity.Validate() != nil || verified.ProviderSnapshotSignedAt.IsZero() {
		return repository.StoreEntitlementApplyResult{}, fmt.Errorf("verified Apple subscription is invalid")
	}
	entitlement := model.StoreEntitlement{
		UserID: verified.Identity.UserID, Store: model.EntitlementStoreApple,
		ProductID: verified.ProductID, Entitlement: verified.Entitlement, Status: verified.Status,
		PeriodStart: verified.PeriodStart, PeriodEnd: verified.PeriodEnd,
		AutoRenewEnabled: verified.AutoRenewEnabled, CanceledAt: verified.CancellationObservedAt,
		RevokedAt: verified.RevokedAt, RevocationReason: verified.RevocationReason,
		Environment: verified.Identity.Environment, LastVerifiedAt: verified.Identity.LastVerifiedAt,
	}
	result, err := p.writer.Apply(ctx, repository.StoreEntitlementUpdate{
		Entitlement: entitlement, AccountIdentifier: verified.Identity.AppAccountToken.String(),
		ProviderRecordID:         verified.Identity.OriginalTransactionID,
		SnapshotObservedAt:       verified.Identity.LastVerifiedAt,
		ProviderSnapshotSignedAt: &verified.ProviderSnapshotSignedAt,
		EventOccurredAt:          eventOccurredAt,
		CurrentBinding:           true,
	}, tx)
	return classifyStorePersistenceResult(result, err)
}

func (p *StoreEntitlementPersistence) PersistGoogle(
	ctx context.Context,
	tx pgx.Tx,
	processed GooglePurchaseProcessingResult,
	eventOccurredAt *time.Time,
) (repository.StoreEntitlementApplyResult, error) {
	verified := processed.Purchase
	if verified.Identity.Validate() != nil {
		return repository.StoreEntitlementApplyResult{}, fmt.Errorf("verified Google purchase is invalid")
	}
	entitlement := model.StoreEntitlement{
		UserID: verified.Identity.UserID, Store: model.EntitlementStoreGoogle,
		ProductID: verified.ProductID, Entitlement: verified.Entitlement, Status: verified.Status,
		PeriodStart: verified.PeriodStart, PeriodEnd: verified.PeriodEnd,
		AutoRenewEnabled: verified.AutoRenewEnabled, CanceledAt: verified.CanceledAt,
		Environment: verified.Identity.Environment, LastVerifiedAt: verified.Identity.LastVerifiedAt,
	}
	acknowledged := verified.Acknowledged
	result, err := p.writer.Apply(ctx, repository.StoreEntitlementUpdate{
		Entitlement:       entitlement,
		AccountIdentifier: verified.Identity.ObfuscatedExternalAccountID,
		ProviderRecordID:  verified.Identity.PurchaseToken, LinkedProviderID: verified.LinkedPurchaseToken,
		ProviderVersion: verified.ProviderETag, Acknowledged: &acknowledged,
		SnapshotObservedAt: verified.Identity.LastVerifiedAt, EventOccurredAt: eventOccurredAt,
		CurrentBinding: !verified.Superseded, HistoricalBinding: verified.Superseded,
		CancellationReason: verified.CancellationReason,
	}, tx)
	return classifyStorePersistenceResult(result, err)
}

func classifyStorePersistenceResult(
	result repository.StoreEntitlementApplyResult,
	err error,
) (repository.StoreEntitlementApplyResult, error) {
	if errors.Is(err, repository.ErrStoreEvidenceOwnershipConflict) {
		return repository.StoreEntitlementApplyResult{Operation: model.EntitlementOperationResult{
			Outcome: model.EntitlementOutcomeRejected,
			Code:    model.EntitlementResultAccountMismatch,
		}}, nil
	}
	return result, err
}
