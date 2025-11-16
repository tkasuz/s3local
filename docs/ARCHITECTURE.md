# Architecture Overview

This document describes the clean architecture implementation of s3local.

## Layered Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Layer (Chi)                      │
│                     main.go                              │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│            Adapters Layer (HTTP Handlers)                │
│  internal/adapters/http/{bucket,object}/handler.go      │
│  - Extract HTTP parameters                               │
│  - Call service methods                                  │
│  - Write HTTP responses                                  │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│            Domain Layer (Business Logic)                 │
│  internal/domain/{bucket,object}/service.go             │
│  - Business rules                                        │
│  - Error translation (sql.ErrNoRows → DomainError)      │
│  - Data validation                                       │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│           Infrastructure Layer (Database)                │
│  internal/sqlc/ (generated queries)                      │
│  internal/storage/ (database adapters)                   │
└─────────────────────────────────────────────────────────┘
```

### Service Layer Responsibility

The service layer translates infrastructure errors to domain errors:

```go
// internal/domain/bucket/service.go
func (s *Service) GetBucket(ctx context.Context, name string) (database.Bucket, error) {
    queries := middleware.GetQueries(ctx)
    bucket, err := queries.GetBucket(ctx, name)

    if err == sql.ErrNoRows {
        return database.Bucket{}, NewNoSuchBucketError(name) // ✅ Domain error
    }
    if err != nil {
        return database.Bucket{}, NewInternalError(err)
    }
    return bucket, nil
}
```

### Adapter Layer Responsibility

The adapter layer handles domain errors and writes HTTP responses:

```go
// internal/adapters/http/bucket/handler.go
func (h *BucketHandler) writeError(w http.ResponseWriter, err error, resource string) {
    var domainErr *bucketdomain.DomainError
    if errors.As(err, &domainErr) {
        // Write S3-compatible error response
        response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, resource)
    } else {
        // Fallback to internal error
        domainErr = bucketdomain.NewInternalError(err)
        response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, resource)
    }
}
```

## Wiring Dependencies (main.go)

```go
func main() {
    // 1. Initialize infrastructure (database)
    dbAdapter, _ := storage.NewAdapter(storage.Config{
        Type:   storage.StorageTypeSQLite,
        DBPath: dbPath,
    })
    db, _ := dbAdapter.Open()
    defer dbAdapter.Close()

    // 2. Create domain services
    bucketService := bucketdomain.NewService()
    objectService := objectdomain.NewService()

    // 3. Create HTTP handlers with injected services
    bucketHandler := bucket.NewBucketHandler(bucketService)
    objectHandler := object.NewObjectHandler(objectService)

    // 4. Register routes
    bucketHandler.RegisterRoutes(r)
    objectHandler.RegisterRoutes(r)
}
```

## Testing Strategy

### Unit Tests (with Mocks)

```go
func TestDeleteBucket(t *testing.T) {
    // Create mock service
    mockService := mockbucket.NewMockServiceInterface(t)

    // Inject mock into handler
    handler := NewBucketHandler(mockService)

    // Setup expectations - service returns domain error
    mockService.EXPECT().
        DeleteBucket(mock.Anything, "nonexistent").
        Return(bucketdomain.NewNoSuchBucketError("nonexistent")).
        Once()

    // Test handler behavior
    handler.DeleteBucket(w, req)
    assert.Equal(t, http.StatusNotFound, w.Code)
}
```

### Integration Tests (with Real AWS SDK)

```go
func setupTestServer(t *testing.T) (*httptest.Server, *s3.Client) {
    // Real database (in-memory SQLite)
    dbAdapter := sqlite.NewAdapter(":memory:")
    db, _ := dbAdapter.Open()

    // Real services
    bucketService := bucketdomain.NewService()
    objectService := objectdomain.NewService()

    // Real handlers
    bucketHandler := bucket.NewBucketHandler(bucketService)
    objectHandler := object.NewObjectHandler(objectService)

    // Test with real AWS S3 SDK client
    client := s3.NewFromConfig(cfg, func(o *s3.Options) {
        o.BaseEndpoint = aws.String(ts.URL)
    })
}
```

## Benefits of This Architecture

### ✅ Separation of Concerns
- **Adapters**: HTTP protocol details
- **Domain**: Business logic and rules
- **Infrastructure**: Database operations

### ✅ Testability
- Unit tests use mocks
- Integration tests use real implementations
- Easy to test each layer independently

### ✅ Maintainability
- Changes to HTTP framework don't affect business logic
- Changes to database don't affect handlers
- Clear boundaries between layers

### ✅ Extensibility
- Easy to add new storage backends (PostgreSQL, MySQL)
- Easy to add new protocols (gRPC, GraphQL)
- Easy to swap implementations via interfaces

## SOLID Principles Applied

1. **Single Responsibility**: Each layer has one reason to change
   - Handlers: HTTP protocol changes
   - Services: Business logic changes
   - Storage: Database changes

2. **Open/Closed**: Open for extension, closed for modification
   - Add new storage adapters without changing existing code
   - Add new error types without changing error handling

3. **Liskov Substitution**: Interfaces enable substitution
   - `ServiceInterface` can be swapped with mocks or different implementations

4. **Interface Segregation**: Focused interfaces
   - `ServiceInterface` contains only necessary methods
   - `DatabaseAdapter` focuses on database operations

5. **Dependency Inversion**: Depend on abstractions
   - Handlers depend on `ServiceInterface`, not concrete `Service`
   - Main function wires concrete implementations

## Error Translation Flow

```
Database Layer          Domain Layer              Adapter Layer
─────────────────       ──────────────────        ──────────────
sql.ErrNoRows    →      NewNoSuchBucketError() →  HTTP 404
                        (DomainError)              + XML body

DB Constraint    →      NewBucketAlreadyExists() → HTTP 409
Violation               (DomainError)              + XML body

Generic Error    →      NewInternalError() →      HTTP 500
                        (DomainError)              + XML body
```

This ensures:
- Database details don't leak to HTTP layer
- HTTP handlers don't need database knowledge
- S3-compatible errors at the boundary
- Domain errors in the middle
