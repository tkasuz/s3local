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
	"github.com/stretchr/testify/require"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/testutil"
)

func TestListObjectsV2(t *testing.T) {
	t.Parallel()
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Get("/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		ListObjectsV2(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	t.Run("Success", func(t *testing.T) {
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   "test-bucket",
			Region: "us-east-1",
		})
		require.NoError(t, err)

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

		// Call ListObjectsV2 via AWS SDK
		out, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket: aws.String("test-bucket"),
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(out.Contents))
		assert.Equal(t, "object1", *out.Contents[0].Key)
		assert.Equal(t, "object2", *out.Contents[1].Key)
	})

	t.Run("Empty result", func(t *testing.T) {
		bucketName := "test-bucket-empty"
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   bucketName,
			Region: "us-east-1",
		})

		require.NoError(t, err)
		out, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(out.Contents))
	})

	t.Run("With prefix", func(t *testing.T) {
		bucketName := "test-bucket-prefix"
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   bucketName,
			Region: "us-east-1",
		})
		assert.NoError(t, err)

		// Create test objects with different prefixes
		testData := []byte("test content")
		_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
			BucketName:   bucketName,
			Key:          "prefix1/object1",
			Data:         testData,
			Size:         int64(len(testData)),
			ETag:         "etag1",
			ContentType:  "text/plain",
			StorageClass: "STANDARD",
		})
		assert.NoError(t, err)

		_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
			BucketName:   bucketName,
			Key:          "prefix2/object2",
			Data:         testData,
			Size:         int64(len(testData)),
			ETag:         "etag2",
			ContentType:  "text/plain",
			StorageClass: "STANDARD",
		})
		assert.NoError(t, err)
		out, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String("prefix1/"),
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(out.Contents))
		assert.Equal(t, "prefix1/object1", *out.Contents[0].Key)
	})

	t.Run("With delimiter", func(t *testing.T) {
		bucketNmae := "test-bucket-delimiter"
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   bucketNmae,
			Region: "us-east-1",
		})
		assert.NoError(t, err)

		// Create test objects with hierarchical structure
		testData := []byte("test content")
		objects := []string{
			"file1.txt",
			"file2.txt",
			"folder1/",
			"folder1/file3.txt",
			"folder1/file4.txt",
			"folder2/file5.txt",
			"folder2/subfolder/file6.txt",
		}

		for i, key := range objects {
			_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
				BucketName:   bucketNmae,
				Key:          key,
				Data:         testData,
				Size:         int64(len(testData)),
				ETag:         "etag" + string(rune(i)),
				ContentType:  "text/plain",
				StorageClass: "STANDARD",
			})
			assert.NoError(t, err)
		}

		// Test 1: List with delimiter "/" - should show files and folder prefixes at root level
		out, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket:    aws.String(bucketNmae),
			Delimiter: aws.String("/"),
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, len(out.Contents), "Should have 2 files at root level")
		assert.Equal(t, 1, len(out.CommonPrefixes), "Should have 2 folder prefixes")
		assert.Equal(t, "file1.txt", *out.Contents[0].Key)
		assert.Equal(t, "file2.txt", *out.Contents[1].Key)
		assert.Equal(t, "folder1/", *out.Contents[2].Key)
		assert.Equal(t, "folder2/", *out.CommonPrefixes[0].Prefix)

		// Test 2: List with prefix and delimiter - should show contents of folder1
		out, err = s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket:    aws.String(bucketNmae),
			Prefix:    aws.String("folder1/"),
			Delimiter: aws.String("/"),
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(out.Contents), "Should have 2 files in folder1")
		assert.Equal(t, 0, len(out.CommonPrefixes), "Should have no subfolders in folder1")
		assert.Equal(t, "folder1/file3.txt", *out.Contents[0].Key)
		assert.Equal(t, "folder1/file4.txt", *out.Contents[1].Key)

		// Test 3: List folder2 with delimiter - should show file and subfolder
		out, err = s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket:    aws.String(bucketNmae),
			Prefix:    aws.String("folder2/"),
			Delimiter: aws.String("/"),
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(out.Contents), "Should have 1 file in folder2")
		assert.Equal(t, 1, len(out.CommonPrefixes), "Should have 1 subfolder in folder2")
		assert.Equal(t, "folder2/file5.txt", *out.Contents[0].Key)
		assert.Equal(t, "folder2/subfolder/", *out.CommonPrefixes[0].Prefix)
	})

	t.Run("With folder marker objects", func(t *testing.T) {
		bucketName := "test-bucket-folder-markers"
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   bucketName,
			Region: "us-east-1",
		})
		assert.NoError(t, err)

		// Create folder marker objects (objects with keys ending in /)
		testData := []byte("")
		folderKeys := []string{
			"folder1/",
			"folder2/",
			"folder1/subfolder/",
		}

		for i, key := range folderKeys {
			_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
				BucketName:   bucketName,
				Key:          key,
				Data:         testData,
				Size:         0,
				ETag:         "etag" + string(rune(i)),
				ContentType:  "application/x-directory",
				StorageClass: "STANDARD",
			})
			assert.NoError(t, err)
		}

		// Also create some regular files
		fileData := []byte("test content")
		_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
			BucketName:   bucketName,
			Key:          "file1.txt",
			Data:         fileData,
			Size:         int64(len(fileData)),
			ETag:         "etag-file1",
			ContentType:  "text/plain",
			StorageClass: "STANDARD",
		})
		assert.NoError(t, err)

		// Test 1: List at root with delimiter - folder markers should appear in Contents
		out, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket:    aws.String(bucketName),
			Delimiter: aws.String("/"),
		})
		assert.NoError(t, err)
		// Should have file1.txt, folder1/, folder2/ in Contents (all direct children at root)
		assert.Equal(t, 3, len(out.Contents), "Should have 3 items at root level")
		// folder1/subfolder/ should NOT appear because it's nested under folder1/
		assert.Equal(t, 0, len(out.CommonPrefixes), "Should have no common prefixes")

		// Verify the keys
		keys := []string{}
		for _, content := range out.Contents {
			keys = append(keys, *content.Key)
		}
		assert.Contains(t, keys, "file1.txt")
		assert.Contains(t, keys, "folder1/")
		assert.Contains(t, keys, "folder2/")

		// Test 2: List under folder1/ prefix with delimiter
		out, err = s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket:    aws.String(bucketName),
			Prefix:    aws.String("folder1/"),
			Delimiter: aws.String("/"),
		})
		assert.NoError(t, err)
		// Should have folder1/subfolder/ in Contents
		assert.Equal(t, 1, len(out.Contents), "Should have 1 item under folder1/")
		assert.Equal(t, "folder1/subfolder/", *out.Contents[0].Key)
		assert.Equal(t, 0, len(out.CommonPrefixes), "Should have no common prefixes")
	})
}
