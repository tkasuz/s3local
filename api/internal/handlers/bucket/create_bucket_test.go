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

func TestCreateBucket(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	req := httptest.NewRequest(http.MethodPut, "/new-bucket", nil)
	req.Header.Set("x-amz-bucket-region", "us-east-1")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bucket", "new-bucket")
	reqCtx := context.WithValue(testCtx, chi.RouteCtxKey, rctx)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	CreateBucket(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "/new-bucket", w.Header().Get("Location"))

	// Verify bucket was created
	bucket, err := store.Queries.GetBucket(context.Background(), "new-bucket")
	assert.NoError(t, err)
	assert.Equal(t, "new-bucket", bucket.Name)
	assert.Equal(t, "us-east-1", bucket.Region)
}

func TestCreateBucket_DefaultRegion(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	req := httptest.NewRequest(http.MethodPut, "/new-bucket", nil)
	// No region header - should default to us-east-1

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bucket", "new-bucket")
	reqCtx := context.WithValue(testCtx, chi.RouteCtxKey, rctx)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	CreateBucket(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify default region
	bucket, err := store.Queries.GetBucket(context.Background(), "new-bucket")
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", bucket.Region)
}

func TestCreateBucket_AlreadyExists(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create the bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "existing-bucket",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/existing-bucket", nil)
	req.Header.Set("x-amz-bucket-region", "us-east-1")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bucket", "existing-bucket")
	reqCtx := context.WithValue(testCtx, chi.RouteCtxKey, rctx)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	CreateBucket(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}
