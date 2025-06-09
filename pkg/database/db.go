package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/edosulai/pt-xyz-multifinance/pkg/config"
	"github.com/edosulai/pt-xyz-multifinance/pkg/logger"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// DB wraps sql.DB to add utility methods
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection
func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	log := logger.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	var db *sql.DB
	var err error

	// Implement retry logic
	maxRetries := 5
	for retry := 0; retry < maxRetries; retry++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			// Test the connection
			err = db.PingContext(ctx)
			if err == nil {
				break
			}
		}

		if retry < maxRetries-1 {
			delay := time.Duration(retry+1) * time.Second
			log.Warn("Failed to connect to database, retrying...",
				zap.Int("attempt", retry+1),
				zap.Duration("delay", delay),
				zap.Error(err))
			time.Sleep(delay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

// WithTx executes function f within a transaction
func (db *DB) WithTx(ctx context.Context, f func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after Rollback
		}
	}()

	err = f(tx)
	if err != nil {
		_ = tx.Rollback() // ignore error; return original error
		return err
	}

	return tx.Commit()
}

// QueryRowContext is a helper function that wraps sql.DB.QueryRowContext
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

// QueryContext is a helper function that wraps sql.DB.QueryContext
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

// ExecContext is a helper function that wraps sql.DB.ExecContext
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}
