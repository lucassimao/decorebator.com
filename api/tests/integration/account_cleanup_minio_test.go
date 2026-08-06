package integration

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/tests/integration/setup"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinIODeletePrefixRemovesVersionsWithoutTouchingCollidingUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	host := os.Getenv("MINIO_HOST")
	port := os.Getenv("MINIO_PORT")
	if host == "" || port == "" {
		t.Skip("MinIO integration environment is unavailable")
	}
	t.Setenv("ENV", "test")
	t.Setenv("MINIO_ROOT_USER", "testuser")
	t.Setenv("MINIO_ROOT_PASSWORD", "testpass123")

	ctx := context.Background()
	client, err := minio.New(fmt.Sprintf("%s:%s", host, port), &minio.Options{
		Creds: credentials.NewStaticV4("testuser", "testpass123", ""),
	})
	require.NoError(t, err)

	bucket := fmt.Sprintf("account-cleanup-%d", time.Now().UnixNano())
	require.NoError(t, client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}))
	defer func() {
		_ = common.MinIODeletePrefix(ctx, bucket, "users/")
		_ = client.RemoveBucket(ctx, bucket)
	}()
	require.NoError(t, client.EnableVersioning(ctx, bucket))

	put := func(object, body string) {
		t.Helper()
		_, putErr := client.PutObject(ctx, bucket, object, bytes.NewBufferString(body), int64(len(body)), minio.PutObjectOptions{})
		require.NoError(t, putErr)
	}
	put("users/42-first.jpg", "version one")
	put("users/42-first.jpg", "version two")
	put("users/42-second.jpg", "another object")
	put("users/420-keep.jpg", "different user")

	require.NoError(t, common.MinIODeletePrefix(ctx, bucket, "users/42-"))
	assert.Empty(t, listObjectVersions(ctx, t, client, bucket, "users/42-"))
	assert.NotEmpty(t, listObjectVersions(ctx, t, client, bucket, "users/420-"))
}

func TestProfileUploadIsCompensatedWhenDatabaseUpdateFails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := configuredTestMinIOClient(t)
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)

	ctx := context.Background()
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	_, err := server.DB.Exec(ctx, fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION test_reject_profile_%d_update()
		RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'forced profile update failure';
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER test_reject_profile_%d_update_trigger
		BEFORE UPDATE ON users
		FOR EACH ROW WHEN (OLD.id = %d)
		EXECUTE FUNCTION test_reject_profile_%d_update();
	`, userID, userID, userID, userID))
	require.NoError(t, err)
	defer func() {
		_, cleanupErr := server.DB.Exec(ctx, fmt.Sprintf(`
			DROP TRIGGER IF EXISTS test_reject_profile_%d_update_trigger ON users;
			DROP FUNCTION IF EXISTS test_reject_profile_%d_update();
		`, userID, userID))
		require.NoError(t, cleanupErr)
	}()

	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{
				"base64Data": "data:image/png;base64,aW1hZ2U=",
				"extension":  "png",
			},
		}).
		Expect().
		Status(http.StatusInternalServerError)

	assert.Empty(t, listObjectVersions(ctx, t, client, "decorebator", fmt.Sprintf("users/%d-", userID)))
}

func configuredTestMinIOClient(t *testing.T) *minio.Client {
	t.Helper()
	host := os.Getenv("MINIO_HOST")
	port := os.Getenv("MINIO_PORT")
	if host == "" || port == "" {
		t.Skip("MinIO integration environment is unavailable")
	}
	t.Setenv("ENV", "test")
	t.Setenv("MINIO_ROOT_USER", "testuser")
	t.Setenv("MINIO_ROOT_PASSWORD", "testpass123")
	client, err := minio.New(fmt.Sprintf("%s:%s", host, port), &minio.Options{
		Creds: credentials.NewStaticV4("testuser", "testpass123", ""),
	})
	require.NoError(t, err)
	return client
}

func listObjectVersions(ctx context.Context, t *testing.T, client *minio.Client, bucket, prefix string) []minio.ObjectInfo {
	t.Helper()
	var objects []minio.ObjectInfo
	for object := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:       prefix,
		Recursive:    true,
		WithVersions: true,
	}) {
		require.NoError(t, object.Err)
		objects = append(objects, object)
	}
	return objects
}
