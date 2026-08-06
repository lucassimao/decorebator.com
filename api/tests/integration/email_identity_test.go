package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailIdentityUsesOneCanonicalContract(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	const canonicalEmail = "mixed.case+identity@example.com"
	server.Expect.POST("/users").
		WithJSON(map[string]string{
			"firstName": "HTTP",
			"lastName":  "Identity",
			"password":  "password123",
			"email":     "  HTTP.Case+Identity@Example.COM\t",
		}).
		Expect().
		Status(http.StatusCreated)
	server.Expect.POST("/login").
		WithJSON(map[string]string{
			"password": "password123",
			"email":    "\thttp.case+IDENTITY@EXAMPLE.COM ",
		}).
		Expect().
		Status(http.StatusOK)

	user, err := server.AppContext.UserService.SaveUser(
		ctx,
		"Email",
		"Owner",
		"password123",
		" \tMixed.Case+Identity@Example.COM\r\n",
		nil,
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, canonicalEmail, user.Email)

	var storedEmail string
	require.NoError(t, server.DB.QueryRow(ctx, `SELECT email FROM users WHERE id=$1`, user.ID).Scan(&storedEmail))
	assert.Equal(t, canonicalEmail, storedEmail)

	token, err := server.AppContext.UserService.LoginUser(ctx, "\tMIXED.CASE+IDENTITY@EXAMPLE.COM ", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	userRepository := &repository.UserRepository{Db: server.DB}
	resetLookup := "  MIXED.CASE+IDENTITY@EXAMPLE.COM\n"
	users, err := userRepository.Find(ctx, repository.FindUserArgs{Email: &resetLookup})
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, user.ID, users[0].ID)
	aliasless, err := server.AppContext.UserService.SaveUser(
		ctx,
		"Aliasless",
		"Identity",
		"password123",
		"mixed.case@example.com",
		nil,
		nil,
	)
	require.NoError(t, err)
	assert.NotEqual(t, user.ID, aliasless.ID, "plus aliases must remain distinct identities")

	_, err = server.AppContext.UserService.SaveUser(ctx, "Unicode", "Rejected", "password123", "usér@example.com", nil, nil)
	var businessErr common.BusinessError
	require.ErrorAs(t, err, &businessErr)
	assert.Equal(t, "Invalid email address", businessErr.Message)

	const concurrentCanonicalEmail = "race.identity+tag@example.com"
	variants := []string{
		concurrentCanonicalEmail,
		"RACE.IDENTITY+TAG@EXAMPLE.COM",
		" race.identity+tag@example.com ",
		"\trace.identity+tag@EXAMPLE.com\n",
	}
	var waitGroup sync.WaitGroup
	errorsByAttempt := make(chan error, len(variants))
	for index, variant := range variants {
		waitGroup.Add(1)
		go func(index int, variant string) {
			defer waitGroup.Done()
			_, saveErr := server.AppContext.UserService.SaveUser(
				ctx,
				fmt.Sprintf("Duplicate%d", index),
				"Identity",
				"password123",
				variant,
				nil,
				nil,
			)
			errorsByAttempt <- saveErr
		}(index, variant)
	}
	waitGroup.Wait()
	close(errorsByAttempt)
	successCount := 0
	duplicateCount := 0
	for saveErr := range errorsByAttempt {
		if saveErr == nil {
			successCount++
			continue
		}
		businessErr = common.BusinessError{}
		require.True(t, errors.As(saveErr, &businessErr), "duplicate error = %T %v", saveErr, saveErr)
		assert.Equal(t, "Email already exists.", businessErr.Message)
		duplicateCount++
	}
	assert.Equal(t, 1, successCount)
	assert.Equal(t, len(variants)-1, duplicateCount)

	var identityCount int
	require.NoError(t, server.DB.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM users WHERE TRANSLATE(BTRIM(email, E' \t\n\r\f\v'), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz')=$1`,
		canonicalEmail,
	).Scan(&identityCount))
	assert.Equal(t, 1, identityCount)
	require.NoError(t, server.DB.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM users WHERE TRANSLATE(BTRIM(email, E' \t\n\r\f\v'), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz')=$1`,
		concurrentCanonicalEmail,
	).Scan(&identityCount))
	assert.Equal(t, 1, identityCount)

	var indexDefinition string
	require.NoError(t, server.DB.QueryRow(
		ctx,
		`SELECT indexdef FROM pg_indexes WHERE schemaname='public' AND indexname='users_email_unique_canonical'`,
	).Scan(&indexDefinition))
	assert.Contains(t, indexDefinition, "translate(btrim")

	tx, err := server.DB.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, tx.Rollback(ctx))
	}()
	_, err = tx.Exec(ctx, `SET LOCAL enable_seqscan=off`)
	require.NoError(t, err)
	rows, err := tx.Query(ctx, `EXPLAIN SELECT id FROM users WHERE TRANSLATE(BTRIM(email, E' \t\n\r\f\v'), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz')=$1`, canonicalEmail)
	require.NoError(t, err)
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		require.NoError(t, rows.Scan(&line))
		plan.WriteString(line)
		plan.WriteByte('\n')
	}
	require.NoError(t, rows.Err())
	assert.Contains(t, plan.String(), "users_email_unique_canonical")
}
