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

func TestListObjectsV2(t *testing.T) {
	t.Parallel()
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create a bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	// Create test objects
	testData := []byte("test content")
	_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
		BucketName:   "test-bucket",
		Key:          "object1",
		Data:         testData,
		Size:         int64(len(testData)),
		ETag:         "etag1",
		ContentType:  "text/plain",
		StorageClass: "STANDARD",
	})
	assert.NoError(t, err)

	_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
		BucketName:   "test-bucket",
		Key:          "object2",
		Data:         testData,
		Size:         int64(len(testData)),
		ETag:         "etag2",
		ContentType:  "text/plain",
		StorageClass: "STANDARD",
	})
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Get("/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		ListObjectsV2(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call ListObjectsV2 via AWS SDK
	out, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String("test-bucket"),
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(out.Contents))
	assert.Equal(t, "object1", *out.Contents[0].Key)
	assert.Equal(t, "object2", *out.Contents[1].Key)
}

func TestListObjectsV2_Empty(t *testing.T) {
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
	r.Get("/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		ListObjectsV2(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call ListObjectsV2 via AWS SDK
	out, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String("test-bucket"),
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(out.Contents))
}

func TestListObjectsV2_WithPrefix(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create a bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	// Create test objects with different prefixes
	testData := []byte("test content")
	_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
		BucketName:   "test-bucket",
		Key:          "prefix1/object1",
		Data:         testData,
		Size:         int64(len(testData)),
		ETag:         "etag1",
		ContentType:  "text/plain",
		StorageClass: "STANDARD",
	})
	assert.NoError(t, err)

	_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
		BucketName:   "test-bucket",
		Key:          "prefix2/object2",
		Data:         testData,
		Size:         int64(len(testData)),
		ETag:         "etag2",
		ContentType:  "text/plain",
		StorageClass: "STANDARD",
	})
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Get("/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		ListObjectsV2(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call ListObjectsV2 via AWS SDK with prefix filter
	out, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String("test-bucket"),
		Prefix: aws.String("prefix1/"),
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(out.Contents))
	assert.Equal(t, "prefix1/object1", *out.Contents[0].Key)
}
