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
}

type AppleSignedDataVerifier interface {
	VerifyAndDecodeTransaction(signedTransaction string) (AppleTransactionPayload, error)
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
	parts := strings.Split(signedTransaction, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return AppleTransactionPayload{}, ErrInvalidAppleSignedData
	}

	header, payload, err := decodeAppleJWS(parts)
	if err != nil {
		return AppleTransactionPayload{}, err
	}
	chain, err := parseAppleCertificateChain(header.Chain)
	if err != nil || v.verifyCertificateChain(chain, payload.SignedDateMS) != nil {
		return AppleTransactionPayload{}, ErrInvalidAppleSignedData
	}
	if verifyAppleJWSSignature(parts, chain[0]) != nil {
		return AppleTransactionPayload{}, ErrInvalidAppleSignedData
	}
	return payload, nil
}

func decodeAppleJWS(parts []string) (appleJWSHeader, AppleTransactionPayload, error) {
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return appleJWSHeader{}, AppleTransactionPayload{}, ErrInvalidAppleSignedData
	}
	var header appleJWSHeader
	unmarshalErr := json.Unmarshal(headerBytes, &header)
	if unmarshalErr != nil || header.Algorithm != "ES256" || len(header.Chain) != 3 {
		return appleJWSHeader{}, AppleTransactionPayload{}, ErrInvalidAppleSignedData
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return appleJWSHeader{}, AppleTransactionPayload{}, ErrInvalidAppleSignedData
	}
	var payload AppleTransactionPayload
	unmarshalErr = json.Unmarshal(payloadBytes, &payload)
	if unmarshalErr != nil || payload.SignedDateMS <= 0 {
		return appleJWSHeader{}, AppleTransactionPayload{}, ErrInvalidAppleSignedData
	}
	return header, payload, nil
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
