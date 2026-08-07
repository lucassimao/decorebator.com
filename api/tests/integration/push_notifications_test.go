package integration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	httphandlers "decorebator.com/internal/http"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	expoPushEndpoint    = "https://exp.host/--/api/v2/push/send"
	expoReceiptEndpoint = "https://exp.host/--/api/v2/push/getReceipts"
)

func TestPushNotificationRegister(t *testing.T) {
	t.Run("stores token metadata on registration", func(t *testing.T) {
		ts, token, userID := newAuthedPushTestServer(t)
		defer ts.Cleanup()

		expoToken := "ExponentPushToken[test-device]"
		locale := "en-US"
		deviceID := "device-123"
		payload := map[string]any{
			"expoPushToken": expoToken,
			"platform":      "ios",
			"timezone":      "America/New_York",
			"locale":        locale,
			"deviceId":      deviceID,
		}

		ts.Expect.POST("/push/register").
			WithHeader("Authorization", token).
			WithJSON(payload).
			Expect().
			Status(http.StatusNoContent)

		var platform, timezone string
		var storedLocale, storedDeviceID sql.NullString
		var isActive bool
		var lastSeenAt time.Time

		err := ts.DB.QueryRow(context.Background(), `
			SELECT platform, timezone, locale, device_id, is_active, last_seen_at
			FROM push_tokens
			WHERE user_id = $1 AND expo_token = $2
		`, userID, expoToken).Scan(&platform, &timezone, &storedLocale, &storedDeviceID, &isActive, &lastSeenAt)
		require.NoError(t, err, "push token should be stored")

		assert.Equal(t, "ios", platform)
		assert.Equal(t, "America/New_York", timezone)
		assert.True(t, isActive, "token should be marked active")
		assert.True(t, storedLocale.Valid)
		assert.Equal(t, locale, storedLocale.String)
		assert.True(t, storedDeviceID.Valid)
		assert.Equal(t, deviceID, storedDeviceID.String)
		assert.WithinDuration(t, time.Now(), lastSeenAt, 5*time.Second, "last_seen_at should be set near now")
	})

	t.Run("rejects invalid platform", func(t *testing.T) {
		ts, token, _ := newAuthedPushTestServer(t)
		defer ts.Cleanup()

		payload := map[string]any{
			"expoPushToken": "ExponentPushToken[invalid-platform]",
			"platform":      "web",
			"timezone":      "America/Chicago",
		}

		ts.Expect.POST("/push/register").
			WithHeader("Authorization", token).
			WithJSON(payload).
			Expect().
			Status(http.StatusBadRequest).
			JSON().
			Object().
			HasValue("error", "invalid platform")
	})

	t.Run("rejects invalid timezone", func(t *testing.T) {
		ts, token, _ := newAuthedPushTestServer(t)
		defer ts.Cleanup()

		payload := map[string]any{
			"expoPushToken": "ExponentPushToken[invalid-timezone]",
			"platform":      "android",
			"timezone":      "Not/A_Real_Timezone",
		}

		ts.Expect.POST("/push/register").
			WithHeader("Authorization", token).
			WithJSON(payload).
			Expect().
			Status(http.StatusBadRequest).
			JSON().
			Object().
			HasValue("error", "invalid timezone")
	})
}

func TestPushNotificationUnregister(t *testing.T) {
	ts, token, userID := newAuthedPushTestServer(t)
	defer ts.Cleanup()

	expoToken := "ExponentPushToken[unregister]"

	// Register first to seed the token
	ts.Expect.POST("/push/register").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"expoPushToken": expoToken,
			"platform":      "android",
			"timezone":      "Europe/London",
		}).
		Expect().
		Status(http.StatusNoContent)

	// Unregister should deactivate the token
	ts.Expect.POST("/push/unregister").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"expoPushToken": expoToken,
		}).
		Expect().
		Status(http.StatusNoContent)

	var isActive bool
	err := ts.DB.QueryRow(context.Background(), `
		SELECT is_active FROM push_tokens WHERE user_id = $1 AND expo_token = $2
	`, userID, expoToken).Scan(&isActive)
	require.NoError(t, err, "push token should exist")
	assert.False(t, isActive, "token should be deactivated after unregister")
}

