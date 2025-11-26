package object

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

func TestDeleteObject(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create a bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	// Create an object
	testData := []byte("test content")
	_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
		BucketName:   "test-bucket",
		Key:          "test-key",
		Data:         testData,
		Size:         int64(len(testData)),
		ETag:         "test-etag",
		ContentType:  "text/plain",
		StorageClass: "STANDARD",
	})
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Delete("/{bucket}/{key}", func(w http.ResponseWriter, r *http.Request) {
		DeleteObject(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call DeleteObject via AWS SDK
	_, err = s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("test-key"),
	})
	assert.NoError(t, err)

	// Verify object was deleted
	_, err = store.Queries.GetObject(context.Background(), db.GetObjectParams{
		BucketName: "test-bucket",
		Key:        "test-key",
	})
	assert.Error(t, err)
}

func TestDeleteObject_NotFound(t *testing.T) {
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
	r.Delete("/{bucket}/{key}", func(w http.ResponseWriter, r *http.Request) {
		DeleteObject(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call DeleteObject via AWS SDK for non-existent key
	// S3 returns success even if object doesn't exist
	_, err = s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("nonexistent-key"),
	})
	assert.NoError(t, err)
}
