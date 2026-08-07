package common

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rollbackStubTx struct {
	pgx.Tx
	rollback func(context.Context) error
}

func (tx rollbackStubTx) Rollback(ctx context.Context) error {
	return tx.rollback(ctx)
}

func TestRollbackTxUsesLiveCleanupContextAndClassifiesErrors(t *testing.T) {
	var logs bytes.Buffer
	originalLogger := Logger
	Logger = slog.New(slog.NewTextHandler(&logs, nil))
	defer func() { Logger = originalLogger }()

	type contextKey string
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey("request"), "request-42"))
	cancel()

	called := false
	RollbackTx(ctx, rollbackStubTx{rollback: func(rollbackCtx context.Context) error {
		called = true
		require.NoError(t, rollbackCtx.Err())
		assert.Equal(t, "request-42", rollbackCtx.Value(contextKey("request")))
		return pgx.ErrTxClosed
	}}, "post-commit cleanup")
	assert.True(t, called)
	assert.Empty(t, logs.String(), "ErrTxClosed after commit must remain silent")

	RollbackTx(ctx, rollbackStubTx{rollback: func(rollbackCtx context.Context) error {
		require.NoError(t, rollbackCtx.Err())
		return errors.New("connection lost during rollback")
	}}, "definition save")
	assert.True(t, strings.Contains(logs.String(), "failed to rollback transaction"))
	assert.True(t, strings.Contains(logs.String(), "operation=\"definition save\""))
	assert.True(t, strings.Contains(logs.String(), "connection lost during rollback"))
}
