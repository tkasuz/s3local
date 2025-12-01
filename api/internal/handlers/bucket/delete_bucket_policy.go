package bucket

import (
	"net/http"

	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// DeleteBucketPolicy handles DELETE /{bucket}?policy
func DeleteBucketPolicy(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := ctx.GetBucketName(r.Context())

	// Check if bucket exists
	exists, err := store.Queries.BucketExists(r.Context(), bucketName)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}
	if !exists {
		s3error.NewNoSuchBucketError(bucketName).WriteError(w)
		return
	}

	// Delete policy
	if err := store.Queries.DeleteBucketPolicy(r.Context(), bucketName); err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
