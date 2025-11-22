package object

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// PutObject handles PUT /{bucket}/{key}
func PutObject(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")

	// Read the body into memory
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}
	data := buf.Bytes()

	// Calculate ETag (MD5 hash)
	hash := md5.Sum(data)
	etag := hex.EncodeToString(hash[:])

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Check if object exists
	exists, err := store.Queries.ObjectExists(r.Context(), db.ObjectExistsParams{
		BucketName: bucket,
		Key:        key,
	})
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	if exists {
		// Update existing object
		err = store.Queries.UpdateObject(r.Context(), db.UpdateObjectParams{
			BucketName:         bucket,
			Key:                key,
			Data:               data,
			Size:               int64(len(data)),
			ETag:               etag,
			ContentType:        contentType,
			ContentEncoding:    toNullString(r.Header.Get("Content-Encoding")),
			ContentDisposition: toNullString(r.Header.Get("Content-Disposition")),
			CacheControl:       toNullString(r.Header.Get("Cache-Control")),
			StorageClass:       "STANDARD",
		})
	} else {
		// Create new object
		_, err = store.Queries.CreateObject(r.Context(), db.CreateObjectParams{
			BucketName:         bucket,
			Key:                key,
			Data:               data,
			Size:               int64(len(data)),
			ETag:               etag,
			ContentType:        contentType,
			ContentEncoding:    toNullString(r.Header.Get("Content-Encoding")),
			ContentDisposition: toNullString(r.Header.Get("Content-Disposition")),
			CacheControl:       toNullString(r.Header.Get("Cache-Control")),
			StorageClass:       "STANDARD",
		})
	}

	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Handle metadata
	metadata := extractMetadata(r.Header)
	if len(metadata) > 0 {
		objectID, err := store.Queries.GetObjectID(r.Context(), db.GetObjectIDParams{
			BucketName: bucket,
			Key:        key,
		})
		if err != nil {
			s3error.NewInternalError(err).WriteError(w)
			return
		}

		// Delete existing metadata and insert new
		_ = store.Queries.DeleteObjectMetadata(r.Context(), objectID)
		for k, v := range metadata {
			_ = store.Queries.CreateObjectMetadata(r.Context(), db.CreateObjectMetadataParams{
				ObjectID: objectID,
				Key:      k,
				Value:    v,
			})
		}
	}

	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, etag))
	w.WriteHeader(http.StatusOK)
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func extractMetadata(headers http.Header) map[string]string {
	metadata := make(map[string]string)
	for key, values := range headers {
		if strings.HasPrefix(strings.ToLower(key), "x-amz-meta-") {
			metaKey := key[len("x-amz-meta-"):]
			if len(values) > 0 {
				metadata[metaKey] = values[0]
			}
		}
	}
	return metadata
}
