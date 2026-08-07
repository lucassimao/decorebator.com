package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestValidatePasswordUsesOneLengthPolicy(t *testing.T) {
	require.NoError(t, ValidatePassword("password"))
	require.NoError(t, ValidatePassword("áéíóúabc"))
	require.EqualError(t, ValidatePassword("1234567"), "password must contain at least 8 characters")
	require.EqualError(t, ValidatePassword(strings.Repeat("a", PasswordMaxBytes+1)),
		"password must contain at most 72 UTF-8 bytes")
	require.EqualError(t, ValidatePassword(strings.Repeat("界", 25)),
		"password must contain at most 72 UTF-8 bytes", "bcrypt's 72-byte boundary must be enforced")
}

func TestProductionBcryptConfigurationRejectsLowOrInvalidOverrides(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("BCRYPT_COST", "4")
	require.Error(t, ValidateBcryptConfiguration("production"))
	assert.Equal(t, bcrypt.DefaultCost, GetBcryptCost(), "runtime hashing must still clamp unsafe overrides")

	t.Setenv("BCRYPT_COST", "invalid")
	require.Error(t, ValidateBcryptConfiguration("production"))

	t.Setenv("BCRYPT_COST", "10")
	require.NoError(t, ValidateBcryptConfiguration("production"))
	assert.Equal(t, 10, GetBcryptCost())
}