func newAuthedPushTestServer(t *testing.T) (*setup.TestServer, string, int64) {
	t.Helper()

	ts := setup.NewTestServer(t)

	signupInput := setup.GenerateSignupInput()
	ts.Expect.POST("/users").
		WithJSON(signupInput).
		Expect().
		Status(http.StatusCreated)
	ts.VerifyTestSignup(t, signupInput.Email, signupInput.Password)

	loginResp := ts.Expect.POST("/login").
		WithJSON(httphandlers.LoginInput{
			Email:    signupInput.Email,
			Password: signupInput.Password,
		}).
		Expect().
		Status(http.StatusOK)

	token := loginResp.Header("Authorization").NotEmpty().Raw()

	var userID int64
	err := ts.DB.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", signupInput.Email).Scan(&userID)
	require.NoError(t, err, "user should exist in database")

	return ts, token, userID
}

func TestNotificationPreferencesToggle(t *testing.T) {
	ts, token, userID := newAuthedPushTestServer(t)
	defer ts.Cleanup()

	// Disable notifications
	disableResp := ts.Expect.PATCH("/users").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"notificationsEnabled": false,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	disableResp.Value("notificationsEnabled").IsEqual(false)

	// Verify persisted state
	var enabled bool
	err := ts.DB.QueryRow(context.Background(), "SELECT notifications_enabled FROM users WHERE id = $1", userID).Scan(&enabled)
	require.NoError(t, err)
	assert.False(t, enabled, "notifications_enabled should be false after toggle off")

	// Re-enable notifications
	enableResp := ts.Expect.PATCH("/users").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"notificationsEnabled": true,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	enableResp.Value("notificationsEnabled").IsEqual(true)

	err = ts.DB.QueryRow(context.Background(), "SELECT notifications_enabled FROM users WHERE id = $1", userID).Scan(&enabled)
	require.NoError(t, err)
	assert.True(t, enabled, "notifications_enabled should be true after toggle on")
}

func TestDailyPracticeReminderWorkerStoresReceiptsAndMarksNotified(t *testing.T) {
	ts, token, userID := newAuthedPushTestServer(t)
	defer ts.Cleanup()

	timezone := timezoneForLocalEleven(time.Now().UTC())

	ts.Expect.POST("/push/register").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"expoPushToken": "ExponentPushToken[worker-success]",
			"platform":      "android",
			"timezone":      timezone,
		}).
		Expect().
		Status(http.StatusNoContent)

	pushService := service.NewPushNotificationService(
		&repository.PushTokenRepository{Db: ts.DB},
		&repository.PushNotificationRepository{Db: ts.DB},
		&repository.PushNotificationEventRepository{Db: ts.DB},
		&repository.PushReceiptRepository{Db: ts.DB},
	)
	pushService.SetHTTPClient(httpClientStub(map[string]stubResponse{
		expoPushEndpoint: {
			Status: http.StatusOK,
			Body:   `{"data":[{"status":"ok","id":"ticket-success"}]}`,
		},
	}))

	worker := service.NewDailyPracticeReminderWorker(pushService)

	err := worker.Work(context.Background(), &river.Job[service.DailyPracticeReminderArgs]{})
	require.NoError(t, err)

	var lastNotified time.Time
	err = ts.DB.QueryRow(context.Background(), `
		SELECT last_notified_at FROM push_tokens WHERE user_id = $1 AND expo_token = $2
	`, userID, "ExponentPushToken[worker-success]").Scan(&lastNotified)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), lastNotified, 5*time.Second)

	var ticketID, expoToken string
	err = ts.DB.QueryRow(context.Background(), `
		SELECT expo_ticket_id, expo_token FROM push_receipts WHERE expo_ticket_id = $1
	`, "ticket-success").Scan(&ticketID, &expoToken)
	require.NoError(t, err)
	assert.Equal(t, "ticket-success", ticketID)
	assert.Equal(t, "ExponentPushToken[worker-success]", expoToken)
}

