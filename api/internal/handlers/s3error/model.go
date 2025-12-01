package s3error

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	// Bucket
	ErrCodeBucketAlreadyExists     ErrorCode = "BucketAlreadyExists"
	ErrCodeBucketAlreadyOwnedByYou ErrorCode = "BucketAlreadyOwnedByYou"
	ErrCodeBucketNotEmpty          ErrorCode = "BucketNotEmpty"
	ErrCodeNoSuchBucket            ErrorCode = "NoSuchBucket"

	// Object
	ErrCodeNoSuchKey ErrorCode = "NoSuchKey"

	// Tagging
	ErrCodeInvalidTag     ErrorCode = "InvalidTag"
	ErrCodeMalformedXML   ErrorCode = "MalformedXML"

	// Policy
	ErrCodeNoSuchBucketPolicy ErrorCode = "NoSuchBucketPolicy"
	ErrCodeMalformedPolicy    ErrorCode = "MalformedPolicy"

	// General
	ErrCodeInternalError ErrorCode = "InternalError"
)

// Error represents the S3 error response
type Error struct {
	XMLName   struct{} `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource,omitempty"`
	RequestId string   `xml:"RequestId,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) WriteError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/xml")
	switch e.Code {
	case string(ErrCodeNoSuchKey):
		w.WriteHeader(http.StatusNotFound)
	case string(ErrCodeNoSuchBucket):
		w.WriteHeader(http.StatusNotFound)
	case string(ErrCodeNoSuchBucketPolicy):
		w.WriteHeader(http.StatusNotFound)
	case string(ErrCodeBucketAlreadyExists), string(ErrCodeBucketAlreadyOwnedByYou):
		w.WriteHeader(http.StatusConflict)
	case string(ErrCodeBucketNotEmpty):
		w.WriteHeader(http.StatusConflict)
	case string(ErrCodeInvalidTag), string(ErrCodeMalformedXML), string(ErrCodeMalformedPolicy):
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	xml.NewEncoder(w).Encode(e)
}

// NewBucketAlreadyExistsError creates a BucketAlreadyExists error with resource
func NewBucketAlreadyExistsError(bucket string) *Error {
	return &Error{
		Code:     string(ErrCodeBucketAlreadyExists),
		Message:  "The requested bucket name is not available. The bucket namespace is shared by all users of the system. Please select a different name and try again.",
		Resource: bucket,
	}
}

// NewBucketAlreadyOwnedByYouError creates a BucketAlreadyOwnedByYou error with resource
func NewBucketAlreadyOwnedByYouError(bucket string) *Error {
	return &Error{
		Code:     string(ErrCodeBucketAlreadyOwnedByYou),
		Message:  "The bucket you tried to create already exists, and you own it.",
		Resource: bucket,
	}
}

// NewBucketNotEmptyError creates a BucketNotEmpty error with resource
func NewBucketNotEmptyError(bucket string) *Error {
	return &Error{
		Code:     string(ErrCodeBucketNotEmpty),
		Message:  "The bucket you tried to delete is not empty.",
		Resource: bucket,
	}
}

// NewNoSuchBucketError creates a NoSuchBucket error with resource
func NewNoSuchBucketError(bucket string) *Error {
	return &Error{
		Code:     string(ErrCodeNoSuchBucket),
		Message:  "The specified bucket does not exist.",
		Resource: bucket,
	}
}

// NewNoSuchKeyError creates a NoSuchKey error with resource
func NewNoSuchKeyError(key string) *Error {
	return &Error{
		Code:     string(ErrCodeNoSuchKey),
		Message:  "The specified key does not exist.",
		Resource: key,
	}
}

// NewInternalError creates an InternalError with optional wrapped message
func NewInternalError(err error) *Error {
	msg := "We encountered an internal error. Please try again."
	if err != nil {
		msg = err.Error()
	}
	return &Error{
		Code:    string(ErrCodeInternalError),
		Message: msg,
	}
}

// NewInvalidTagError creates an InvalidTag error
func NewInvalidTagError(message string) *Error {
	return &Error{
		Code:    string(ErrCodeInvalidTag),
		Message: message,
	}
}

// NewMalformedXMLError creates a MalformedXML error
func NewMalformedXMLError() *Error {
	return &Error{
		Code:    string(ErrCodeMalformedXML),
		Message: "The XML you provided was not well-formed or did not validate against our published schema.",
	}
}

// NewNoSuchBucketPolicyError creates a NoSuchBucketPolicy error
func NewNoSuchBucketPolicyError(bucket string) *Error {
	return &Error{
		Code:     string(ErrCodeNoSuchBucketPolicy),
		Message:  "The bucket policy does not exist.",
		Resource: bucket,
	}
}

// NewMalformedPolicyError creates a MalformedPolicy error
func NewMalformedPolicyError(message string) *Error {
	return &Error{
		Code:    string(ErrCodeMalformedPolicy),
		Message: message,
	}
}
