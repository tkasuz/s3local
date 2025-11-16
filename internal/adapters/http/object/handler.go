package object

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/adapters/http/response"
	objectdomain "github.com/tkasuz/s3local/internal/domain/object"
)

// ObjectHandler handles S3 object operations
// Business logic is delegated to the service layer
type ObjectHandler struct {
	objectService objectdomain.ServiceInterface
}

// NewObjectHandler creates a new object handler with dependency injection
func NewObjectHandler(objectService objectdomain.ServiceInterface) *ObjectHandler {
	return &ObjectHandler{
		objectService: objectService,
	}
}

// RegisterRoutes registers object-related routes
func (h *ObjectHandler) RegisterRoutes(r chi.Router) {
	r.Get("/{bucket}", h.ListObjects)
	r.Put("/{bucket}/{key:.*}", h.PutObject)
	r.Get("/{bucket}/{key:.*}", h.GetObject)
	r.Head("/{bucket}/{key:.*}", h.HeadObject)
	r.Delete("/{bucket}/{key:.*}", h.DeleteObject)
}

// ListObjects handles GET /{bucket}
func (h *ObjectHandler) ListObjects(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	// Check if this is a bucket-only request (no key)
	if r.URL.Path == "/"+bucket || r.URL.Path == "/"+bucket+"/" {
		query := r.URL.Query()

		maxKeys, _ := strconv.ParseInt(query.Get("max-keys"), 10, 32)
		if maxKeys == 0 {
			maxKeys = 1000
		}

		marker := query.Get("marker")
		prefix := query.Get("prefix")

		// Query objects from database
		objects, err := h.objectService.ListObjects(r.Context(), bucket, marker, prefix, maxKeys)
		if err != nil {
			domainErr := objectdomain.NewInternalError(err)
			response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, "/"+bucket)
			return
		}

		xmlResult := objectdomain.ListBucketResult{
			Name:        bucket,
			Prefix:      prefix,
			Marker:      marker,
			MaxKeys:     int32(maxKeys),
			IsTruncated: int64(len(objects)) >= maxKeys,
			Contents:    make([]objectdomain.Contents, 0, len(objects)),
		}

		for _, obj := range objects {
			xmlResult.Contents = append(xmlResult.Contents, objectdomain.Contents{
				Key:          obj.Key,
				LastModified: obj.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
				ETag:         fmt.Sprintf(`"%s"`, obj.ETag),
				Size:         obj.Size,
				StorageClass: obj.StorageClass,
			})
		}

		if xmlResult.IsTruncated && len(objects) > 0 {
			xmlResult.NextMarker = objects[len(objects)-1].Key
		}

		response.WriteXML(w, http.StatusOK, xmlResult)
		return
	}

	// If we reach here, it's actually an object GET request
	h.GetObject(w, r)
}

// PutObject handles PUT /{bucket}/{key}
func (h *ObjectHandler) PutObject(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")

	// Read the body into memory
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		domainErr := objectdomain.NewInternalError(err)
		response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, "/"+bucket+"/"+key)
		return
	}
	data := buf.Bytes()

	contentType := r.Header.Get("Content-Type")
	contentEncoding := r.Header.Get("Content-Encoding")
	contentDisposition := r.Header.Get("Content-Disposition")
	cacheControl := r.Header.Get("Cache-Control")

	// Store metadata if any
	metadata := extractMetadata(r.Header)

	// Create or update the object
	result, err := h.objectService.PutObject(r.Context(), objectdomain.PutObjectParams{
		BucketName:         bucket,
		Key:                key,
		Data:               data,
		ContentType:        contentType,
		ContentEncoding:    contentEncoding,
		ContentDisposition: contentDisposition,
		CacheControl:       cacheControl,
		Metadata:           metadata,
	})

	if err != nil {
		domainErr := objectdomain.NewInternalError(err)
		response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, "/"+bucket+"/"+key)
		return
	}

	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, result.ETag))
	w.WriteHeader(http.StatusOK)
}

// GetObject handles GET /{bucket}/{key}
func (h *ObjectHandler) GetObject(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")

	obj, err := h.objectService.GetObject(r.Context(), bucket, key)
	if err == sql.ErrNoRows {
		domainErr := objectdomain.NewNoSuchKeyError(key)
		response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, "/"+bucket+"/"+key)
		return
	}
	if err != nil {
		domainErr := objectdomain.NewInternalError(err)
		response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, "/"+bucket+"/"+key)
		return
	}

	// Get metadata
	metadataRows, err := h.objectService.GetObjectMetadataByID(r.Context(), obj.ID)
	if err != nil && err != sql.ErrNoRows {
		domainErr := objectdomain.NewInternalError(err)
		response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, "/"+bucket+"/"+key)
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

// HeadObject handles HEAD /{bucket}/{key}
func (h *ObjectHandler) HeadObject(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")

	metadata, err := h.objectService.GetObjectMetadata(r.Context(), bucket, key)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get custom metadata
	metadataRows, _ := h.objectService.GetObjectMetadataByID(r.Context(), metadata.ID)

	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.Size, 10))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, metadata.ETag))
	w.Header().Set("Last-Modified", metadata.UpdatedAt.Format(http.TimeFormat))
	w.Header().Set("Accept-Ranges", "bytes")

	if metadata.ContentEncoding.Valid {
		w.Header().Set("Content-Encoding", metadata.ContentEncoding.String)
	}
	if metadata.ContentDisposition.Valid {
		w.Header().Set("Content-Disposition", metadata.ContentDisposition.String)
	}

	for _, meta := range metadataRows {
		w.Header().Set("x-amz-meta-"+meta.Key, meta.Value)
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteObject handles DELETE /{bucket}/{key}
func (h *ObjectHandler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")

	err := h.objectService.DeleteObject(r.Context(), bucket, key)
	if err != nil {
		domainErr := objectdomain.NewInternalError(err)
		response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, "/"+bucket+"/"+key)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

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
