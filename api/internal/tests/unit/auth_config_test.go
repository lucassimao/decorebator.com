package unit

import (
	"testing"

	"decorebator.com/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAuthConfigRejectsUnsafeStartupConfiguration(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "missing key", env: map[string]string{"ENV": "test"}, want: "JWT_KEY is required"},
		{name: "short key", env: map[string]string{"ENV": "test", "JWT_KEY": "too-short"}, want: "at least 32 bytes"},
		{name: "placeholder", env: map[string]string{"ENV": "test", "JWT_KEY": "your_jwt_secret_here_must_be_longer"}, want: "placeholder"},
		{name: "whitespace", env: map[string]string{"ENV": "test", "JWT_KEY": " 0123456789abcdef0123456789abcdef"}, want: "surrounding whitespace"},
		{name: "missing environment", env: map[string]string{"JWT_KEY": "0123456789abcdef0123456789abcdef"}, want: "ENV is required"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := config.LoadAuthConfigFrom(authMapLookup(test.env))
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestLoadAuthConfigReturnsBoundImmutableContract(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	result, err := config.LoadAuthConfigFrom(authMapLookup(map[string]string{"ENV": "staging", "JWT_KEY": key}))
	require.NoError(t, err)
	assert.Equal(t, []byte(key), result.SigningKey)
	assert.Equal(t, "staging", result.Environment)
	assert.Equal(t, config.AuthIssuer, result.Issuer)
	assert.Equal(t, config.AuthAudience, result.Audience)
	assert.Equal(t, config.AccessTokenDuration, result.AccessTTL)
}

func authMapLookup(values map[string]string) config.EnvironmentLookup {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
