package model

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

const ProviderKeyMaxLength = 128

const (
	appleTransactionNamespace  = "apple_transaction"
	appleNotificationNamespace = "apple_notification"
	googlePurchaseNamespace    = "google_purchase"
	googleTransportNamespace   = "google_rtdn_transport"
	googleSemanticNamespace    = "google_rtdn_semantic"
)

// ProviderRecordKey identifies a mutable purchase/transaction record. It must
// never be used to deduplicate provider event deliveries.
type ProviderRecordKey string

func (k ProviderRecordKey) String() string {
	return string(k)
}

// ProviderEventKey identifies a provider notification or semantic event. Raw
// provider identifiers and purchase tokens never appear in this safe digest.
type ProviderEventKey string

func (k ProviderEventKey) String() string {
	return string(k)
}

func AppleTransactionRecordKey(
	environment StoreEnvironment,
	transactionID string,
) (ProviderRecordKey, error) {
	if err := validateOpaqueStoreIdentifier("apple transaction ID", transactionID, 0); err != nil {
		return "", err
	}
	key, err := newProviderKey(appleTransactionNamespace, string(environment), environment.Valid(), transactionID)
	return ProviderRecordKey(key), err
}

func AppleNotificationIdempotencyKey(
	environment StoreEnvironment,
	notificationUUID uuid.UUID,
) (ProviderEventKey, error) {
	if notificationUUID == uuid.Nil {
		return "", fmt.Errorf("apple notification UUID is required")
	}
	key, err := newProviderKey(appleNotificationNamespace, string(environment), environment.Valid(), notificationUUID.String())
	return ProviderEventKey(key), err
}

// GooglePurchaseTokenRecordKey is the digest used by the unique purchase
// binding constraint. The protected raw token remains available for Play API
// refresh calls but is never part of this safe key.
func GooglePurchaseTokenRecordKey(
	environment StoreEnvironment,
	purchaseToken string,
) (ProviderRecordKey, error) {
	if err := validateOpaqueStoreIdentifier("google purchase token", purchaseToken, 0); err != nil {
		return "", err
	}
	key, err := newProviderKey(googlePurchaseNamespace, string(environment), environment.Valid(), purchaseToken)
	return ProviderRecordKey(key), err
}

// GoogleRTDNTransportIdempotencyKey deduplicates Pub/Sub redelivery of the same
// message. RTDN has no environment before the Play API refresh, so topic scope
// and message ID are the complete pre-verification transport identity.
func GoogleRTDNTransportIdempotencyKey(
	topic string,
	messageID string,
) (ProviderEventKey, error) {
	if err := validateOpaqueStoreIdentifier("google Pub/Sub topic", topic, 0); err != nil {
		return "", err
	}
	if err := validateOpaqueStoreIdentifier("google Pub/Sub message ID", messageID, 0); err != nil {
		return "", err
	}
	key, err := newProviderKey(googleTransportNamespace, "", true, topic, messageID)
	return ProviderEventKey(key), err
}

// GoogleRTDNSemanticEvent contains the provider fields that identify one RTDN
// business event independently of its Pub/Sub delivery envelope. Environment
// is deliberately absent because it is known only after provider refresh.
type GoogleRTDNSemanticEvent struct {
	PackageName      string
	PurchaseToken    string
	NotificationKind string
	NotificationType int32
	EventTimeMillis  int64
}

func (e GoogleRTDNSemanticEvent) Validate() error {
	if err := validateOpaqueStoreIdentifier("google package name", e.PackageName, 0); err != nil {
		return err
	}
	if err := validateOpaqueStoreIdentifier("google purchase token", e.PurchaseToken, 0); err != nil {
		return err
	}
	if e.NotificationKind != "subscription" && e.NotificationKind != "voided_subscription" {
		return fmt.Errorf("unsupported google notification kind")
	}
	if e.NotificationType <= 0 {
		return fmt.Errorf("google notification type must be positive")
	}
	if e.EventTimeMillis <= 0 {
		return fmt.Errorf("google event time must be positive")
	}

	return nil
}

func GoogleRTDNSemanticIdempotencyKey(
	event GoogleRTDNSemanticEvent,
) (ProviderEventKey, error) {
	if err := event.Validate(); err != nil {
		return "", err
	}
	key, err := newProviderKey(
		googleSemanticNamespace,
		"",
		true,
		event.PackageName,
		event.PurchaseToken,
		event.NotificationKind,
		fmt.Sprintf("%d", event.NotificationType),
		fmt.Sprintf("%d", event.EventTimeMillis),
	)
	return ProviderEventKey(key), err
}

func newProviderKey(namespace, scope string, validScope bool, parts ...string) (string, error) {
	if !validScope {
		return "", fmt.Errorf("unsupported store environment %q", scope)
	}

	hasher := sha256.New()
	for _, part := range parts {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(part)))
		_, _ = hasher.Write(length[:])
		_, _ = hasher.Write([]byte(part))
	}

	digest := hex.EncodeToString(hasher.Sum(nil))
	if scope == "" {
		return namespace + ":" + digest, nil
	}

	return namespace + ":" + scope + ":" + digest, nil
}
