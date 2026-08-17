package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinIODeletePrefixRemovesVersionsWithoutTouchingCollidingUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	client := configuredTestMinIOClient(t)

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
	bucket := common.MinIOBucketName()
	require.NoError(t, client.EnableVersioning(context.Background(), bucket))
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
				"base64Data": validProfilePNG(t),
			},
		}).
		Expect().
		Status(http.StatusInternalServerError)

	assert.Empty(t, listObjectVersions(ctx, t, client, bucket, fmt.Sprintf("users/%d-", userID)))
	var intentVersion string
	var deleteAllVersions bool
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT COALESCE(args->>'object_version_id', ''),
		       COALESCE((args->>'delete_all_versions')::boolean, false)
		FROM river_job
		WHERE kind='profile_upload_reconciliation_v1'
		ORDER BY id DESC LIMIT 1
	`).Scan(&intentVersion, &deleteAllVersions))
	assert.Empty(t, intentVersion)
	assert.True(t, deleteAllVersions)
}

func TestProfileUploadUsesValidatedServerOwnedObjectMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := configuredTestMinIOClient(t)
	bucket := common.MinIOBucketName()
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)

	ctx := context.Background()
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	prefix := fmt.Sprintf("users/%d-", userID)
	defer func() { require.NoError(t, common.MinIODeletePrefix(ctx, bucket, prefix)) }()

	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{
				"base64Data": "data:image/svg+xml;base64," + validProfilePNG(t),
				"extension":  "../../attacker.svg",
			},
		}).
		Expect().
		Status(http.StatusBadRequest)
	assert.Empty(t, listObjectVersions(ctx, t, client, bucket, prefix))

	for _, extension := range []string{"../../attacker.svg", "jpg"} {
		server.Expect.PATCH("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			WithJSON(map[string]any{
				"updateProfilePicture": map[string]any{
					"base64Data": "data:image/png;base64," + validProfilePNG(t),
					"extension":  extension,
				},
			}).
			Expect().
			Status(http.StatusBadRequest)
		assert.Empty(t, listObjectVersions(ctx, t, client, bucket, prefix))
	}

	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{
				"base64Data": hostileProfileJPEGWithOversizedDHT(t),
			},
		}).
		Expect().
		Status(http.StatusBadRequest)
	assert.Empty(t, listObjectVersions(ctx, t, client, bucket, prefix))

	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{
				"base64Data": "data:image/png;base64," + validProfilePNG(t),
				"extension":  "png",
			},
		}).
		Expect().
		Status(http.StatusOK)

	objects := listObjectVersions(ctx, t, client, bucket, prefix)
	require.Len(t, objects, 1)
	assert.True(t, strings.HasSuffix(objects[0].Key, ".png"), objects[0].Key)
	assert.NotContains(t, objects[0].Key, "attacker")
	info, err := client.StatObject(ctx, bucket, objects[0].Key, minio.StatObjectOptions{})
	require.NoError(t, err)
	assert.Equal(t, "image/png", info.ContentType)
}

func TestNativeProfileImageFixturesUploadAndPersist(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	fixtureVariables := []string{"PROFILE_IMAGE_ANDROID_FIXTURE", "PROFILE_IMAGE_IOS_FIXTURE"}
	found := false
	for _, variable := range fixtureVariables {
		if os.Getenv(variable) != "" {
			found = true
		}
	}
	if !found {
		t.Skip("native compatibility fixtures were not supplied")
	}

	client := configuredTestMinIOClient(t)
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	ctx := context.Background()
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	bucket := common.MinIOBucketName()
	prefix := fmt.Sprintf("users/%d-", userID)
	defer func() { require.NoError(t, common.MinIODeletePrefix(ctx, bucket, prefix)) }()

	accepted := 0
	for _, variable := range fixtureVariables {
		fixturePath := os.Getenv(variable)
		if fixturePath == "" {
			continue
		}
		fixture, err := os.ReadFile(fixturePath)
		require.NoError(t, err)
		response := server.Expect.PATCH("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			WithJSON(map[string]any{
				"updateProfilePicture": map[string]any{
					"base64Data": base64.StdEncoding.EncodeToString(fixture),
					"extension":  "jpg",
				},
			}).
			Expect().
			Status(http.StatusOK)
		profileURL := response.JSON().Object().Value("profilePictureUrl").String().Raw()
		var persistedURL string
		require.NoError(t, server.DB.QueryRow(
			ctx, `SELECT profile_picture_url FROM users WHERE id=$1`, userID,
		).Scan(&persistedURL))
		assert.Equal(t, profileURL, persistedURL)
		accepted++
	}
	assert.Len(t, listObjectVersions(ctx, t, client, bucket, prefix), accepted)
}

func TestProfileUploadFinalizationDeadlineReturnsRequestTimeout(t *testing.T) {
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
	bucket := common.MinIOBucketName()
	prefix := fmt.Sprintf("users/%d-", userID)
	defer func() { require.NoError(t, common.MinIODeletePrefix(ctx, bucket, prefix)) }()
	_, err := server.DB.Exec(ctx, `
		CREATE OR REPLACE FUNCTION test_delay_profile_intent_update()
		RETURNS trigger AS $$
		BEGIN
			PERFORM pg_sleep(20);
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER test_delay_profile_intent_update_trigger
		BEFORE UPDATE OF args ON river_job
		FOR EACH ROW WHEN (OLD.kind = 'profile_upload_reconciliation_v1')
		EXECUTE FUNCTION test_delay_profile_intent_update();
	`)
	require.NoError(t, err)
	defer func() {
		_, cleanupErr := server.DB.Exec(ctx, `
			DROP TRIGGER IF EXISTS test_delay_profile_intent_update_trigger ON river_job;
			DROP FUNCTION IF EXISTS test_delay_profile_intent_update();
		`)
		require.NoError(t, cleanupErr)
	}()

	started := time.Now()
	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{"base64Data": validProfilePNG(t)},
		}).
		Expect().
		Status(http.StatusRequestTimeout)
	assert.GreaterOrEqual(t, time.Since(started), 11*time.Second)
	require.Len(t, listObjectVersions(ctx, t, client, bucket, prefix), 1)
	var encodedArgs []byte
	require.NoError(t, server.DB.QueryRow(ctx, `
		SELECT args FROM river_job
		WHERE kind='profile_upload_reconciliation_v1'
		ORDER BY id DESC LIMIT 1
	`).Scan(&encodedArgs))
	var args service.ProfileUploadReconciliationArgs
	require.NoError(t, json.Unmarshal(encodedArgs, &args))
	assert.True(t, args.DeleteAllVersions)
	assert.Empty(t, args.ObjectVersionID)
	require.NoError(t, service.NewProfileUploadReconciliationWorker(server.DB).Work(
		ctx,
		&river.Job[service.ProfileUploadReconciliationArgs]{Args: args},
	))
	assert.Empty(t, listObjectVersions(ctx, t, client, bucket, prefix))
}

func TestProfileUploadUsesConfiguredNonDefaultBucketEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	controlBucket := os.Getenv("MINIO_BUCKET")
	if controlBucket == "" {
		controlBucket = "decorebator"
	}
	bucket := fmt.Sprintf("profile-upload-%d", time.Now().UnixNano())
	t.Setenv("MINIO_BUCKET", bucket)
	client := configuredTestMinIOClient(t)
	ctx := context.Background()
	require.NoError(t, client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}))
	defer func() {
		require.NoError(t, common.MinIODeletePrefix(ctx, bucket, "users/"))
		require.NoError(t, client.RemoveBucket(ctx, bucket))
	}()

	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	prefix := fmt.Sprintf("users/%d-", userID)

	response := server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{"base64Data": validProfilePNG(t)},
		}).
		Expect().
		Status(http.StatusOK)

	objects := listObjectVersions(ctx, t, client, bucket, prefix)
	require.Len(t, objects, 1)
	profileURL := response.JSON().Object().Value("profilePictureUrl").String().Raw()
	assert.Contains(t, profileURL, "/"+bucket+"/")
	assert.Contains(t, profileURL, objects[0].Key)
	assert.Empty(t, listObjectVersions(ctx, t, client, controlBucket, prefix))
}

func TestInvalidProfileFieldsDoNotCreateAnObject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := configuredTestMinIOClient(t)
	bucket := common.MinIOBucketName()
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	ctx := context.Background()
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	prefix := fmt.Sprintf("users/%d-", userID)

	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"dateOfBirth":          "not-a-date",
			"updateProfilePicture": map[string]any{"base64Data": validProfilePNG(t)},
		}).
		Expect().
		Status(http.StatusBadRequest)
	assert.Empty(t, listObjectVersions(ctx, t, client, bucket, prefix))
}

func TestProfileMetadataPolyglotsReturnBadRequestWithoutCreatingObjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := configuredTestMinIOClient(t)
	bucket := common.MinIOBucketName()
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	ctx := context.Background()
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	prefix := fmt.Sprintf("users/%d-", userID)

	for name, payload := range map[string]string{
		"jpeg-app":    profileJPEGWithMetadata(t, 0xe2, []byte("PK\x03\x04embedded.zip")),
		"jpeg-com":    profileJPEGWithMetadata(t, 0xfe, []byte("<script>embedded</script>")),
		"png-private": profilePNGWithMetadata(t, "raNd", []byte("PK\x03\x04embedded.zip")),
		"png-iccp":    profilePNGWithMetadata(t, "iCCP", []byte("profile\x00\x00not-zlib")),
		"png-ztxt":    profilePNGWithMetadata(t, "zTXt", []byte("Comment\x00\x00not-zlib")),
	} {
		t.Run(name, func(t *testing.T) {
			server.Expect.PATCH("/users").
				WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
				WithJSON(map[string]any{
					"updateProfilePicture": map[string]any{"base64Data": payload},
				}).
				Expect().
				Status(http.StatusBadRequest)
			assert.Empty(t, listObjectVersions(ctx, t, client, bucket, prefix))
		})
	}
}

func TestLegacyMobileCameraUploadIsDownscaledServerSide(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	if raceEnabled {
		t.Skip("wall-clock compatibility is exercised without race/atomic-coverage instrumentation by make test")
	}

	client := configuredTestMinIOClient(t)
	bucket := common.MinIOBucketName()
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	ctx := context.Background()
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	prefix := fmt.Sprintf("users/%d-", userID)
	defer func() { require.NoError(t, common.MinIODeletePrefix(ctx, bucket, prefix)) }()

	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{
				"base64Data": legacyCameraJPEG(t, 3000, 3000),
				"extension":  "jpg",
			},
		}).
		Expect().
		Status(http.StatusOK)

	objects := listObjectVersions(ctx, t, client, bucket, prefix)
	require.Len(t, objects, 1)
	object, err := client.GetObject(ctx, bucket, objects[0].Key, minio.GetObjectOptions{})
	require.NoError(t, err)
	config, err := jpeg.DecodeConfig(object)
	require.NoError(t, err)
	require.NoError(t, object.Close())
	assert.Equal(t, common.MaxProfileImageOutputDimension, config.Width)
	assert.Equal(t, common.MaxProfileImageOutputDimension, config.Height)
}

func TestLegacyAndroidEditedPNGWithEXIFUploadsAndPersists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := configuredTestMinIOClient(t)
	bucket := common.MinIOBucketName()
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	ctx := context.Background()
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	prefix := fmt.Sprintf("users/%d-", userID)
	defer func() { require.NoError(t, common.MinIODeletePrefix(ctx, bucket, prefix)) }()

	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{
				"base64Data": legacyAndroidPNGWithEXIF(t),
				"extension":  "png",
			},
		}).
		Expect().
		Status(http.StatusOK)

	objects := listObjectVersions(ctx, t, client, bucket, prefix)
	require.Len(t, objects, 1)
	assert.True(t, strings.HasSuffix(objects[0].Key, ".png"))
}

func TestProfileUploadIntentAndReferenceCommitBeforeWorkerClaim(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := setup.NewTestServer(t)
	defer server.Cleanup()
	server.WithTestUser(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	objectName := fmt.Sprintf("users/%d-0123456789abcdef0123456789abcdef.jpg", userID)
	objectURL, err := common.MinIOObjectURL(common.MinIOBucketName(), objectName)
	require.NoError(t, err)
	jobID, err := server.AppContext.UserService.ScheduleUncertainProfileUploadCleanup(
		ctx, userID, common.MinIOBucketName(), objectName, objectURL,
	)
	require.NoError(t, err)
	receipt := common.UploadReceipt{
		Bucket: common.MinIOBucketName(), ObjectName: objectName,
		URL: objectURL, VersionID: "version-1",
	}

	blocker, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = blocker.Rollback(context.Background()) }()
	var lockedID int64
	require.NoError(t, blocker.QueryRow(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&lockedID))

	var updateUser *service.User
	var updateErr error
	updateDone := make(chan struct{})
	go func() {
		defer close(updateDone)
		updateUser, updateErr = server.AppContext.UserService.UpdateProfileWithUploadIntent(
			ctx, userID, nil, nil, nil, nil, &objectURL, nil, nil, nil, jobID, receipt,
		)
	}()

	require.Eventually(t, func() bool {
		var waiting bool
		queryErr := server.DB.QueryRow(ctx, `SELECT EXISTS(
			SELECT 1 FROM pg_stat_activity
			WHERE datname=current_database() AND wait_event_type='Lock'
			  AND query LIKE '%UPDATE users%'
		)`).Scan(&waiting)
		return queryErr == nil && waiting
	}, 3*time.Second, 20*time.Millisecond)

	workerStarted := make(chan struct{})
	workerDone := make(chan error, 1)
	var workerReferenced bool
	go func() {
		workerTx, beginErr := server.DB.Begin(ctx)
		if beginErr != nil {
			workerDone <- beginErr
			return
		}
		defer func() { _ = workerTx.Rollback(context.Background()) }()
		close(workerStarted)
		var claimedID int64
		if claimErr := workerTx.QueryRow(ctx, `SELECT id FROM river_job WHERE id=$1 FOR UPDATE`, jobID).Scan(&claimedID); claimErr != nil {
			workerDone <- claimErr
			return
		}
		if referenceErr := workerTx.QueryRow(ctx, `SELECT EXISTS(
			SELECT 1 FROM users WHERE id=$1 AND profile_picture_url=$2
		)`, userID, objectURL).Scan(&workerReferenced); referenceErr != nil {
			workerDone <- referenceErr
			return
		}
		workerDone <- workerTx.Commit(ctx)
	}()
	<-workerStarted
	select {
	case claimErr := <-workerDone:
		t.Fatalf("worker claimed uncommitted reconciliation intent: %v", claimErr)
	case <-time.After(100 * time.Millisecond):
	}

	require.NoError(t, blocker.Commit(ctx))
	<-updateDone
	require.NoError(t, updateErr)
	require.NotNil(t, updateUser)
	require.NoError(t, <-workerDone)
	assert.True(t, workerReferenced)
}

func TestMinIOExactVersionDeletePreservesOlderSameKeyVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := configuredTestMinIOClient(t)
	ctx := context.Background()
	bucket := fmt.Sprintf("profile-version-%d", time.Now().UnixNano())
	require.NoError(t, client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}))
	defer func() {
		require.NoError(t, common.MinIODeletePrefix(ctx, bucket, "users/"))
		require.NoError(t, client.RemoveBucket(ctx, bucket))
	}()
	require.NoError(t, client.EnableVersioning(ctx, bucket))

	first, err := common.MinIOPUTWithReceipt(ctx, []byte("first"), bucket, "users/42-object.png", "image/png")
	require.NoError(t, err)
	second, err := common.MinIOPUTWithReceipt(ctx, []byte("second"), bucket, "users/42-object.png", "image/png")
	require.NoError(t, err)
	require.NotEmpty(t, first.VersionID)
	require.NotEmpty(t, second.VersionID)
	require.NotEqual(t, first.VersionID, second.VersionID)

	require.NoError(t, common.MinIODeleteObjectVersion(ctx, bucket, second.ObjectName, second.VersionID))
	object, err := client.GetObject(ctx, bucket, first.ObjectName, minio.GetObjectOptions{})
	require.NoError(t, err)
	body, err := io.ReadAll(object)
	require.NoError(t, err)
	require.NoError(t, object.Close())
	assert.Equal(t, "first", string(body))

	_, err = common.MinIOPUTWithReceipt(ctx, []byte("sibling"), bucket, first.ObjectName+".sibling", "image/png")
	require.NoError(t, err)
	require.NoError(t, common.MinIODeleteObjectAllVersions(ctx, bucket, first.ObjectName))
	for _, remaining := range listObjectVersions(ctx, t, client, bucket, first.ObjectName) {
		assert.NotEqual(t, first.ObjectName, remaining.Key)
	}
	sibling, err := client.GetObject(ctx, bucket, first.ObjectName+".sibling", minio.GetObjectOptions{})
	require.NoError(t, err)
	siblingBody, err := io.ReadAll(sibling)
	require.NoError(t, err)
	require.NoError(t, sibling.Close())
	assert.Equal(t, "sibling", string(siblingBody))
}

func TestProfileUploadReconciliationHandlesUnversionedBucketWithoutTouchingSibling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := configuredTestMinIOClient(t)
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	server.WithTestUser(t)
	ctx := context.Background()
	bucket := fmt.Sprintf("profile-unversioned-%d", time.Now().UnixNano())
	require.NoError(t, client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}))
	defer func() {
		_ = common.MinIODeletePrefix(ctx, bucket, "users/")
		require.NoError(t, client.RemoveBucket(ctx, bucket))
		require.NoError(t, common.ConfigureMinIOFromEnvironment("test"))
	}()

	host := os.Getenv("MINIO_HOST")
	port := os.Getenv("MINIO_PORT")
	accessKey := os.Getenv("MINIO_ROOT_USER")
	if accessKey == "" {
		accessKey = os.Getenv("MINIO_ACCESS_KEY")
	}
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	if secretKey == "" {
		secretKey = os.Getenv("MINIO_SECRET_KEY")
	}
	require.NoError(t, common.ConfigureMinIO(common.MinIOConfig{
		Environment: "test", Endpoint: fmt.Sprintf("%s:%s", host, port),
		AccessKey: accessKey, SecretKey: secretKey, Bucket: bucket,
	}))

	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	target := fmt.Sprintf("users/%d-0123456789abcdef0123456789abcdef.jpg", userID)
	sibling := fmt.Sprintf("users/%d-fedcba9876543210fedcba9876543210.jpg", userID)
	for objectName, body := range map[string]string{target: "target", sibling: "sibling"} {
		_, err := client.PutObject(
			ctx, bucket, objectName, bytes.NewBufferString(body), int64(len(body)), minio.PutObjectOptions{},
		)
		require.NoError(t, err)
	}
	targetURL, err := common.MinIOObjectURL(bucket, target)
	require.NoError(t, err)
	_, err = server.DB.Exec(ctx, `UPDATE users SET profile_picture_url=$1 WHERE id=$2`, targetURL, userID)
	require.NoError(t, err)
	worker := service.NewProfileUploadReconciliationWorker(server.DB)
	job := &river.Job[service.ProfileUploadReconciliationArgs]{Args: service.ProfileUploadReconciliationArgs{
		UserID: userID, Bucket: bucket, ObjectName: target, ObjectURL: targetURL,
	}}

	// Empty VersionID is the normal receipt for an unversioned bucket. A live
	// database reference must preserve it.
	require.NoError(t, worker.Work(ctx, job))
	kept, err := client.GetObject(ctx, bucket, target, minio.GetObjectOptions{})
	require.NoError(t, err)
	keptBody, err := io.ReadAll(kept)
	require.NoError(t, err)
	require.NoError(t, kept.Close())
	assert.Equal(t, "target", string(keptBody))

	_, err = server.DB.Exec(ctx, `UPDATE users SET profile_picture_url=NULL WHERE id=$1`, userID)
	require.NoError(t, err)
	require.NoError(t, worker.Work(ctx, job))
	_, err = client.StatObject(ctx, bucket, target, minio.StatObjectOptions{})
	assert.Equal(t, "NoSuchKey", minio.ToErrorResponse(err).Code)
	siblingObject, err := client.GetObject(ctx, bucket, sibling, minio.GetObjectOptions{})
	require.NoError(t, err)
	siblingBody, err := io.ReadAll(siblingObject)
	require.NoError(t, err)
	require.NoError(t, siblingObject.Close())
	assert.Equal(t, "sibling", string(siblingBody))
}

func TestProfileUpdateRejectsOversizedRequestBeforeDecoding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)

	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{
				"base64Data": strings.Repeat("A", (8<<20)+1),
			},
		}).
		Expect().
		Status(http.StatusRequestEntityTooLarge)
}

func TestProfileUpdateMapsDecodedImageSizeViolationToPayloadTooLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	server.Expect.PATCH("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithJSON(map[string]any{
			"updateProfilePicture": map[string]any{
				"base64Data": strings.Repeat("A", base64.StdEncoding.EncodedLen(common.MaxProfileImageBytes+1)),
			},
		}).
		Expect().
		Status(http.StatusRequestEntityTooLarge)
}

func TestProfileUpdateRejectsChunkedOversizedTrailingContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	body := io.NopCloser(strings.NewReader(`{}` + strings.Repeat(" ", (8<<20)+1)))
	request, err := http.NewRequest(http.MethodPatch, server.BaseURL+"/users", body)
	require.NoError(t, err)
	request.ContentLength = -1
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, http.StatusRequestEntityTooLarge, response.StatusCode)
}

func TestProfileUpdateRejectsASecondJSONValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	request, err := http.NewRequest(http.MethodPatch, server.BaseURL+"/users", strings.NewReader(`{} {}`))
	require.NoError(t, err)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestCanceledProfileUpdateDoesNotMutateAndSchedulesExactReconciliation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	configuredTestMinIOClient(t)
	bucket := common.MinIOBucketName()
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	token := server.WithTestUser(t)
	ctx := context.Background()
	var userID int64
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID))
	prefix := fmt.Sprintf("users/%d-", userID)
	defer func() { require.NoError(t, common.MinIODeletePrefix(ctx, bucket, prefix)) }()

	_, err := server.DB.Exec(ctx, fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION test_slow_profile_%d_update()
		RETURNS trigger AS $$
		BEGIN
			PERFORM pg_sleep(5);
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER test_slow_profile_%d_update_trigger
		BEFORE UPDATE ON users
		FOR EACH ROW WHEN (OLD.id = %d)
		EXECUTE FUNCTION test_slow_profile_%d_update();
	`, userID, userID, userID, userID))
	require.NoError(t, err)
	defer func() {
		_, cleanupErr := server.DB.Exec(ctx, fmt.Sprintf(`
			DROP TRIGGER IF EXISTS test_slow_profile_%d_update_trigger ON users;
			DROP FUNCTION IF EXISTS test_slow_profile_%d_update();
		`, userID, userID))
		require.NoError(t, cleanupErr)
	}()

	payload, err := json.Marshal(map[string]any{
		"updateProfilePicture": map[string]any{"base64Data": validProfilePNG(t)},
	})
	require.NoError(t, err)
	requestContext, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodPatch, server.BaseURL+"/users", bytes.NewReader(payload))
	require.NoError(t, err)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")
	response, requestErr := http.DefaultClient.Do(request)
	if response != nil {
		require.NoError(t, response.Body.Close())
	}
	require.ErrorIs(t, requestErr, context.DeadlineExceeded)

	require.Eventually(t, func() bool {
		var profileURL string
		if queryErr := server.DB.QueryRow(ctx, `SELECT COALESCE(profile_picture_url, '') FROM users WHERE id=$1`, userID).Scan(&profileURL); queryErr != nil || profileURL != "" {
			return false
		}
		var jobs int
		if queryErr := server.DB.QueryRow(ctx, `
			SELECT COUNT(*) FROM river_job
			WHERE kind='profile_upload_reconciliation_v1'
			  AND args->>'user_id'=$1
			  AND args->>'object_name' LIKE $2
		`, fmt.Sprint(userID), prefix+"%").Scan(&jobs); queryErr != nil {
			return false
		}
		return jobs == 1
	}, 3*time.Second, 50*time.Millisecond)
}