func TestDailyPracticeReminderWorkerDeactivatesInvalidTokens(t *testing.T) {
	ts, token, userID := newAuthedPushTestServer(t)
	defer ts.Cleanup()

	timezone := timezoneForLocalEleven(time.Now().UTC())
	expoToken := "ExponentPushToken[worker-device-not-registered]"

	ts.Expect.POST("/push/register").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"expoPushToken": expoToken,
			"platform":      "ios",
			"timezone":      timezone,
		}).
		Expect().
		Status(http.StatusNoContent)

	pushService := service.NewPushNotificationService(
		&repository.PushTokenRepository{Db: ts.DB},
		&repository.PushNotificationRepository{Db: ts.DB},
		&repository.PushNotificationEventRepository{Db: ts.DB},
		&repository.PushReceiptRepository{Db: ts.DB},
	)
	pushService.SetHTTPClient(httpClientStub(map[string]stubResponse{
		expoPushEndpoint: {
			Status: http.StatusOK,
			Body:   `{"data":[{"status":"error","message":"Device not registered","details":{"error":"DeviceNotRegistered"}}]}`,
		},
	}))

	worker := service.NewDailyPracticeReminderWorker(pushService)

	err := worker.Work(context.Background(), &river.Job[service.DailyPracticeReminderArgs]{})
	require.NoError(t, err)

	var isActive bool
	err = ts.DB.QueryRow(context.Background(), `
		SELECT is_active FROM push_tokens WHERE user_id = $1 AND expo_token = $2
	`, userID, expoToken).Scan(&isActive)
	require.NoError(t, err)
	assert.False(t, isActive, "token should be deactivated after DeviceNotRegistered error")
}

func TestDueItemsReminderWorkerStoresReceiptsAndMarksNotified(t *testing.T) {
	ts, token, userID := newAuthedPushTestServer(t)
	defer ts.Cleanup()

	now := time.Now().UTC()
	timezone := timezoneForLocalHour(now, 12)

	ts.Expect.POST("/push/register").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"expoPushToken": "ExponentPushToken[due-items-success]",
			"platform":      "android",
			"timezone":      timezone,
		}).
		Expect().
		Status(http.StatusNoContent)

	seedDueItem(t, ts, userID, now.Add(-5*time.Minute))

	pushService := service.NewPushNotificationService(
		&repository.PushTokenRepository{Db: ts.DB},
		&repository.PushNotificationRepository{Db: ts.DB},
		&repository.PushNotificationEventRepository{Db: ts.DB},
		&repository.PushReceiptRepository{Db: ts.DB},
	)
	pushService.SetHTTPClient(httpClientStub(map[string]stubResponse{
		expoPushEndpoint: {
			Status: http.StatusOK,
			Body:   `{"data":[{"status":"ok","id":"ticket-due-items"}]}`,
		},
	}))

	worker := service.NewDueItemsReminderWorker(pushService)

	err := worker.Work(context.Background(), &river.Job[service.DueItemsReminderArgs]{})
	require.NoError(t, err)

	var lastNotified time.Time
	err = ts.DB.QueryRow(context.Background(), `
		SELECT last_notified_at FROM push_tokens WHERE user_id = $1 AND expo_token = $2
	`, userID, "ExponentPushToken[due-items-success]").Scan(&lastNotified)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), lastNotified, 5*time.Second)

	var ticketID, expoToken string
	err = ts.DB.QueryRow(context.Background(), `
		SELECT expo_ticket_id, expo_token FROM push_receipts WHERE expo_ticket_id = $1
	`, "ticket-due-items").Scan(&ticketID, &expoToken)
	require.NoError(t, err)
	assert.Equal(t, "ticket-due-items", ticketID)
	assert.Equal(t, "ExponentPushToken[due-items-success]", expoToken)
}

