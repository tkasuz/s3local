# Storage Adapter Pattern

This package provides a database adapter pattern that decouples the application from specific database implementations.

## Architecture

```
storage/
├── storage.go           # DatabaseAdapter interface and factory
└── sqlite/
    ├── adapter.go       # SQLite implementation
    ├── db.go           # Migration utilities
    └── migrations/     # SQL migration files
```

## Usage

### Using the Factory (Recommended)

```go
import "github.com/tkasuz/s3local/internal/storage"

// Create adapter using factory
dbAdapter, err := storage.NewAdapter(storage.Config{
    Type:   storage.StorageTypeSQLite,
    DBPath: "s3local.db",
})
if err != nil {
    log.Fatal(err)
}

db, err := dbAdapter.Open()
if err != nil {
    log.Fatal(err)
}
defer dbAdapter.Close()
```

### Direct Adapter Usage

```go
import "github.com/tkasuz/s3local/internal/storage/sqlite"

// Create SQLite adapter directly
dbAdapter := sqlite.NewAdapter("s3local.db")

db, err := dbAdapter.Open()
if err != nil {
    log.Fatal(err)
}
defer dbAdapter.Close()
```

### In-Memory Database (Testing)

```go
// Use in-memory SQLite for tests
dbAdapter := sqlite.NewAdapter(":memory:")
db, _ := dbAdapter.Open()
defer dbAdapter.Close()
```

## Benefits

1. **Decoupling**: Main application doesn't depend on specific database implementation
2. **Testability**: Easy to swap implementations for testing
3. **Extensibility**: Add new storage backends by implementing `DatabaseAdapter`
4. **Encapsulation**: Database-specific logic stays in adapter packages

## Adding New Storage Backends

To add a new storage backend (e.g., PostgreSQL):

1. Create package: `internal/storage/postgres/`
2. Implement `storage.DatabaseAdapter` interface
3. Add new type to `StorageType` enum
4. Update factory in `storage.NewAdapter()`

Example:

```go
// internal/storage/postgres/adapter.go
package postgres

import "database/sql"

type Adapter struct {
    connStr string
    db      *sql.DB
}

func NewAdapter(connStr string) *Adapter {
    return &Adapter{connStr: connStr}
}

func (a *Adapter) Open() (*sql.DB, error) {
    db, err := sql.Open("postgres", a.connStr)
    if err != nil {
        return nil, err
    }
    // Run migrations...
    a.db = db
    return db, nil
}

func (a *Adapter) Close() error {
    if a.db != nil {
        return a.db.Close()
    }
    return nil
}

func (a *Adapter) Name() string {
    return "postgres"
}
```

Then update `storage.go`:

```go
const (
    StorageTypeSQLite   StorageType = "sqlite"
    StorageTypePostgres StorageType = "postgres" // Add new type
)

type Config struct {
    Type    StorageType
    DBPath  string // For SQLite
    ConnStr string // For PostgreSQL
}

func NewAdapter(cfg Config) (DatabaseAdapter, error) {
    switch cfg.Type {
    case StorageTypeSQLite:
        return sqlite.NewAdapter(cfg.DBPath), nil
    case StorageTypePostgres:
        return postgres.NewAdapter(cfg.ConnStr), nil
    default:
        return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
    }
}
```

## Interface Definition

```go
type DatabaseAdapter interface {
    // Open initializes and opens a database connection
    Open() (*sql.DB, error)

    // Close closes the database connection
    Close() error

    // Name returns the name of the storage adapter
    Name() string
}
```