func validProfilePNG(t *testing.T) string {
	t.Helper()
	var encoded bytes.Buffer
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.NRGBA{R: 255, A: 255})
	require.NoError(t, png.Encode(&encoded, img))
	return base64.StdEncoding.EncodeToString(encoded.Bytes())
}

func profileJPEGWithMetadata(t *testing.T, marker byte, payload []byte) string {
	t.Helper()
	var encoded bytes.Buffer
	require.NoError(t, jpeg.Encode(&encoded, image.NewNRGBA(image.Rect(0, 0, 2, 2)), &jpeg.Options{Quality: 80}))
	segment := []byte{0xff, marker, 0, 0}
	binary.BigEndian.PutUint16(segment[2:4], uint16(len(payload)+2))
	result := append([]byte(nil), encoded.Bytes()[:2]...)
	result = append(result, segment...)
	result = append(result, payload...)
	result = append(result, encoded.Bytes()[2:]...)
	return base64.StdEncoding.EncodeToString(result)
}

func profilePNGWithMetadata(t *testing.T, chunkType string, payload []byte) string {
	t.Helper()
	encoded, err := base64.StdEncoding.DecodeString(validProfilePNG(t))
	require.NoError(t, err)
	chunk := make([]byte, 12+len(payload))
	binary.BigEndian.PutUint32(chunk[:4], uint32(len(payload)))
	copy(chunk[4:8], chunkType)
	copy(chunk[8:8+len(payload)], payload)
	binary.BigEndian.PutUint32(chunk[8+len(payload):], crc32.ChecksumIEEE(chunk[4:8+len(payload)]))
	result := append([]byte(nil), encoded[:33]...)
	result = append(result, chunk...)
	result = append(result, encoded[33:]...)
	return base64.StdEncoding.EncodeToString(result)
}

