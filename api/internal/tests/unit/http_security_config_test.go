package unit

import (
	"testing"

	"decorebator.com/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPSecurityConfigRejectsUnsafeProductionSettings(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "missing origins", env: map[string]string{"STATIC_AUTHENTICATION": strongStaticToken}, want: "CORS_ALLOWED_ORIGINS"},
		{name: "insecure origin", env: map[string]string{"CORS_ALLOWED_ORIGINS": "http://decorebator.com", "STATIC_AUTHENTICATION": strongStaticToken}, want: "HTTPS"},
		{name: "origin path", env: map[string]string{"CORS_ALLOWED_ORIGINS": "https://decorebator.com/path", "STATIC_AUTHENTICATION": strongStaticToken}, want: "origin"},
		{name: "origin missing hostname", env: map[string]string{"CORS_ALLOWED_ORIGINS": "https://:443", "STATIC_AUTHENTICATION": strongStaticToken}, want: "origin"},
		{name: "origin empty query", env: map[string]string{"CORS_ALLOWED_ORIGINS": "https://decorebator.com?", "STATIC_AUTHENTICATION": strongStaticToken}, want: "origin"},
		{name: "origin empty fragment", env: map[string]string{"CORS_ALLOWED_ORIGINS": "https://decorebator.com#", "STATIC_AUTHENTICATION": strongStaticToken}, want: "origin"},
		{name: "origin invalid port", env: map[string]string{"CORS_ALLOWED_ORIGINS": "https://decorebator.com:bad", "STATIC_AUTHENTICATION": strongStaticToken}, want: "origin"},
		{name: "missing static token", env: map[string]string{"CORS_ALLOWED_ORIGINS": "https://decorebator.com"}, want: "STATIC_AUTHENTICATION"},
		{name: "short static token", env: map[string]string{"CORS_ALLOWED_ORIGINS": "https://decorebator.com", "STATIC_AUTHENTICATION": "short"}, want: "32 bytes"},
		{name: "placeholder static token", env: map[string]string{"CORS_ALLOWED_ORIGINS": "https://decorebator.com", "STATIC_AUTHENTICATION": "your_static_auth_token_change_me_now"}, want: "placeholder"},
		{name: "documented placeholder static token", env: map[string]string{"CORS_ALLOWED_ORIGINS": "https://decorebator.com", "STATIC_AUTHENTICATION": "replace-with-at-least-32-random-bytes"}, want: "placeholder"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := config.LoadHTTPSecurityConfigFrom("production", authMapLookup(test.env))
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestHTTPSecurityConfigReturnsExplicitImmutablePolicy(t *testing.T) {
	result, err := config.LoadHTTPSecurityConfigFrom("production", authMapLookup(map[string]string{
		"CORS_ALLOWED_ORIGINS":  "https://decorebator.com",
		"STATIC_AUTHENTICATION": strongStaticToken,
	}))
	require.NoError(t, err)
	assert.Equal(t, []string{"https://decorebator.com"}, result.AllowedOrigins)
	assert.Equal(t, strongStaticToken, result.StaticAuthentication)
	assert.True(t, result.SecureCookies)
	assert.Empty(t, result.CookieDomain, "cookies must remain host-only")
	assert.Equal(t, "decorebator.com", result.LegacyCookieDomain)
}

func TestHTTPSecurityConfigRejectsMultipleDeployedCredentialedOrigins(t *testing.T) {
	_, err := config.LoadHTTPSecurityConfigFrom("production", authMapLookup(map[string]string{
		"CORS_ALLOWED_ORIGINS":  "https://decorebator.com,https://www.decorebator.com",
		"STATIC_AUTHENTICATION": strongStaticToken,
	}))
	require.ErrorContains(t, err, "exactly one")
}

func TestHTTPSecurityConfigUsesBoundedLocalOriginsWithoutReflection(t *testing.T) {
	result, err := config.LoadHTTPSecurityConfigFrom("test", authMapLookup(nil))
	require.NoError(t, err)
	assert.NotEmpty(t, result.AllowedOrigins)
	assert.NotContains(t, result.AllowedOrigins, "https://attacker.example")
	assert.False(t, result.SecureCookies)
	assert.Empty(t, result.StaticAuthentication)
}

func TestHTTPSecurityConfigTreatsStagingAsDeployed(t *testing.T) {
	_, err := config.LoadHTTPSecurityConfigFrom("staging", authMapLookup(nil))
	require.ErrorContains(t, err, "CORS_ALLOWED_ORIGINS")

	result, err := config.LoadHTTPSecurityConfigFrom("staging", authMapLookup(map[string]string{
		"CORS_ALLOWED_ORIGINS":  "https://staging.decorebator.com",
		"STATIC_AUTHENTICATION": strongStaticToken,
	}))
	require.NoError(t, err)
	assert.True(t, result.SecureCookies)
	assert.Empty(t, result.LegacyCookieDomain)
}

const strongStaticToken = "0123456789abcdef0123456789abcdef"
