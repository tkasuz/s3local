package bucket

import (
	"database/sql"
	"net/http"

	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// DeleteBucket handles DELETE /{bucket}
func DeleteBucket(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := ctx.GetBucketName(r.Context())

	// Check if bucket exists
	_, err := store.Queries.GetBucket(r.Context(), bucketName)
	if err == sql.ErrNoRows {
		s3error.NewNoSuchBucketError(bucketName).WriteError(w)
		return
	}
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Check if bucket has objects
	count, err := store.Queries.CountObjectsInBucket(r.Context(), bucketName)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}
	if count > 0 {
		s3error.NewBucketNotEmptyError(bucketName).WriteError(w)
		return
	}

	err = store.Queries.DeleteBucket(r.Context(), bucketName)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusNoContent)
}

// DeleteBucketRequest represents the S3 DeleteBucket request
type DeleteBucketRequest struct {
	// Path parameter
	Bucket string
}

// DeleteBucket returns 204 No Content on success with no body or special headers
