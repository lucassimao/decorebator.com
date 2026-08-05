package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionRepositoryReturnsDeferredCommitFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setup.CreateTestDB(t)
	defer db.Close()
	ctx := context.Background()
	require.NoError(t, setup.RunMigrations(db))
	require.NoError(t, setup.CleanTestData(db))

	_, err := db.Exec(ctx, `
		CREATE OR REPLACE FUNCTION test_fail_subscription_event_commit()
		RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'forced deferred subscription event failure';
		END;
		$$ LANGUAGE plpgsql;
		CREATE CONSTRAINT TRIGGER test_fail_subscription_event_commit_trigger
		AFTER INSERT ON subscription_events
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW
		WHEN (NEW.event_type LIKE 'force_commit_failure_%')
		EXECUTE FUNCTION test_fail_subscription_event_commit();
	`)
	require.NoError(t, err)
	defer func() {
		_, cleanupErr := db.Exec(ctx, `
			DROP TRIGGER IF EXISTS test_fail_subscription_event_commit_trigger ON subscription_events;
			DROP FUNCTION IF EXISTS test_fail_subscription_event_commit();
		`)
		require.NoError(t, cleanupErr)
	}()

	var userID int64
	require.NoError(t, db.QueryRow(ctx, `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES ('Commit', 'Test', $1, 'not-a-real-password') RETURNING id
	`, fmt.Sprintf("subscription-commit-%d@example.com", time.Now().UnixNano())).Scan(&userID))

	repo := repository.NewSubscriptionRepository(db)
	failedCreate := testSubscription(userID, "commit-failure-create")
	failedCreateID, err := repo.CreateSubscription(ctx, failedCreate, model.SubscriptionEvent{
		ExternalEventID: uniqueCommitEventID("failed_create"),
		Provider:        model.ProviderRevenueCat,
		EventType:       "force_commit_failure_create",
		EventData:       `{}`,
	})
	require.ErrorContains(t, err, "failed to commit transaction")
	assert.Zero(t, failedCreateID)
	var createCount int
	require.NoError(t, db.QueryRow(ctx,
		"SELECT count(*) FROM subscriptions WHERE revenuecat_subscription_id = $1",
		*failedCreate.RevenueCatSubscriptionID,
	).Scan(&createCount))
	assert.Zero(t, createCount)

	subscription := testSubscription(userID, "commit-failure-update")
	createdID, err := repo.CreateSubscription(ctx, subscription, model.SubscriptionEvent{
		ExternalEventID: uniqueCommitEventID("setup"),
		Provider:        model.ProviderRevenueCat,
		EventType:       "test_setup",
		EventData:       `{}`,
	})
	require.NoError(t, err)
	subscription.ID = createdID
	originalEnd := subscription.CurrentPeriodEnd
	subscription.CurrentPeriodEnd = originalEnd.Add(30 * 24 * time.Hour)

	err = repo.UpdateSubscription(ctx, subscription, model.SubscriptionEvent{
		ExternalEventID: uniqueCommitEventID("failed_update"),
		Provider:        model.ProviderRevenueCat,
		EventType:       "force_commit_failure_update",
		EventData:       `{}`,
	})
	require.ErrorContains(t, err, "failed to commit transaction")
	var persistedEnd time.Time
	require.NoError(t, db.QueryRow(ctx,
		"SELECT current_period_end FROM subscriptions WHERE id = $1", createdID,
	).Scan(&persistedEnd))
	assert.WithinDuration(t, originalEnd, persistedEnd, time.Microsecond)
}

func uniqueCommitEventID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func testSubscription(userID int64, suffix string) *model.Subscription {
	providerID := "rc_" + suffix
	productID := "com.decorebator.premium.monthly"
	platform := model.PlatformIOS
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &model.Subscription{
		UserID: userID, Provider: model.ProviderRevenueCat,
		RevenueCatSubscriptionID: &providerID, AppStoreProductID: &productID, Platform: &platform,
		Plan: model.PlanMonthly, Status: model.StatusActive,
		CurrentPeriodStart: now, CurrentPeriodEnd: now.Add(30 * 24 * time.Hour),
		AmountCents: 699, Currency: "USD",
	}
}
