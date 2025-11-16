package object

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

// Domain errors for object operations

// ErrorCode represents an object error code type
type ErrorCode string

const (
	ErrCodeNoSuchKey        ErrorCode = "NoSuchKey"
	ErrCodeNoSuchBucket     ErrorCode = "NoSuchBucket"
	ErrCodeInvalidArgument  ErrorCode = "InvalidArgument"
	ErrCodeEntityTooLarge   ErrorCode = "EntityTooLarge"
	ErrCodeEntityTooSmall   ErrorCode = "EntityTooSmall"
	ErrCodeInternalError    ErrorCode = "InternalError"
)

// DomainError represents an object domain error with S3-compatible error information
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

// Predefined object errors based on S3 REST API error responses

var (
	// ErrNoSuchKey is returned when the specified key does not exist
	ErrNoSuchKey = &DomainError{
		Code:       ErrCodeNoSuchKey,
		Message:    "The specified key does not exist.",
		StatusCode: http.StatusNotFound,
	}

	// ErrNoSuchBucket is returned when the specified bucket does not exist
	ErrNoSuchBucket = &DomainError{
		Code:       ErrCodeNoSuchBucket,
		Message:    "The specified bucket does not exist.",
		StatusCode: http.StatusNotFound,
	}

	// ErrInvalidArgument is returned when an argument is invalid
	ErrInvalidArgument = &DomainError{
		Code:       ErrCodeInvalidArgument,
		Message:    "Invalid Argument",
		StatusCode: http.StatusBadRequest,
	}

	// ErrEntityTooLarge is returned when the upload exceeds maximum allowed object size
	ErrEntityTooLarge = &DomainError{
		Code:       ErrCodeEntityTooLarge,
		Message:    "Your proposed upload exceeds the maximum allowed object size.",
		StatusCode: http.StatusBadRequest,
	}

	// ErrEntityTooSmall is returned when the upload is smaller than minimum allowed size
	ErrEntityTooSmall = &DomainError{
		Code:       ErrCodeEntityTooSmall,
		Message:    "Your proposed upload is smaller than the minimum allowed object size.",
		StatusCode: http.StatusBadRequest,
	}

	// ErrInternalError is returned when an internal error occurs
	ErrInternalError = &DomainError{
		Code:       ErrCodeInternalError,
		Message:    "We encountered an internal error. Please try again.",
		StatusCode: http.StatusInternalServerError,
	}
)

// NewNoSuchKeyError creates a NoSuchKey error with context
func NewNoSuchKeyError(key string) *DomainError {
	return &DomainError{
		Code:       ErrCodeNoSuchKey,
		Message:    fmt.Sprintf("The specified key '%s' does not exist.", key),
		StatusCode: http.StatusNotFound,
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

// NewInvalidArgumentError creates an InvalidArgument error with custom message
func NewInvalidArgumentError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeInvalidArgument,
		Message:    message,
		StatusCode: http.StatusBadRequest,
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

// PutObjectParams contains parameters for putting an object
type PutObjectParams struct {
	BucketName         string
	Key                string
	Data               []byte
	ContentType        string
	ContentEncoding    string
	ContentDisposition string
	CacheControl       string
	Metadata           map[string]string
}

// XML response structures for object operations

// ListBucketResult represents the S3 ListObjects response
type ListBucketResult struct {
	XMLName        xml.Name       `xml:"ListBucketResult"`
	Name           string         `xml:"Name"`
	Prefix         string         `xml:"Prefix"`
	Marker         string         `xml:"Marker"`
	NextMarker     string         `xml:"NextMarker,omitempty"`
	MaxKeys        int32          `xml:"MaxKeys"`
	IsTruncated    bool           `xml:"IsTruncated"`
	Contents       []Contents     `xml:"Contents"`
	CommonPrefixes []CommonPrefix `xml:"CommonPrefixes,omitempty"`
}

// Contents represents an object in the list
type Contents struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

// CommonPrefix represents a common prefix in the list
type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// Error represents an S3 error response
type Error struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource"`
	RequestId string   `xml:"RequestId"`
}