func legacyCameraJPEG(t *testing.T, width, height int) string {
	t.Helper()
	var encoded bytes.Buffer
	require.NoError(t, jpeg.Encode(
		&encoded, image.NewNRGBA(image.Rect(0, 0, width, height)), &jpeg.Options{Quality: 80},
	))
	return base64.StdEncoding.EncodeToString(addLegacyAndroidEXIF(t, encoded.Bytes()))
}

func addLegacyAndroidEXIF(t *testing.T, source []byte) []byte {
	t.Helper()
	const rootOffset, makeOffset, modelOffset, exifOffset = 8, 62, 68, 74
	const dateOffset, makerOffset, commentOffset = 116, 136, 144
	tiff := make([]byte, 160)
	copy(tiff[:2], "II")
	binary.LittleEndian.PutUint16(tiff[2:4], 42)
	binary.LittleEndian.PutUint32(tiff[4:8], rootOffset)
	binary.LittleEndian.PutUint16(tiff[rootOffset:rootOffset+2], 4)
	writeLegacyEXIFEntry(tiff[10:22], 0x0112, 3, 1, 6)
	writeLegacyEXIFEntry(tiff[22:34], 0x010f, 2, 6, makeOffset)
	writeLegacyEXIFEntry(tiff[34:46], 0x0110, 2, 6, modelOffset)
	writeLegacyEXIFEntry(tiff[46:58], 0x8769, 4, 1, exifOffset)
	copy(tiff[makeOffset:modelOffset], "Canon\x00")
	copy(tiff[modelOffset:exifOffset], "Pixel\x00")
	binary.LittleEndian.PutUint16(tiff[exifOffset:exifOffset+2], 3)
	writeLegacyEXIFEntry(tiff[76:88], 0x9003, 2, 20, dateOffset)
	writeLegacyEXIFEntry(tiff[88:100], 0x927c, 7, 8, makerOffset)
	writeLegacyEXIFEntry(tiff[100:112], 0x9286, 7, 16, commentOffset)
	copy(tiff[dateOffset:makerOffset], "2026:08:08 12:00:00\x00")
	copy(tiff[makerOffset:commentOffset], "MakerNot")
	copy(tiff[commentOffset:], "ASCII\x00\x00\x00ordinary")
	payload := append([]byte("Exif\x00\x00"), tiff...)
	segment := []byte{0xff, 0xe1, 0, 0}
	binary.BigEndian.PutUint16(segment[2:4], uint16(len(payload)+2))
	result := append([]byte(nil), source[:2]...)
	result = append(result, segment...)
	result = append(result, payload...)
	return append(result, source[2:]...)
}

