package bucket

import (
	"encoding/xml"
	"net/http"
	"strconv"

	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// ListBuckets handles GET /
func ListBuckets(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())

	// Parse query parameters
	bucketRegion := r.URL.Query().Get("bucket-region")
	prefix := r.URL.Query().Get("prefix")
	maxBuckets := 10000 // Default max-buckets value

	if maxBucketsStr := r.URL.Query().Get("max-buckets"); maxBucketsStr != "" {
		if parsed, err := strconv.Atoi(maxBucketsStr); err == nil && parsed > 0 {
			maxBuckets = parsed
		}
	}

	// Prepare query parameters
	params := db.ListBucketsFilteredParams{
		Limit: int64(maxBuckets),
	}

	if bucketRegion != "" {
		params.Region = bucketRegion
	}
	if prefix != "" {
		params.Prefix = prefix
	}

	// Query with filters
	dbBuckets, err := store.Queries.ListBucketsFiltered(r.Context(), params)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Convert to response format
	buckets := make([]Bucket, len(dbBuckets))
	for i, b := range dbBuckets {
		buckets[i] = Bucket{
			Name:         b.Name,
			BucketRegion: b.Region,
			CreationDate: b.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		}
	}

	result := ListAllMyBucketsResult{
		Buckets: Buckets{Bucket: buckets},
		Owner: Owner{
			DisplayName: "s3local",
			ID:          "s3local",
		},
		Prefix: prefix,
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(result)
}

// ListBucketsRequest represents the S3 ListBuckets request
type ListBucketsRequest struct {
	// Query parameters
	BucketRegion      string // bucket-region
	ContinuationToken string // continuation-token
	MaxBuckets        int    // max-buckets (default: 10000)
	Prefix            string // prefix
}

// ListAllMyBucketsResult represents the S3 ListBuckets response body
type ListAllMyBucketsResult struct {
	XMLName           struct{} `xml:"ListAllMyBucketsResult"`
	Buckets           Buckets  `xml:"Buckets"`
	Owner             Owner    `xml:"Owner"`
	ContinuationToken string   `xml:"ContinuationToken,omitempty"`
	Prefix            string   `xml:"Prefix,omitempty"`
}

// Buckets wraps a list of buckets
type Buckets struct {
	Bucket []Bucket `xml:"Bucket"`
}

// Bucket represents a single bucket in the response
type Bucket struct {
	BucketArn    string `xml:"BucketArn,omitempty"`
	BucketRegion string `xml:"BucketRegion,omitempty"`
	CreationDate string `xml:"CreationDate"`
	Name         string `xml:"Name"`
}

// Owner represents the bucket owner
type Owner struct {
	DisplayName string `xml:"DisplayName"`
	ID          string `xml:"ID"`
}