func TestDailyPracticeReminderWorkerSkipsWhenDueItemsExist(t *testing.T) {
	ts, token, userID := newAuthedPushTestServer(t)
	defer ts.Cleanup()

	now := time.Now().UTC()
	timezone := timezoneForLocalHour(now, 11)

	ts.Expect.POST("/push/register").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"expoPushToken": "ExponentPushToken[skip-daily-when-due]",
			"platform":      "ios",
			"timezone":      timezone,
		}).
		Expect().
		Status(http.StatusNoContent)

	seedDueItem(t, ts, userID, now.Add(-10*time.Minute))

	pushService := service.NewPushNotificationService(
		&repository.PushTokenRepository{Db: ts.DB},
		&repository.PushNotificationRepository{Db: ts.DB},
		&repository.PushNotificationEventRepository{Db: ts.DB},
		&repository.PushReceiptRepository{Db: ts.DB},
	)
	pushService.SetHTTPClient(httpClientStub(map[string]stubResponse{
		expoPushEndpoint: {
			Status: http.StatusOK,
			Body:   `{"data":[{"status":"ok","id":"ticket-should-not-send"}]}`,
		},
	}))

	worker := service.NewDailyPracticeReminderWorker(pushService)

	err := worker.Work(context.Background(), &river.Job[service.DailyPracticeReminderArgs]{})
	require.NoError(t, err)

	var lastNotified sql.NullTime
	err = ts.DB.QueryRow(context.Background(), `
		SELECT last_notified_at FROM push_tokens WHERE user_id = $1 AND expo_token = $2
	`, userID, "ExponentPushToken[skip-daily-when-due]").Scan(&lastNotified)
	require.NoError(t, err)
	assert.False(t, lastNotified.Valid, "daily reminder should not send when due items exist")
}

func TestDueItemsReminderHonorsWeeklyCap(t *testing.T) {
	ts, token, userID := newAuthedPushTestServer(t)
	defer ts.Cleanup()

	now := time.Now().UTC()
	timezone := timezoneForLocalHour(now, 12)

	ts.Expect.POST("/push/register").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"expoPushToken": "ExponentPushToken[due-items-cap]",
			"platform":      "android",
			"timezone":      timezone,
		}).
		Expect().
		Status(http.StatusNoContent)

	seedDueItem(t, ts, userID, now.Add(-5*time.Minute))

	pushService := service.NewPushNotificationService(
		&repository.PushTokenRepository{Db: ts.DB},
		&repository.PushNotificationRepository{Db: ts.DB},
		&repository.PushNotificationEventRepository{Db: ts.DB},
		&repository.PushReceiptRepository{Db: ts.DB},
	)
	pushService.SetHTTPClient(httpClientStub(map[string]stubResponse{
		expoPushEndpoint: {
			Status: http.StatusOK,
			Body:   `{"data":[{"status":"ok","id":"ticket-cap"}]}`,
		},
	}))

	worker := service.NewDueItemsReminderWorker(pushService)

	// Send two reminders to reach weekly cap.
	err := worker.Work(context.Background(), &river.Job[service.DueItemsReminderArgs]{})
	require.NoError(t, err)
	err = worker.Work(context.Background(), &river.Job[service.DueItemsReminderArgs]{})
	require.NoError(t, err)

	var afterSecond time.Time
	err = ts.DB.QueryRow(context.Background(), `
		SELECT last_notified_at FROM push_tokens WHERE user_id = $1 AND expo_token = $2
	`, userID, "ExponentPushToken[due-items-cap]").Scan(&afterSecond)
	require.NoError(t, err)

	// Third attempt should be blocked by rolling weekly cap.
	err = worker.Work(context.Background(), &river.Job[service.DueItemsReminderArgs]{})
	require.NoError(t, err)

	var afterThird time.Time
	err = ts.DB.QueryRow(context.Background(), `
		SELECT last_notified_at FROM push_tokens WHERE user_id = $1 AND expo_token = $2
	`, userID, "ExponentPushToken[due-items-cap]").Scan(&afterThird)
	require.NoError(t, err)
	assert.Equal(t, afterSecond, afterThird, "weekly cap should prevent additional reminders")
}

