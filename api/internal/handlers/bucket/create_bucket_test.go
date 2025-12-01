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
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/testutil"
)

func TestCreateBucket(t *testing.T) {
	t.Parallel()
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	r := chi.NewRouter()
	r.Route("/{bucket}", func(r chi.Router) {
		r.Use(ctx.WithBucketName())
		r.Use(ctx.WithStore(store))
		r.Put("/", func(w http.ResponseWriter, req *http.Request) {
			CreateBucket(w, req)
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	t.Run("Success", func(t *testing.T) {
		// Call CreateBucket via AWS SDK
		_, err := s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
			Bucket: aws.String("new-bucket"),
		})
		assert.NoError(t, err)

		// Verify bucket was created
		bucket, err := store.Queries.GetBucket(context.Background(), "new-bucket")
		assert.NoError(t, err)
		assert.Equal(t, "new-bucket", bucket.Name)
		assert.Equal(t, "us-east-1", bucket.Region)
	})
}
