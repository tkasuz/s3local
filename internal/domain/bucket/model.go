package bucket

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

// Domain errors for bucket operations

// ErrorCode represents a bucket error code type
type ErrorCode string

const (
	ErrCodeBucketAlreadyExists     ErrorCode = "BucketAlreadyExists"
	ErrCodeBucketAlreadyOwnedByYou ErrorCode = "BucketAlreadyOwnedByYou"
	ErrCodeBucketNotEmpty          ErrorCode = "BucketNotEmpty"
	ErrCodeNoSuchBucket            ErrorCode = "NoSuchBucket"
	ErrCodeInternalError           ErrorCode = "InternalError"
)

// DomainError represents a bucket domain error with S3-compatible error information
type DomainError struct {
	Code       ErrorCode
	Message    string
	StatusCode int
	Err        error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Predefined bucket errors based on S3 REST API error responses

var (
	// ErrBucketAlreadyExists is returned when creating a bucket that already exists
	ErrBucketAlreadyExists = &DomainError{
		Code:       ErrCodeBucketAlreadyExists,
		Message:    "The requested bucket name is not available. The bucket namespace is shared by all users of the system. Please select a different name and try again.",
		StatusCode: http.StatusConflict,
	}

	// ErrBucketAlreadyOwnedByYou is returned when creating a bucket that you already own
	ErrBucketAlreadyOwnedByYou = &DomainError{
		Code:       ErrCodeBucketAlreadyOwnedByYou,
		Message:    "Your previous request to create the named bucket succeeded and you already own it.",
		StatusCode: http.StatusConflict,
	}

	// ErrBucketNotEmpty is returned when trying to delete a non-empty bucket
	ErrBucketNotEmpty = &DomainError{
		Code:       ErrCodeBucketNotEmpty,
		Message:    "The bucket you tried to delete is not empty.",
		StatusCode: http.StatusConflict,
	}

	// ErrNoSuchBucket is returned when the specified bucket does not exist
	ErrNoSuchBucket = &DomainError{
		Code:       ErrCodeNoSuchBucket,
		Message:    "The specified bucket does not exist.",
		StatusCode: http.StatusNotFound,
	}

	// ErrInternalError is returned when an internal error occurs
	ErrInternalError = &DomainError{
		Code:       ErrCodeInternalError,
		Message:    "We encountered an internal error. Please try again.",
		StatusCode: http.StatusInternalServerError,
	}
)

// NewBucketAlreadyExistsError creates a BucketAlreadyExists error with context
func NewBucketAlreadyExistsError(bucketName string) *DomainError {
	return &DomainError{
		Code:       ErrCodeBucketAlreadyExists,
		Message:    fmt.Sprintf("The requested bucket name '%s' is not available.", bucketName),
		StatusCode: http.StatusConflict,
	}
}

// NewNoSuchBucketError creates a NoSuchBucket error with context
func NewNoSuchBucketError(bucketName string) *DomainError {
	return &DomainError{
		Code:       ErrCodeNoSuchBucket,
		Message:    fmt.Sprintf("The specified bucket '%s' does not exist.", bucketName),
		StatusCode: http.StatusNotFound,
	}
}

// NewInternalError creates an InternalError with wrapped error
func NewInternalError(err error) *DomainError {
	return &DomainError{
		Code:       ErrCodeInternalError,
		Message:    "We encountered an internal error. Please try again.",
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// XML response structures for bucket operations

// ListAllMyBucketsResult represents the S3 ListBuckets response
type ListAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Owner   Owner    `xml:"Owner"`
	Buckets Buckets  `xml:"Buckets"`
}

// Owner represents the bucket owner
type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// Buckets wraps a list of buckets
type Buckets struct {
	Bucket []Bucket `xml:"Bucket"`
}

// Bucket represents a single bucket in the list
type Bucket struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

// Error represents an S3 error response
type Error struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource"`
	RequestId string   `xml:"RequestId"`
}
