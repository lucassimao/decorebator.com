package http

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type blockingProfileBody struct {
	closed chan struct{}
	once   sync.Once
}

func (b *blockingProfileBody) Read([]byte) (int, error) {
	<-b.closed
	return 0, errors.New("profile request body closed")
}

func (b *blockingProfileBody) Close() error {
	b.once.Do(func() { close(b.closed) })
	return nil
}

func TestProfileUploadAdmissionIsFailFastPerUserAndGloballyBounded(t *testing.T) {
	t.Parallel()
	admission := newProfileUploadAdmission(2)

	releaseOne, ok := admission.tryAcquire(1)
	require.True(t, ok)
	_, duplicateUser := admission.tryAcquire(1)
	assert.False(t, duplicateUser)

	releaseTwo, ok := admission.tryAcquire(2)
	require.True(t, ok)
	_, overGlobalLimit := admission.tryAcquire(3)
	assert.False(t, overGlobalLimit)

	releaseOne()
	releaseThree, ok := admission.tryAcquire(3)
	require.True(t, ok)
	releaseThree()
	releaseTwo()
}

func TestProfileInputRejectsLanguageLongerThanStorageBeforeUpload(t *testing.T) {
	language := "language-tag"
	require.EqualError(t, validateProfileInputBounds(UpdateProfileInput{
		PreferredLanguage: &language,
	}), "preferred language is too long")
}

func TestUpdateProfileMapsBodyAdmissionSaturationTo503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routes := &UserRoutes{
		profileBodies:  make(chan struct{}, 1),
		profileUploads: newProfileUploadAdmission(1),
	}
	routes.profileBodies <- struct{}{}
	request := httptest.NewRequest(http.MethodPatch, "/users", strings.NewReader(`{"firstName":"safe"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = request
	ginContext.Set("user", &model.User{ID: 42})

	routes.UpdateProfile(ginContext)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.Equal(t, "1", recorder.Header().Get("Retry-After"))
	assert.Len(t, routes.profileBodies, 1)
	<-routes.profileBodies
}

func TestUpdateProfileMapsPerUserUploadAdmissionSaturationTo429AndReleasesBodySlot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routes := &UserRoutes{
		profileBodies:  make(chan struct{}, 1),
		profileUploads: newProfileUploadAdmission(1),
	}
	release, acquired := routes.profileUploads.tryAcquire(42)
	require.True(t, acquired)
	defer release()
	request := httptest.NewRequest(http.MethodPatch, "/users", strings.NewReader(
		`{"updateProfilePicture":{"base64Data":"aW1hZ2U="}}`,
	))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = request
	ginContext.Set("user", &model.User{ID: 42})

	routes.UpdateProfile(ginContext)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Equal(t, "1", recorder.Header().Get("Retry-After"))
	assert.Empty(t, routes.profileBodies)
}

func TestUpdateProfileMapsRasterAdmissionSaturationTo429WithoutPersistence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routes := &UserRoutes{
		profileBodies:     make(chan struct{}, 1),
		profileUploads:    newProfileUploadAdmission(1),
		profileUserExists: func(context.Context, int64) (bool, error) { return true, nil },
		normalizeProfileImage: func(context.Context, string) (common.ProfileImage, error) {
			return common.ProfileImage{}, common.ErrProfileImageBusy
		},
	}
	request := httptest.NewRequest(http.MethodPatch, "/users", strings.NewReader(
		`{"updateProfilePicture":{"base64Data":"aW1hZ2U="}}`,
	))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = request
	ginContext.Set("user", &model.User{ID: 42})

	routes.UpdateProfile(ginContext)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Equal(t, "1", recorder.Header().Get("Retry-After"))
	assert.Empty(t, routes.profileBodies)
	release, acquired := routes.profileUploads.tryAcquire(42)
	require.True(t, acquired)
	release()
}

func TestTrailingJSONRejectionUsesFixedAllocation(t *testing.T) {
	payload := []byte(`{"firstName":"safe"}` + strings.Repeat(`["nested-value"],`, 50_000))
	allocations := testing.AllocsPerRun(5, func() {
		body := bytes.NewReader(payload)
		decoder := json.NewDecoder(body)
		var input UpdateProfileInput
		require.NoError(t, decoder.Decode(&input))
		require.Error(t, requireJSONWhitespaceEOF(decoder, body))
	})
	assert.Less(t, allocations, float64(50))
}

func TestProfileBodyReadStopsAtRouteContextDeadlineAndReleasesAdmission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := &blockingProfileBody{closed: make(chan struct{})}
	request := httptest.NewRequest(http.MethodPatch, "/users", body)
	ctx, cancel := context.WithTimeout(request.Context(), 25*time.Millisecond)
	defer cancel()
	request = request.WithContext(ctx)
	request.ContentLength = -1
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = request
	ginContext.Set("user", &model.User{ID: 42})
	routes := &UserRoutes{
		profileBodies:  make(chan struct{}, 1),
		profileUploads: newProfileUploadAdmission(1),
	}

	started := time.Now()
	routes.UpdateProfile(ginContext)

	assert.Equal(t, http.StatusRequestTimeout, recorder.Code)
	assert.Less(t, time.Since(started), 500*time.Millisecond)
	assert.Empty(t, routes.profileBodies)
}

func TestRealConnectionPartialProfileBodyReturnsAtRouteDeadline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routes := &UserRoutes{
		profileBodies:  make(chan struct{}, 1),
		profileUploads: newProfileUploadAdmission(1),
	}
	engine := gin.New()
	engine.Use(TimeoutMiddleware(75 * time.Millisecond))
	engine.PATCH("/users", func(c *gin.Context) {
		c.Set("user", &model.User{ID: 42})
		routes.UpdateProfile(c)
	})
	server := httptest.NewServer(engine)
	defer server.Close()

	address := strings.TrimPrefix(server.URL, "http://")
	connection, err := net.Dial("tcp", address)
	require.NoError(t, err)
	defer connection.Close()
	started := time.Now()
	_, err = fmt.Fprintf(connection, "PATCH /users HTTP/1.1\r\nHost: %s\r\nContent-Type: application/json\r\nContent-Length: 100\r\n\r\n{", address)
	require.NoError(t, err)
	require.NoError(t, connection.SetReadDeadline(time.Now().Add(time.Second)))
	response, err := http.ReadResponse(bufio.NewReader(connection), nil)
	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusRequestTimeout, response.StatusCode)
	assert.Less(t, time.Since(started), 500*time.Millisecond)
	require.Eventually(t, func() bool { return len(routes.profileBodies) == 0 }, time.Second, 5*time.Millisecond)
}

func TestProfileCleanupCannotOutliveRouteDeadlineAndMapsTimeoutAfterCleanup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := runProfileCleanup(ctx, func(cleanupContext context.Context) error {
		<-cleanupContext.Done()
		return cleanupContext.Err()
	})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(started), 250*time.Millisecond)
	assert.True(t, profileRequestTimedOut(ctx, errors.New("persistence rejected")))
}

func TestProfilePersistenceFailureClassificationDefaultsUnknownErrorsToAmbiguous(t *testing.T) {
	assert.False(t, isDefinitiveProfilePersistenceFailure(errors.New("connection reset after write")))
	assert.True(t, isDefinitiveProfilePersistenceFailure(common.BusinessError{Message: "invalid update"}))
	assert.True(t, isDefinitiveProfilePersistenceFailure(&pgconn.PgError{Code: "23514"}))
}
