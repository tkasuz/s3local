package bucket

import (
	"encoding/xml"
	"net/http"

	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// ListBuckets handles GET /
func ListBuckets(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())

	dbBuckets, err := store.Queries.ListBuckets(r.Context())
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

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
