package object

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// DeleteObject handles DELETE /{bucket}/{key}
func DeleteObject(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")

	err := store.Queries.DeleteObject(r.Context(), db.DeleteObjectParams{
		BucketName: bucket,
		Key:        key,
	})
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// S3 returns 204 No Content even if the object didn't exist
	w.WriteHeader(http.StatusNoContent)
}
