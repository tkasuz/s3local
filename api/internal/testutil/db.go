package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
)

var (
	templateDBPath string
	setupOnce      sync.Once
	setupErr       error
)

// initTemplateDB creates a template database with migrations applied once
// This avoids Atlas migration lock conflicts when running parallel tests
func initTemplateDB() {
	setupOnce.Do(func() {
		// Create a temporary template database file
		tmpfile, err := os.CreateTemp("", "template-*.db")
		if err != nil {
			setupErr = fmt.Errorf("failed to create template file: %w", err)
			return
		}
		tmpfile.Close()
		templateDBPath = tmpfile.Name()

		// Run migrations on the template database
		dbURL := fmt.Sprintf("sqlite://%s?_fk=1", templateDBPath)
		if err := db.RunMigrations(dbURL); err != nil {
			setupErr = fmt.Errorf("failed to run migrations on template: %w", err)
			os.Remove(templateDBPath)
			return
		}
	})
}

// SetupTestDB creates a temporary SQLite database with migrations applied
// for use in tests. It returns a context with the store injected.
// The database file is automatically cleaned up when the test completes.
func SetupTestDB(t *testing.T) context.Context {
	t.Helper()

	// Initialize template database once (thread-safe)
	initTemplateDB()
	if setupErr != nil {
		t.Fatalf("failed to initialize template database: %v", setupErr)
	}

	// Create a temporary database file for this test
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

	// Copy the template database to the test database
	if err := copyFile(templateDBPath, dbPath); err != nil {
		t.Fatalf("failed to copy template database: %v", err)
	}

	// Open the database
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

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
