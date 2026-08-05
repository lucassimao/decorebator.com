package unit

import (
	"strings"
	"testing"

	"decorebator.com/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderIdempotencyKeysAreDeterministicAndDomainSeparated(t *testing.T) {
	environment := model.StoreEnvironmentSandbox
	providerID := "2000000123456789"

	transactionKey, err := model.AppleTransactionRecordKey(environment, providerID)
	require.NoError(t, err)
	repeatedTransactionKey, err := model.AppleTransactionRecordKey(environment, providerID)
	require.NoError(t, err)
	notificationKey, err := model.AppleNotificationIdempotencyKey(environment, uuid.MustParse("3ee168c7-75a9-49a7-8879-d97ca09f0263"))
	require.NoError(t, err)
	productionKey, err := model.AppleTransactionRecordKey(model.StoreEnvironmentProduction, providerID)
	require.NoError(t, err)

	assert.Equal(t, transactionKey, repeatedTransactionKey)
	assert.NotEqual(t, transactionKey, notificationKey)
	assert.NotEqual(t, transactionKey, productionKey)
	assert.NotContains(t, transactionKey.String(), providerID)
}

func TestGoogleIdempotencyKeysNeverExposePurchaseTokens(t *testing.T) {
	token := "sensitive-google-purchase-token"
	purchaseKey, err := model.GooglePurchaseTokenRecordKey(model.StoreEnvironmentSandbox, token)
	require.NoError(t, err)
	transportKey, err := model.GoogleRTDNTransportIdempotencyKey(
		"projects/decorebator/topics/play-rtdn",
		"pubsub-message-123",
	)
	require.NoError(t, err)
	semanticKey, err := model.GoogleRTDNSemanticIdempotencyKey(model.GoogleRTDNSemanticEvent{
		PackageName:      "com.decorebator.app",
		PurchaseToken:    token,
		NotificationKind: "subscription",
		NotificationType: 4,
		EventTimeMillis:  1770076800000,
	})
	require.NoError(t, err)

	for _, key := range []string{purchaseKey.String(), transportKey.String(), semanticKey.String()} {
		assert.NotContains(t, key, token)
		assert.LessOrEqual(t, len(key), model.ProviderKeyMaxLength)
	}
	assert.NotEqual(t, purchaseKey, transportKey)
	assert.NotEqual(t, purchaseKey, semanticKey)
}

func TestGoogleSemanticKeyChangesForDistinctBusinessEvents(t *testing.T) {
	base := model.GoogleRTDNSemanticEvent{
		PackageName:      "com.decorebator.app",
		PurchaseToken:    "purchase-token",
		NotificationKind: "subscription",
		NotificationType: 2,
		EventTimeMillis:  1770076800000,
	}
	baseKey, err := model.GoogleRTDNSemanticIdempotencyKey(base)
	require.NoError(t, err)

	mutations := []func(*model.GoogleRTDNSemanticEvent){
		func(event *model.GoogleRTDNSemanticEvent) { event.PurchaseToken = "replacement-token" },
		func(event *model.GoogleRTDNSemanticEvent) { event.NotificationKind = "voided_subscription" },
		func(event *model.GoogleRTDNSemanticEvent) { event.NotificationType++ },
		func(event *model.GoogleRTDNSemanticEvent) { event.EventTimeMillis++ },
		func(event *model.GoogleRTDNSemanticEvent) { event.PackageName = "com.decorebator.other" },
	}

	for _, mutate := range mutations {
		event := base
		mutate(&event)
		key, keyErr := model.GoogleRTDNSemanticIdempotencyKey(event)
		require.NoError(t, keyErr)
		assert.NotEqual(t, baseKey, key)
	}
}

func TestIdempotencyKeyEncodingPreservesComponentBoundaries(t *testing.T) {
	left, err := model.GoogleRTDNTransportIdempotencyKey("ab", "c")
	require.NoError(t, err)
	right, err := model.GoogleRTDNTransportIdempotencyKey("a", "bc")
	require.NoError(t, err)

	assert.NotEqual(t, left, right)
}

func TestProviderIdempotencyKeysRejectMalformedInputs(t *testing.T) {
	_, err := model.AppleTransactionRecordKey("staging", "transaction")
	require.ErrorContains(t, err, "environment")

	_, err = model.AppleTransactionRecordKey(model.StoreEnvironmentSandbox, " transaction ")
	require.ErrorContains(t, err, "whitespace")

	_, err = model.AppleNotificationIdempotencyKey(model.StoreEnvironmentSandbox, uuid.Nil)
	require.ErrorContains(t, err, "notification")

	_, err = model.GooglePurchaseTokenRecordKey(model.StoreEnvironmentProduction, "")
	require.ErrorContains(t, err, "purchase token")

	_, err = model.GoogleRTDNTransportIdempotencyKey("", "message")
	require.ErrorContains(t, err, "topic")
	_, err = model.GoogleRTDNTransportIdempotencyKey("topic", "")
	require.ErrorContains(t, err, "message ID")

	_, err = model.GoogleRTDNSemanticIdempotencyKey(model.GoogleRTDNSemanticEvent{
		PackageName:      "",
		PurchaseToken:    "purchase-token",
		NotificationType: 2,
		EventTimeMillis:  1770076800000,
	})
	require.ErrorContains(t, err, "package name")

	_, err = model.GoogleRTDNSemanticIdempotencyKey(model.GoogleRTDNSemanticEvent{
		PackageName:      "com.decorebator.app",
		PurchaseToken:    "",
		NotificationType: 2,
		EventTimeMillis:  1770076800000,
	})
	require.ErrorContains(t, err, "purchase token")

	_, err = model.GoogleRTDNSemanticIdempotencyKey(model.GoogleRTDNSemanticEvent{
		PackageName:      "com.decorebator.app",
		PurchaseToken:    "purchase-token",
		NotificationKind: "subscription",
		NotificationType: 0,
		EventTimeMillis:  1770076800000,
	})
	require.ErrorContains(t, err, "notification type")

	_, err = model.GoogleRTDNSemanticIdempotencyKey(model.GoogleRTDNSemanticEvent{
		PackageName:      "com.decorebator.app",
		PurchaseToken:    strings.Repeat("x", 16),
		NotificationKind: "subscription",
		NotificationType: 2,
		EventTimeMillis:  0,
	})
	require.ErrorContains(t, err, "event time")
}
