package http_test

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tkasuz/s3local/internal/adapters/http/bucket"
	"github.com/tkasuz/s3local/internal/adapters/http/object"
	bucketdomain "github.com/tkasuz/s3local/internal/domain/bucket"
	objectdomain "github.com/tkasuz/s3local/internal/domain/object"
	appmiddleware "github.com/tkasuz/s3local/internal/middleware"
	"github.com/tkasuz/s3local/internal/storage/sqlite"
)

// setupTestServer creates a test server with all handlers and middleware
func setupTestServer(t *testing.T) (*httptest.Server, *s3.Client) {
	// Create in-memory SQLite database using adapter
	dbAdapter := sqlite.NewAdapter(":memory:")
	db, err := dbAdapter.Open()
	require.NoError(t, err)
	t.Cleanup(func() { dbAdapter.Close() })

	// Create domain services
	bucketService := bucketdomain.NewService()
	objectService := objectdomain.NewService()

	// Create handlers with dependency injection
	bucketHandler := bucket.NewBucketHandler(bucketService)
	objectHandler := object.NewObjectHandler(objectService)

	// Create router
	r := chi.NewRouter()
	r.Use(appmiddleware.QueriesMiddleware(db))

	// Register routes
	bucketHandler.RegisterRoutes(r)
	objectHandler.RegisterRoutes(r)

	// Create test server
	ts := httptest.NewServer(r)
	t.Cleanup(ts.Close)

	// Create S3 client pointing to test server
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test-access-key",
			"test-secret-key",
			"",
		)),
	)
	require.NoError(t, err)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(ts.URL)
		o.UsePathStyle = true // Use path-style URLs (host/bucket) instead of virtual-hosted (bucket.host)
	})

	return ts, client
}

func TestIntegration_BucketOperations(t *testing.T) {
	_, client := setupTestServer(t)
	ctx := context.Background()

	t.Run("ListBuckets empty", func(t *testing.T) {
		output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Empty(t, output.Buckets)
	})

	t.Run("CreateBucket", func(t *testing.T) {
		_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String("test-bucket"),
		})
		require.NoError(t, err)
	})

	t.Run("ListBuckets after create", func(t *testing.T) {
		output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
		require.NoError(t, err)
		assert.Len(t, output.Buckets, 1)
		assert.Equal(t, "test-bucket", *output.Buckets[0].Name)
		assert.NotNil(t, output.Buckets[0].CreationDate)
	})

	t.Run("HeadBucket existing", func(t *testing.T) {
		_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String("test-bucket"),
		})
		require.NoError(t, err)
	})

	t.Run("HeadBucket non-existing", func(t *testing.T) {
		_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String("nonexistent-bucket"),
		})
		require.Error(t, err)
	})

	t.Run("CreateBucket duplicate", func(t *testing.T) {
		_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String("test-bucket"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BucketAlreadyExists")
	})

	t.Run("DeleteBucket non-empty fails", func(t *testing.T) {
		// First create a bucket and put an object
		bucketName := "bucket-with-objects"
		_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		require.NoError(t, err)

		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("test-file.txt"),
			Body:   strings.NewReader("test content"),
		})
		require.NoError(t, err)

		// Try to delete the bucket (should fail)
		_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BucketNotEmpty")
	})

	t.Run("DeleteBucket empty succeeds", func(t *testing.T) {
		bucketName := "empty-bucket"
		_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		require.NoError(t, err)

		_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		require.NoError(t, err)

		// Verify bucket is deleted
		_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		})
		require.Error(t, err)
	})
}

