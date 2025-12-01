package bucket

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/testutil"
)

func TestPutBucketPolicy(t *testing.T) {
	t.Parallel()
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create a bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Route("/{bucket}", func(r chi.Router) {
		r.Use(ctx.WithBucketName())
		r.Use(ctx.WithStore(store))
		r.Put("/", func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Query().Has("policy") {
				PutBucketPolicy(w, req)
			}
		})
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Query().Has("policy") {
				GetBucketPolicy(w, req)
			}
		})
		r.Delete("/", func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Query().Has("policy") {
				DeleteBucketPolicy(w, req)
			}
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	t.Run("Success", func(t *testing.T) {
		// Test PUT bucket policy
		policyJSON := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": "*",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::test-bucket/*"
			}
		]
	}`

		req, _ := http.NewRequest("PUT", ts.URL+"/test-bucket?policy", bytes.NewBufferString(policyJSON))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

		// Verify policy was created in database
		policy, err := store.Queries.GetBucketPolicy(context.Background(), "test-bucket")
		assert.NoError(t, err)
		assert.Contains(t, policy.Policy, "2012-10-17")
		assert.Contains(t, policy.Policy, "s3:GetObject")

		// Test GET bucket policy
		req, _ = http.NewRequest("GET", ts.URL+"/test-bucket?policy", nil)
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "2012-10-17")
		assert.Contains(t, string(body), "s3:GetObject")
		resp.Body.Close()

		// Test DELETE bucket policy
		req, _ = http.NewRequest("DELETE", ts.URL+"/test-bucket?policy", nil)
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

		// Verify policy was deleted
		_, err = store.Queries.GetBucketPolicy(context.Background(), "test-bucket")
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		policyJSON := `{
		"Version": "2012-10-17",
		"Statement": []
	}`

		req, _ := http.NewRequest("PUT", ts.URL+"/nonexistent-bucket?policy", bytes.NewBufferString(policyJSON))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})

}
