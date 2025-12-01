package bucket

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// PutBucketTagging handles PUT /{bucket}?tagging
func PutBucketTagging(w http.ResponseWriter, r *http.Request) {
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

	// Validate tag count (AWS limit is 50 tags per bucket)
	if len(tagging.TagSet.Tag) > 50 {
		s3error.NewInvalidTagError("Tag set cannot contain more than 50 tags").WriteError(w)
		return
	}

	// Delete existing tags
	if err := store.Queries.DeleteBucketTags(r.Context(), bucketName); err != nil {
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

		err := store.Queries.CreateBucketTag(r.Context(), db.CreateBucketTagParams{
			BucketName: bucketName,
			Key:        tag.Key,
			Value:      tag.Value,
		})
		if err != nil {
			s3error.NewInternalError(err).WriteError(w)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// Tagging represents the tagging XML structure
type Tagging struct {
	XMLName struct{} `xml:"Tagging"`
	TagSet  TagSet   `xml:"TagSet"`
}
