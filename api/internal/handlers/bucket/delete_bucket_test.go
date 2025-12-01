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

	r := chi.NewRouter()
	r.Route("/{bucket}", func(r chi.Router) {
		r.Use(ctx.WithBucketName())
		r.Use(ctx.WithStore(store))
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			DeleteBucket(w, r)
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	t.Run("Successfully delete bucket", func(t *testing.T) {
		// Create a bucket first
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   "test-bucket",
			Region: "us-east-1",
		})
		assert.NoError(t, err)

		_, err = s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
			Bucket: aws.String("test-bucket"),
		})
		assert.NoError(t, err)

		// Verify bucket is deleted
		_, err = store.Queries.GetBucket(context.Background(), "test-bucket")
		assert.Error(t, err)
	})

	t.Run("Not found", func(t *testing.T) {
		_, err := s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
			Bucket: aws.String("nonexistent"),
		})
		assert.Error(t, err)

	})
}
