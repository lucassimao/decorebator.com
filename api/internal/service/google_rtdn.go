package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"google.golang.org/api/idtoken"
)

const (
	googleRTDNMaxPushBytes  = 64 * 1024
	googleRTDNLease         = 30 * time.Second
	googleRTDNRetryDelay    = time.Minute
	googleRTDNFutureSkew    = 5 * time.Minute
	googleRTDNAuthTimeout   = 15 * time.Second
	googleRTDNRefreshLimit  = 20 * time.Second
	googleRTDNVersion       = "1.0"
	googleRTDNTestEventKind = "test"
)

var (
	ErrInvalidGooglePubSubAuthentication = errors.New("invalid Google Pub/Sub authentication")
	ErrInvalidGoogleRTDNEnvelope         = errors.New("invalid Google RTDN envelope")
)

type GoogleRTDNDeliveryDisposition string

const (
	GoogleRTDNAcknowledge  GoogleRTDNDeliveryDisposition = "acknowledge"
	GoogleRTDNRetry        GoogleRTDNDeliveryDisposition = "retry"
	GoogleRTDNUnauthorized GoogleRTDNDeliveryDisposition = "unauthorized"
)

type GoogleRTDNDeliveryError struct {
	Disposition GoogleRTDNDeliveryDisposition
	Cause       error
}

func (e *GoogleRTDNDeliveryError) Error() string {
	return "Google RTDN delivery failed: " + string(e.Disposition)
}

func (e *GoogleRTDNDeliveryError) Unwrap() error { return e.Cause }

// GoogleRTDNDispositionOf maps Ingest errors to the push response decision.
// A nil error is terminal and acknowledged; unknown errors fail safe to retry.
func GoogleRTDNDispositionOf(err error) GoogleRTDNDeliveryDisposition {
	if err == nil {
		return GoogleRTDNAcknowledge
	}
	var deliveryError *GoogleRTDNDeliveryError
	if errors.As(err, &deliveryError) {
		return deliveryError.Disposition
	}
	return GoogleRTDNRetry
}

type GooglePubSubClaims struct {
	Issuer        string
	Email         string
	EmailVerified bool
}

type GooglePubSubTokenVerifier interface {
	Verify(context.Context, string, string) (GooglePubSubClaims, error)
}

type GooglePubSubTokenError struct {
	Retryable bool
	Cause     error
}

func (e *GooglePubSubTokenError) Error() string { return "Google Pub/Sub token verification failed" }
func (e *GooglePubSubTokenError) Unwrap() error { return e.Cause }

// GoogleOIDCTokenVerifier validates the Google signature, audience, and expiry.
// GoogleRTDNIngestor additionally binds the verified issuer and service-account
// claims to the configured Pub/Sub push subscription.
type GoogleOIDCTokenVerifier struct{}

func (GoogleOIDCTokenVerifier) Verify(
	ctx context.Context,
	token string,
	audience string,
) (GooglePubSubClaims, error) {
	validationContext, cancel := context.WithTimeout(ctx, googleRTDNAuthTimeout)
	defer cancel()
	payload, err := idtoken.Validate(validationContext, token, audience)
	if err != nil {
		return GooglePubSubClaims{}, &GooglePubSubTokenError{Retryable: true, Cause: err}
	}
	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	return GooglePubSubClaims{Issuer: payload.Issuer, Email: email, EmailVerified: emailVerified}, nil
}

type GoogleRTDNInbox interface {
	Claim(context.Context, repository.ProviderEventInboxInput, time.Time, time.Duration) (repository.ProviderEventInboxResult, error)
	Complete(context.Context, repository.ProviderEventClaim, repository.ProviderEventProcessor) (repository.ProviderEventInboxResult, error)
	MarkRetryable(context.Context, repository.ProviderEventClaim, time.Time) error
}

type GoogleRTDNEvent struct {
	Kind             string
	PackageName      string
	PurchaseToken    string `json:"-"`
	NotificationType int32
	RefundType       int32
	OccurredAt       time.Time
}

func (e GoogleRTDNEvent) String() string {
	return fmt.Sprintf("GoogleRTDNEvent{Kind:%s PackageName:%s NotificationType:%d RefundType:%d OccurredAt:%s}",
		e.Kind, e.PackageName, e.NotificationType, e.RefundType, e.OccurredAt.UTC().Format(time.RFC3339Nano))
}

// GoogleRTDNRefresher performs the purchase-token lookup, Play API refresh,
// and account binding outside the inbox transaction. The returned processor
// may only apply the already-verified local database mutation.
type GoogleRTDNRefresher func(context.Context, GoogleRTDNEvent) (repository.ProviderEventProcessor, error)