func TestDueItemsReminderSendsBelowWeeklyCap(t *testing.T) {
	ts, token, userID := newAuthedPushTestServer(t)
	defer ts.Cleanup()

	now := time.Now().UTC()
	timezone := timezoneForLocalHour(now, 12)

	ts.Expect.POST("/push/register").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"expoPushToken": "ExponentPushToken[due-items-under-cap]",
			"platform":      "android",
			"timezone":      timezone,
		}).
		Expect().
		Status(http.StatusNoContent)

	seedDueItem(t, ts, userID, now.Add(-5*time.Minute))

	_, err := ts.DB.Exec(context.Background(), `
		INSERT INTO push_notification_events (user_id, notification_type, sent_at)
		VALUES ($1, 'daily_practice_reminder', $2)
	`, userID, now.Add(-2*24*time.Hour))
	require.NoError(t, err)

	pushService := service.NewPushNotificationService(
		&repository.PushTokenRepository{Db: ts.DB},
		&repository.PushNotificationRepository{Db: ts.DB},
		&repository.PushNotificationEventRepository{Db: ts.DB},
		&repository.PushReceiptRepository{Db: ts.DB},
	)
	pushService.SetHTTPClient(httpClientStub(map[string]stubResponse{
		expoPushEndpoint: {
			Status: http.StatusOK,
			Body:   `{"data":[{"status":"ok","id":"ticket-under-cap"}]}`,
		},
	}))

	worker := service.NewDueItemsReminderWorker(pushService)

	err = worker.Work(context.Background(), &river.Job[service.DueItemsReminderArgs]{})
	require.NoError(t, err)

	var lastNotified time.Time
	err = ts.DB.QueryRow(context.Background(), `
		SELECT last_notified_at FROM push_tokens WHERE user_id = $1 AND expo_token = $2
	`, userID, "ExponentPushToken[due-items-under-cap]").Scan(&lastNotified)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), lastNotified, 5*time.Second)
}

func TestPushReceiptWorkerUpdatesStatusAndDeactivatesTokens(t *testing.T) {
	ts, token, userID := newAuthedPushTestServer(t)
	defer ts.Cleanup()

	expoToken := "ExponentPushToken[receipt-check]"
	timezone := timezoneForLocalEleven(time.Now().UTC())

	ts.Expect.POST("/push/register").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"expoPushToken": expoToken,
			"platform":      "android",
			"timezone":      timezone,
		}).
		Expect().
		Status(http.StatusNoContent)

	_, err := ts.DB.Exec(context.Background(), `
		INSERT INTO push_receipts (expo_ticket_id, expo_token, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`, "receipt-123", expoToken)
	require.NoError(t, err)

	pushService := service.NewPushNotificationService(
		&repository.PushTokenRepository{Db: ts.DB},
		&repository.PushNotificationRepository{Db: ts.DB},
		&repository.PushNotificationEventRepository{Db: ts.DB},
		&repository.PushReceiptRepository{Db: ts.DB},
	)
	pushService.SetHTTPClient(httpClientStub(map[string]stubResponse{
		expoReceiptEndpoint: {
			Status: http.StatusOK,
			Body: `{
				"data": {
					"receipt-123": {
						"status": "error",
						"message": "DeviceNotRegistered",
						"details": {"error": "DeviceNotRegistered"}
					}
				}
			}`,
		},
	}))

	worker := service.NewPushReceiptWorker(pushService)

	err = worker.Work(context.Background(), &river.Job[service.PushReceiptArgs]{})
	require.NoError(t, err)

	var status sql.NullString
	var checkedAt time.Time
	err = ts.DB.QueryRow(context.Background(), `
		SELECT status, checked_at FROM push_receipts WHERE expo_ticket_id = $1
	`, "receipt-123").Scan(&status, &checkedAt)
	require.NoError(t, err)
	assert.True(t, status.Valid)
	assert.Equal(t, "error", status.String)
	assert.WithinDuration(t, time.Now(), checkedAt, 5*time.Second)

	var isActive bool
	err = ts.DB.QueryRow(context.Background(), `
		SELECT is_active FROM push_tokens WHERE user_id = $1 AND expo_token = $2
	`, userID, expoToken).Scan(&isActive)
	require.NoError(t, err)
	assert.False(t, isActive, "token should be deactivated when receipt reports DeviceNotRegistered")
}

