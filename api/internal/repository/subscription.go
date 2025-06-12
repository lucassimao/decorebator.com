package repository

import (
	"context"
	"fmt"
	"time"

	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// CreateSubscription creates a new subscription record
func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, subscription *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			user_id, stripe_subscription_id, stripe_customer_id,
			plan, status, current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, trial_end,
			amount_cents, currency
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		subscription.UserID,
		subscription.StripeSubscriptionID,
		subscription.StripeCustomerID,
		subscription.Plan,
		subscription.Status,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelAtPeriodEnd,
		subscription.CancelledAt,
		subscription.TrialEnd,
		subscription.AmountCents,
		subscription.Currency,
	).Scan(&subscription.ID, &subscription.CreatedAt, &subscription.UpdatedAt)

	return err
}

// GetSubscriptionByStripeID retrieves a subscription by Stripe subscription ID
func (r *SubscriptionRepository) GetSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string) (*model.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_customer_id,
			   plan, status, current_period_start, current_period_end,
			   cancel_at_period_end, cancelled_at, trial_end,
			   amount_cents, currency, created_at, updated_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	subscription := &model.Subscription{}
	err := r.db.QueryRow(ctx, query, stripeSubscriptionID).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.StripeSubscriptionID,
		&subscription.StripeCustomerID,
		&subscription.Plan,
		&subscription.Status,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&subscription.CancelledAt,
		&subscription.TrialEnd,
		&subscription.AmountCents,
		&subscription.Currency,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetActiveSubscriptionForUser retrieves the active subscription for a user
func (r *SubscriptionRepository) GetActiveSubscriptionForUser(ctx context.Context, userID int64) (*model.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_customer_id,
			   plan, status, current_period_start, current_period_end,
			   cancel_at_period_end, cancelled_at, trial_end,
			   amount_cents, currency, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1 AND status IN ('active', 'trialing')
		ORDER BY created_at DESC
		LIMIT 1
	`

	subscription := &model.Subscription{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.StripeSubscriptionID,
		&subscription.StripeCustomerID,
		&subscription.Plan,
		&subscription.Status,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&subscription.CancelledAt,
		&subscription.TrialEnd,
		&subscription.AmountCents,
		&subscription.Currency,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// UpdateSubscription updates an existing subscription
func (r *SubscriptionRepository) UpdateSubscription(ctx context.Context, subscription *model.Subscription) error {
	query := `
		UPDATE subscriptions
		SET status = $2,
			current_period_start = $3,
			current_period_end = $4,
			cancel_at_period_end = $5,
			cancelled_at = $6,
			trial_end = $7,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		subscription.ID,
		subscription.Status,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelAtPeriodEnd,
		subscription.CancelledAt,
		subscription.TrialEnd,
	).Scan(&subscription.UpdatedAt)

	return err
}

// CreateSubscriptionEvent records a Stripe webhook event
func (r *SubscriptionRepository) CreateSubscriptionEvent(ctx context.Context, event *model.SubscriptionEvent) error {
	query := `
		INSERT INTO subscription_events (
			subscription_id, stripe_event_id, event_type, event_data
		) VALUES ($1, $2, $3, $4)
		RETURNING id, processed_at
	`

	err := r.db.QueryRow(
		ctx, query,
		event.SubscriptionID,
		event.StripeEventID,
		event.EventType,
		event.EventData,
	).Scan(&event.ID, &event.ProcessedAt)

	return err
}

