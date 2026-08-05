package service

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var ErrInvalidAppleSignedData = errors.New("invalid Apple signed data")

var (
	appleLeafCertificateOID         = []int{1, 2, 840, 113635, 100, 6, 11, 1}
	appleIntermediateCertificateOID = []int{1, 2, 840, 113635, 100, 6, 2, 1}
)

type AppleTransactionPayload struct {
	OriginalTransactionID string `json:"originalTransactionId"`
	TransactionID         string `json:"transactionId"`
	BundleID              string `json:"bundleId"`
	ProductID             string `json:"productId"`
	PurchaseDateMS        int64  `json:"purchaseDate"`
	ExpiresDateMS         int64  `json:"expiresDate"`
	ProductType           string `json:"type"`
	AppAccountToken       string `json:"appAccountToken"`
	Environment           string `json:"environment"`
	SignedDateMS          int64  `json:"signedDate"`
	RevocationDateMS      *int64 `json:"revocationDate,omitempty"`
	RevocationReason      *int32 `json:"revocationReason,omitempty"`
	AutoRenewProductID    string `json:"autoRenewProductId,omitempty"`
}

type AppleNotificationData struct {
	Environment           string `json:"environment"`
	AppAppleID            int64  `json:"appAppleId"`
	BundleID              string `json:"bundleId"`
	BundleVersion         string `json:"bundleVersion"`
	SignedTransactionInfo string `json:"signedTransactionInfo"`
	SignedRenewalInfo     string `json:"signedRenewalInfo"`
	Status                int32  `json:"status"`
}

type AppleNotificationPayload struct {
	NotificationType string                 `json:"notificationType"`
	Subtype          string                 `json:"subtype"`
	NotificationUUID string                 `json:"notificationUUID"`
	Data             *AppleNotificationData `json:"data"`
	Version          string                 `json:"version"`
	SignedDateMS     int64                  `json:"signedDate"`
}

type AppleRenewalInfoPayload struct {
	TransactionID            string `json:"transactionId"`
	OriginalTransactionID    string `json:"originalTransactionId"`
	ProductID                string `json:"productId"`
	AutoRenewProductID       string `json:"autoRenewProductId"`
	AutoRenewStatus          int32  `json:"autoRenewStatus"`
	IsInBillingRetryPeriod   bool   `json:"isInBillingRetryPeriod"`
	GracePeriodExpiresDateMS int64  `json:"gracePeriodExpiresDate"`
	ExpirationIntent         int32  `json:"expirationIntent"`
	Environment              string `json:"environment"`
	SignedDateMS             int64  `json:"signedDate"`
}

type AppleSignedDataVerifier interface {
	VerifyAndDecodeTransaction(signedTransaction string) (AppleTransactionPayload, error)
}

type AppleNotificationSignedDataVerifier interface {
	VerifyAndDecodeNotification(signedNotification string) (AppleNotificationPayload, error)
	VerifyAndDecodeTransaction(signedTransaction string) (AppleTransactionPayload, error)
	VerifyAndDecodeRenewalInfo(signedRenewalInfo string) (AppleRenewalInfoPayload, error)
}

type appleJWSHeader struct {
	Algorithm string   `json:"alg"`
	Chain     []string `json:"x5c"`
}

// AppleJWSVerifier follows Apple's published App Store Server Library profile:
// an ES256 compact JWS, an x5c chain of exactly three certificates, the Apple
// leaf/intermediate profile OIDs, and a chain rooted in an explicitly supplied
// Apple trust anchor. It deliberately accepts roots through dependency
// injection so startup configuration can pin current Apple PKI certificates.
type AppleJWSVerifier struct {
	roots []*x509.Certificate
}

func NewAppleJWSVerifier(roots []*x509.Certificate) (*AppleJWSVerifier, error) {
	if len(roots) == 0 {
		return nil, fmt.Errorf("Apple root certificates are required")
	}
	for _, root := range roots {
		if root == nil || !root.IsCA {
			return nil, fmt.Errorf("Apple root certificate must be a CA")
		}
	}
	return &AppleJWSVerifier{roots: append([]*x509.Certificate(nil), roots...)}, nil
}

func (v *AppleJWSVerifier) VerifyAndDecodeTransaction(signedTransaction string) (AppleTransactionPayload, error) {
	var payload AppleTransactionPayload
	if err := v.verifyAndDecode(signedTransaction, &payload); err != nil {
		return AppleTransactionPayload{}, err
	}
	if payload.TransactionID == "" || payload.AutoRenewProductID != "" {
		return AppleTransactionPayload{}, ErrInvalidAppleSignedData
	}
	return payload, nil
}

