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

func TestHeadObject(t *testing.T) {
	t.Parallel()
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Route("/{bucket}/{key}", func(r chi.Router) {
		r.Use(ctx.WithBucketName())
		r.Use(ctx.WithObjectKey())
		r.Head("/", func(w http.ResponseWriter, r *http.Request) {
			HeadObject(w, r)
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Create a bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket",
		Region: "us-east-1",
	})
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		// Create an object with metadata via store.Queries
		testData := []byte("test data")
		obj, err := store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
			BucketName:   "test-bucket",
			Key:          "test-key",
			Data:         testData,
			Size:         int64(len(testData)),
			ETag:         "098f6bcd4621d373cade4e832627b4f6",
			ContentType:  "text/plain",
			StorageClass: "STANDARD",
		})
		require.NoError(t, err)

		// Add metadata
		err = store.Queries.CreateObjectMetadata(context.Background(), db.CreateObjectMetadataParams{
			ObjectID: obj.ID,
			Key:      "custom-key",
			Value:    "custom-value",
		})
		require.NoError(t, err)

		// Call HeadObject via AWS SDK
		resp, err := s3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("test-key"),
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, aws.String("text/plain"), resp.ContentType)
		assert.Equal(t, int64(9), *resp.ContentLength)
		assert.NotNil(t, resp.ETag)
		assert.NotNil(t, resp.LastModified)
		assert.Equal(t, "bytes", *resp.AcceptRanges)
		assert.Equal(t, "custom-value", resp.Metadata["custom-key"])
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := s3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("nonexistent-key"),
		})
		assert.Error(t, err)
	})
}
