package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	"decorebator.com/internal/model"
)

const storeEvidenceKeySize = 32

type StoreEvidenceKind string

const (
	StoreEvidenceAccountIdentifier StoreEvidenceKind = "account_identifier"
	StoreEvidenceProviderRecord    StoreEvidenceKind = "provider_record"
	StoreEvidenceLinkedRecord      StoreEvidenceKind = "linked_provider_record"
)

type StoreEvidenceContext struct {
	UserID      int64
	Store       model.EntitlementStore
	Environment *model.StoreEnvironment
	Kind        StoreEvidenceKind
	Lookup      string
}

type ProtectedStoreEvidence struct {
	Ciphertext []byte
	KeyVersion int16
}

// StoreEvidenceProtector encrypts recoverable provider identifiers and creates
// blind indexes for exact lookup. The lookup key is an installation invariant:
// it must only change during an offline migration that rewrites every digest.
type StoreEvidenceProtector struct {
	lookupKey     []byte
	activeVersion int16
	keys          map[int16][]byte
	random        io.Reader
}

func NewStoreEvidenceProtector(
	lookupKey []byte,
	activeVersion int16,
	keys map[int16][]byte,
) (*StoreEvidenceProtector, error) {
	return newStoreEvidenceProtector(lookupKey, activeVersion, keys, rand.Reader)
}

func newStoreEvidenceProtector(
	lookupKey []byte,
	activeVersion int16,
	keys map[int16][]byte,
	random io.Reader,
) (*StoreEvidenceProtector, error) {
	if len(lookupKey) != storeEvidenceKeySize || activeVersion <= 0 || len(keys) == 0 || random == nil {
		return nil, fmt.Errorf("store evidence protector configuration is invalid")
	}
	keyCopy := make(map[int16][]byte, len(keys))
	for version, key := range keys {
		if version <= 0 || len(key) != storeEvidenceKeySize {
			return nil, fmt.Errorf("store evidence encryption key configuration is invalid")
		}
		keyCopy[version] = append([]byte(nil), key...)
	}
	if _, ok := keyCopy[activeVersion]; !ok {
		return nil, fmt.Errorf("active store evidence encryption key is missing")
	}
	return &StoreEvidenceProtector{
		lookupKey: append([]byte(nil), lookupKey...), activeVersion: activeVersion,
		keys: keyCopy, random: random,
	}, nil
}

func (p *StoreEvidenceProtector) LookupDigest(
	store model.EntitlementStore,
	environment *model.StoreEnvironment,
	kind StoreEvidenceKind,
	plaintext string,
) (string, error) {
	if p == nil || len(p.lookupKey) != storeEvidenceKeySize || strings.TrimSpace(plaintext) == "" ||
		!validStoreEvidenceScope(store, environment, kind) {
		return "", fmt.Errorf("store evidence lookup input is invalid")
	}
	mac := hmac.New(sha256.New, p.lookupKey)
	writeEvidencePart(mac, string(store))
	writeEvidencePart(mac, environmentValue(environment))
	writeEvidencePart(mac, string(kind))
	writeEvidencePart(mac, plaintext)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (p *StoreEvidenceProtector) Encrypt(
	context StoreEvidenceContext,
	plaintext string,
) (ProtectedStoreEvidence, error) {
	if p == nil || strings.TrimSpace(plaintext) == "" {
		return ProtectedStoreEvidence{}, fmt.Errorf("store evidence plaintext is invalid")
	}
	if err := validateStoreEvidenceContext(context); err != nil {
		return ProtectedStoreEvidence{}, err
	}
	key := p.keys[p.activeVersion]
	block, err := aes.NewCipher(key)
	if err != nil {
		return ProtectedStoreEvidence{}, fmt.Errorf("initialize store evidence cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return ProtectedStoreEvidence{}, fmt.Errorf("initialize store evidence AEAD: %w", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(p.random, nonce); err != nil {
		return ProtectedStoreEvidence{}, fmt.Errorf("generate store evidence nonce: %w", err)
	}
	ciphertext := aead.Seal(nonce, nonce, []byte(plaintext), evidenceAAD(context))
	return ProtectedStoreEvidence{Ciphertext: ciphertext, KeyVersion: p.activeVersion}, nil
}

func (p *StoreEvidenceProtector) Decrypt(
	context StoreEvidenceContext,
	protected ProtectedStoreEvidence,
) (string, error) {
	if p == nil || len(protected.Ciphertext) == 0 || protected.KeyVersion <= 0 {
		return "", fmt.Errorf("protected store evidence is invalid")
	}
	if err := validateStoreEvidenceContext(context); err != nil {
		return "", err
	}
	key, ok := p.keys[protected.KeyVersion]
	if !ok {
		return "", fmt.Errorf("store evidence encryption key version is unavailable")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("initialize store evidence cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("initialize store evidence AEAD: %w", err)
	}
	if len(protected.Ciphertext) < aead.NonceSize() {
		return "", fmt.Errorf("protected store evidence is truncated")
	}
	nonce := protected.Ciphertext[:aead.NonceSize()]
	ciphertext := protected.Ciphertext[aead.NonceSize():]
	plaintext, err := aead.Open(nil, nonce, ciphertext, evidenceAAD(context))
	if err != nil {
		return "", fmt.Errorf("authenticate protected store evidence")
	}
	return string(plaintext), nil
}

func (p *StoreEvidenceProtector) ActiveVersion() int16 {
	if p == nil {
		return 0
	}
	return p.activeVersion
}

func validateStoreEvidenceContext(context StoreEvidenceContext) error {
	if context.UserID <= 0 || context.Lookup == "" || !validStoreEvidenceScope(context.Store, context.Environment, context.Kind) {
		return fmt.Errorf("store evidence context is invalid")
	}
	return nil
}

func validStoreEvidenceScope(
	store model.EntitlementStore,
	environment *model.StoreEnvironment,
	kind StoreEvidenceKind,
) bool {
	if !store.Valid() {
		return false
	}
	if kind != StoreEvidenceAccountIdentifier && kind != StoreEvidenceProviderRecord && kind != StoreEvidenceLinkedRecord {
		return false
	}
	if kind == StoreEvidenceAccountIdentifier {
		return environment == nil
	}
	return environment != nil && environment.Valid()
}

func evidenceAAD(context StoreEvidenceContext) []byte {
	buffer := make([]byte, 0, 128)
	buffer = appendEvidencePart(buffer, strconv.FormatInt(context.UserID, 10))
	buffer = appendEvidencePart(buffer, string(context.Store))
	buffer = appendEvidencePart(buffer, environmentValue(context.Environment))
	buffer = appendEvidencePart(buffer, string(context.Kind))
	return appendEvidencePart(buffer, context.Lookup)
}

func environmentValue(environment *model.StoreEnvironment) string {
	if environment == nil {
		return ""
	}
	return string(*environment)
}

type evidenceWriter interface {
	Write([]byte) (int, error)
}

func writeEvidencePart(writer evidenceWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write([]byte(value))
}

func appendEvidencePart(buffer []byte, value string) []byte {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	buffer = append(buffer, length[:]...)
	return append(buffer, value...)
}
