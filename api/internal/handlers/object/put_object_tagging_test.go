package object

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/testutil"
)

func TestPutObjectTagging(t *testing.T) {
	t.Parallel()
	testCtx := testutil.SetupTestDB(t)
	store := ctx.GetStore(testCtx)

	r := chi.NewRouter()
	r.Use(ctx.WithStore(store))
	r.Route("/{bucket}/{key}", func(r chi.Router) {
		r.Use(ctx.WithBucketName())
		r.Use(ctx.WithObjectKey())
		r.Put("/", func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Query().Has("tagging") {
				PutObjectTagging(w, req)
			}
		})
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Query().Has("tagging") {
				GetObjectTagging(w, req)
			}
		})
		r.Delete("/", func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Query().Has("tagging") {
				DeleteObjectTagging(w, req)
			}
		})
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

		// Create an object
		testData := []byte("test content")
		_, err = store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
			BucketName:   "test-bucket",
			Key:          "test-key",
			Data:         testData,
			Size:         int64(len(testData)),
			ETag:         "test-etag",
			ContentType:  "text/plain",
			StorageClass: "STANDARD",
		})
		require.NoError(t, err)

		// Test PUT object tagging
		_, err = s3Client.PutObjectTagging(context.Background(), &s3.PutObjectTaggingInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("test-key"),
			Tagging: &types.Tagging{
				TagSet: []types.Tag{
					{
						Key:   aws.String("Environment"),
						Value: aws.String("Production"),
					},
					{
						Key:   aws.String("Team"),
						Value: aws.String("Engineering"),
					},
				},
			},
		})
		assert.NoError(t, err)

		// Verify tags were created in database
		objectID, err := store.Queries.GetObjectID(context.Background(), db.GetObjectIDParams{
			BucketName: "test-bucket",
			Key:        "test-key",
		})
		require.NoError(t, err)

		tags, err := store.Queries.GetObjectTags(context.Background(), objectID)
		assert.NoError(t, err)
		assert.Len(t, tags, 2)

		// Test GET object tagging
		getResp, err := s3Client.GetObjectTagging(context.Background(), &s3.GetObjectTaggingInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("test-key"),
		})
		assert.NoError(t, err)
		assert.Len(t, getResp.TagSet, 2)
		assert.Equal(t, "Environment", *getResp.TagSet[0].Key)
		assert.Equal(t, "Production", *getResp.TagSet[0].Value)
		assert.Equal(t, "Team", *getResp.TagSet[1].Key)
		assert.Equal(t, "Engineering", *getResp.TagSet[1].Value)

		// Test DELETE object tagging
		_, err = s3Client.DeleteObjectTagging(context.Background(), &s3.DeleteObjectTaggingInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("test-key"),
		})
		assert.NoError(t, err)

		// Verify tags were deleted
		tags, err = store.Queries.GetObjectTags(context.Background(), objectID)
		assert.NoError(t, err)
		assert.Len(t, tags, 0)
	})

	t.Run("Object Not Found", func(t *testing.T) {
		_, err := s3Client.PutObjectTagging(context.Background(), &s3.PutObjectTaggingInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("nonexistent-key"),
			Tagging: &types.Tagging{
				TagSet: []types.Tag{
					{
						Key:   aws.String("Environment"),
						Value: aws.String("Production"),
					},
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("Update Existing Tags", func(t *testing.T) {
		// Create a bucket
		err := store.Queries.CreateBucket(context.Background(), db.CreateBucketParams{
			Name:   "test-bucket-update",
			Region: "us-east-1",
		})
		require.NoError(t, err)

		// Create an object
		testData := []byte("test content")
		obj, err := store.Queries.CreateObject(context.Background(), db.CreateObjectParams{
			BucketName:   "test-bucket-update",
			Key:          "test-key",
			Data:         testData,
			Size:         int64(len(testData)),
			ETag:         "test-etag",
			ContentType:  "text/plain",
			StorageClass: "STANDARD",
		})
		require.NoError(t, err)

		// Add initial tags
		err = store.Queries.CreateObjectTag(context.Background(), db.CreateObjectTagParams{
			ObjectID: obj.ID,
			Key:      "OldTag",
			Value:    "OldValue",
		})
		require.NoError(t, err)

		// Update tags via PutObjectTagging
		_, err = s3Client.PutObjectTagging(context.Background(), &s3.PutObjectTaggingInput{
			Bucket: aws.String("test-bucket-update"),
			Key:    aws.String("test-key"),
			Tagging: &types.Tagging{
				TagSet: []types.Tag{
					{
						Key:   aws.String("NewTag"),
						Value: aws.String("NewValue"),
					},
				},
			},
		})
		assert.NoError(t, err)

		// Verify old tags were replaced
		tags, err := store.Queries.GetObjectTags(context.Background(), obj.ID)
		assert.NoError(t, err)
		assert.Len(t, tags, 1)
		assert.Equal(t, "NewTag", tags[0].Key)
		assert.Equal(t, "NewValue", tags[0].Value)
	})
}
