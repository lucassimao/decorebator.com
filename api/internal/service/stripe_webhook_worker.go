package service

import (
	"context"
	"fmt"

	"decorebator.com/internal/common"
	"github.com/getsentry/sentry-go"
	"github.com/riverqueue/river"
	"github.com/stripe/stripe-go/v82"
)

type StripeWebhookArgs struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	EventData []byte `json:"event_data"`
}

func (StripeWebhookArgs) Kind() string { return "stripe-webhook" }

type StripeWebhookWorker struct {
	river.WorkerDefaults[StripeWebhookArgs]
	subscriptionService *SubscriptionService
}

func NewStripeWebhookWorker(subscriptionService *SubscriptionService) *StripeWebhookWorker {
	return &StripeWebhookWorker{
		subscriptionService: subscriptionService,
	}
}

func (w *StripeWebhookWorker) Work(ctx context.Context, job *river.Job[StripeWebhookArgs]) error {
	logger := common.Logger.With("worker", "stripe-webhook", "event_id", job.Args.EventID)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("stripe_event_id", job.Args.EventID)
			scope.SetTag("stripe_event_type", job.Args.EventType)
			scope.SetExtra("stripe_event_payload", string(job.Args.EventData))
		})
	}

	// Reconstruct the Stripe event from the job args
	event := stripe.Event{
		ID:   job.Args.EventID,
		Type: stripe.EventType(job.Args.EventType),
		Data: &stripe.EventData{
			Raw: job.Args.EventData,
		},
	}

	// Check if we've already processed this event
	processed, err := w.subscriptionService.subRepo.HasProcessedEvent(ctx, event.ID)
	if err != nil {
		logger.Error("failed to check event processing status", "error", err)
		return err
	}
	if processed {
		logger.Info("Event already processed, skipping")
		return river.JobCancel(fmt.Errorf("%s event already processed", event.ID))
	}

	// Handle different event types using existing service methods
	switch event.Type {
	case "customer.subscription.created":
		err = w.subscriptionService.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.updated":
		err = w.subscriptionService.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		err = w.subscriptionService.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_failed":
		err = w.subscriptionService.handleInvoicePaymentFailed(ctx, event)
	default:
		// Log unhandled event types
		logger.Warn("Unhandled webhook event type", "event_type", event.Type)
		return nil
	}

	if err != nil {
		logger.Error("failed to process stripe webhook event", "error", err)
		return err
	}

	logger.Info("Stripe webhook processed successfully")
	return nil
}
