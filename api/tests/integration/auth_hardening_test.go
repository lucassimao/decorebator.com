package integration

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"decorebator.com/internal/app"
	httphandlers "decorebator.com/internal/http"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type failingResetEmailJobService struct {
	service.JobService
}

func (f failingResetEmailJobService) ScheduleResetPasswordEmailJob(
	context.Context,
	string,
	...pgx.Tx,
) error {
	return errors.New("forced reset-email job failure")
}

func TestPendingSignupRequiresRealResetActivationBeforeLogin(t *testing.T) {
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", strings.Repeat("k", 32))
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	signup := setup.GenerateSignupInput()
	server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusCreated)
	server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
		Email: signup.Email, Password: signup.Password,
	}).Expect().Status(http.StatusBadRequest)

	var userID int64
	var storedPasswordHash string
	require.NoError(t, server.DB.QueryRow(
		t.Context(), `SELECT id,password_hash FROM users WHERE email=$1`, signup.Email,
	).Scan(&userID, &storedPasswordHash))
	require.Error(t, bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(signup.Password)),
		"a pre-AUTH-3 binary must not authenticate the submitted pending password")
	token, err := mail.IssueResetPasswordToken(t.Context(), server.DB, userID)
	require.NoError(t, err)
	newPassword := "activated-password"
	server.Expect.PATCH("/password/reset").WithJSON(map[string]string{
		"token": token, "password": newPassword,
	}).Expect().Status(http.StatusOK)
	server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
		Email: signup.Email, Password: newPassword,
	}).Expect().Status(http.StatusOK)

	var pending bool
	require.NoError(t, server.DB.QueryRow(t.Context(), `
		SELECT EXISTS(SELECT 1 FROM pending_email_verifications WHERE user_id=$1)
	`, userID).Scan(&pending))
	require.False(t, pending)
}

func TestNewBinaryConsumesLegacyResetLinkOnce(t *testing.T) {
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", strings.Repeat("k", 32))
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	user, err := server.AppContext.UserService.SaveUser(
		t.Context(), "Legacy", "Reset", "original-password", "legacy-reset@example.com", nil, nil,
	)
	require.NoError(t, err)
	legacyToken := createLegacyResetToken(t, user.ID)
	payload, err := mail.ValidateResetPasswordPayload(legacyToken)
	require.NoError(t, err)
	require.True(t, payload.Legacy)

	server.Expect.PATCH("/password/reset").WithJSON(map[string]string{
		"token": legacyToken, "password": "legacy-updated-password",
	}).Expect().Status(http.StatusOK)
	server.Expect.PATCH("/password/reset").WithJSON(map[string]string{
		"token": legacyToken, "password": "legacy-replayed-password",
	}).Expect().Status(http.StatusBadRequest).
		JSON().Object().Value("error").String().IsEqual("token_invalid")
	caseVariant := strings.ToUpper(legacyToken)
	require.NotEqual(t, legacyToken, caseVariant, "fixture must exercise an alternate hex representation")
	server.Expect.PATCH("/password/reset").WithJSON(map[string]string{
		"token": caseVariant, "password": "case-replayed-password",
	}).Expect().Status(http.StatusBadRequest)
	server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
		Email: user.Email, Password: "legacy-updated-password",
	}).Expect().Status(http.StatusOK)
}

func TestConcurrentOldPasswordLoginCannotOutlivePasswordReset(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	user, err := server.AppContext.UserService.SaveUser(
		t.Context(), "Login", "Reset Race", "old-password", "login-reset-race@example.com", nil, nil,
	)
	require.NoError(t, err)

	blocker, err := server.DB.Begin(t.Context())
	require.NoError(t, err)
	require.NoError(t, func() error {
		_, lockErr := blocker.Exec(t.Context(), `LOCK TABLE auth_session_families IN ACCESS EXCLUSIVE MODE`)
		return lockErr
	}())
	defer func() { _ = blocker.Rollback(context.WithoutCancel(t.Context())) }()

	type loginResult struct {
		credentials service.SessionCredentials
		err         error
	}
	loginDone := make(chan loginResult, 1)
	go func() {
		credentials, loginErr := server.AppContext.UserService.LoginUser(
			t.Context(), user.Email, "old-password",
		)
		loginDone <- loginResult{credentials: credentials, err: loginErr}
	}()
	waitForBlockedQuery(t, server.DB, "INSERT INTO auth_session_families")

	resetDone := make(chan error, 1)
	go func() {
		resetDone <- server.AppContext.UserService.UpdatePassword(t.Context(), user.ID, "new-password")
	}()
	waitForBlockedQuery(t, server.DB, "UPDATE users SET password_hash")

	require.NoError(t, blocker.Commit(t.Context()))
	login := <-loginDone
	require.NoError(t, login.err)
	require.NotEmpty(t, login.credentials.AccessToken)
	require.NoError(t, <-resetDone)
	_, err = server.AppContext.AuthSessions.ValidateAccess(t.Context(), login.credentials.AccessToken)
	require.Error(t, err, "the reset must revoke a session created by the racing old-password login")
}

