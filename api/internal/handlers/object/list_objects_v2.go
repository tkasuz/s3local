package object

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// ListBucketResultV2 represents the S3 ListObjectsV2 response
type ListBucketResultV2 struct {
	XMLName               xml.Name   `xml:"ListBucketResult"`
	Xmlns                 string     `xml:"xmlns,attr"`
	Name                  string     `xml:"Name"`
	Prefix                string     `xml:"Prefix"`
	MaxKeys               int64      `xml:"MaxKeys"`
	KeyCount              int        `xml:"KeyCount"`
	IsTruncated           bool       `xml:"IsTruncated"`
	ContinuationToken     string     `xml:"ContinuationToken,omitempty"`
	NextContinuationToken string     `xml:"NextContinuationToken,omitempty"`
	StartAfter            string     `xml:"StartAfter,omitempty"`
	Contents              []Contents `xml:"Contents"`
}

// Contents represents an object in the list response
type Contents struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

// ListObjectsV2 handles GET /{bucket}?list-type=2
func ListObjectsV2(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucket := chi.URLParam(r, "bucket")

	query := r.URL.Query()

	maxKeys, _ := strconv.ParseInt(query.Get("max-keys"), 10, 64)
	if maxKeys == 0 || maxKeys > 1000 {
		maxKeys = 1000
	}

	prefix := query.Get("prefix")
	startAfter := query.Get("start-after")
	continuationToken := query.Get("continuation-token")

	// Use continuation-token if provided, otherwise use start-after
	marker := continuationToken
	if marker == "" {
		marker = startAfter
	}

	objects, err := store.Queries.ListObjects(r.Context(), db.ListObjectsParams{
		BucketName: bucket,
		Column2:    marker,
		Key:        marker,
		Column4:    prefix,
		Column5:    sql.NullString{String: prefix, Valid: prefix != ""},
		Limit:      maxKeys + 1, // Fetch one extra to check if truncated
	})
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	isTruncated := int64(len(objects)) > maxKeys
	if isTruncated {
		objects = objects[:maxKeys]
	}

	result := ListBucketResultV2{
		Xmlns:             "http://s3.amazonaws.com/doc/2006-03-01/",
		Name:              bucket,
		Prefix:            prefix,
		MaxKeys:           maxKeys,
		KeyCount:          len(objects),
		IsTruncated:       isTruncated,
		ContinuationToken: continuationToken,
		StartAfter:        startAfter,
		Contents:          make([]Contents, 0, len(objects)),
	}

	for _, obj := range objects {
		result.Contents = append(result.Contents, Contents{
			Key:          obj.Key,
			LastModified: obj.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
			ETag:         fmt.Sprintf(`"%s"`, obj.ETag),
			Size:         obj.Size,
			StorageClass: obj.StorageClass,
		})
	}

	if isTruncated && len(objects) > 0 {
		result.NextContinuationToken = objects[len(objects)-1].Key
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(result)
}

// ListObjectsV2Request represents the S3 ListObjectsV2 request
type ListObjectsV2Request struct {
	// Path parameters
	Bucket string

	// Query parameters
	QueryParams ListObjectsV2QueryParams
}

// ListObjectsV2QueryParams represents query parameters for ListObjectsV2
type ListObjectsV2QueryParams struct {
	Delimiter         string // delimiter
	EncodingType      string // encoding-type
	MaxKeys           int64  // max-keys
	Prefix            string // prefix
	ContinuationToken string // continuation-token
	FetchOwner        bool   // fetch-owner
	StartAfter        string // start-after
}
