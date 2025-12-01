package object

import (
	"net/http"

	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// DeleteObjectTagging handles DELETE /{bucket}/{key}?tagging
func DeleteObjectTagging(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := ctx.GetBucketName(r.Context())
	objectKey := ctx.GetObjectKey(r.Context())

	// Check if object exists and get object ID
	objectID, err := store.Queries.GetObjectID(r.Context(), db.GetObjectIDParams{
		BucketName: bucketName,
		Key:        objectKey,
	})
	if err != nil {
		s3error.NewNoSuchKeyError(objectKey).WriteError(w)
		return
	}

	// Delete all tags
	if err := store.Queries.DeleteObjectTags(r.Context(), objectID); err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
