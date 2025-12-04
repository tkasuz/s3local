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
	t.Parallel()
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Route("/{bucket}", func(r chi.Router) {
		r.Use(ctx.WithBucketName())

		// Use wildcard to match any object key path including nested paths
		r.With(ctx.WithObjectKey()).Put("/*", func(w http.ResponseWriter, r *http.Request) {
			PutObject(w, r)
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	t.Run("Successfully upload new object", func(t *testing.T) {
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   "test-bucket",
			Region: "us-east-1",
		})
		assert.NoError(t, err)

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
	})

	t.Run("Successfully upload with the same object key", func(t *testing.T) {
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   "test-bucket-update",
			Region: "us-east-1",
		})
		assert.NoError(t, err)

		// Create initial object
		initialData := []byte("initial content")
		_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
			BucketName:   "test-bucket-update",
			Key:          "test-key",
			Data:         initialData,
			Size:         int64(len(initialData)),
			ETag:         "initial-etag",
			ContentType:  "text/plain",
			StorageClass: "STANDARD",
		})
		assert.NoError(t, err)
		// Update object via AWS SDK
		updatedData := []byte("updated content")
		_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String("test-bucket-update"),
			Key:    aws.String("test-key"),
			Body:   bytes.NewReader(updatedData),
		})

		assert.NoError(t, err)

		// Verify object was updated
		obj, err := store.Queries.GetObject(context.Background(), db.GetObjectParams{
			BucketName: "test-bucket-update",
			Key:        "test-key",
		})
		assert.NoError(t, err)
		assert.Equal(t, updatedData, obj.Data)
	})

	t.Run("Successfully upload object with trailing slash (folder marker)", func(t *testing.T) {
		bucketName := "test-bucket-trailing-slash"
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   bucketName,
			Region: "us-east-1",
		})
		assert.NoError(t, err)

		// Upload object with trailing slash (folder marker)
		testData := []byte("")
		_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("folder1/"),
			Body:   bytes.NewReader(testData),
		})
		assert.NoError(t, err)

		// Verify object was created with the trailing slash
		obj, err := store.Queries.GetObject(context.Background(), db.GetObjectParams{
			BucketName: bucketName,
			Key:        "folder1/",
		})
		assert.NoError(t, err)
		assert.Equal(t, "folder1/", obj.Key, "Key should include the trailing slash")
		assert.Equal(t, int64(0), obj.Size)
	})

	t.Run("Successfully upload object with nested path and trailing slash", func(t *testing.T) {
		bucketName := "test-bucket-nested-trailing-slash"
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   bucketName,
			Region: "us-east-1",
		})
		assert.NoError(t, err)

		// Upload object with nested path and trailing slash
		testData := []byte("")
		_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("parent/child/subfolder/"),
			Body:   bytes.NewReader(testData),
		})
		assert.NoError(t, err)

		// Verify object was created with the full key including trailing slash
		obj, err := store.Queries.GetObject(context.Background(), db.GetObjectParams{
			BucketName: bucketName,
			Key:        "parent/child/subfolder/",
		})
		assert.NoError(t, err)
		assert.Equal(t, "parent/child/subfolder/", obj.Key, "Key should include the full path with trailing slash")
	})
}
