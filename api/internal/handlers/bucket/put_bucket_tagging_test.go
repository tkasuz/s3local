package bucket

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/testutil"
)

func TestPutBucketTagging(t *testing.T) {
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
			if req.URL.Query().Has("tagging") {
				PutBucketTagging(w, req)
			}
		})
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Query().Has("tagging") {
				GetBucketTagging(w, req)
			}
		})
		r.Delete("/", func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Query().Has("tagging") {
				DeleteBucketTagging(w, req)
			}
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	t.Run("Success", func(t *testing.T) {
		// Test PUT bucket tagging
		taggingXML := `<?xml version="1.0" encoding="UTF-8"?>
<Tagging xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
   <TagSet>
      <Tag>
         <Key>Environment</Key>
         <Value>Production</Value>
      </Tag>
      <Tag>
         <Key>Team</Key>
         <Value>DevOps</Value>
      </Tag>
   </TagSet>
</Tagging>`

		req, _ := http.NewRequest("PUT", ts.URL+"/test-bucket?tagging", bytes.NewBufferString(taggingXML))
		req.Header.Set("Content-Type", "application/xml")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

		// Verify tags were created in database
		tags, err := store.Queries.GetBucketTags(context.Background(), "test-bucket")
		assert.NoError(t, err)
		assert.Len(t, tags, 2)
		assert.Equal(t, "Environment", tags[0].Key)
		assert.Equal(t, "Production", tags[0].Value)
		assert.Equal(t, "Team", tags[1].Key)
		assert.Equal(t, "DevOps", tags[1].Value)

		// Test GET bucket tagging
		req, _ = http.NewRequest("GET", ts.URL+"/test-bucket?tagging", nil)
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// Test DELETE bucket tagging
		req, _ = http.NewRequest("DELETE", ts.URL+"/test-bucket?tagging", nil)
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

		// Verify tags were deleted
		tags, err = store.Queries.GetBucketTags(context.Background(), "test-bucket")
		assert.NoError(t, err)
		assert.Len(t, tags, 0)
	})

	t.Run("Not Found", func(t *testing.T) {
		taggingXML := `<?xml version="1.0" encoding="UTF-8"?>
<Tagging xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
   <TagSet>
      <Tag>
         <Key>Environment</Key>
         <Value>Production</Value>
      </Tag>
   </TagSet>
</Tagging>`

		req, _ := http.NewRequest("PUT", ts.URL+"/nonexistent-bucket?tagging", bytes.NewBufferString(taggingXML))
		req.Header.Set("Content-Type", "application/xml")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})
}
