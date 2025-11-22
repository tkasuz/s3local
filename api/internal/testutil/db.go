package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
)

// SetupTestDB creates a temporary SQLite database with migrations applied
// for use in tests. It returns a context with the store injected.
// The database file is automatically cleaned up when the test completes.
func SetupTestDB(t *testing.T) context.Context {
	t.Helper()

	// Create a temporary database file
	tmpfile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpfile.Close()
	dbPath := tmpfile.Name()

	// Clean up the temp file when test completes
	t.Cleanup(func() {
		os.Remove(dbPath)
	})

	// Run migrations using Atlas
	dbURL := fmt.Sprintf("sqlite://%s?_fk=1", dbPath)
	if err := db.RunMigrations(dbURL); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Open the database (use the same file that was migrated)
	database, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	queries := db.New(database)
	store := db.NewStore(database, queries)

	t.Cleanup(func() {
		database.Close()
	})

	// Create a context with the store injected
	testCtx := context.Background()
	testCtx = context.WithValue(testCtx, ctx.StoreKey, store)

	return testCtx
}
