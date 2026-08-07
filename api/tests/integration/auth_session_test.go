package integration

import (
	"bytes"
	"net/http"
	"sync"
	"testing"

	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nativeSessionCredentials struct {
	access   string
	refresh  string
	password string
}

func TestAuthSessionRotationReplayAndLogout(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	credentials := createNativeSession(t, server)
	var storedHash []byte
	require.NoError(t, server.DB.QueryRow(t.Context(), `
		SELECT token_hash FROM auth_refresh_tokens WHERE consumed_at IS NULL
	`).Scan(&storedHash))
	assert.Len(t, storedHash, 32)
	assert.False(t, bytes.Equal(storedHash, []byte(credentials.refresh)))

	rotated := refreshNativeSession(t, server, credentials.refresh, http.StatusOK)
	assert.NotEqual(t, credentials.refresh, rotated.refresh)
	server.Expect.GET("/users").WithHeader("Authorization", rotated.access).Expect().Status(http.StatusOK)

	refreshNativeSession(t, server, credentials.refresh, http.StatusUnauthorized)
	server.Expect.GET("/users").WithHeader("Authorization", rotated.access).Expect().Status(http.StatusUnauthorized)

	other := createNativeSession(t, server)
	server.Expect.POST("/logout").
		WithHeader("X-Auth-Client", "native").
		WithHeader("X-Refresh-Token", other.refresh).
		Expect().Status(http.StatusOK)
	server.Expect.GET("/users").WithHeader("Authorization", other.access).Expect().Status(http.StatusUnauthorized)

	passwordSession := createNativeSession(t, server)
	server.Expect.PATCH("/users").
		WithHeader("Authorization", passwordSession.access).
		WithJSON(map[string]any{"updatePassword": map[string]string{
			"currentPassword": passwordSession.password,
			"newPassword":     "replacement-password-123",
		}}).
		Expect().Status(http.StatusOK)
	server.Expect.GET("/users").WithHeader("Authorization", passwordSession.access).Expect().Status(http.StatusUnauthorized)
	refreshNativeSession(t, server, passwordSession.refresh, http.StatusUnauthorized)
}

func TestConcurrentRefreshReplayRevokesWinningReplacement(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	credentials := createNativeSession(t, server)

	type result struct {
		status int
		access string
		err    error
	}
	results := make(chan result, 2)
	var start sync.WaitGroup
	start.Add(1)
	for range 2 {
		go func() {
			start.Wait()
			request, err := http.NewRequest(http.MethodPost, server.BaseURL+"/session/refresh", nil)
			if err != nil {
				results <- result{err: err}
				return
			}
			request.Header.Set("X-Auth-Client", "native")
			request.Header.Set("X-Refresh-Token", credentials.refresh)
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				results <- result{err: err}
				return
			}
			defer response.Body.Close()
			results <- result{status: response.StatusCode, access: response.Header.Get("Authorization")}
		}()
	}
	start.Done()
	first, second := <-results, <-results
	require.NoError(t, first.err)
	require.NoError(t, second.err)
	statuses := []int{first.status, second.status}
	assert.ElementsMatch(t, []int{http.StatusOK, http.StatusUnauthorized}, statuses)
	winner := first
	if winner.status != http.StatusOK {
		winner = second
	}
	require.NotEmpty(t, winner.access)
	server.Expect.GET("/users").WithHeader("Authorization", winner.access).Expect().Status(http.StatusUnauthorized)
}

func TestExpiredRefreshFamilyFailsClosed(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	credentials := createNativeSession(t, server)
	claims := setup.DecodeJWT(t, credentials.access)
	familyID, ok := claims["sid"].(string)
	require.True(t, ok)
	_, err := server.DB.Exec(t.Context(), `
		UPDATE auth_session_families
		SET idle_expires_at=created_at,absolute_expires_at=created_at
		WHERE id=$1
	`, familyID)
	require.Error(t, err, "database constraints must prevent invalid in-place expiry corruption")
	_, err = server.DB.Exec(t.Context(), `
		UPDATE auth_session_families
		SET created_at=created_at-INTERVAL '91 days',
			idle_expires_at=NOW()-INTERVAL '1 second',
			absolute_expires_at=NOW()-INTERVAL '1 second'
		WHERE id=$1
	`, familyID)
	require.NoError(t, err)
	refreshNativeSession(t, server, credentials.refresh, http.StatusUnauthorized)
	server.Expect.GET("/users").WithHeader("Authorization", credentials.access).Expect().Status(http.StatusUnauthorized)
}