func waitForBlockedQuery(t *testing.T, db *pgxpool.Pool, queryFragment string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var blocked bool
		err := db.QueryRow(t.Context(), `
			SELECT EXISTS(
				SELECT 1 FROM pg_stat_activity
				WHERE datname=current_database()
					AND wait_event_type='Lock'
					AND position($1 in query) > 0
			)
		`, queryFragment).Scan(&blocked)
		require.NoError(t, err)
		if blocked {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("query did not block as expected: %s", queryFragment)
}

func createLegacyResetToken(t *testing.T, userID int64) string {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"userId": userID, "expiresAt": time.Now().UTC().Add(30 * time.Minute),
	})
	require.NoError(t, err)
	block, err := aes.NewCipher([]byte(strings.Repeat("k", 32)))
	require.NoError(t, err)
	aead, err := cipher.NewGCM(block)
	require.NoError(t, err)
	nonce := make([]byte, aead.NonceSize())
	ciphertext := aead.Seal(nonce, nonce, payload, nil)
	return hex.EncodeToString(ciphertext)
}

func TestPendingSignupRollsBackWhenActivationJobCannotCommit(t *testing.T) {
	server := setup.NewTestServer(t, func(builder *app.ContextBuilder) *app.ContextBuilder {
		return builder.WithJobService(failingResetEmailJobService{})
	})
	defer server.Cleanup()
	signup := setup.GenerateSignupInput()
	server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusServiceUnavailable)
	var users, pending int
	require.NoError(t, server.DB.QueryRow(
		t.Context(), `SELECT count(*) FROM users WHERE email=$1`, signup.Email,
	).Scan(&users))
	require.NoError(t, server.DB.QueryRow(
		t.Context(), `SELECT count(*) FROM pending_email_verifications`,
	).Scan(&pending))
	require.Zero(t, users)
	require.Zero(t, pending)
}

func TestAuthHardeningMigrationDownQuiescesWritersAndReUp(t *testing.T) {
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", strings.Repeat("k", 32))
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	downSQL, err := os.ReadFile("../../cmd/migrate/migrations/000080_add_pending_email_verifications.down.sql")
	require.NoError(t, err)
	upSQL, err := os.ReadFile("../../cmd/migrate/migrations/000080_add_pending_email_verifications.up.sql")
	require.NoError(t, err)

	pending := setup.GenerateSignupInput()
	server.Expect.POST("/users").WithJSON(pending).Expect().Status(http.StatusCreated)
	active, err := server.AppContext.UserService.SaveUser(
		t.Context(), "Rollback", "Active", "active-password", "rollback-active@example.com", nil, nil,
	)
	require.NoError(t, err)
	_, err = mail.IssueResetPasswordToken(t.Context(), server.DB, active.ID, "rollback-drill")
	require.NoError(t, err)

	restored := false
	defer func() {
		if !restored {
			_ = execMigrationSQL(context.WithoutCancel(t.Context()), server.DB, upSQL)
		}
	}()
	require.NoError(t, execMigrationSQL(t.Context(), server.DB, downSQL))

	var pendingUsers, pendingRows, activeTokens int
	require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT count(*) FROM users WHERE email=$1`, pending.Email).Scan(&pendingUsers))
	require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT count(*) FROM pending_email_verifications`).Scan(&pendingRows))
	require.NoError(t, server.DB.QueryRow(t.Context(), `
		SELECT count(*) FROM password_reset_tokens
		WHERE user_id=$1 AND consumed_at IS NULL
	`, active.ID).Scan(&activeTokens))
	require.Zero(t, pendingUsers)
	require.Zero(t, pendingRows)
	require.Zero(t, activeTokens)

	blocked := setup.GenerateSignupInput()
	err = server.AppContext.UserService.RegisterPendingUser(
		t.Context(), blocked.FirstName, blocked.LastName, blocked.Password, blocked.Email,
		nil, nil,
	)
	require.ErrorIs(t, err, service.ErrAuthHardeningWritesDisabled)
	_, err = mail.IssueResetPasswordToken(t.Context(), server.DB, active.ID, "blocked-after-down")
	require.ErrorIs(t, err, mail.ErrAuthHardeningWritesDisabled)

	require.NoError(t, execMigrationSQL(t.Context(), server.DB, upSQL))
	var writesEnabled bool
	require.NoError(t, server.DB.QueryRow(t.Context(), `
		SELECT writes_enabled FROM auth_hardening_rollout_state WHERE singleton=TRUE
	`).Scan(&writesEnabled))
	require.False(t, writesEnabled, "migration up must wait for the old-binary drain owner gate")
	require.NoError(t, enableAuthHardeningWrites(t.Context(), server.DB))
	restored = true
	require.NoError(t, server.AppContext.UserService.RegisterPendingUser(
		t.Context(), blocked.FirstName, blocked.LastName, blocked.Password, blocked.Email,
		nil, nil,
	))
}