func TestIntegration_ObjectOperations(t *testing.T) {
	_, client := setupTestServer(t)
	ctx := context.Background()

	// Setup: Create a test bucket
	bucketName := "test-objects-bucket"
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	require.NoError(t, err)

	t.Run("ListObjects empty bucket", func(t *testing.T) {
		output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		require.NoError(t, err)
		assert.Empty(t, output.Contents)
	})

	t.Run("PutObject", func(t *testing.T) {
		content := "Hello, S3!"
		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String("hello.txt"),
			Body:        strings.NewReader(content),
			ContentType: aws.String("text/plain"),
		})
		require.NoError(t, err)
	})

	t.Run("GetObject", func(t *testing.T) {
		output, err := client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("hello.txt"),
		})
		require.NoError(t, err)
		defer output.Body.Close()

		body, err := io.ReadAll(output.Body)
		require.NoError(t, err)
		assert.Equal(t, "Hello, S3!", string(body))
		assert.Equal(t, "text/plain", *output.ContentType)
		assert.NotNil(t, output.ETag)
		assert.NotNil(t, output.LastModified)
	})

	t.Run("HeadObject", func(t *testing.T) {
		output, err := client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("hello.txt"),
		})
		require.NoError(t, err)
		assert.Equal(t, int64(10), *output.ContentLength)
		assert.Equal(t, "text/plain", *output.ContentType)
		assert.NotNil(t, output.ETag)
		assert.NotNil(t, output.LastModified)
	})

	t.Run("GetObject non-existing", func(t *testing.T) {
		_, err := client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("nonexistent.txt"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NoSuchKey")
	})

	t.Run("PutObject with metadata", func(t *testing.T) {
		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("with-metadata.txt"),
			Body:   strings.NewReader("content with metadata"),
			Metadata: map[string]string{
				"author": "test-user",
				"version": "1.0",
			},
		})
		require.NoError(t, err)

		// Verify metadata
		output, err := client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("with-metadata.txt"),
		})
		require.NoError(t, err)
		assert.Equal(t, "test-user", output.Metadata["author"])
		assert.Equal(t, "1.0", output.Metadata["version"])
	})

	t.Run("ListObjects after puts", func(t *testing.T) {
		output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		require.NoError(t, err)
		assert.Len(t, output.Contents, 2)

		keys := []string{*output.Contents[0].Key, *output.Contents[1].Key}
		assert.Contains(t, keys, "hello.txt")
		assert.Contains(t, keys, "with-metadata.txt")
	})

	t.Run("ListObjects with prefix", func(t *testing.T) {
		// Put objects with different prefixes
		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("docs/readme.md"),
			Body:   strings.NewReader("readme"),
		})
		require.NoError(t, err)

		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("docs/guide.md"),
			Body:   strings.NewReader("guide"),
		})
		require.NoError(t, err)

		// List with prefix
		output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String("docs/"),
		})
		require.NoError(t, err)
		assert.Len(t, output.Contents, 2)
		assert.Equal(t, "docs/readme.md", *output.Contents[0].Key)
		assert.Equal(t, "docs/guide.md", *output.Contents[1].Key)
	})

	t.Run("DeleteObject", func(t *testing.T) {
		_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("hello.txt"),
		})
		require.NoError(t, err)

		// Verify object is deleted
		_, err = client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("hello.txt"),
		})
		require.Error(t, err)
	})

	t.Run("Update object (overwrite)", func(t *testing.T) {
		key := "update-test.txt"

		// Put initial version
		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
			Body:   strings.NewReader("version 1"),
		})
		require.NoError(t, err)

		// Get first version
		output1, err := client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		require.NoError(t, err)
		body1, _ := io.ReadAll(output1.Body)
		output1.Body.Close()
		etag1 := *output1.ETag

		// Wait a moment to ensure different timestamp
		time.Sleep(10 * time.Millisecond)

		// Put updated version
		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
			Body:   strings.NewReader("version 2"),
		})
		require.NoError(t, err)

		// Get updated version
		output2, err := client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		require.NoError(t, err)
		body2, _ := io.ReadAll(output2.Body)
		output2.Body.Close()
		etag2 := *output2.ETag

		assert.Equal(t, "version 1", string(body1))
		assert.Equal(t, "version 2", string(body2))
		assert.NotEqual(t, etag1, etag2)
	})
}

func TestIntegration_CompleteWorkflow(t *testing.T) {
	_, client := setupTestServer(t)
	ctx := context.Background()

	// Create bucket
	bucketName := "workflow-test"
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	require.NoError(t, err)

	// Upload multiple files
	files := map[string]string{
		"images/photo1.jpg":   "photo1 binary data",
		"images/photo2.jpg":   "photo2 binary data",
		"documents/readme.md": "# Readme\nHello world",
		"data.json":           `{"key": "value"}`,
	}

	for key, content := range files {
		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
			Body:   strings.NewReader(content),
		})
		require.NoError(t, err)
	}

	// List all objects
	listOutput, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	require.NoError(t, err)
	assert.Len(t, listOutput.Contents, 4)

	// List with prefix filter
	listOutput, err = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String("images/"),
	})
	require.NoError(t, err)
	assert.Len(t, listOutput.Contents, 2)

	// Download a specific file
	getOutput, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("data.json"),
	})
	require.NoError(t, err)
	body, _ := io.ReadAll(getOutput.Body)
	getOutput.Body.Close()
	assert.Equal(t, `{"key": "value"}`, string(body))

	// Delete objects one by one
	for key := range files {
		_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		require.NoError(t, err)
	}

	// Verify bucket is empty
	listOutput, err = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	require.NoError(t, err)
	assert.Empty(t, listOutput.Contents)

	// Delete empty bucket
	_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	require.NoError(t, err)
}

func TestIntegration_ErrorCases(t *testing.T) {
	_, client := setupTestServer(t)
	ctx := context.Background()

	t.Run("GetObject from non-existing bucket", func(t *testing.T) {
		_, err := client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String("non-existing-bucket"),
			Key:    aws.String("test.txt"),
		})
		require.Error(t, err)
	})

	t.Run("DeleteBucket non-existing", func(t *testing.T) {
		_, err := client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String("non-existing-bucket"),
		})
		require.Error(t, err)
	})

	t.Run("HeadObject non-existing", func(t *testing.T) {
		// Create bucket first
		bucketName := "error-test-bucket"
		_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		require.NoError(t, err)

		_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("non-existing.txt"),
		})
		require.Error(t, err)
	})
}
