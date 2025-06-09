package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/edosulai/pt-xyz-multifinance/pkg/config"
	"github.com/edosulai/pt-xyz-multifinance/pkg/database"
	_ "github.com/lib/pq"
)

// TestDB represents a test database instance
type TestDB struct {
	DB *database.DB
}

// NewTestDB creates a new test database connection
func NewTestDB(t *testing.T) *TestDB {
	// Initialize logger first
	if err := InitTestLogger(t); err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	cfg := getTestConfig()
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Configure connection pool for tests
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Minute)

	return &TestDB{DB: db}
}

// Cleanup cleans up the test database
func (tdb *TestDB) Cleanup() error {
	return tdb.DB.Close()
}

// TruncateTables truncates all tables in the test database
func (tdb *TestDB) TruncateTables(ctx context.Context) error {
	_, err := tdb.DB.ExecContext(ctx, "TRUNCATE TABLE users, loans, documents CASCADE")
	return err
}

// SetupTestDB initializes a test database and returns cleanup function
func SetupTestDB(t *testing.T) (*TestDB, context.Context, func()) {
	ctx := context.Background()
	testDB := NewTestDB(t)

	// Clean up old data
	err := testDB.TruncateTables(ctx)
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		if err := testDB.Cleanup(); err != nil {
			t.Errorf("failed to cleanup test db: %v", err)
		}
	}

	return testDB, ctx, cleanup
}

// getTestConfig returns configuration for test database
func getTestConfig() *config.Config {
	return &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "root",
			Name:     "xyz_multifinance_test",
			SSLMode:  "disable",
		},
	}
}
