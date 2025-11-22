package bucket

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// CreateBucket handles PUT /{bucket}
func CreateBucket(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := chi.URLParam(r, "bucket")
	region := r.Header.Get("x-amz-bucket-region")
	if region == "" {
		region = "us-east-1"
	}

	err := store.Queries.CreateBucket(r.Context(), db.CreateBucketParams{
		Name:   bucketName,
		Region: region,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			s3error.NewBucketAlreadyExistsError(bucketName).WriteError(w)
			return
		}
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	w.Header().Set("Location", "/"+bucketName)
	w.WriteHeader(http.StatusOK)
}

// CreateBucketRequest represents the S3 CreateBucket request
type CreateBucketRequest struct {
	// Path parameter
	Bucket string

	// Request body (optional)
	Body *CreateBucketConfiguration
}

// CreateBucketConfiguration represents the S3 CreateBucket request body
type CreateBucketConfiguration struct {
	XMLName            struct{}        `xml:"CreateBucketConfiguration"`
	LocationConstraint string          `xml:"LocationConstraint,omitempty"`
	Location           *BucketLocation `xml:"Location,omitempty"`
	Bucket             *BucketInfo     `xml:"Bucket,omitempty"`
	Tags               *TagSet         `xml:"Tags,omitempty"`
}

// BucketLocation specifies the location for a directory bucket
type BucketLocation struct {
	Name string `xml:"Name"`
	Type string `xml:"Type"`
}

// BucketInfo specifies bucket configuration
type BucketInfo struct {
	DataRedundancy string `xml:"DataRedundancy"`
	Type           string `xml:"Type"`
}

// TagSet represents a collection of tags
type TagSet struct {
	Tag []Tag `xml:"Tag"`
}

// Tag represents a key-value tag
type Tag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

// CreateBucketResponseHeaders represents response headers for CreateBucket
type CreateBucketResponseHeaders struct {
	Location  string // Location header: /{bucket}
	BucketArn string // x-amz-bucket-arn header
}
