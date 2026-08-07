package common

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// RollbackTx closes an owned transaction without hiding the operation's primary
// error. Rollback after a successful or failed commit is expected to report
// ErrTxClosed; any other cleanup failure remains observable.
func RollbackTx(ctx context.Context, tx pgx.Tx, operation string) {
	if tx == nil {
		return
	}
	rollbackCtx := context.WithoutCancel(ctx)
	if err := tx.Rollback(rollbackCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		Logger.ErrorContext(rollbackCtx, "failed to rollback transaction", "operation", operation, "error", err)
	}
}