type stubResponse struct {
	Status int
	Body   string
}

type stubRoundTripper struct {
	responses map[string]stubResponse
}

func (s stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, ok := s.responses[req.URL.String()]
	if !ok {
		return nil, errors.New("unexpected request to " + req.URL.String())
	}

	return &http.Response{
		StatusCode: resp.Status,
		Body:       io.NopCloser(strings.NewReader(resp.Body)),
		Header:     make(http.Header),
	}, nil
}

func httpClientStub(responses map[string]stubResponse) *http.Client {
	return &http.Client{
		Transport: stubRoundTripper{responses: responses},
	}
}

func timezoneForLocalEleven(now time.Time) string {
	offset := now.Hour() - 11
	return fmt.Sprintf("Etc/GMT%+d", offset)
}

func timezoneForLocalHour(now time.Time, hour int) string {
	offset := now.Hour() - hour
	return fmt.Sprintf("Etc/GMT%+d", offset)
}

func seedDueItem(t *testing.T, ts *setup.TestServer, userID int64, nextReviewAt time.Time) {
	t.Helper()

	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	var wordlistID int64
	err := ts.DB.QueryRow(ctx, `
		INSERT INTO wordlists (name, description, user_id, language_code, pronunciation_system)
		VALUES ($1, $2, $3, 'en', 'ipa')
		RETURNING id
	`, "Due Wordlist "+suffix, "Due items wordlist", userID).Scan(&wordlistID)
	require.NoError(t, err)

	var wordID int64
	err = ts.DB.QueryRow(ctx, `
		INSERT INTO words (name, user_id, wordlist_id, processing_status)
		VALUES ($1, $2, $3, 'completed')
		RETURNING id
	`, "dueword-"+suffix, userID, wordlistID).Scan(&wordID)
	require.NoError(t, err)

	var definitionID int64
	err = ts.DB.QueryRow(ctx, `
		INSERT INTO definitions (token, language, part_of_speech, meaning, source, part_of_speech_normalized)
		VALUES ($1, 'en', 'noun', $2, 'test', 'noun')
		RETURNING id
	`, "dueword-"+suffix, "Test meaning").Scan(&definitionID)
	require.NoError(t, err)

	_, err = ts.DB.Exec(ctx, `
		INSERT INTO word_definitions (word_id, definition_id)
		VALUES ($1, $2)
	`, wordID, definitionID)
	require.NoError(t, err)

	_, err = ts.DB.Exec(ctx, `
		INSERT INTO leitner_system_tracking (user_id, definition_id, word_id, box_id, updated_at, next_review_at)
		VALUES ($1, $2, $3, 2, $4, $5)
	`, userID, definitionID, wordID, nextReviewAt.Add(-6*time.Hour), nextReviewAt)
	require.NoError(t, err)
}
