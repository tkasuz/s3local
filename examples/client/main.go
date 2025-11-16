package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
		config.WithBaseEndpoint("http://localhost:8080"),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create S3 client with path-style addressing (required for s3local)
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Example 1: Create a bucket
	fmt.Println("=== Creating bucket ===")
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String("example-bucket"),
	})
	if err != nil {
		log.Printf("CreateBucket error: %v", err)
	} else {
		fmt.Println("Bucket 'example-bucket' created successfully")
	}

	// Example 2: List buckets
	fmt.Println("\n=== Listing buckets ===")
	listBucketsResp, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("ListBuckets error: %v", err)
	}
	for _, bucket := range listBucketsResp.Buckets {
		fmt.Printf("- %s (created: %s)\n", *bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
	}

	// Example 3: Upload an object
	fmt.Println("\n=== Uploading object ===")
	content := []byte("Hello from AWS SDK Go v2 client!\nThis is a test file for s3local.")
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String("example-bucket"),
		Key:         aws.String("example.txt"),
		Body:        bytes.NewReader(content),
		ContentType: aws.String("text/plain"),
		Metadata: map[string]string{
			"author": "aws-sdk-client",
		},
	})
	if err != nil {
		log.Fatalf("PutObject error: %v", err)
	}
	fmt.Println("Object 'example.txt' uploaded successfully")

	// Example 4: List objects
	fmt.Println("\n=== Listing objects ===")
	listObjectsResp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String("example-bucket"),
	})
	if err != nil {
		log.Fatalf("ListObjectsV2 error: %v", err)
	}
	fmt.Printf("Found %d objects:\n", listObjectsResp.KeyCount)
	for _, obj := range listObjectsResp.Contents {
		fmt.Printf("- %s (size: %d bytes, modified: %s)\n",
			*obj.Key, obj.Size, obj.LastModified.Format("2006-01-02 15:04:05"))
	}

	// Example 5: Head object (get metadata)
	fmt.Println("\n=== Getting object metadata ===")
	headResp, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String("example-bucket"),
		Key:    aws.String("example.txt"),
	})
	if err != nil {
		log.Fatalf("HeadObject error: %v", err)
	}
	fmt.Printf("Object metadata:\n")
	fmt.Printf("  ETag: %s\n", *headResp.ETag)
	fmt.Printf("  Content-Type: %s\n", *headResp.ContentType)
	fmt.Printf("  Size: %d bytes\n", headResp.ContentLength)
	fmt.Printf("  Last Modified: %s\n", headResp.LastModified.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Custom Metadata: %v\n", headResp.Metadata)

	// Example 6: Download object
	fmt.Println("\n=== Downloading object ===")
	getResp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("example-bucket"),
		Key:    aws.String("example.txt"),
	})
	if err != nil {
		log.Fatalf("GetObject error: %v", err)
	}
	defer getResp.Body.Close()

	downloadedData, err := io.ReadAll(getResp.Body)
	if err != nil {
		log.Fatalf("Failed to read object data: %v", err)
	}
	fmt.Printf("Downloaded data:\n%s\n", string(downloadedData))

	// Example 7: Tag the object
	fmt.Println("\n=== Tagging object ===")
	_, err = client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket: aws.String("example-bucket"),
		Key:    aws.String("example.txt"),
		Tagging: &types.Tagging{
			TagSet: []types.Tag{
				{Key: aws.String("environment"), Value: aws.String("development")},
				{Key: aws.String("project"), Value: aws.String("s3local-example")},
			},
		},
	})
	if err != nil {
		log.Fatalf("PutObjectTagging error: %v", err)
	}
	fmt.Println("Tags added successfully")

	// Get tags
	getTagsResp, err := client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String("example-bucket"),
		Key:    aws.String("example.txt"),
	})
	if err != nil {
		log.Fatalf("GetObjectTagging error: %v", err)
	}
	fmt.Println("Object tags:")
	for _, tag := range getTagsResp.TagSet {
		fmt.Printf("  %s = %s\n", *tag.Key, *tag.Value)
	}

	// Example 8: Delete object
	fmt.Println("\n=== Deleting object ===")
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String("example-bucket"),
		Key:    aws.String("example.txt"),
	})
	if err != nil {
		log.Fatalf("DeleteObject error: %v", err)
	}
	fmt.Println("Object deleted successfully")

	// Example 9: Delete bucket
	fmt.Println("\n=== Deleting bucket ===")
	_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String("example-bucket"),
	})
	if err != nil {
		log.Fatalf("DeleteBucket error: %v", err)
	}
	fmt.Println("Bucket deleted successfully")

	fmt.Println("\n=== Example completed successfully! ===")
}
