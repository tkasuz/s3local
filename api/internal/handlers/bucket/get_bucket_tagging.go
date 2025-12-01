package bucket

import (
	"encoding/xml"
	"net/http"

	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// GetBucketTagging handles GET /{bucket}?tagging
func GetBucketTagging(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := ctx.GetBucketName(r.Context())

	// Check if bucket exists
	exists, err := store.Queries.BucketExists(r.Context(), bucketName)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}
	if !exists {
		s3error.NewNoSuchBucketError(bucketName).WriteError(w)
		return
	}

	// Get tags
	tags, err := store.Queries.GetBucketTags(r.Context(), bucketName)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Build response
	tagging := Tagging{
		TagSet: TagSet{
			Tag: make([]Tag, len(tags)),
		},
	}

	for i, tag := range tags {
		tagging.TagSet.Tag[i] = Tag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}

	// Marshal to XML
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)

	output, err := xml.MarshalIndent(tagging, "", "  ")
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	w.Write([]byte(xml.Header))
	w.Write(output)
}
