package db

import (
	"database/sql"
	"fmt"

	"github.com/tkasuz/s3local/internal/adapters/db/sqlite"
)

// DatabaseAdapter defines the interface for database initialization
type DatabaseAdapter interface {
	// Open initializes and opens a database connection
	// Returns a sql.DB instance and any error encountered
	Open() (*sql.DB, error)

	// Close closes the database connection
	Close() error

	// Name returns the name of the storage adapter
	Name() string
}

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeSQLite StorageType = "sqlite"
)

// Config holds configuration for storage initialization
type Config struct {
	Type   StorageType
	DBPath string // For SQLite: path to database file (use ":memory:" for in-memory)
}

// NewAdapter creates a new database adapter based on the configuration
func NewAdapter(cfg Config) (DatabaseAdapter, error) {
	switch cfg.Type {
	case StorageTypeSQLite:
		return sqlite.NewAdapter(cfg.DBPath), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}
