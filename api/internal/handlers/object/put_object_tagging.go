package object

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// PutObjectTagging handles PUT /{bucket}/{key}?tagging
func PutObjectTagging(w http.ResponseWriter, r *http.Request) {
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

	// Parse XML body
	if r.Body == nil || r.ContentLength == 0 {
		s3error.NewMalformedXMLError().WriteError(w)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	var tagging Tagging
	if err := xml.Unmarshal(body, &tagging); err != nil {
		s3error.NewMalformedXMLError().WriteError(w)
		return
	}

	// Validate tag count (AWS limit is 10 tags per object)
	if len(tagging.TagSet.Tag) > 10 {
		s3error.NewInvalidTagError("Tag set cannot contain more than 10 tags").WriteError(w)
		return
	}

	// Delete existing tags
	if err := store.Queries.DeleteObjectTags(r.Context(), objectID); err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Insert new tags
	for _, tag := range tagging.TagSet.Tag {
		// Validate tag key and value
		if tag.Key == "" {
			s3error.NewInvalidTagError("Tag key cannot be empty").WriteError(w)
			return
		}
		if len(tag.Key) > 128 {
			s3error.NewInvalidTagError("Tag key cannot be longer than 128 characters").WriteError(w)
			return
		}
		if len(tag.Value) > 256 {
			s3error.NewInvalidTagError("Tag value cannot be longer than 256 characters").WriteError(w)
			return
		}

		err := store.Queries.CreateObjectTag(r.Context(), db.CreateObjectTagParams{
			ObjectID: objectID,
			Key:      tag.Key,
			Value:    tag.Value,
		})
		if err != nil {
			s3error.NewInternalError(err).WriteError(w)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Tagging represents the tagging XML structure
type Tagging struct {
	XMLName struct{} `xml:"Tagging"`
	TagSet  TagSet   `xml:"TagSet"`
}

// TagSet represents the tag set
type TagSet struct {
	Tag []Tag `xml:"Tag"`
}

// Tag represents a single tag
type Tag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}
