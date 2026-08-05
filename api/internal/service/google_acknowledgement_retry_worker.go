package service

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/riverqueue/river"
)

const GoogleAcknowledgementRetryQueue = "store-acknowledgement"

type GoogleAcknowledgementRetryArgs struct {
	BindingID int64 `json:"binding_id" river:"unique"`
}

func (GoogleAcknowledgementRetryArgs) Kind() string { return "google-store-acknowledgement" }

type GoogleAcknowledgementBindingReader interface {
	GetPurchaseBindingByID(context.Context, int64) (repository.StorePurchaseBinding, bool, error)
}

type GoogleAcknowledgementSweepReader interface {
	ListUnacknowledgedGoogleBindings(context.Context, model.StoreEnvironment, int) ([]repository.StoreAcknowledgementRetryCandidate, error)
}

type GoogleAcknowledgementRetryWorker struct {
	river.WorkerDefaults[GoogleAcknowledgementRetryArgs]
	bindings    GoogleAcknowledgementBindingReader
	processor   GoogleProviderPurchaseProcessor
	persistence *StoreEntitlementPersistence
	environment model.StoreEnvironment
	now         func() time.Time
}

func NewGoogleAcknowledgementRetryWorker(
	bindings GoogleAcknowledgementBindingReader,
	processor GoogleProviderPurchaseProcessor,
	persistence *StoreEntitlementPersistence,
	environment model.StoreEnvironment,
	now func() time.Time,
) (*GoogleAcknowledgementRetryWorker, error) {
	if bindings == nil || processor == nil || persistence == nil || !environment.Valid() {
		return nil, fmt.Errorf("Google acknowledgement retry dependencies are required")
	}
	if now == nil {
		now = time.Now
	}
	return &GoogleAcknowledgementRetryWorker{
		bindings: bindings, processor: processor, persistence: persistence,
		environment: environment, now: now,
	}, nil
}

func (w *GoogleAcknowledgementRetryWorker) Work(
	ctx context.Context,
	job *river.Job[GoogleAcknowledgementRetryArgs],
) error {
	return w.retryBinding(ctx, job.Args.BindingID, job.CreatedAt)
}

func (w *GoogleAcknowledgementRetryWorker) retryBinding(
	ctx context.Context,
	bindingID int64,
	createdAt time.Time,
) error {
	age := w.now().UTC().Sub(createdAt.UTC())
	if age >= 60*time.Hour {
		common.Logger.ErrorContext(ctx, "Google acknowledgement retry exhausted",
			"binding_id", bindingID, "age_hours", int(age.Hours()))
		return nil
	}
	binding, found, err := w.bindings.GetPurchaseBindingByID(ctx, bindingID)
	if err != nil {
		return fmt.Errorf("resolve Google acknowledgement binding: %w", err)
	}
	if !found || !binding.Current || binding.Store != model.EntitlementStoreGoogle ||
		binding.Environment != w.environment || (binding.Acknowledged != nil && *binding.Acknowledged) {
		return nil
	}
	account := model.GoogleAccountIdentity{
		UserID: binding.UserID, ObfuscatedExternalAccountID: binding.AccountIdentifier,
	}
	processed, err := w.processor.VerifyAndAcknowledge(ctx, GooglePurchaseVerificationInput{
		AuthenticatedUserID: binding.UserID, PurchaseToken: binding.ProviderRecordID, Account: account,
	})
	if err != nil {
		return fmt.Errorf("retry Google acknowledgement: %w", err)
	}
	if processed.Purchase.Identity.Environment != w.environment {
		common.Logger.ErrorContext(ctx, "Google acknowledgement retry rejected environment mismatch",
			"binding_id", bindingID)
		return nil
	}
	if _, err = w.persistence.PersistGoogle(ctx, nil, processed, nil); err != nil {
		return fmt.Errorf("persist Google acknowledgement retry: %w", err)
	}
	if processed.Acknowledgement == GoogleAcknowledgementRetry {
		if age >= 24*time.Hour {
			common.Logger.ErrorContext(ctx, "Google acknowledgement retry remains pending",
				"binding_id", bindingID, "age_hours", int(age.Hours()))
		}
		return fmt.Errorf("Google acknowledgement remains retryable")
	}
	if processed.Acknowledgement == GoogleAcknowledgementFailed {
		common.Logger.ErrorContext(ctx, "Google acknowledgement permanently rejected",
			"binding_id", bindingID)
	}
	return nil
}

type GoogleAcknowledgementSweepArgs struct{}

func (GoogleAcknowledgementSweepArgs) Kind() string { return "google-store-acknowledgement-sweep" }

type GoogleAcknowledgementSweepWorker struct {
	river.WorkerDefaults[GoogleAcknowledgementSweepArgs]
	repository  GoogleAcknowledgementSweepReader
	retry       *GoogleAcknowledgementRetryWorker
	environment model.StoreEnvironment
}

func NewGoogleAcknowledgementSweepWorker(
	repository GoogleAcknowledgementSweepReader,
	retry *GoogleAcknowledgementRetryWorker,
	environment model.StoreEnvironment,
) (*GoogleAcknowledgementSweepWorker, error) {
	if repository == nil || retry == nil || !environment.Valid() {
		return nil, fmt.Errorf("Google acknowledgement sweep dependencies are required")
	}
	return &GoogleAcknowledgementSweepWorker{repository: repository, retry: retry, environment: environment}, nil
}

func (w *GoogleAcknowledgementSweepWorker) Work(
	ctx context.Context,
	_ *river.Job[GoogleAcknowledgementSweepArgs],
) error {
	candidates, err := w.repository.ListUnacknowledgedGoogleBindings(ctx, w.environment, 100)
	if err != nil {
		return fmt.Errorf("list unacknowledged Google purchases: %w", err)
	}
	failures := 0
	for _, candidate := range candidates {
		if err = w.retry.retryBinding(ctx, candidate.BindingID, candidate.CreatedAt); err != nil {
			failures++
		}
	}
	if failures > 0 {
		// Keep the sweep retryable after attempting every candidate;
		// acknowledged candidates are safe no-ops on batch redelivery.
		return fmt.Errorf("retry %d Google acknowledgement candidates", failures)
	}
	return nil
}

func (w *GoogleAcknowledgementRetryWorker) NextRetry(job *river.Job[GoogleAcknowledgementRetryArgs]) time.Time {
	delay := time.Minute
	for attempt := 1; attempt < job.Attempt && delay < 6*time.Hour; attempt++ {
		delay *= 2
	}
	if delay > 6*time.Hour {
		delay = 6 * time.Hour
	}
	base := w.now().UTC()
	if job.AttemptedAt != nil {
		base = job.AttemptedAt.UTC()
	}
	return base.Add(delay)
}
