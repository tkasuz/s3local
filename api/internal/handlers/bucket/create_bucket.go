package bucket

import (
	"net/http"
	"strings"

	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/s3error"
)

// CreateBucket handles PUT /{bucket}
func CreateBucket(w http.ResponseWriter, r *http.Request) {
	store := ctx.GetStore(r.Context())
	bucketName := ctx.GetBucketName(r.Context())

	// Parse headers
	// TODO: use parsed headers
	_ = CreateBucketHeaders{
		ACL:                        r.Header.Get("x-amz-acl"),
		GrantFullControl:           r.Header.Get("x-amz-grant-full-control"),
		GrantRead:                  r.Header.Get("x-amz-grant-read"),
		GrantReadACP:               r.Header.Get("x-amz-grant-read-acp"),
		GrantWrite:                 r.Header.Get("x-amz-grant-write"),
		GrantWriteACP:              r.Header.Get("x-amz-grant-write-acp"),
		ObjectLockEnabledForBucket: r.Header.Get("x-amz-bucket-object-lock-enabled"),
		ObjectOwnership:            r.Header.Get("x-amz-object-ownership"),
	}

	region := "us-east-1"

	// TODO: use parsed body
	// Parse XML body
	// var reqBody *CreateBucketBody
	// if r.Body != nil && r.ContentLength > 0 {
	// 	defer r.Body.Close()
	// 	body, err := io.ReadAll(r.Body)
	// 	if err == nil && len(body) > 0 {
	// 		var parsed CreateBucketBody
	// 		if err := xml.Unmarshal(body, &parsed); err == nil {
	// 			reqBody = &parsed
	// 			if parsed.LocationConstraint != "" {
	// 				region = parsed.LocationConstraint
	// 			}
	// 		}
	// 	}
	// }

	// Fallback to x-amz-bucket-region header if present
	if r.Header.Get("x-amz-bucket-region") != "" {
		region = r.Header.Get("x-amz-bucket-region")
	}

	err := store.Queries.CreateBucket(r.Context(), db.CreateBucketParams{
		Name:   bucketName,
		Region: region,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			s3error.NewBucketAlreadyExistsError(bucketName).WriteError(w)
			return
		}
		s3error.NewInternalError(err).WriteError(w)
		return
	}

	w.Header().Set("Location", "/"+bucketName)
	w.Header().Set("x-amz-bucket-region", region)
	w.Header().Set("x-amz-bucket-arn", "arn:aws:s3:::"+bucketName)
	w.WriteHeader(http.StatusOK)
}

type CreateBucketBody struct {
	XMLName            struct{}        `xml:"CreateBucketConfiguration"`
	LocationConstraint string          `xml:"LocationConstraint,omitempty"`
	Location           *BucketLocation `xml:"Location,omitempty"`
	Bucket             *BucketInfo     `xml:"Bucket,omitempty"`
	Tags               *TagSet         `xml:"Tags,omitempty"`
}

// BucketLocation specifies the location for a directory bucket
type BucketLocation struct {
	Name string `xml:"Name"`
	Type string `xml:"Type"`
}

// BucketInfo specifies bucket configuration
type BucketInfo struct {
	DataRedundancy string `xml:"DataRedundancy"`
	Type           string `xml:"Type"`
}

// TagSet represents a collection of tags
type TagSet struct {
	Tag []Tag `xml:"Tag"`
}

// Tag represents a key-value tag
type Tag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

type CreateBucketHeaders struct {
	ACL                        string // x-amz-acl
	GrantFullControl           string // x-amz-grant-full-control
	GrantRead                  string // x-amz-grant-read
	GrantReadACP               string // x-amz-grant-read-acp
	GrantWrite                 string // x-amz-grant-write
	GrantWriteACP              string // x-amz-grant-write-acp
	ObjectLockEnabledForBucket string // x-amz-bucket-object-lock-enabled
	ObjectOwnership            string // x-amz-object-ownership
}