func (v *AppleJWSVerifier) VerifyAndDecodeNotification(signedNotification string) (AppleNotificationPayload, error) {
	var payload AppleNotificationPayload
	if err := v.verifyAndDecode(signedNotification, &payload); err != nil {
		return AppleNotificationPayload{}, err
	}
	if payload.Version == "" || payload.Data == nil {
		return AppleNotificationPayload{}, ErrInvalidAppleSignedData
	}
	return payload, nil
}

func (v *AppleJWSVerifier) VerifyAndDecodeRenewalInfo(signedRenewalInfo string) (AppleRenewalInfoPayload, error) {
	var payload AppleRenewalInfoPayload
	if err := v.verifyAndDecode(signedRenewalInfo, &payload); err != nil {
		return AppleRenewalInfoPayload{}, err
	}
	if payload.OriginalTransactionID == "" || payload.SignedDateMS <= 0 || payload.TransactionID != "" {
		return AppleRenewalInfoPayload{}, ErrInvalidAppleSignedData
	}
	return payload, nil
}

func (v *AppleJWSVerifier) verifyAndDecode(signedData string, destination any) error {
	parts := strings.Split(signedData, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return ErrInvalidAppleSignedData
	}

	header, payloadBytes, signedDateMS, err := decodeAppleJWS(parts)
	if err != nil {
		return err
	}
	chain, err := parseAppleCertificateChain(header.Chain)
	if err != nil || v.verifyCertificateChain(chain, signedDateMS) != nil {
		return ErrInvalidAppleSignedData
	}
	if verifyAppleJWSSignature(parts, chain[0]) != nil {
		return ErrInvalidAppleSignedData
	}
	if err := json.Unmarshal(payloadBytes, destination); err != nil {
		return ErrInvalidAppleSignedData
	}
	return nil
}

func decodeAppleJWS(parts []string) (appleJWSHeader, []byte, int64, error) {
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return appleJWSHeader{}, nil, 0, ErrInvalidAppleSignedData
	}
	var header appleJWSHeader
	unmarshalErr := json.Unmarshal(headerBytes, &header)
	if unmarshalErr != nil || header.Algorithm != "ES256" || len(header.Chain) != 3 {
		return appleJWSHeader{}, nil, 0, ErrInvalidAppleSignedData
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return appleJWSHeader{}, nil, 0, ErrInvalidAppleSignedData
	}
	var signedPayload struct {
		SignedDateMS int64 `json:"signedDate"`
	}
	unmarshalErr = json.Unmarshal(payloadBytes, &signedPayload)
	if unmarshalErr != nil || signedPayload.SignedDateMS <= 0 {
		return appleJWSHeader{}, nil, 0, ErrInvalidAppleSignedData
	}
	return header, payloadBytes, signedPayload.SignedDateMS, nil
}

func (v *AppleJWSVerifier) verifyCertificateChain(chain []*x509.Certificate, signedDateMS int64) error {
	if !certificateHasExtension(chain[0], appleLeafCertificateOID) ||
		!certificateHasExtension(chain[1], appleIntermediateCertificateOID) {
		return ErrInvalidAppleSignedData
	}
	if !v.chainEndsInTrustedRoot(chain[2]) {
		return ErrInvalidAppleSignedData
	}

	rootPool := x509.NewCertPool()
	for _, root := range v.roots {
		rootPool.AddCert(root)
	}
	intermediatePool := x509.NewCertPool()
	intermediatePool.AddCert(chain[1])
	_, verifyErr := chain[0].Verify(x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: intermediatePool,
		CurrentTime:   time.UnixMilli(signedDateMS),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	if verifyErr != nil {
		return ErrInvalidAppleSignedData
	}
	return nil
}

func verifyAppleJWSSignature(parts []string, leaf *x509.Certificate) error {
	publicKey, ok := leaf.PublicKey.(*ecdsa.PublicKey)
	if !ok || publicKey.Curve.Params().BitSize != 256 {
		return ErrInvalidAppleSignedData
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(signature) != 64 {
		return ErrInvalidAppleSignedData
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])
	if !ecdsa.Verify(publicKey, digest[:], r, s) {
		return ErrInvalidAppleSignedData
	}
	return nil
}

func parseAppleCertificateChain(encoded []string) ([]*x509.Certificate, error) {
	chain := make([]*x509.Certificate, 0, len(encoded))
	for _, value := range encoded {
		der, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		certificate, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, err
		}
		chain = append(chain, certificate)
	}
	return chain, nil
}

func certificateHasExtension(certificate *x509.Certificate, oid []int) bool {
	for _, extension := range certificate.Extensions {
		if extension.Id.Equal(oid) {
			return true
		}
	}
	return false
}

func (v *AppleJWSVerifier) chainEndsInTrustedRoot(candidate *x509.Certificate) bool {
	for _, root := range v.roots {
		if candidate.Equal(root) {
			return true
		}
	}
	return false
}
