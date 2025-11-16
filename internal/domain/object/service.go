package object

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"

	database "github.com/tkasuz/s3local/internal/sqlc"
	"github.com/tkasuz/s3local/internal/middleware"
)

// ServiceInterface defines the interface for object operations
type ServiceInterface interface {
	PutObject(ctx context.Context, params PutObjectParams) (database.CreateObjectRow, error)
	GetObject(ctx context.Context, bucketName, key string) (database.Object, error)
	GetObjectMetadata(ctx context.Context, bucketName, key string) (database.GetObjectMetadataRow, error)
	GetObjectMetadataByID(ctx context.Context, objectID int64) ([]database.GetObjectMetadataByObjectIDRow, error)
	ListObjects(ctx context.Context, bucketName, marker, prefix string, maxKeys int64) ([]database.ListObjectsRow, error)
	DeleteObject(ctx context.Context, bucketName, key string) error
}

// Service handles object business logic
type Service struct{}

// NewService creates a new object service
func NewService() *Service {
	return &Service{}
}

// PutObject stores an object
func (s *Service) PutObject(ctx context.Context, params PutObjectParams) (database.CreateObjectRow, error) {
	queries := middleware.GetQueries(ctx)

	// Calculate MD5 for ETag
	hash := md5.Sum(params.Data)
	etag := hex.EncodeToString(hash[:])

	if params.ContentType == "" {
		params.ContentType = "application/octet-stream"
	}

	// Create or update the object
	obj, err := queries.CreateObject(ctx, database.CreateObjectParams{
		BucketName:         params.BucketName,
		Key:                params.Key,
		Data:               params.Data,
		Size:               int64(len(params.Data)),
		ETag:               etag,
		ContentType:        params.ContentType,
		ContentEncoding:    sql.NullString{String: params.ContentEncoding, Valid: params.ContentEncoding != ""},
		ContentDisposition: sql.NullString{String: params.ContentDisposition, Valid: params.ContentDisposition != ""},
		CacheControl:       sql.NullString{String: params.CacheControl, Valid: params.CacheControl != ""},
		StorageClass:       "STANDARD",
	})

	if err != nil {
		return database.CreateObjectRow{}, err
	}

	// Store metadata if any
	for k, v := range params.Metadata {
		err := queries.CreateObjectMetadata(ctx, database.CreateObjectMetadataParams{
			ObjectID: obj.ID,
			Key:      k,
			Value:    v,
		})
		if err != nil {
			// Log but don't fail the request
			continue
		}
	}

	return obj, nil
}

// GetObject retrieves an object
func (s *Service) GetObject(ctx context.Context, bucketName, key string) (database.Object, error) {
	queries := middleware.GetQueries(ctx)
	return queries.GetObject(ctx, database.GetObjectParams{
		BucketName: bucketName,
		Key:        key,
	})
}

// GetObjectMetadata retrieves object metadata (without data)
func (s *Service) GetObjectMetadata(ctx context.Context, bucketName, key string) (database.GetObjectMetadataRow, error) {
	queries := middleware.GetQueries(ctx)
	return queries.GetObjectMetadata(ctx, database.GetObjectMetadataParams{
		BucketName: bucketName,
		Key:        key,
	})
}

// GetObjectMetadataByID retrieves custom metadata for an object
func (s *Service) GetObjectMetadataByID(ctx context.Context, objectID int64) ([]database.GetObjectMetadataByObjectIDRow, error) {
	queries := middleware.GetQueries(ctx)
	return queries.GetObjectMetadataByObjectID(ctx, objectID)
}

// ListObjects lists objects in a bucket
func (s *Service) ListObjects(ctx context.Context, bucketName, marker, prefix string, maxKeys int64) ([]database.ListObjectsRow, error) {
	queries := middleware.GetQueries(ctx)
	return queries.ListObjects(ctx, database.ListObjectsParams{
		BucketName: bucketName,
		Column2:    marker,
		Key:        marker,
		Column4:    prefix,
		Column5:    sql.NullString{String: prefix, Valid: prefix != ""},
		Limit:      maxKeys,
	})
}

// DeleteObject deletes an object
func (s *Service) DeleteObject(ctx context.Context, bucketName, key string) error {
	queries := middleware.GetQueries(ctx)
	return queries.DeleteObject(ctx, database.DeleteObjectParams{
		BucketName: bucketName,
		Key:        key,
	})
}
