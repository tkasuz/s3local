package object

import (
	"encoding/xml"
	"net/http"

	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// GetObjectTagging handles GET /{bucket}/{key}?tagging
func GetObjectTagging(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := ctx.GetBucketName(r.Context())
	objectKey := ctx.GetObjectKey(r.Context())

	// Check if object exists and get object ID
	objectID, err := store.Queries.GetObjectID(r.Context(), db.GetObjectIDParams{
		BucketName: bucketName,
		Key:        objectKey,
	})
	if err != nil {
		s3error.NewNoSuchKeyError(objectKey).WriteError(w)
		return
	}

	// Get tags
	tags, err := store.Queries.GetObjectTags(r.Context(), objectID)
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
