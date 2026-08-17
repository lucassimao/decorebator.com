package common

import (
	"errors"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReusableMinIOClientReusesConfigurationAndRebuildsAfterChange(t *testing.T) {
	t.Setenv("ENV", TestEnv)
	t.Setenv("MINIO_HOST", "minio")
	t.Setenv("MINIO_PORT", "9000")
	t.Setenv("MINIO_ROOT_USER", "test-user")
	t.Setenv("MINIO_ROOT_PASSWORD", "test-password")

	require.NoError(t, ConfigureMinIOFromEnvironment(TestEnv))
	first, _, err := reusableMinIOClient()
	require.NoError(t, err)
	second, _, err := reusableMinIOClient()
	require.NoError(t, err)
	assert.Same(t, first, second)

	t.Setenv("MINIO_PORT", "9001")
	require.NoError(t, ConfigureMinIOFromEnvironment(TestEnv))
	changed, _, err := reusableMinIOClient()
	require.NoError(t, err)
	assert.NotSame(t, first, changed)
}

func TestDrainMinIOListingCompletesProducerAfterFirstDeleteFailure(t *testing.T) {
	objects := make(chan minio.ObjectInfo)
	producerDone := make(chan struct{})
	go func() {
		defer close(producerDone)
		defer close(objects)
		for _, key := range []string{"first", "second", "third"} {
			objects <- minio.ObjectInfo{Key: key}
		}
	}()
	want := errors.New("delete unavailable")
	err := drainMinIOListing(func() {}, objects, func(minio.ObjectInfo) error {
		return want
	}, func(minio.ObjectInfo) bool { return true }, "prefix", "users/42-")
	require.ErrorIs(t, err, want)
	select {
	case <-producerDone:
	case <-time.After(time.Second):
		t.Fatal("listing producer remained blocked after delete failure")
	}
}

func TestStagingMinIOConfigurationUsesTLSAndPublicHTTPSURL(t *testing.T) {
	t.Setenv("MINIO_HOST", "objects.example.test")
	t.Setenv("MINIO_PORT", "")
	t.Setenv("MINIO_ROOT_USER", "staging-user")
	t.Setenv("MINIO_ROOT_PASSWORD", "staging-secret")
	t.Setenv("MINIO_BUCKET", "decorebator-staging")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("MINIO_PUBLIC_BASE_URL", "https://media-staging.example.test")

	config, err := LoadMinIOConfig(StagingEnv)
	require.NoError(t, err)
	assert.True(t, config.Secure)
	assert.Equal(t, "objects.example.test", config.Endpoint)
	assert.Equal(t,
		"https://media-staging.example.test/users/42-photo.jpg",
		config.publicObjectURL("decorebator-staging", "users/42-photo.jpg"),
	)
}

func TestDeployedMinIOConfigurationRequiresPublicHTTPSURL(t *testing.T) {
	t.Setenv("MINIO_HOST", "objects.internal.example.test")
	t.Setenv("MINIO_PORT", "9443")
	t.Setenv("MINIO_ROOT_USER", "production-user")
	t.Setenv("MINIO_ROOT_PASSWORD", "production-secret")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("MINIO_BUCKET", "production-media")
	t.Setenv("MINIO_PUBLIC_BASE_URL", "")

	_, err := LoadMinIOConfig(ProductionEnv)
	require.EqualError(t, err, "deployed storage requires MINIO_PUBLIC_BASE_URL")
}

func TestDeployedMinIOConfigurationRejectsCredentialsInPublicURL(t *testing.T) {
	t.Setenv("MINIO_HOST", "objects.internal.example.test")
	t.Setenv("MINIO_ROOT_USER", "production-user")
	t.Setenv("MINIO_ROOT_PASSWORD", "production-secret")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("MINIO_BUCKET", "production-media")
	t.Setenv("MINIO_PUBLIC_BASE_URL", "https://public-user:public-password@media.example.test")

	_, err := LoadMinIOConfig(ProductionEnv)
	require.EqualError(t, err, "MINIO_PUBLIC_BASE_URL must be an absolute URL (HTTPS when deployed) without credentials, query, or fragment")
}

func TestDeployedMinIOConfigurationRequiresValidExplicitBuckets(t *testing.T) {
	t.Setenv("MINIO_HOST", "objects.example.test")
	t.Setenv("MINIO_ROOT_USER", "production-user")
	t.Setenv("MINIO_ROOT_PASSWORD", "production-secret")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("MINIO_PUBLIC_BASE_URL", "https://media.example.test")
	t.Setenv("MINIO_BUCKET", "")
	_, err := LoadMinIOConfig(ProductionEnv)
	require.EqualError(t, err, "deployed storage requires MINIO_BUCKET")

	t.Setenv("MINIO_BUCKET", "Invalid_Bucket")
	_, err = LoadMinIOConfig(ProductionEnv)
	require.EqualError(t, err, "MINIO_BUCKET is invalid")

	t.Setenv("MINIO_BUCKET", "production-media")
	t.Setenv("MINIO_LEGACY_BUCKETS", "former-media,Invalid_Bucket")
	_, err = LoadMinIOConfig(ProductionEnv)
	require.EqualError(t, err, "MINIO_LEGACY_BUCKETS contains an invalid bucket")
}

func TestConfiguredAndExplicitLegacyBucketsAreAllowlisted(t *testing.T) {
	require.NoError(t, ConfigureMinIO(MinIOConfig{
		Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret",
		Bucket: "current-media", LegacyBuckets: []string{"former-media"},
	}))
	t.Cleanup(func() {
		require.NoError(t, ConfigureMinIO(MinIOConfig{
			Endpoint: "localhost:9000", AccessKey: "test-user", SecretKey: "test-secret", Bucket: "decorebator",
		}))
	})

	assert.True(t, MinIOBucketAllowed("current-media"))
	assert.True(t, MinIOBucketAllowed("former-media"))
	assert.False(t, MinIOBucketAllowed("attacker-media"))
}
