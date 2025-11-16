package bucket

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/adapters/http/response"
	bucketdomain "github.com/tkasuz/s3local/internal/domain/bucket"
)

// BucketHandler handles S3 bucket operations
type BucketHandler struct {
	bucketService bucketdomain.ServiceInterface
}

// NewBucketHandler creates a new bucket handler with dependency injection
func NewBucketHandler(bucketService bucketdomain.ServiceInterface) *BucketHandler {
	return &BucketHandler{
		bucketService: bucketService,
	}
}

// RegisterRoutes registers bucket-related routes
func (h *BucketHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListBuckets)
	r.Put("/{bucket}", h.CreateBucket)
	r.Delete("/{bucket}", h.DeleteBucket)
	r.Head("/{bucket}", h.HeadBucket)
}

// writeError writes a domain error or internal error response
func (h *BucketHandler) writeError(w http.ResponseWriter, err error, resource string) {
	var domainErr *bucketdomain.DomainError
	if errors.As(err, &domainErr) {
		response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, resource)
	} else {
		// If it's not a domain error, treat it as an internal error
		domainErr = bucketdomain.NewInternalError(err)
		response.WriteDomainError(w, string(domainErr.Code), domainErr.Message, domainErr.StatusCode, resource)
	}
}

// ListBuckets handles GET / (list all buckets)
func (h *BucketHandler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	buckets, err := h.bucketService.ListBuckets(r.Context())
	if err != nil {
		h.writeError(w, err, "/")
		return
	}

	result := bucketdomain.ListAllMyBucketsResult{
		Owner: bucketdomain.Owner{
			ID:          "local-user",
			DisplayName: "Local User",
		},
		Buckets: bucketdomain.Buckets{
			Bucket: make([]bucketdomain.Bucket, 0, len(buckets)),
		},
	}

	for _, b := range buckets {
		result.Buckets.Bucket = append(result.Buckets.Bucket, bucketdomain.Bucket{
			Name:         b.Name,
			CreationDate: b.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		})
	}

	response.WriteXML(w, http.StatusOK, result)
}

// CreateBucket handles PUT /{bucket}
func (h *BucketHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	bucketName := chi.URLParam(r, "bucket")
	region := r.Header.Get("x-amz-bucket-region")
	if region == "" {
		region = "us-east-1"
	}

	err := h.bucketService.CreateBucket(r.Context(), bucketName, region)
	if err != nil {
		h.writeError(w, err, "/"+bucketName)
		return
	}

	w.Header().Set("Location", "/"+bucketName)
	w.WriteHeader(http.StatusOK)
}

// DeleteBucket handles DELETE /{bucket}
func (h *BucketHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	bucketName := chi.URLParam(r, "bucket")

	err := h.bucketService.DeleteBucket(r.Context(), bucketName)
	if err != nil {
		h.writeError(w, err, "/"+bucketName)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HeadBucket handles HEAD /{bucket}
func (h *BucketHandler) HeadBucket(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	_, err := h.bucketService.GetBucket(r.Context(), bucket)
	if err != nil {
		// For HEAD requests, we only return status codes, not error bodies
		var domainErr *bucketdomain.DomainError
		if errors.As(err, &domainErr) {
			w.WriteHeader(domainErr.StatusCode)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

