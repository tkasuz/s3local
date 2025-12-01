package bucket

import (
	"database/sql"
	"net/http"

	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// GetBucketPolicy handles GET /{bucket}?policy
func GetBucketPolicy(w http.ResponseWriter, r *http.Request) {
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

	// Get policy
	policy, err := store.Queries.GetBucketPolicy(r.Context(), bucketName)
	if err != nil {
		if err == sql.ErrNoRows {
			s3error.NewNoSuchBucketPolicyError(bucketName).WriteError(w)
			return
		}
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Return JSON policy
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(policy.Policy))
}
