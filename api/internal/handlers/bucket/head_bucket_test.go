package bucket

import (
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

func TestHeadBucket(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create a bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "existing-bucket",
		Region: "us-west-2",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodHead, "/existing-bucket", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bucket", "existing-bucket")
	reqCtx := context.WithValue(testCtx, chi.RouteCtxKey, rctx)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	HeadBucket(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "us-west-2", w.Header().Get("x-amz-bucket-region"))
}

func TestHeadBucket_NotFound(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)

	req := httptest.NewRequest(http.MethodHead, "/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bucket", "nonexistent")
	reqCtx := context.WithValue(testCtx, chi.RouteCtxKey, rctx)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	HeadBucket(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	// HEAD requests should not have a body
	assert.Empty(t, w.Body.String())
}
