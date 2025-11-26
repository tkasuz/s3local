package object

import (
	"bytes"
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

func TestPutObject(t *testing.T) {
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
	r.Put("/{bucket}/{key}", func(w http.ResponseWriter, r *http.Request) {
		PutObject(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call PutObject via AWS SDK
	testData := []byte("test content")
	_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("test-key"),
		Body:   bytes.NewReader(testData),
	})
	assert.NoError(t, err)

	// Verify object was created
	obj, err := store.Queries.GetObject(context.Background(), db.GetObjectParams{
		BucketName: "test-bucket",
		Key:        "test-key",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test-key", obj.Key)
	assert.Equal(t, testData, obj.Data)
}

func TestPutObject_Update(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create a bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	// Create initial object
	initialData := []byte("initial content")
	_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
		BucketName:   "test-bucket",
		Key:          "test-key",
		Data:         initialData,
		Size:         int64(len(initialData)),
		ETag:         "initial-etag",
		ContentType:  "text/plain",
		StorageClass: "STANDARD",
	})
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Put("/{bucket}/{key}", func(w http.ResponseWriter, r *http.Request) {
		PutObject(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Update object via AWS SDK
	updatedData := []byte("updated content")
	_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("test-key"),
		Body:   bytes.NewReader(updatedData),
	})
	assert.NoError(t, err)

	// Verify object was updated
	obj, err := store.Queries.GetObject(context.Background(), db.GetObjectParams{
		BucketName: "test-bucket",
		Key:        "test-key",
	})
	assert.NoError(t, err)
	assert.Equal(t, updatedData, obj.Data)
}