func legacyAndroidPNGWithEXIF(t *testing.T) string {
	t.Helper()
	var encodedPNG bytes.Buffer
	require.NoError(t, png.Encode(&encodedPNG, image.NewNRGBA(image.Rect(0, 0, 2, 2))))
	var encodedJPEG bytes.Buffer
	require.NoError(t, jpeg.Encode(&encodedJPEG, image.NewNRGBA(image.Rect(0, 0, 2, 2)), &jpeg.Options{Quality: 80}))
	jpegWithEXIF := addLegacyAndroidEXIF(t, encodedJPEG.Bytes())
	tiff := jpegWithEXIF[12:172]
	chunk := make([]byte, 12+len(tiff))
	binary.BigEndian.PutUint32(chunk[:4], uint32(len(tiff)))
	copy(chunk[4:8], "eXIf")
	copy(chunk[8:8+len(tiff)], tiff)
	binary.BigEndian.PutUint32(chunk[8+len(tiff):], crc32.ChecksumIEEE(chunk[4:8+len(tiff)]))
	result := append([]byte(nil), encodedPNG.Bytes()[:33]...)
	result = append(result, chunk...)
	result = append(result, encodedPNG.Bytes()[33:]...)
	return base64.StdEncoding.EncodeToString(result)
}

func writeLegacyEXIFEntry(target []byte, tag, valueType uint16, count, value uint32) {
	binary.LittleEndian.PutUint16(target[0:2], tag)
	binary.LittleEndian.PutUint16(target[2:4], valueType)
	binary.LittleEndian.PutUint32(target[4:8], count)
	if valueType == 3 && count == 1 {
		binary.LittleEndian.PutUint16(target[8:10], uint16(value))
		return
	}
	binary.LittleEndian.PutUint32(target[8:12], value)
}

