package bucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/testutil"
)

func TestDeleteBucket(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create a bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Delete("/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		DeleteBucket(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call DeleteBucket via AWS SDK
	_, err = s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String("test-bucket"),
	})
	assert.NoError(t, err)

	// Verify bucket is deleted
	_, err = store.Queries.GetBucket(context.Background(), "test-bucket")
	assert.Error(t, err)
}

func TestDeleteBucket_NotFound(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(ctx.GetStore(testCtx)))
	r.Delete("/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		DeleteBucket(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call DeleteBucket via AWS SDK
	_, err := s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String("nonexistent"),
	})
	assert.Error(t, err)
}

func TestDeleteBucket_NotEmpty(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create a bucket
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "nonempty-bucket",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	// Insert an object directly to make the bucket non-empty
	_, err = store.DB.Exec(`INSERT INTO objects (bucket_name, key, data, size, content_type, etag) VALUES (?, ?, ?, ?, ?, ?)`,
		"nonempty-bucket", "test-key", []byte("test"), 4, "text/plain", "abc123")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/nonempty-bucket", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bucket", "nonempty-bucket")
	reqCtx := context.WithValue(testCtx, chi.RouteCtxKey, rctx)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	DeleteBucket(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}
