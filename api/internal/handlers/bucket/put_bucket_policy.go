package bucket

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// PutBucketPolicy handles PUT /{bucket}?policy
func PutBucketPolicy(w http.ResponseWriter, r *http.Request) {
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

	// Parse JSON body
	if r.Body == nil || r.ContentLength == 0 {
		s3error.NewMalformedPolicyError("Policy document is required").WriteError(w)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	// Validate JSON
	var policy map[string]interface{}
	if err := json.Unmarshal(body, &policy); err != nil {
		s3error.NewMalformedPolicyError("Policy document is not valid JSON").WriteError(w)
		return
	}

	// Basic validation - check required fields
	if _, ok := policy["Version"]; !ok {
		s3error.NewMalformedPolicyError("Policy document must contain a Version field").WriteError(w)
		return
	}
	if _, ok := policy["Statement"]; !ok {
		s3error.NewMalformedPolicyError("Policy document must contain a Statement field").WriteError(w)
		return
	}

	// Store the policy
	err = store.Queries.PutBucketPolicy(r.Context(), db.PutBucketPolicyParams{
		BucketName: bucketName,
		Policy:     string(body),
	})
	if err != nil {
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
