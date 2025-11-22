package object

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// GetObject handles GET /{bucket}/{key}
func GetObject(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")

	obj, err := store.Queries.GetObject(r.Context(), db.GetObjectParams{
		BucketName: bucket,
		Key:        key,
	})
	if err == sql.ErrNoRows {
		s3error.NewNoSuchKeyError(key).WriteError(w)
		return
	}
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Get custom metadata
	metadataRows, err := store.Queries.GetObjectMetadataByObjectID(r.Context(), obj.ID)
	if err != nil && err != sql.ErrNoRows {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(obj.Size, 10))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, obj.ETag))
	w.Header().Set("Last-Modified", obj.UpdatedAt.Format(http.TimeFormat))
	w.Header().Set("Accept-Ranges", "bytes")

	if obj.ContentEncoding.Valid {
		w.Header().Set("Content-Encoding", obj.ContentEncoding.String)
	}
	if obj.ContentDisposition.Valid {
		w.Header().Set("Content-Disposition", obj.ContentDisposition.String)
	}
	if obj.CacheControl.Valid {
		w.Header().Set("Cache-Control", obj.CacheControl.String)
	}

	// Copy metadata to response headers
	for _, meta := range metadataRows {
		w.Header().Set("x-amz-meta-"+meta.Key, meta.Value)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(obj.Data)
}