func hostileProfileJPEGWithOversizedDHT(t *testing.T) string {
	t.Helper()
	var encoded bytes.Buffer
	require.NoError(t, jpeg.Encode(&encoded, image.NewNRGBA(image.Rect(0, 0, 2, 2)), &jpeg.Options{Quality: 80}))
	counts := make([]byte, 16)
	counts[8] = 255
	counts[9] = 2
	payload := append([]byte{0}, counts...)
	payload = append(payload, make([]byte, 257)...)
	dht := []byte{0xff, 0xc4, 0, 0}
	binary.BigEndian.PutUint16(dht[2:4], uint16(len(payload)+2))
	dht = append(dht, payload...)
	result := append(append(append([]byte(nil), encoded.Bytes()[:2]...), dht...), encoded.Bytes()[2:]...)
	return base64.StdEncoding.EncodeToString(result)
}

func configuredTestMinIOClient(t *testing.T) *minio.Client {
	t.Helper()
	host := os.Getenv("MINIO_HOST")
	port := os.Getenv("MINIO_PORT")
	accessKey := os.Getenv("MINIO_ROOT_USER")
	if accessKey == "" {
		accessKey = os.Getenv("MINIO_ACCESS_KEY")
	}
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	if secretKey == "" {
		secretKey = os.Getenv("MINIO_SECRET_KEY")
	}
	if host == "" || port == "" || accessKey == "" || secretKey == "" {
		t.Skip("MinIO integration environment is unavailable")
	}
	require.NoError(t, common.ConfigureMinIOFromEnvironment("test"))
	client, err := minio.New(fmt.Sprintf("%s:%s", host, port), &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
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
