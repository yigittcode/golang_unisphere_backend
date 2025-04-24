package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/config"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// PostgresDB database connection structure
type PostgresDB struct {
	Pool *pgxpool.Pool
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	// Create a context with timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connection string
	connString := cfg.GetPostgresConnectionString()

	// Configuration
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgxpool config: %w", err)
	}

	// Connection pool configuration
	poolConfig.MaxConns = int32(cfg.Database.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.Database.MaxIdleConns)

	// Parse max lifetime duration
	maxLifetime, err := time.ParseDuration(cfg.Database.ConnMaxLifetime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection max lifetime: %w", err)
	}
	poolConfig.MaxConnLifetime = maxLifetime

	// Add health check for connections
	poolConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		err := conn.Ping(ctx)
		if err != nil {
			logger.Warn().Err(err).Msg("Unhealthy connection detected")
			return false
		}
		return true
	}

	// Create connection pool with context for timeout
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	// Test connection with context
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}

	return &PostgresDB{Pool: pool}, nil
}

// Close closing method
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// TransactionFn is a function that executes within a transaction
type TransactionFn func(ctx context.Context, tx pgx.Tx) error

// WithTransaction runs a function within a transaction
func (db *PostgresDB) WithTransaction(ctx context.Context, fn TransactionFn) error {
	// Add timeout to context if not already present
	_, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// Begin transaction
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Rollback on panic
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r) // Re-throw panic after rollback
		}
	}()

	// Execute function within transaction
	if err := fn(ctx, tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			logger.Error().Err(rbErr).Msg("Failed to rollback transaction")
			return fmt.Errorf("error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