func TestAuthHardeningMigrationDownSerializesWithPendingActivation(t *testing.T) {
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", strings.Repeat("k", 32))
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	downSQL, err := os.ReadFile("../../cmd/migrate/migrations/000080_add_pending_email_verifications.down.sql")
	require.NoError(t, err)
	upSQL, err := os.ReadFile("../../cmd/migrate/migrations/000080_add_pending_email_verifications.up.sql")
	require.NoError(t, err)
	restored := false
	defer func() {
		if !restored {
			_ = execMigrationSQL(context.WithoutCancel(t.Context()), server.DB, upSQL)
		}
	}()

	signup := setup.GenerateSignupInput()
	server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusCreated)
	var userID int64
	require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT id FROM users WHERE email=$1`, signup.Email).Scan(&userID))
	token, err := mail.IssueResetPasswordToken(t.Context(), server.DB, userID, "rollback-activation")
	require.NoError(t, err)
	payload, err := mail.ValidateResetPasswordPayload(token)
	require.NoError(t, err)

	tokenBlocker, err := server.DB.Begin(t.Context())
	require.NoError(t, err)
	defer func() { _ = tokenBlocker.Rollback(context.WithoutCancel(t.Context())) }()
	var lockedHash []byte
	require.NoError(t, tokenBlocker.QueryRow(t.Context(), `
		SELECT token_hash FROM password_reset_tokens
		WHERE delivery_key='rollback-activation'
		FOR UPDATE
	`).Scan(&lockedHash))

	activationDone := make(chan error, 1)
	go func() {
		activationDone <- server.AppContext.UserService.ResetPasswordAndVerifyEmail(
			t.Context(), userID, payload.TokenID, "activated-during-rollback",
		)
	}()
	waitForRolloutSharedLock(t, server.DB)

	downDone := make(chan error, 1)
	go func() { downDone <- execMigrationSQL(t.Context(), server.DB, downSQL) }()
	select {
	case downErr := <-downDone:
		t.Fatalf("down migration bypassed in-flight activation lock: %v", downErr)
	case <-time.After(100 * time.Millisecond):
	}
	require.NoError(t, tokenBlocker.Rollback(t.Context()))
	require.NoError(t, <-activationDone)
	require.NoError(t, <-downDone)

	var userExists, pending bool
	require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&userExists))
	require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT EXISTS(SELECT 1 FROM pending_email_verifications WHERE user_id=$1)`, userID).Scan(&pending))
	require.True(t, userExists, "activation committed before cleanup must preserve the activated account")
	require.False(t, pending)
	require.NoError(t, execMigrationSQL(t.Context(), server.DB, upSQL))
	require.NoError(t, enableAuthHardeningWrites(t.Context(), server.DB))
	restored = true
}

