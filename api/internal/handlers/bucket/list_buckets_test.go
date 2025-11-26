package bucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/testutil"
)

func TestListBuckets(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	// Create test buckets
	err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket-1",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	err = store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
		Name:   "test-bucket-2",
		Region: "us-west-2",
	})
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ListBuckets(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	// Call ListBuckets via AWS SDK
	out, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(out.Buckets))
	assert.Equal(t, "test-bucket-1", *out.Buckets[0].Name)
	assert.Equal(t, "test-bucket-2", *out.Buckets[1].Name)
}

func TestListBuckets_Empty(t *testing.T) {
	testCtx := testutil.SetupTestDB(t)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(ctx.GetStore(testCtx)))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ListBuckets(w, r)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	s3Client := testutil.CreateNewS3Client(ts)

	out, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(out.Buckets))
}