type GoogleRTDNIngestor struct {
	tokens              GooglePubSubTokenVerifier
	inbox               GoogleRTDNInbox
	refresher           GoogleRTDNRefresher
	audience            string
	serviceAccountEmail string
	topic               string
	subscription        string
	packageName         string
	now                 func() time.Time
}

func NewGoogleRTDNIngestor(
	tokens GooglePubSubTokenVerifier,
	inbox GoogleRTDNInbox,
	refresher GoogleRTDNRefresher,
	audience string,
	serviceAccountEmail string,
	topic string,
	subscription string,
	packageName string,
	now func() time.Time,
) (*GoogleRTDNIngestor, error) {
	if tokens == nil || inbox == nil || refresher == nil ||
		strings.TrimSpace(audience) == "" || strings.TrimSpace(serviceAccountEmail) == "" ||
		!validGoogleResource(topic, "topics") || !validGoogleResource(subscription, "subscriptions") ||
		strings.TrimSpace(packageName) == "" {
		return nil, fmt.Errorf("Google RTDN configuration is incomplete")
	}
	if now == nil {
		now = time.Now
	}
	return &GoogleRTDNIngestor{
		tokens: tokens, inbox: inbox, refresher: refresher, audience: audience,
		serviceAccountEmail: serviceAccountEmail, topic: topic, subscription: subscription,
		packageName: packageName, now: now,
	}, nil
}

func (i *GoogleRTDNIngestor) Ingest(
	ctx context.Context,
	authorization string,
	body []byte,
) (repository.ProviderEventInboxResult, error) {
	if err := i.authenticate(ctx, authorization); err != nil {
		disposition := GoogleRTDNUnauthorized
		var tokenError *GooglePubSubTokenError
		if errors.As(err, &tokenError) && tokenError.Retryable {
			disposition = GoogleRTDNRetry
		}
		return repository.ProviderEventInboxResult{}, googleRTDNDeliveryError(disposition, err)
	}
	if len(body) == 0 || len(body) > googleRTDNMaxPushBytes {
		return repository.ProviderEventInboxResult{}, googleRTDNDeliveryError(GoogleRTDNAcknowledge, ErrInvalidGoogleRTDNEnvelope)
	}
	envelope, err := i.decodeEnvelope(body)
	if err != nil {
		return repository.ProviderEventInboxResult{}, googleRTDNDeliveryError(GoogleRTDNAcknowledge, err)
	}
	transportKey, err := model.GoogleRTDNTransportIdempotencyKey(i.topic, envelope.Message.MessageID)
	if err != nil {
		return repository.ProviderEventInboxResult{}, googleRTDNDeliveryError(GoogleRTDNAcknowledge, ErrInvalidGoogleRTDNEnvelope)
	}

	now := i.now().UTC()
	event, semanticKey, rejection := i.decodeDeveloperNotification(envelope.Message.Data, now)
	input := repository.ProviderEventInboxInput{
		Store: model.EntitlementStoreGoogle, IdempotencyKey: transportKey,
		ProviderEventType: "invalid", ProviderOccurredAt: now,
	}
	if rejection == nil {
		input.ProviderEventType = event.Kind
		input.ProviderOccurredAt = event.OccurredAt
		input.SemanticKey = semanticKey
	}
	claimResult, err := i.inbox.Claim(ctx, input, now, googleRTDNLease)
	if err != nil {
		return claimResult, googleRTDNDeliveryError(GoogleRTDNRetry, err)
	}
	if claimResult.Duplicate {
		if claimResult.Outcome == model.EntitlementOutcomeRetry {
			return claimResult, googleRTDNDeliveryError(GoogleRTDNRetry, errors.New("Google RTDN is already processing or waiting for retry"))
		}
		return claimResult, nil
	}
	if claimResult.Claim == nil {
		return repository.ProviderEventInboxResult{}, googleRTDNDeliveryError(GoogleRTDNRetry, errors.New("Google RTDN claim is missing"))
	}
	claim := *claimResult.Claim
	if rejection != nil {
		return i.complete(ctx, claim, rejectedGoogleRTDN)
	}
	if event.Kind == googleRTDNTestEventKind {
		return i.complete(ctx, claim, acceptedGoogleRTDNTest)
	}
	refreshContext, cancel := context.WithTimeout(ctx, googleRTDNRefreshLimit)
	defer cancel()
	processor, err := i.refresher(refreshContext, event)
	if err != nil {
		if processorForFailure, permanent := permanentGoogleRTDNRefreshFailure(err); permanent {
			return i.complete(ctx, claim, processorForFailure)
		}
		if retryErr := i.inbox.MarkRetryable(ctx, claim, i.now().UTC().Add(googleRTDNRetryDelay)); retryErr != nil {
			return repository.ProviderEventInboxResult{}, googleRTDNDeliveryError(GoogleRTDNRetry, retryErr)
		}
		return repository.ProviderEventInboxResult{
			Outcome: model.EntitlementOutcomeRetry, Code: model.EntitlementResultRetryableProvider,
		}, googleRTDNDeliveryError(GoogleRTDNRetry, err)
	}
	return i.complete(ctx, claim, processor)
}

