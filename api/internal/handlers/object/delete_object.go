package object

import (
	"net/http"

	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// DeleteObject handles DELETE /{bucket}/{key}
func DeleteObject(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := ctx.GetBucketName(r.Context())
	objectKey := ctx.GetObjectKey(r.Context())

	err := store.Queries.DeleteObject(r.Context(), db.DeleteObjectParams{
		BucketName: bucketName,
		Key:        objectKey,
	})
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// S3 returns 204 No Content even if the object didn't exist
	w.WriteHeader(http.StatusNoContent)
}

// DeleteObjectRequest represents the S3 DeleteObject request
type DeleteObjectRequest struct {
	// Path parameters
	Bucket string
	Key    string

	// Request headers
	Headers DeleteObjectRequestHeaders
}

// DeleteObjectRequestHeaders represents request headers for DeleteObject
type DeleteObjectRequestHeaders struct {
	MFA                       string // x-amz-mfa
	VersionId                 string // versionId (query parameter)
	BypassGovernanceRetention string // x-amz-bypass-governance-retention
}
