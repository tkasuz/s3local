package bucket

import (
	"context"
	"database/sql"
	"errors"

	database "github.com/tkasuz/s3local/internal/sqlc"
	"github.com/tkasuz/s3local/internal/middleware"
)

// ServiceInterface defines the interface for bucket operations
type ServiceInterface interface {
	ListBuckets(ctx context.Context) ([]database.Bucket, error)
	CreateBucket(ctx context.Context, name, region string) error
	DeleteBucket(ctx context.Context, name string) error
	GetBucket(ctx context.Context, name string) (database.Bucket, error)
}

// Service handles bucket business logic
type Service struct{}

// NewService creates a new bucket service
func NewService() *Service {
	return &Service{}
}

// ListBuckets returns all buckets
func (s *Service) ListBuckets(ctx context.Context) ([]database.Bucket, error) {
	queries := middleware.GetQueries(ctx)
	return queries.ListBuckets(ctx)
}

// CreateBucket creates a new bucket
func (s *Service) CreateBucket(ctx context.Context, name, region string) error {
	queries := middleware.GetQueries(ctx)
	err := queries.CreateBucket(ctx, database.CreateBucketParams{
		Name:   name,
		Region: region,
	})
	if err != nil {
		// Check for unique constraint violation (bucket already exists)
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "UNIQUE constraint failed: buckets.name" {
			return NewBucketAlreadyExistsError(name)
		}
		return NewInternalError(err)
	}
	return nil
}

// DeleteBucket deletes a bucket if it's empty
func (s *Service) DeleteBucket(ctx context.Context, name string) error {
	queries := middleware.GetQueries(ctx)

	// Check if bucket exists first
	_, err := queries.GetBucket(ctx, name)
	if err == sql.ErrNoRows {
		return NewNoSuchBucketError(name)
	}
	if err != nil {
		return NewInternalError(err)
	}

	// Check if bucket has objects
	count, err := queries.CountObjectsInBucket(ctx, name)
	if err != nil {
		return NewInternalError(err)
	}
	if count > 0 {
		return ErrBucketNotEmpty
	}

	err = queries.DeleteBucket(ctx, name)
	if err != nil {
		return NewInternalError(err)
	}
	return nil
}

// GetBucket retrieves a bucket by name
func (s *Service) GetBucket(ctx context.Context, name string) (database.Bucket, error) {
	queries := middleware.GetQueries(ctx)
	bucket, err := queries.GetBucket(ctx, name)
	if err == sql.ErrNoRows {
		return database.Bucket{}, NewNoSuchBucketError(name)
	}
	if err != nil {
		return database.Bucket{}, NewInternalError(err)
	}
	return bucket, nil
}