func enableAuthHardeningWrites(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		UPDATE auth_hardening_rollout_state
		SET writes_enabled=TRUE,updated_at=NOW()
		WHERE singleton=TRUE
	`)
	return err
}

func waitForRolloutSharedLock(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		tx, err := db.Begin(t.Context())
		require.NoError(t, err)
		var enabled bool
		err = tx.QueryRow(t.Context(), `
			SELECT writes_enabled FROM auth_hardening_rollout_state
			WHERE singleton=TRUE FOR UPDATE NOWAIT
		`).Scan(&enabled)
		_ = tx.Rollback(context.WithoutCancel(t.Context()))
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "55P03" {
			return
		}
		if err != nil {
			t.Fatalf("inspect rollout lock: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("activation did not acquire the rollout shared lock")
}

func execMigrationSQL(ctx context.Context, db *pgxpool.Pool, source []byte) error {
	connection, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer connection.Release()
	_, err = connection.Conn().PgConn().Exec(ctx, string(source)).ReadAll()
	return err
}

func TestResetPasswordTokenIsSingleUseUnderReplayAndConcurrency(t *testing.T) {
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", strings.Repeat("k", 32))
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	user, err := server.AppContext.UserService.SaveUser(
		t.Context(), "Reset", "Replay", "original-password", "reset-replay@example.com", nil, nil,
	)
	require.NoError(t, err)
	token, err := mail.IssueResetPasswordToken(t.Context(), server.DB, user.ID)
	require.NoError(t, err)
	payload, err := mail.ValidateResetPasswordPayload(token)
	require.NoError(t, err)

	start := make(chan struct{})
	errorsByAttempt := make(chan error, 2)
	for index := range 2 {
		go func() {
			<-start
			errorsByAttempt <- server.AppContext.UserService.ResetPasswordAndVerifyEmail(
				t.Context(), user.ID, payload.TokenID, fmt.Sprintf("concurrent-password-%d", index),
			)
		}()
	}
	close(start)
	var successes, replays int
	for range 2 {
		consumeErr := <-errorsByAttempt
		if consumeErr == nil {
			successes++
			continue
		}
		if strings.Contains(consumeErr.Error(), "token_invalid") {
			replays++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, replays)
	require.ErrorContains(t, server.AppContext.UserService.ResetPasswordAndVerifyEmail(
		t.Context(), user.ID, payload.TokenID, "replayed-password",
	), "token_invalid")
}

func TestOlderResetDeliveryRetryCannotInvalidateNewerLink(t *testing.T) {
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", strings.Repeat("k", 32))
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	user, err := server.AppContext.UserService.SaveUser(
		t.Context(), "Reset", "Ordering", "original-password", "reset-ordering@example.com", nil, nil,
	)
	require.NoError(t, err)

	first, err := mail.IssueResetPasswordToken(t.Context(), server.DB, user.ID, "river-job-a")
	require.NoError(t, err)
	firstRetry, err := mail.IssueResetPasswordToken(t.Context(), server.DB, user.ID, "river-job-a")
	require.NoError(t, err)
	require.Equal(t, first, firstRetry)
	firstPayload, err := mail.ValidateResetPasswordPayload(first)
	require.NoError(t, err)

	second, err := mail.IssueResetPasswordToken(t.Context(), server.DB, user.ID, "river-job-b")
	require.NoError(t, err)
	secondPayload, err := mail.ValidateResetPasswordPayload(second)
	require.NoError(t, err)
	_, err = mail.IssueResetPasswordToken(t.Context(), server.DB, user.ID, "river-job-a")
	require.ErrorIs(t, err, mail.ErrResetDeliverySuperseded)

	require.NoError(t, server.AppContext.UserService.ResetPasswordAndVerifyEmail(
		t.Context(), user.ID, secondPayload.TokenID, "newer-link-password",
	))
	require.ErrorContains(t, server.AppContext.UserService.ResetPasswordAndVerifyEmail(
		t.Context(), user.ID, firstPayload.TokenID, "older-link-password",
	), "token_invalid")
}

func TestLateResetWorkerRetryCancelsBeforeProviderAndPreservesNewerLink(t *testing.T) {
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", strings.Repeat("k", 32))
	t.Setenv("DISABLE_EMAILS", "false")
	t.Setenv("RESEND_API_KEY", "")
	server := setup.NewTestServer(t)
	defer server.Cleanup()
	user, err := server.AppContext.UserService.SaveUser(
		t.Context(), "Reset", "Worker", "original-password", "reset-worker-ordering@example.com", nil, nil,
	)
	require.NoError(t, err)
	worker := service.NewResetPasswordEmailWorker(server.AppContext.MailService)
	job := func(id int64) *river.Job[service.ResetPasswordEmailArgs] {
		return &river.Job[service.ResetPasswordEmailArgs]{
			JobRow: &rivertype.JobRow{ID: id},
			Args:   service.ResetPasswordEmailArgs{Email: user.Email},
		}
	}

	require.ErrorContains(t, worker.Work(t.Context(), job(1001)), "RESEND_API_KEY is required")
	require.ErrorContains(t, worker.Work(t.Context(), job(1002)), "RESEND_API_KEY is required")
	newerToken, err := mail.IssueResetPasswordToken(t.Context(), server.DB, user.ID, "1002")
	require.NoError(t, err)
	newerPayload, err := mail.ValidateResetPasswordPayload(newerToken)
	require.NoError(t, err)

	lateError := worker.Work(t.Context(), job(1001))
	var cancelError *river.JobCancelError
	require.ErrorAs(t, lateError, &cancelError)
	require.ErrorIs(t, lateError, mail.ErrResetDeliverySuperseded)
	require.NoError(t, server.AppContext.UserService.ResetPasswordAndVerifyEmail(
		t.Context(), user.ID, newerPayload.TokenID, "newer-worker-password",
	))
}

func TestResetPasswordMailBoundaryHandlesAbsentSentinelProviderFailureAndCancellation(t *testing.T) {
	t.Setenv("DISABLE_EMAILS", "false")
	t.Setenv("RESEND_API_KEY", "")
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	require.NoError(t, server.AppContext.MailService.SendResetPasswordEmail(t.Context(), ""))
	require.NoError(t, server.AppContext.MailService.SendResetPasswordEmail(
		t.Context(), "absent-reset-recipient@example.com",
	))
	user, err := server.AppContext.UserService.SaveUser(
		t.Context(), "Reset", "Mail", "original-password", "known-reset-recipient@example.com", nil, nil,
	)
	require.NoError(t, err)
	require.ErrorContains(t, server.AppContext.MailService.SendResetPasswordEmail(
		t.Context(), user.Email, "river-job-42",
	), "RESEND_API_KEY is required")
	var firstEncryptedToken string
	require.NoError(t, server.DB.QueryRow(t.Context(), `
		SELECT delivery_token_ciphertext FROM password_reset_tokens
		WHERE user_id=$1 AND delivery_key='river-job-42'
	`, user.ID).Scan(&firstEncryptedToken))
	require.ErrorContains(t, server.AppContext.MailService.SendResetPasswordEmail(
		t.Context(), user.Email, "river-job-42",
	), "RESEND_API_KEY is required")
	var retriedEncryptedToken string
	require.NoError(t, server.DB.QueryRow(t.Context(), `
		SELECT delivery_token_ciphertext FROM password_reset_tokens
		WHERE user_id=$1 AND delivery_key='river-job-42'
	`, user.ID).Scan(&retriedEncryptedToken))
	require.Equal(t, firstEncryptedToken, retriedEncryptedToken,
		"a retry of one River job must preserve an ambiguously delivered link")
	var activeTokens int
	require.NoError(t, server.DB.QueryRow(t.Context(), `
		SELECT count(*) FROM password_reset_tokens
		WHERE user_id=$1 AND consumed_at IS NULL AND expires_at > NOW()
	`, user.ID).Scan(&activeTokens))
	require.Equal(t, 1, activeTokens)

	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	require.Error(t, server.AppContext.MailService.SendResetPasswordEmail(canceled, user.Email))
}

func TestAuthRateLimitsAccountAndSourceWithRetryAfter(t *testing.T) {
	t.Run("signup source", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		for range 100 {
			server.Expect.POST("/users").
				WithJSON(setup.GenerateSignupInput()).
				Expect().Status(http.StatusCreated)
		}
		limited := server.Expect.POST("/users").
			WithJSON(setup.GenerateSignupInput()).
			Expect().Status(http.StatusTooManyRequests)
		limited.Header("Retry-After").NotEmpty()
		assertRetryAfterWithin(t, limited.Header("Retry-After").Raw())
		limited.JSON().Object().Value("error").String().IsEqual("Too many requests")
	})

	t.Run("signup account and enumeration parity", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		signup := setup.GenerateSignupInput()
		for range 3 {
			response := server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusCreated)
			response.Header("Authorization").IsEmpty()
			response.Header("X-Refresh-Token").IsEmpty()
			response.JSON().Object().Value("message").String().IsEqual(
				"If the address can be used, account instructions will be sent.",
			)
		}
		limited := server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusTooManyRequests)
		assertRetryAfterWithin(t, limited.Header("Retry-After").Raw())

		var users, jobs int
		require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT count(*) FROM users WHERE email=$1`, signup.Email).Scan(&users))
		require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT count(*) FROM river_job WHERE kind='reset_password_email'`).Scan(&jobs))
		require.Equal(t, 1, users)
		require.Equal(t, 3, jobs)
	})

	t.Run("login account and enumeration parity", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		signup := setup.GenerateSignupInput()
		server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusCreated)

		assertLoginLimit := func(email string) {
			for range 5 {
				server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
					Email: email, Password: "wrong-password",
				}).Expect().Status(http.StatusBadRequest).
					JSON().Object().Value("error").String().IsEqual("Invalid email or password")
			}
			limited := server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
				Email: email, Password: "wrong-password",
			}).Expect().Status(http.StatusTooManyRequests)
			limited.Header("Retry-After").NotEmpty()
			limited.JSON().Object().Value("error").String().IsEqual("Too many requests")
		}
		assertLoginLimit(signup.Email)
		assertLoginLimit("missing-account@example.com")
	})

	t.Run("successful authentication clears a reserved account bucket", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		user, err := server.AppContext.UserService.SaveUser(
			t.Context(), "Lockout", "Target", "correct-password", "lockout-target@example.com", nil, nil,
		)
		require.NoError(t, err)
		require.Positive(t, user.ID)
		for range 4 {
			server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
				Email: user.Email, Password: "wrong-password",
			}).Expect().Status(http.StatusBadRequest)
		}
		server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
			Email: user.Email, Password: "correct-password",
		}).Expect().Status(http.StatusOK)
		for range 5 {
			server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
				Email: user.Email, Password: "wrong-password",
			}).Expect().Status(http.StatusBadRequest)
		}
		server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
			Email: user.Email, Password: "wrong-password",
		}).Expect().Status(http.StatusTooManyRequests)
	})

	t.Run("infrastructure failures do not poison the account bucket", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		user, err := server.AppContext.UserService.SaveUser(
			t.Context(), "Infrastructure", "Failure", "correct-password", "infra-login@example.com", nil, nil,
		)
		require.NoError(t, err)
		require.NoError(t, func() error {
			_, renameErr := server.DB.Exec(t.Context(), `
				ALTER TABLE auth_session_families RENAME TO auth_session_families_unavailable
			`)
			return renameErr
		}())
		restored := false
		defer func() {
			if !restored {
				_, _ = server.DB.Exec(context.WithoutCancel(t.Context()), `
					ALTER TABLE auth_session_families_unavailable RENAME TO auth_session_families
				`)
			}
		}()
		for range 5 {
			server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
				Email: user.Email, Password: "correct-password",
			}).Expect().Status(http.StatusServiceUnavailable)
		}
		_, err = server.DB.Exec(t.Context(), `
			ALTER TABLE auth_session_families_unavailable RENAME TO auth_session_families
		`)
		require.NoError(t, err)
		restored = true
		server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
			Email: user.Email, Password: "wrong-password",
		}).Expect().Status(http.StatusBadRequest)
	})

	t.Run("login account keys canonicalize valid email variants and bound invalid identifiers", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		for _, email := range []string{
			" Person@Example.com ", "person@example.com", "PERSON@EXAMPLE.COM",
			"person@example.com ", " person@example.com",
		} {
			server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
				Email: email, Password: "wrong-password",
			}).Expect().Status(http.StatusBadRequest)
		}
		server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
			Email: "person@example.com", Password: "wrong-password",
		}).Expect().Status(http.StatusTooManyRequests)

		for range 5 {
			server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
				Email: "not-an-email", Password: "wrong-password",
			}).Expect().Status(http.StatusBadRequest)
		}
		server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
			Email: "not-an-email", Password: "wrong-password",
		}).Expect().Status(http.StatusTooManyRequests)
	})

	t.Run("actual login routes perform one bcrypt comparison", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		signup := setup.GenerateSignupInput()
		server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusCreated)
		for _, email := range []string{signup.Email, "unknown-bcrypt@example.com", "malformed"} {
			before := server.AppContext.UserService.AuthPasswordComparisonCount()
			beforePending := server.AppContext.UserService.AuthPendingLookupCount()
			server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
				Email: email, Password: "wrong-password",
			}).Expect().Status(http.StatusBadRequest)
			require.Equal(t, before+1, server.AppContext.UserService.AuthPasswordComparisonCount())
			require.Equal(t, beforePending+1, server.AppContext.UserService.AuthPendingLookupCount())
		}
	})

	t.Run("signup correlation performs the same pending lookup for mismatching accounts", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		signup := setup.GenerateSignupInput()
		server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusCreated)
		_, err := server.AppContext.UserService.SaveUser(
			t.Context(), "Existing", "Account", "different-password", "existing-correlation@example.com", nil, nil,
		)
		require.NoError(t, err)
		for _, email := range []string{signup.Email, "existing-correlation@example.com", "missing-correlation@example.com"} {
			before := server.AppContext.UserService.AuthPendingLookupCount()
			server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
				Email: email, Password: signup.Password,
			}).Expect().Status(http.StatusBadRequest)
			require.Equal(t, before+1, server.AppContext.UserService.AuthPendingLookupCount())
		}
	})

	t.Run("successful password matches perform one pending lookup", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		pending, err := server.AppContext.UserService.SaveUser(
			t.Context(), "Pending", "Match", "matching-password", "pending-match@example.com", nil, nil,
		)
		require.NoError(t, err)
		_, err = server.DB.Exec(t.Context(), `INSERT INTO pending_email_verifications (user_id) VALUES ($1)`, pending.ID)
		require.NoError(t, err)

		beforeComparisons := server.AppContext.UserService.AuthPasswordComparisonCount()
		beforeLookups := server.AppContext.UserService.AuthPendingLookupCount()
		_, err = server.AppContext.UserService.LoginUser(t.Context(), pending.Email, "matching-password")
		require.ErrorIs(t, err, service.ErrInvalidLoginCredentials)
		require.Equal(t, beforeComparisons+1, server.AppContext.UserService.AuthPasswordComparisonCount())
		require.Equal(t, beforeLookups+1, server.AppContext.UserService.AuthPendingLookupCount())

		active, err := server.AppContext.UserService.SaveUser(
			t.Context(), "Active", "Match", "matching-password", "active-match@example.com", nil, nil,
		)
		require.NoError(t, err)
		beforeComparisons = server.AppContext.UserService.AuthPasswordComparisonCount()
		beforeLookups = server.AppContext.UserService.AuthPendingLookupCount()
		credentials, err := server.AppContext.UserService.LoginUser(t.Context(), active.Email, "matching-password")
		require.NoError(t, err)
		require.NotEmpty(t, credentials.AccessToken)
		require.Equal(t, beforeComparisons+1, server.AppContext.UserService.AuthPasswordComparisonCount())
		require.Equal(t, beforeLookups+1, server.AppContext.UserService.AuthPendingLookupCount())
	})

	t.Run("reset request account", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		signup := setup.GenerateSignupInput()
		server.Expect.POST("/users").WithJSON(signup).Expect().Status(http.StatusCreated)
		server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]string{"email": signup.Email}).Expect().Status(http.StatusAccepted)
		server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]string{"email": "unknown-parity@example.com"}).Expect().Status(http.StatusAccepted)
		for range 3 {
			server.Expect.POST("/password/send-reset-email").
				WithJSON(map[string]string{"email": "limited-account@example.com"}).
				Expect().Status(http.StatusAccepted)
		}
		limited := server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]string{"email": "limited-account@example.com"}).
			Expect().Status(http.StatusTooManyRequests)
		limited.Header("Retry-After").NotEmpty()
	})

	t.Run("malformed reset identifier does not cross the durable queue boundary", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]string{"email": "not an email with private input"}).
			Expect().Status(http.StatusAccepted)
		var args []byte
		require.NoError(t, server.DB.QueryRow(t.Context(), `
			SELECT args FROM river_job WHERE kind='reset_password_email' ORDER BY id DESC LIMIT 1
		`).Scan(&args))
		var decoded service.ResetPasswordEmailArgs
		require.NoError(t, json.Unmarshal(args, &decoded))
		require.Empty(t, decoded.Email)
		require.NotContains(t, string(args), "private input")
	})

	t.Run("reset request source uses the same queued response for unknown accounts", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		for attempt := range 10 {
			response := server.Expect.POST("/password/send-reset-email").
				WithJSON(map[string]string{"email": fmt.Sprintf("unknown-%d@example.com", attempt)}).
				Expect().Status(http.StatusAccepted)
			response.JSON().Object().Value("message").String().IsEqual(
				"If the account exists, reset instructions will be sent.",
			)
		}
		limited := server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]string{"email": "unknown-final@example.com"}).
			Expect().Status(http.StatusTooManyRequests)
		assertRetryAfterWithin(t, limited.Header("Retry-After").Raw())
		var jobs int
		require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT count(*) FROM river_job WHERE kind='reset_password_email'`).Scan(&jobs))
		require.Equal(t, 10, jobs)
	})

	t.Run("reset consume source", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		for range 10 {
			server.Expect.PATCH("/password/reset").
				WithJSON(map[string]string{"token": "invalid-token", "password": "valid-password"}).
				Expect().Status(http.StatusBadRequest)
		}
		limited := server.Expect.PATCH("/password/reset").
			WithJSON(map[string]string{"token": "invalid-token", "password": "valid-password"}).
			Expect().Status(http.StatusTooManyRequests)
		limited.Header("Retry-After").NotEmpty()
	})

	t.Run("reset consume account uses the verified token owner", func(t *testing.T) {
		t.Setenv("RESET_PASSWORD_PRIVATE_KEY", "test-reset-password-key-32-chars")
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		user, err := server.AppContext.UserService.SaveUser(
			t.Context(), "Reset", "Owner", "valid-password", "reset-owner@example.com", nil, nil,
		)
		require.NoError(t, err)
		for attempt := range 5 {
			unstored, issueErr := mail.CreateResetPasswordToken(user.ID)
			require.NoError(t, issueErr)
			server.Expect.PATCH("/password/reset").WithJSON(map[string]string{
				"token": unstored.Encrypted, "password": fmt.Sprintf("valid-password-%d", attempt),
			}).Expect().Status(http.StatusBadRequest)
		}
		token, err := mail.IssueResetPasswordToken(t.Context(), server.DB, user.ID)
		require.NoError(t, err)
		server.Expect.PATCH("/password/reset").WithJSON(map[string]string{
			"token": token, "password": "valid-password-after-failures",
		}).Expect().Status(http.StatusOK)
		unstored, err := mail.CreateResetPasswordToken(user.ID)
		require.NoError(t, err)
		limited := server.Expect.PATCH("/password/reset").WithJSON(map[string]string{
			"token": unstored.Encrypted, "password": "valid-password-final",
		}).Expect().Status(http.StatusTooManyRequests)
		assertRetryAfterWithin(t, limited.Header("Retry-After").Raw())
	})

	t.Run("forwarding headers cannot bypass the transport source limit", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		for attempt := range 30 {
			server.Expect.POST("/login").
				WithHeader("X-Forwarded-For", "198.51.100."+strconv.Itoa(attempt+1)).
				WithJSON(httphandlers.LoginInput{
					Email: "missing-" + strconv.Itoa(attempt) + "@example.com", Password: "wrong-password",
				}).Expect().Status(http.StatusBadRequest)
		}
		limited := server.Expect.POST("/login").
			WithHeader("X-Forwarded-For", "203.0.113.200").
			WithJSON(httphandlers.LoginInput{
				Email: "missing-final@example.com", Password: "wrong-password",
			}).Expect().Status(http.StatusTooManyRequests)
		limited.Header("Retry-After").NotEmpty()
	})

	t.Run("explicit trusted proxy separates forwarded client sources", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXY_CIDRS", "127.0.0.1/32")
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		for attempt := range 30 {
			server.Expect.POST("/login").
				WithHeader("X-Forwarded-For", "198.51.100.20").
				WithJSON(httphandlers.LoginInput{
					Email: "trusted-proxy-" + strconv.Itoa(attempt) + "@example.com", Password: "wrong-password",
				}).Expect().Status(http.StatusBadRequest)
		}
		server.Expect.POST("/login").
			WithHeader("X-Forwarded-For", "203.0.113.20").
			WithJSON(httphandlers.LoginInput{
				Email: "other-source@example.com", Password: "wrong-password",
			}).Expect().Status(http.StatusBadRequest)
	})

	t.Run("trusted multi-hop chain stops at the first untrusted hop", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXY_CIDRS", "127.0.0.1/32,10.0.0.0/8")
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		for attempt := range 30 {
			server.Expect.POST("/login").
				WithHeader("X-Forwarded-For", fmt.Sprintf("198.51.100.%d, 10.0.0.5", attempt+1)).
				WithJSON(httphandlers.LoginInput{
					Email: fmt.Sprintf("chain-%d@example.com", attempt), Password: "wrong-password",
				}).Expect().Status(http.StatusBadRequest)
		}
		server.Expect.POST("/login").
			WithHeader("X-Forwarded-For", "203.0.113.20, 10.0.0.5").
			WithJSON(httphandlers.LoginInput{Email: "chain-other@example.com", Password: "wrong-password"}).
			Expect().Status(http.StatusBadRequest)
	})

	t.Run("equivalent forwarded IPv6 forms share one source bucket", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXY_CIDRS", "127.0.0.1/32")
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		forms := []string{"2001:0db8:0:0:0:0:0:1", "2001:db8::1"}
		for attempt := range 30 {
			server.Expect.POST("/login").WithHeader("X-Forwarded-For", forms[attempt%2]).
				WithJSON(httphandlers.LoginInput{
					Email: fmt.Sprintf("ipv6-%d@example.com", attempt), Password: "wrong-password",
				}).Expect().Status(http.StatusBadRequest)
		}
		server.Expect.POST("/login").WithHeader("X-Forwarded-For", forms[1]).
			WithJSON(httphandlers.LoginInput{Email: "ipv6-final@example.com", Password: "wrong-password"}).
			Expect().Status(http.StatusTooManyRequests)
	})

	t.Run("two fleet instances share Redis counters and expiry", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		environment := "fleet-" + uuid.NewString()
		first, err := service.NewAuthRateLimiter(server.AppContext.RedisClient, true, environment, nil)
		require.NoError(t, err)
		second, err := service.NewAuthRateLimiter(server.AppContext.RedisClient, true, environment, nil)
		require.NoError(t, err)
		for attempt := range 5 {
			limiter := first
			if attempt%2 == 1 {
				limiter = second
			}
			decision, checkErr := limiter.Check(
				t.Context(), service.AuthLimitLogin, service.AuthLimitAccount, "fleet@example.com",
			)
			require.NoError(t, checkErr)
			require.True(t, decision.Allowed)
		}
		decision, err := second.Check(
			t.Context(), service.AuthLimitLogin, service.AuthLimitAccount, "fleet@example.com",
		)
		require.NoError(t, err)
		require.False(t, decision.Allowed)
		keys, err := server.AppContext.RedisClient.Keys(
			t.Context(), "auth_limit:v1:"+environment+":*",
		).Result()
		require.NoError(t, err)
		require.Len(t, keys, 1)
		ttl, err := server.AppContext.RedisClient.PTTL(t.Context(), keys[0]).Result()
		require.NoError(t, err)
		require.Positive(t, ttl)
		require.LessOrEqual(t, ttl, 10*time.Minute)
	})

	t.Run("Redis increment is atomic under concurrency", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		var allowed atomic.Int64
		var failures atomic.Int64
		var wait sync.WaitGroup
		for range 40 {
			wait.Add(1)
			go func() {
				defer wait.Done()
				decision, err := server.AppContext.AuthRateLimiter.Check(
					t.Context(), service.AuthLimitLogin, service.AuthLimitAccount, "concurrent@example.com",
				)
				if err != nil {
					failures.Add(1)
					return
				}
				if decision.Allowed {
					allowed.Add(1)
				}
			}()
		}
		wait.Wait()
		require.Equal(t, int64(0), failures.Load())
		require.Equal(t, int64(5), allowed.Load())
	})

	t.Run("Redis outage fails closed", func(t *testing.T) {
		server := setup.NewTestServer(t)
		defer server.Cleanup()
		require.NoError(t, server.AppContext.RedisClient.Close())
		response := server.Expect.POST("/login").WithJSON(httphandlers.LoginInput{
			Email: "outage@example.com", Password: "wrong-password",
		}).Expect().Status(http.StatusServiceUnavailable)
		response.JSON().Object().Value("error").String().IsEqual("Authentication temporarily unavailable")
	})
}

