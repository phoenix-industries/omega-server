// Package database provides database connection and transaction management.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"

	"github.com/phoenix-industries/omega-server/internal/buildinfo"
)

const defaultTimeout = 3 * time.Second

// Database contains Postgres database connection.
type Database struct {
	pool   *pgxpool.Pool
	config *Config
	logger *slog.Logger
}

// Connect to Postgres database.
func Connect(ctx context.Context, config *Config) (*Database, error) {
	return ConnectWithLogger(ctx, config, slog.Default())
}

// ConnectWithLogger to Postgres database with custom logger.
func ConnectWithLogger(ctx context.Context, config *Config, logger *slog.Logger) (*Database, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	conn := config.ConnectionString()
	pgxConfig, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection config: %w", err)
	}

	pgxConfig.PrepareConn = func(ctx context.Context, conn *pgx.Conn) (bool, error) {
		err := conn.Ping(ctx)
		return err == nil, err
	}

	logLevel := tracelog.LogLevelWarn
	if buildinfo.DevMode() {
		logLevel = tracelog.LogLevelTrace
	}

	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   NewLogger(logger),
		LogLevel: logLevel,
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &Database{
		config: config,
		pool:   pool,
		logger: logger,
	}, nil
}

func (db *Database) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *Database) Config() *Config {
	return db.config
}

// Close closes database connection pool.
func (db *Database) Close() {
	db.logger.Info("database: closing connection pool")
	db.pool.Close()
}
