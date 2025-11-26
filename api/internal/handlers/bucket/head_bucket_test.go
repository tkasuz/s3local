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

func TestHeadBucket(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create a bucket first
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "existing-bucket",
		Region: "us-west-2",
	})
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Head("/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		HeadBucket(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call HeadBucket via AWS SDK
	out, err := s3Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String("existing-bucket"),
	})
	assert.NoError(t, err)
	assert.Equal(t, "us-west-2", aws.ToString(out.BucketRegion))
}

func TestHeadBucket_NotFound(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(ctx.GetStore(testCtx)))
	r.Head("/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		HeadBucket(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call HeadBucket via AWS SDK
	_, err := s3Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String("nonexistent"),
	})
	assert.Error(t, err)
}
