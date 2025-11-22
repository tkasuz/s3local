package bucket

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
)

// HeadBucket handles HEAD /{bucket}
func HeadBucket(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := chi.URLParam(r, "bucket")

	bucket, err := store.Queries.GetBucket(r.Context(), bucketName)
	if err != nil {
		// For HEAD requests, we only return status codes, not error bodies
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("x-amz-bucket-region", bucket.Region)
	w.WriteHeader(http.StatusOK)
}

// HeadBucketRequest represents the S3 HeadBucket request
type HeadBucketRequest struct {
	// Path parameter
	Bucket string

	// Request headers
	ExpectedBucketOwner string // x-amz-expected-bucket-owner
}

// HeadBucketResponseHeaders represents response headers for HeadBucket
type HeadBucketResponseHeaders struct {
	BucketRegion       string // x-amz-bucket-region
	AccessPointAlias   bool   // x-amz-access-point-alias
	BucketLocationName string // x-amz-bucket-location-name
	BucketLocationType string // x-amz-bucket-location-type
}
