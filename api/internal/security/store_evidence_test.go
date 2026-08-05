package security

import (
	"bytes"
	"testing"

	"decorebator.com/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreEvidenceProtectorEncryptsWithBoundAADAndStableBlindIndex(t *testing.T) {
	lookupKey := bytes.Repeat([]byte{1}, storeEvidenceKeySize)
	keys := map[int16][]byte{1: bytes.Repeat([]byte{2}, storeEvidenceKeySize)}
	protector, err := newStoreEvidenceProtector(lookupKey, 1, keys, bytes.NewReader(bytes.Repeat([]byte{3}, 128)))
	require.NoError(t, err)
	environment := model.StoreEnvironmentProduction
	lookup, err := protector.LookupDigest(model.EntitlementStoreGoogle, &environment, StoreEvidenceProviderRecord, "provider-secret")
	require.NoError(t, err)
	assert.NotContains(t, lookup, "provider-secret")
	context := StoreEvidenceContext{
		UserID: 42, Store: model.EntitlementStoreGoogle, Environment: &environment,
		Kind: StoreEvidenceProviderRecord, Lookup: lookup,
	}
	protected, err := protector.Encrypt(context, "provider-secret")
	require.NoError(t, err)
	assert.NotContains(t, string(protected.Ciphertext), "provider-secret")
	plaintext, err := protector.Decrypt(context, protected)
	require.NoError(t, err)
	assert.Equal(t, "provider-secret", plaintext)

	otherUser := context
	otherUser.UserID = 43
	_, err = protector.Decrypt(otherUser, protected)
	assert.Error(t, err)
	otherField := context
	otherField.Kind = StoreEvidenceLinkedRecord
	_, err = protector.Decrypt(otherField, protected)
	assert.Error(t, err)
}

func TestStoreEvidenceProtectorSupportsEncryptionKeyRotation(t *testing.T) {
	lookupKey := bytes.Repeat([]byte{1}, storeEvidenceKeySize)
	oldKey := bytes.Repeat([]byte{2}, storeEvidenceKeySize)
	newKey := bytes.Repeat([]byte{3}, storeEvidenceKeySize)
	environment := model.StoreEnvironmentSandbox
	context := StoreEvidenceContext{
		UserID: 42, Store: model.EntitlementStoreApple, Environment: &environment,
		Kind: StoreEvidenceProviderRecord,
	}
	oldProtector, err := newStoreEvidenceProtector(lookupKey, 1, map[int16][]byte{1: oldKey}, bytes.NewReader(bytes.Repeat([]byte{4}, 64)))
	require.NoError(t, err)
	context.Lookup, err = oldProtector.LookupDigest(context.Store, context.Environment, context.Kind, "transaction")
	require.NoError(t, err)
	oldEvidence, err := oldProtector.Encrypt(context, "transaction")
	require.NoError(t, err)

	rotated, err := newStoreEvidenceProtector(lookupKey, 2, map[int16][]byte{1: oldKey, 2: newKey}, bytes.NewReader(bytes.Repeat([]byte{5}, 64)))
	require.NoError(t, err)
	plaintext, err := rotated.Decrypt(context, oldEvidence)
	require.NoError(t, err)
	assert.Equal(t, "transaction", plaintext)
	newEvidence, err := rotated.Encrypt(context, plaintext)
	require.NoError(t, err)
	assert.Equal(t, int16(2), newEvidence.KeyVersion)
	assert.NotEqual(t, oldEvidence.Ciphertext, newEvidence.Ciphertext)
}

func TestStoreEvidenceProtectorRejectsInvalidConfigurationAndScope(t *testing.T) {
	_, err := NewStoreEvidenceProtector([]byte("short"), 1, map[int16][]byte{1: bytes.Repeat([]byte{1}, storeEvidenceKeySize)})
	assert.Error(t, err)
	protector, err := NewStoreEvidenceProtector(
		bytes.Repeat([]byte{1}, storeEvidenceKeySize), 1,
		map[int16][]byte{1: bytes.Repeat([]byte{2}, storeEvidenceKeySize)},
	)
	require.NoError(t, err)
	_, err = protector.LookupDigest(model.EntitlementStoreApple, nil, StoreEvidenceProviderRecord, "record")
	assert.Error(t, err)
}
