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
	"github.com/stretchr/testify/require"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/testutil"
)

func TestHeadBucket(t *testing.T) {
	t.Parallel()
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	r := chi.NewRouter()
	r.Route("/{bucket}", func(r chi.Router) {
		r.Use(ctx.WithBucketName())
		r.Use(ctx.WithStore(store))
		r.Head("/", func(w http.ResponseWriter, r *http.Request) {
			HeadBucket(w, r)
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	t.Run("Success", func(t *testing.T) {
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   "existing-bucket",
			Region: "us-west-2",
		})
		require.NoError(t, err)
		out, err := s3Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
			Bucket: aws.String("existing-bucket"),
		})
		assert.NoError(t, err)
		assert.Equal(t, "us-west-2", aws.ToString(out.BucketRegion))
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := s3Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
			Bucket: aws.String("nonexistent"),
		})
		assert.Error(t, err)
	})
}
