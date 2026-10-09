package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/phoenix-industries/omega-server/assets"
)

// Migrate runs database migrations using goose.
func (db *Database) Migrate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	goose.SetBaseFS(assets.Migrations)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(db.pool)
	if err := goose.UpContext(ctx, sqlDB, "migrations"); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}