// HasProcessedEvent checks if a Stripe event has already been processed
func (r *SubscriptionRepository) HasProcessedEvent(ctx context.Context, stripeEventID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM subscription_events WHERE stripe_event_id = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, stripeEventID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetUserSubscriptionHistory retrieves all subscriptions for a user
func (r *SubscriptionRepository) GetUserSubscriptionHistory(ctx context.Context, userID int64) ([]*model.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_customer_id,
			   plan, status, current_period_start, current_period_end,
			   cancel_at_period_end, cancelled_at, trial_end,
			   amount_cents, currency, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*model.Subscription
	for rows.Next() {
		subscription := &model.Subscription{}
		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.StripeSubscriptionID,
			&subscription.StripeCustomerID,
			&subscription.Plan,
			&subscription.Status,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&subscription.CancelAtPeriodEnd,
			&subscription.CancelledAt,
			&subscription.TrialEnd,
			&subscription.AmountCents,
			&subscription.Currency,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

// UpdateUserStripeCustomerID updates the user's Stripe customer ID
func (r *SubscriptionRepository) UpdateUserStripeCustomerID(ctx context.Context, userID int64, stripeCustomerID string) error {
	query := `UPDATE users SET stripe_customer_id = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, stripeCustomerID)
	return err
}

// CountUserWordlists counts the number of wordlists for a user
func (r *SubscriptionRepository) CountUserWordlists(ctx context.Context, userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM wordlists WHERE user_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count wordlists: %w", err)
	}

	return count, nil
}

// CountWordsInWordlist counts the number of words in a specific wordlist
func (r *SubscriptionRepository) CountWordsInWordlist(ctx context.Context, wordlistID int64) (int, error) {
	query := `SELECT COUNT(*) FROM words WHERE wordlist_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, wordlistID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count words: %w", err)
	}

	return count, nil
}

// GetSubscriptionByID retrieves a subscription by ID
func (r *SubscriptionRepository) GetSubscriptionByID(ctx context.Context, id int64) (*model.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_customer_id,
			   plan, status, current_period_start, current_period_end,
			   cancel_at_period_end, cancelled_at, trial_end,
			   amount_cents, currency, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	subscription := &model.Subscription{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.StripeSubscriptionID,
		&subscription.StripeCustomerID,
		&subscription.Plan,
		&subscription.Status,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&subscription.CancelledAt,
		&subscription.TrialEnd,
		&subscription.AmountCents,
		&subscription.Currency,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetSubscriptionsRenewingBetween retrieves subscriptions with renewal dates between start and end
func (r *SubscriptionRepository) GetSubscriptionsRenewingBetween(ctx context.Context, start, end time.Time) ([]*model.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_customer_id,
			   plan, status, current_period_start, current_period_end,
			   cancel_at_period_end, cancelled_at, trial_end,
			   amount_cents, currency, created_at, updated_at
		FROM subscriptions
		WHERE status = 'active' 
		  AND cancel_at_period_end = false
		  AND current_period_end >= $1 
		  AND current_period_end <= $2
		ORDER BY current_period_end ASC
	`

	rows, err := r.db.Query(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*model.Subscription
	for rows.Next() {
		subscription := &model.Subscription{}
		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.StripeSubscriptionID,
			&subscription.StripeCustomerID,
			&subscription.Plan,
			&subscription.Status,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&subscription.CancelAtPeriodEnd,
			&subscription.CancelledAt,
			&subscription.TrialEnd,
			&subscription.AmountCents,
			&subscription.Currency,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

// HasSentRenewalReminder checks if a renewal reminder has been sent for a subscription
func (r *SubscriptionRepository) HasSentRenewalReminder(ctx context.Context, subscriptionID int64, reminderWindow time.Time) (bool, error) {
	// Check in subscription_events for a renewal_reminder event within the last 7 days
	query := `
		SELECT EXISTS(
			SELECT 1 FROM subscription_events 
			WHERE subscription_id = $1 
			  AND event_type = 'renewal_reminder_sent'
			  AND processed_at >= $2
		)
	`

	var exists bool
	sevenDaysAgo := reminderWindow.AddDate(0, 0, -7)
	err := r.db.QueryRow(ctx, query, subscriptionID, sevenDaysAgo).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// MarkRenewalReminderSent records that a renewal reminder was sent
func (r *SubscriptionRepository) MarkRenewalReminderSent(ctx context.Context, subscriptionID int64) error {
	query := `
		INSERT INTO subscription_events (
			subscription_id, stripe_event_id, event_type, event_data
		) VALUES ($1, $2, 'renewal_reminder_sent', '{}')
	`

	// Generate a unique event ID for this internal event
	eventID := fmt.Sprintf("reminder_%d_%d", subscriptionID, time.Now().Unix())
	_, err := r.db.Exec(ctx, query, subscriptionID, eventID)
	return err
}
