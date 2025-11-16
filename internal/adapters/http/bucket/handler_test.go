package bucket

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	bucketdomain "github.com/tkasuz/s3local/internal/domain/bucket"
	database "github.com/tkasuz/s3local/internal/sqlc"
	mockbucket "github.com/tkasuz/s3local/mocks/bucket"
)

func TestListBuckets(t *testing.T) {
	// Create a mock service
	mockService := mockbucket.NewMockServiceInterface(t)

	// Create handler with mock service
	handler := &BucketHandler{
		bucketService: mockService,
	}

	// Setup mock expectations
	expectedBuckets := []database.Bucket{
		{
			Name:      "test-bucket-1",
			Region:    "us-east-1",
			CreatedAt: time.Now(),
		},
		{
			Name:      "test-bucket-2",
			Region:    "us-west-2",
			CreatedAt: time.Now(),
		},
	}

	mockService.EXPECT().
		ListBuckets(mock.Anything).
		Return(expectedBuckets, nil).
		Once()

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ListBuckets(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))

	// Parse response
	var result bucketdomain.ListAllMyBucketsResult
	err := xml.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Buckets.Bucket))
	assert.Equal(t, "test-bucket-1", result.Buckets.Bucket[0].Name)
	assert.Equal(t, "test-bucket-2", result.Buckets.Bucket[1].Name)
}

func TestCreateBucket(t *testing.T) {
	mockService := mockbucket.NewMockServiceInterface(t)
	handler := &BucketHandler{
		bucketService: mockService,
	}

	mockService.EXPECT().
		CreateBucket(mock.Anything, "new-bucket", "us-east-1").
		Return(nil).
		Once()

	req := httptest.NewRequest(http.MethodPut, "/new-bucket", nil)
	req.Header.Set("x-amz-bucket-region", "us-east-1")
	w := httptest.NewRecorder()

	// Setup chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bucket", "new-bucket")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.CreateBucket(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "/new-bucket", w.Header().Get("Location"))
}

func TestDeleteBucket(t *testing.T) {
	tests := []struct {
		name           string
		bucketName     string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful deletion",
			bucketName:     "test-bucket",
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "bucket not found",
			bucketName:     "nonexistent",
			mockError:      bucketdomain.NewNoSuchBucketError("nonexistent"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "bucket not empty",
			bucketName:     "nonempty-bucket",
			mockError:      bucketdomain.ErrBucketNotEmpty,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mockbucket.NewMockServiceInterface(t)
			handler := NewBucketHandler(mockService)

			mockService.EXPECT().
				DeleteBucket(mock.Anything, tt.bucketName).
				Return(tt.mockError).
				Once()

			req := httptest.NewRequest(http.MethodDelete, "/"+tt.bucketName, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("bucket", tt.bucketName)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.DeleteBucket(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHeadBucket(t *testing.T) {
	tests := []struct {
		name           string
		bucketName     string
		mockReturn     database.Bucket
		mockError      error
		expectedStatus int
	}{
		{
			name:       "bucket exists",
			bucketName: "existing-bucket",
			mockReturn: database.Bucket{
				Name:      "existing-bucket",
				Region:    "us-east-1",
				CreatedAt: time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "bucket not found",
			bucketName:     "nonexistent",
			mockReturn:     database.Bucket{},
			mockError:      bucketdomain.NewNoSuchBucketError("nonexistent"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mockbucket.NewMockServiceInterface(t)
			handler := NewBucketHandler(mockService)

			mockService.EXPECT().
				GetBucket(mock.Anything, tt.bucketName).
				Return(tt.mockReturn, tt.mockError).
				Once()

			req := httptest.NewRequest(http.MethodHead, "/"+tt.bucketName, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("bucket", tt.bucketName)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.HeadBucket(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