func TestSignupRollsBackWhenSessionPersistenceFails(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	_, err := server.DB.Exec(t.Context(), `
		CREATE FUNCTION fail_auth_refresh_insert() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'forced auth refresh failure'; END $$;
		CREATE TRIGGER fail_auth_refresh_insert
		BEFORE INSERT ON auth_refresh_tokens
		FOR EACH ROW EXECUTE FUNCTION fail_auth_refresh_insert();
	`)
	require.NoError(t, err)
	defer func() {
		_, _ = server.DB.Exec(t.Context(), `
			DROP TRIGGER IF EXISTS fail_auth_refresh_insert ON auth_refresh_tokens;
			DROP FUNCTION IF EXISTS fail_auth_refresh_insert();
		`)
	}()
	signup := setup.GenerateSignupInput()
	server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusInternalServerError)
	var users int
	require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT COUNT(*) FROM users WHERE email=$1`, signup.Email).Scan(&users))
	assert.Zero(t, users)
}

func TestRefreshInfrastructureFailurePreservesCredential(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	credentials := createNativeSession(t, server)
	_, err := server.DB.Exec(t.Context(), `
		CREATE FUNCTION fail_rotated_refresh_insert() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			IF EXISTS (SELECT 1 FROM auth_refresh_tokens WHERE family_id=NEW.family_id) THEN
				RAISE EXCEPTION 'forced rotated refresh failure';
			END IF;
			RETURN NEW;
		END $$;
		CREATE TRIGGER fail_rotated_refresh_insert
		BEFORE INSERT ON auth_refresh_tokens
		FOR EACH ROW EXECUTE FUNCTION fail_rotated_refresh_insert();
	`)
	require.NoError(t, err)
	defer func() {
		_, _ = server.DB.Exec(t.Context(), `
			DROP TRIGGER IF EXISTS fail_rotated_refresh_insert ON auth_refresh_tokens;
			DROP FUNCTION IF EXISTS fail_rotated_refresh_insert();
		`)
	}()

	refreshNativeSession(t, server, credentials.refresh, http.StatusServiceUnavailable)

	_, err = server.DB.Exec(t.Context(), `
		DROP TRIGGER IF EXISTS fail_rotated_refresh_insert ON auth_refresh_tokens;
		DROP FUNCTION IF EXISTS fail_rotated_refresh_insert();
	`)
	require.NoError(t, err)
	rotated := refreshNativeSession(t, server, credentials.refresh, http.StatusOK)
	server.Expect.GET("/users").WithHeader("Authorization", rotated.access).Expect().Status(http.StatusOK)
}

func createNativeSession(t *testing.T, server *setup.TestServer) nativeSessionCredentials {
	t.Helper()
	signup := setup.GenerateSignupInput()
	response := server.Expect.POST("/users").
		WithHeader("X-Auth-Client", "native").
		WithJSON(signup).
		Expect().Status(http.StatusCreated)
	return nativeSessionCredentials{
		access:   response.Header("Authorization").NotEmpty().Raw(),
		refresh:  response.Header("X-Refresh-Token").NotEmpty().Raw(),
		password: signup.Password,
	}
}

func refreshNativeSession(
	t *testing.T,
	server *setup.TestServer,
	refresh string,
	status int,
) nativeSessionCredentials {
	t.Helper()
	response := server.Expect.POST("/session/refresh").
		WithHeader("X-Auth-Client", "native").
		WithHeader("X-Refresh-Token", refresh).
		Expect().Status(status)
	if status != http.StatusOK {
		return nativeSessionCredentials{}
	}
	return nativeSessionCredentials{
		access:  response.Header("Authorization").NotEmpty().Raw(),
		refresh: response.Header("X-Refresh-Token").NotEmpty().Raw(),
	}
}