func (i *GoogleRTDNIngestor) complete(
	ctx context.Context,
	claim repository.ProviderEventClaim,
	processor repository.ProviderEventProcessor,
) (repository.ProviderEventInboxResult, error) {
	result, err := i.inbox.Complete(ctx, claim, processor)
	if err != nil {
		return result, googleRTDNDeliveryError(GoogleRTDNRetry, err)
	}
	return result, nil
}

func (i *GoogleRTDNIngestor) authenticate(ctx context.Context, authorization string) error {
	fields := strings.Fields(authorization)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") || fields[1] == "" {
		return ErrInvalidGooglePubSubAuthentication
	}
	claims, err := i.tokens.Verify(ctx, fields[1], i.audience)
	if err != nil {
		return err
	}
	if (claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com") ||
		claims.Email != i.serviceAccountEmail || !claims.EmailVerified {
		return ErrInvalidGooglePubSubAuthentication
	}
	return nil
}

type googlePubSubPushEnvelope struct {
	Message struct {
		Data      string `json:"data"`
		MessageID string `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func (i *GoogleRTDNIngestor) decodeEnvelope(body []byte) (googlePubSubPushEnvelope, error) {
	var envelope googlePubSubPushEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil || envelope.Subscription != i.subscription ||
		envelope.Message.MessageID == "" || envelope.Message.Data == "" {
		return googlePubSubPushEnvelope{}, ErrInvalidGoogleRTDNEnvelope
	}
	return envelope, nil
}

type googleDeveloperNotification struct {
	Version                         string                    `json:"version"`
	PackageName                     string                    `json:"packageName"`
	EventTimeMillis                 json.RawMessage           `json:"eventTimeMillis"`
	SubscriptionNotification        *googleSubscriptionRTDN   `json:"subscriptionNotification"`
	VoidedPurchaseNotification      *googleVoidedPurchaseRTDN `json:"voidedPurchaseNotification"`
	TestNotification                *googleTestRTDN           `json:"testNotification"`
	OneTimeProductNotification      json.RawMessage           `json:"oneTimeProductNotification"`
	PendingRefundReviewNotification json.RawMessage           `json:"pendingRefundReviewNotification"`
}

type googleSubscriptionRTDN struct {
	Version          string `json:"version"`
	NotificationType int32  `json:"notificationType"`
	PurchaseToken    string `json:"purchaseToken"`
}

type googleVoidedPurchaseRTDN struct {
	PurchaseToken string `json:"purchaseToken"`
	ProductType   int32  `json:"productType"`
	RefundType    int32  `json:"refundType"`
}

type googleTestRTDN struct {
	Version string `json:"version"`
}

func (i *GoogleRTDNIngestor) decodeDeveloperNotification(
	encoded string,
	now time.Time,
) (GoogleRTDNEvent, *model.ProviderEventKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(decoded) == 0 || len(decoded) > googleRTDNMaxPushBytes {
		return GoogleRTDNEvent{}, nil, ErrInvalidGoogleRTDNEnvelope
	}
	var notification googleDeveloperNotification
	if err = json.Unmarshal(decoded, &notification); err != nil {
		return GoogleRTDNEvent{}, nil, ErrInvalidGoogleRTDNEnvelope
	}
	occurredAt, err := googleRTDNOccurredAt(notification.EventTimeMillis)
	if err != nil || notification.Version != googleRTDNVersion || notification.PackageName != i.packageName ||
		occurredAt.Before(time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)) || occurredAt.After(now.Add(googleRTDNFutureSkew)) {
		return GoogleRTDNEvent{}, nil, ErrInvalidGoogleRTDNEnvelope
	}
	event, err := googleRTDNEventFromNotification(notification, occurredAt)
	if err != nil || event.Kind == googleRTDNTestEventKind {
		return event, nil, err
	}
	semantic, err := model.GoogleRTDNSemanticIdempotencyKey(model.GoogleRTDNSemanticEvent{
		PackageName: event.PackageName, PurchaseToken: event.PurchaseToken, NotificationKind: event.Kind,
		NotificationType: event.NotificationType, EventTimeMillis: event.OccurredAt.UnixMilli(),
	})
	if err != nil {
		return GoogleRTDNEvent{}, nil, ErrInvalidGoogleRTDNEnvelope
	}
	return event, &semantic, nil
}

func googleRTDNEventFromNotification(
	notification googleDeveloperNotification,
	occurredAt time.Time,
) (GoogleRTDNEvent, error) {
	typeCount := boolCount(
		notification.SubscriptionNotification != nil,
		notification.VoidedPurchaseNotification != nil,
		notification.TestNotification != nil,
		rawGoogleRTDNPresent(notification.OneTimeProductNotification),
		rawGoogleRTDNPresent(notification.PendingRefundReviewNotification),
	)
	if typeCount != 1 {
		return GoogleRTDNEvent{}, ErrInvalidGoogleRTDNEnvelope
	}
	if notification.TestNotification != nil {
		if notification.TestNotification.Version != googleRTDNVersion {
			return GoogleRTDNEvent{}, ErrInvalidGoogleRTDNEnvelope
		}
		return GoogleRTDNEvent{Kind: googleRTDNTestEventKind, PackageName: notification.PackageName, OccurredAt: occurredAt}, nil
	}
	if rawGoogleRTDNPresent(notification.OneTimeProductNotification) ||
		rawGoogleRTDNPresent(notification.PendingRefundReviewNotification) {
		return GoogleRTDNEvent{}, ErrInvalidGoogleRTDNEnvelope
	}

	event := GoogleRTDNEvent{PackageName: notification.PackageName, OccurredAt: occurredAt}
	if subscription := notification.SubscriptionNotification; subscription != nil {
		if subscription.Version != googleRTDNVersion || subscription.NotificationType <= 0 ||
			!validGooglePurchaseToken(subscription.PurchaseToken) {
			return GoogleRTDNEvent{}, ErrInvalidGoogleRTDNEnvelope
		}
		event.Kind, event.NotificationType, event.PurchaseToken = "subscription", subscription.NotificationType, subscription.PurchaseToken
	} else {
		voided := notification.VoidedPurchaseNotification
		if voided.ProductType == 2 || voided.ProductType < 0 || voided.RefundType < 0 ||
			!validGooglePurchaseToken(voided.PurchaseToken) {
			return GoogleRTDNEvent{}, ErrInvalidGoogleRTDNEnvelope
		}
		event.Kind, event.NotificationType, event.RefundType, event.PurchaseToken =
			"voided_subscription", 1, voided.RefundType, voided.PurchaseToken
	}
	return event, nil
}

func googleRTDNOccurredAt(raw json.RawMessage) (time.Time, error) {
	if len(raw) == 0 {
		return time.Time{}, ErrInvalidGoogleRTDNEnvelope
	}
	value := strings.Trim(string(raw), `"`)
	millis, err := strconv.ParseInt(value, 10, 64)
	if err != nil || millis <= 0 {
		return time.Time{}, ErrInvalidGoogleRTDNEnvelope
	}
	return time.UnixMilli(millis).UTC(), nil
}

func validGoogleResource(value string, resourceType string) bool {
	parts := strings.Split(value, "/")
	return len(parts) == 4 && parts[0] == "projects" && parts[1] != "" &&
		parts[2] == resourceType && parts[3] != "" && strings.TrimSpace(value) == value
}

func googleRTDNDeliveryError(disposition GoogleRTDNDeliveryDisposition, cause error) error {
	return &GoogleRTDNDeliveryError{Disposition: disposition, Cause: cause}
}

func permanentGoogleRTDNRefreshFailure(err error) (repository.ProviderEventProcessor, bool) {
	var verificationError *GooglePurchaseVerificationError
	if errors.As(err, &verificationError) {
		switch verificationError.Failure {
		case model.EntitlementFailureInvalidEvidence:
			return rejectedGoogleRTDN, true
		case model.EntitlementFailureAccountMismatch:
			return fixedGoogleRTDNResult(model.EntitlementResultAccountMismatch), true
		case model.EntitlementFailureUnknownProduct:
			return fixedGoogleRTDNResult(model.EntitlementResultUnknownProduct), true
		case model.EntitlementFailureProviderUnavailable:
			return nil, false
		}
	}
	var apiError *GooglePlayAPIError
	if errors.As(err, &apiError) && !apiError.Retryable && apiError.StatusCode != 0 &&
		apiError.StatusCode != 401 && apiError.StatusCode != 403 {
		return rejectedGoogleRTDN, true
	}
	return nil, false
}

func fixedGoogleRTDNResult(code model.EntitlementResultCode) repository.ProviderEventProcessor {
	return func(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
		return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeRejected, Code: code}, nil
	}
}

func boolCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

func rawGoogleRTDNPresent(value json.RawMessage) bool {
	return len(value) > 0 && string(value) != "null"
}

func rejectedGoogleRTDN(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
	return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeRejected, Code: model.EntitlementResultInvalidPurchase}, nil
}

func acceptedGoogleRTDNTest(context.Context, pgx.Tx) (model.EntitlementOperationResult, error) {
	return model.EntitlementOperationResult{Outcome: model.EntitlementOutcomeUnchanged, Code: model.EntitlementResultProviderTest}, nil
}
