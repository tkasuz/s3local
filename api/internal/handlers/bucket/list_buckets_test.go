package bucket

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/testutil"
)

func TestListBuckets(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create test buckets
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket-1",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	err = store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket-2",
		Region: "us-west-2",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(testCtx)
	w := httptest.NewRecorder()

	ListBuckets(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))

	var result ListAllMyBucketsResult
	err = xml.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Buckets.Bucket))
	assert.Equal(t, "test-bucket-1", result.Buckets.Bucket[0].Name)
	assert.Equal(t, "test-bucket-2", result.Buckets.Bucket[1].Name)
}

func TestListBuckets_Empty(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(testCtx)
	w := httptest.NewRecorder()

	ListBuckets(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result ListAllMyBucketsResult
	err := xml.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Buckets.Bucket))
}
