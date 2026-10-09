package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// InTx runs the given function f within a transaction.
func (db *Database) InTx(ctx context.Context, f func(tx pgx.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := f(tx); err != nil {
		if rerr := tx.Rollback(ctx); rerr != nil {
			db.logger.ErrorContext(ctx, "failed to rollback transaction", "error", rerr, "original_error", err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