func assertRetryAfterWithin(t *testing.T, raw string) {
	t.Helper()
	seconds, err := strconv.ParseInt(raw, 10, 64)
	require.NoError(t, err)
	require.GreaterOrEqual(t, seconds, int64(1))
	require.LessOrEqual(t, seconds, int64(time.Hour/time.Second))
}

func TestPasswordPolicyIsSharedBySignupResetAndChange(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	tooShort := setup.GenerateSignupInput()
	tooShort.Password = "1234567"
	server.Expect.POST("/users").WithJSON(tooShort).Expect().Status(http.StatusBadRequest)

	tooLong := setup.GenerateSignupInput()
	tooLong.Password = strings.Repeat("a", 73)
	server.Expect.POST("/users").WithJSON(tooLong).Expect().Status(http.StatusBadRequest)

	valid := setup.GenerateSignupInput()
	valid.Password = "12345678"
	server.Expect.POST("/users").
		WithHeader("X-Auth-Client", "native").
		WithJSON(valid).Expect().Status(http.StatusCreated)
	server.VerifyTestSignup(t, valid.Email, valid.Password)
	response := server.Expect.POST("/login").
		WithHeader("X-Auth-Client", "native").
		WithJSON(httphandlers.LoginInput{Email: valid.Email, Password: valid.Password}).
		Expect().Status(http.StatusOK)
	access := response.Header("Authorization").NotEmpty().Raw()

	server.Expect.PATCH("/users").
		WithHeader("Authorization", access).
		WithJSON(map[string]any{"updatePassword": map[string]string{
			"currentPassword": valid.Password,
			"newPassword":     "1234567",
		}}).
		Expect().Status(http.StatusBadRequest)
	require.Error(t, server.AppContext.UserService.UpdatePassword(t.Context(), 1, "1234567"),
		"reset service path must enforce the same policy")

	credentials, err := server.AppContext.UserService.LoginUser(t.Context(), valid.Email, valid.Password)
	require.NoError(t, err, "rejected password change must preserve the prior password")
	require.NotEmpty(t, credentials.AccessToken)
}
