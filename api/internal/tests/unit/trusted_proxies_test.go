package unit

import (
	"testing"

	"decorebator.com/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustedProxyConfigIsExplicitAndValidated(t *testing.T) {
	lookup := func(values map[string]string) config.EnvironmentLookup {
		return func(name string) (string, bool) {
			value, ok := values[name]
			return value, ok
		}
	}

	empty, err := config.LoadTrustedProxyConfigFrom(lookup(nil))
	require.NoError(t, err)
	assert.Empty(t, empty.CIDRs)

	configured, err := config.LoadTrustedProxyConfigFrom(lookup(map[string]string{
		"TRUSTED_PROXY_CIDRS": "10.0.0.0/8, 2001:db8::1,10.0.0.0/8",
	}))
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.0/8", "2001:db8::1"}, configured.CIDRs)

	direct, err := config.LoadTrustedProxyConfigFrom(lookup(map[string]string{
		"ENV": "production", "DIRECT_CLIENT_TRAFFIC": "true",
	}))
	require.NoError(t, err)
	assert.True(t, direct.DirectMode)

	_, err = config.LoadTrustedProxyConfigFrom(lookup(map[string]string{"ENV": "production"}))
	require.ErrorContains(t, err, "requires TRUSTED_PROXY_CIDRS")

	for _, value := range []string{"not-an-address", "10.0.0.0/8,", "0.0.0.0/0", "203.0.113.7/0", "::/0", "2001:db8::1/0"} {
		_, loadErr := config.LoadTrustedProxyConfigFrom(lookup(map[string]string{
			"TRUSTED_PROXY_CIDRS": value,
		}))
		require.Error(t, loadErr, value)
	}
}
