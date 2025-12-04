package object

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// ListBucketResultV2 represents the S3 ListObjectsV2 response
type ListBucketResultV2 struct {
	XMLName               xml.Name        `xml:"ListBucketResult"`
	Xmlns                 string          `xml:"xmlns,attr"`
	Name                  string          `xml:"Name"`
	Prefix                string          `xml:"Prefix"`
	Delimiter             string          `xml:"Delimiter,omitempty"`
	MaxKeys               int64           `xml:"MaxKeys"`
	KeyCount              int             `xml:"KeyCount"`
	IsTruncated           bool            `xml:"IsTruncated"`
	ContinuationToken     string          `xml:"ContinuationToken,omitempty"`
	NextContinuationToken string          `xml:"NextContinuationToken,omitempty"`
	StartAfter            string          `xml:"StartAfter,omitempty"`
	Contents              []Contents      `xml:"Contents"`
	CommonPrefixes        []CommonPrefix  `xml:"CommonPrefixes"`
}

// Contents represents an object in the list response
type Contents struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

// CommonPrefix represents a common prefix in the list response
type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// processObjectsWithDelimiter processes objects and groups them by common prefixes
// This function extracts common prefixes based on the delimiter and returns
// both the filtered contents and the common prefixes.
func processObjectsWithDelimiter(objects []db.ListObjectsRow, prefix, delimiter string) ([]Contents, []CommonPrefix) {
	contents := make([]Contents, 0)
	folderMarkers := make(map[string]bool) // Track folder marker objects
	processedPrefixes := make(map[string]bool)

	// First pass: identify folder marker objects and add them to contents
	for _, obj := range objects {
		// Skip the prefix itself if it matches exactly
		if obj.Key == prefix {
			continue
		}

		// Get the part of the key after the prefix
		keyAfterPrefix := obj.Key
		if prefix != "" && len(obj.Key) > len(prefix) {
			keyAfterPrefix = obj.Key[len(prefix):]
		}

		// Find the first occurrence of delimiter in the key after prefix
		delimiterIdx := strings.Index(keyAfterPrefix, delimiter)

		if delimiterIdx >= 0 {
			// Check if the delimiter is at the end (folder marker object)
			isLastDelimiter := delimiterIdx == len(keyAfterPrefix)-len(delimiter)

			if isLastDelimiter {
				// This is a folder marker object at the current level
				contents = append(contents, Contents{
					Key:          obj.Key,
					LastModified: obj.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
					ETag:         fmt.Sprintf(`"%s"`, obj.ETag),
					Size:         obj.Size,
					StorageClass: obj.StorageClass,
				})
				// Mark this as a folder marker so we don't create a CommonPrefix for it
				folderMarkers[obj.Key] = true
			}
		} else {
			// No delimiter found, add as regular content
			contents = append(contents, Contents{
				Key:          obj.Key,
				LastModified: obj.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
				ETag:         fmt.Sprintf(`"%s"`, obj.ETag),
				Size:         obj.Size,
				StorageClass: obj.StorageClass,
			})
		}
	}

	// Second pass: create common prefixes for nested objects, but skip if we have a folder marker
	commonPrefixes := make([]CommonPrefix, 0)
	for _, obj := range objects {
		// Skip the prefix itself if it matches exactly
		if obj.Key == prefix {
			continue
		}

		// Get the part of the key after the prefix
		keyAfterPrefix := obj.Key
		if prefix != "" && len(obj.Key) > len(prefix) {
			keyAfterPrefix = obj.Key[len(prefix):]
		}

		// Find the first occurrence of delimiter in the key after prefix
		delimiterIdx := strings.Index(keyAfterPrefix, delimiter)

		if delimiterIdx >= 0 {
			// Check if this is NOT a folder marker at the current level
			isLastDelimiter := delimiterIdx == len(keyAfterPrefix)-len(delimiter)

			if !isLastDelimiter {
				// This key has a delimiter in the middle (nested object)
				commonPrefix := prefix + keyAfterPrefix[:delimiterIdx+len(delimiter)]

				// Only add as CommonPrefix if we don't have a folder marker for it
				if !folderMarkers[commonPrefix] && !processedPrefixes[commonPrefix] {
					commonPrefixes = append(commonPrefixes, CommonPrefix{
						Prefix: commonPrefix,
					})
					processedPrefixes[commonPrefix] = true
				}
			}
		}
	}

	return contents, commonPrefixes
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
	delimiter := query.Get("delimiter")
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
		Delimiter:         delimiter,
		MaxKeys:           maxKeys,
		IsTruncated:       isTruncated,
		ContinuationToken: continuationToken,
		StartAfter:        startAfter,
		Contents:          make([]Contents, 0),
		CommonPrefixes:    make([]CommonPrefix, 0),
	}

	// Process objects and group by common prefixes if delimiter is specified
	if delimiter != "" {
		result.Contents, result.CommonPrefixes = processObjectsWithDelimiter(objects, prefix, delimiter)
	} else {
		// No delimiter, return all objects as contents
		for _, obj := range objects {
			result.Contents = append(result.Contents, Contents{
				Key:          obj.Key,
				LastModified: obj.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
				ETag:         fmt.Sprintf(`"%s"`, obj.ETag),
				Size:         obj.Size,
				StorageClass: obj.StorageClass,
			})
		}
	}

	if isTruncated && len(objects) > 0 {
		result.NextContinuationToken = objects[len(objects)-1].Key
	}

	// KeyCount should include both Contents and CommonPrefixes
	result.KeyCount = len(result.Contents) + len(result.CommonPrefixes)

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
